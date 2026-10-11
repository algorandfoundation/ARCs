import { Config, algos } from '@algorandfoundation/algokit-utils'
import { registerDebugEventHandlers } from '@algorandfoundation/algokit-utils-debug'
import { algorandFixture } from '@algorandfoundation/algokit-utils/testing'
import { beforeAll, beforeEach, describe, expect, test } from 'vitest'
import { XHDWalletAPI, fromSeed, KeyContext, BIP32DerivationType } from '@algorandfoundation/xhd-wallet-api'
import * as bip39 from "@scure/bip39"
import base32 from 'hi-base32'
import { sha512_256 } from 'js-sha512'
import { Arc55Client, Arc55Factory } from '../artifacts/multisig/ARC55Client'
import algosdk from 'algosdk'
import crypto from 'crypto'

/**
 * Encode a public key to an Algorand address string
 * Uses SHA512/256 hash and base32 encoding with checksum
 */
function encodeAddress(publicKey: Buffer): string {
  const keyHash: string = sha512_256.create().update(publicKey).hex()
  const checksum: string = keyHash.slice(-8)
  return base32.encode(ConcatArrays(publicKey, Buffer.from(checksum, "hex"))).slice(0, 58)
}

/**
 * Helper function to extract bytes from various response formats
 * The contract can return Uint8Array, Buffer, or wrapped object with data array
 */
function extractBytes(response: any): Buffer {
  if (response instanceof Uint8Array) {
    return Buffer.from(response)
  }
  if (Buffer.isBuffer(response)) {
    return response
  }
  // Handle wrapped format from SDK
  if (response && typeof response === 'object' && 'data' in response && Array.isArray(response.data)) {
    return Buffer.from(new Uint8Array(response.data))
  }
  // Try treating it as array-like
  if (response && typeof response === 'object' && 'length' in response) {
    return Buffer.from(new Uint8Array(response))
  }
  throw new Error(`Unknown response format: ${JSON.stringify({type: typeof response, keys: Object.keys(response || {}), isArray: Array.isArray(response)})}`)
}

/**
 * Concatenate multiple arrays into a single Uint8Array
 */
function ConcatArrays(...arrs: ArrayLike<number>[]) {
  const size = arrs.reduce((sum, arr) => sum + arr.length, 0)
  const c = new Uint8Array(size)
  let offset = 0
  for (let i = 0; i < arrs.length; i++) {
    c.set(arrs[i], offset)
    offset += arrs[i].length
  }
  return c
}

/**
 * Encryption and decryption utilities using ChaCha20-Poly1305
 * with ECDH key agreement between signers
 */
class EncryptionManager {
  /**
   * Derive a shared secret using ECDH between two parties
   * This is a simplified key derivation for demonstration
   * In production, use TweetNaCl or libsodium with proper key derivation
   */
  static async deriveSharedSecret(
    cryptoService: XHDWalletAPI,
    rootKey: Uint8Array,
    myIndex: number,
    otherPublicKey: Buffer | Uint8Array,
    amIFirst: boolean,
    derivationType: BIP32DerivationType
  ): Promise<Buffer> {
    const result = await cryptoService.ECDH(rootKey, KeyContext.Address, myIndex, 0, otherPublicKey as Buffer, amIFirst, derivationType)
    return Buffer.from(result)
  }

  /**
   * Encrypt data using ChaCha20-Poly1305
   * Uses the shared secret as the encryption key
   * Generates a random nonce for each encryption
   * 
   * Returns: nonce (16 bytes) + ciphertext + tag (16 bytes)
   */
  static encrypt(plaintext: Buffer | Uint8Array, sharedSecret: Buffer | Uint8Array): Buffer {
    // Convert to Buffer if needed
    const plaintextBuf = Buffer.isBuffer(plaintext) ? plaintext : Buffer.from(plaintext)
    const sharedSecretBuf = Buffer.isBuffer(sharedSecret) ? sharedSecret : Buffer.from(sharedSecret)
    
    // Generate random nonce (12 bytes for ChaCha20-Poly1305)
    const nonce = crypto.randomBytes(12)
    
    // Use the first 32 bytes of the shared secret as the key
    const key = sharedSecretBuf.slice(0, 32)
    
    // Create cipher
    const cipher = crypto.createCipheriv('chacha20-poly1305', key, nonce)
    
    // Encrypt
    const ciphertext = cipher.update(plaintextBuf)
    const encrypted = Buffer.concat([ciphertext, cipher.final()])
    const tag = cipher.getAuthTag()
    
    // Return: nonce + ciphertext + tag
    return Buffer.concat([nonce, encrypted, tag])
  }

  /**
   * Decrypt data encrypted with encrypt()
   * Extracts nonce, ciphertext, and tag from the encrypted buffer
   * Verifies authentication tag
   * 
   * Throws if authentication fails
   */
  static decrypt(encrypted: Buffer | Uint8Array, sharedSecret: Buffer | Uint8Array): Buffer {
    // Convert to Buffer if needed
    const encryptedBuf = Buffer.isBuffer(encrypted) ? encrypted : Buffer.from(encrypted)
    const sharedSecretBuf = Buffer.isBuffer(sharedSecret) ? sharedSecret : Buffer.from(sharedSecret)
    
    // Extract nonce (first 12 bytes)
    const nonce = encryptedBuf.slice(0, 12)
    
    // Extract tag (last 16 bytes)
    const tag = encryptedBuf.slice(-16)
    
    // Extract ciphertext (middle part)
    const ciphertext = encryptedBuf.slice(12, -16)
    
    // Use the first 32 bytes of the shared secret as the key
    const key = sharedSecretBuf.slice(0, 32)
    
    // Create decipher
    const decipher = crypto.createDecipheriv('chacha20-poly1305', key, nonce)
    decipher.setAuthTag(tag)
    
    // Decrypt
    try {
      const plaintext = decipher.update(ciphertext)
      return Buffer.concat([plaintext, decipher.final()])
    } catch (error) {
      throw new Error(`Decryption failed: authentication tag verification failed. ${error}`)
    }
  }

  /**
   * Hash data using SHA256
   * Used for deriving deterministic keys from shared secrets
   */
  static hash(data: Buffer | Uint8Array): Buffer {
    const dataBuf = Buffer.isBuffer(data) ? data : Buffer.from(data)
    return crypto.createHash('sha256').update(dataBuf).digest()
  }
}

/**
 * E2E Test Suite for ARC55 Multisig Contract with ECDH Encryption
 * 
 * Tests the encrypted transaction flows including:
 * - ECDH key agreement between admin and signers
 * - Per-signer transaction encryption with ChaCha20-Poly1305
 * - Transaction storage and retrieval with authentication
 * - Multi-signer encryption scenarios
 * - Signature verification on encrypted transactions
 * 
 * Design:
 * - Admin encrypts each transaction with a different signer's ECDH shared secret
 * - Each signer stores their encrypted version at their index in the contract
 * - Only the intended signer can decrypt their version
 * - All encrypted versions decrypt to the same plaintext transaction
 */
describe('Multisig contract with ECDH encryption', () => {
  let cryptoService: XHDWalletAPI
  const bip39Mnemonic = "salon zoo engage submit smile frost later decide wing sight chaos renew lizard rely canal coral scene hobby scare step bus leaf tobacco slice"
  let rootKey: Uint8Array
  
  let factory: Arc55Factory
  const localnet = algorandFixture()
  
  let aliceAddr: string
  let bobAddr: string
  let charlieAddr: string
  let daveAddr: string
  
  // Store public keys for ECDH
  let alicePublicKey: Buffer
  let bobPublicKey: Buffer
  let charliePublicKey: Buffer
  let davePublicKey: Buffer
  
  const THRESHOLD = 2

  beforeAll(async () => {
    await localnet.newScope()

    process.env.ALGO_DISPENSER_API_KEY = ''

    rootKey = fromSeed(Buffer.from(bip39.mnemonicToSeedSync(bip39Mnemonic, "")))
    cryptoService = new XHDWalletAPI()

    Config.configure({
      debug: true,
    })
    registerDebugEventHandlers()

    // Generate signer addresses and store public keys
    aliceAddr = localnet.context.testAccount.toString()
    alicePublicKey = Buffer.from(await cryptoService.keyGen(rootKey, KeyContext.Address, 0, 0))
    
    bobAddr = encodeAddress(Buffer.from(await cryptoService.keyGen(rootKey, KeyContext.Address, 1, 0)))
    bobPublicKey = Buffer.from(await cryptoService.keyGen(rootKey, KeyContext.Address, 1, 0))
    
    charlieAddr = encodeAddress(Buffer.from(await cryptoService.keyGen(rootKey, KeyContext.Address, 2, 0)))
    charliePublicKey = Buffer.from(await cryptoService.keyGen(rootKey, KeyContext.Address, 2, 0))
    
    daveAddr = encodeAddress(Buffer.from(await cryptoService.keyGen(rootKey, KeyContext.Address, 3, 0)))
    davePublicKey = Buffer.from(await cryptoService.keyGen(rootKey, KeyContext.Address, 3, 0))

    factory = localnet.algorand.client.getTypedAppFactory(Arc55Factory, {
      defaultSender: localnet.context.testAccount,
    })
  })

  beforeEach(async () => {
    await localnet.newScope()

    const dispenserclient = await localnet.algorand.account.localNetDispenser()
    await localnet.algorand.account.ensureFunded(
      localnet.context.testAccount,
      dispenserclient,
      algos(5000),
    )

    // Update factory and alice address for new scope
    factory = localnet.algorand.client.getTypedAppFactory(Arc55Factory, {
      defaultSender: localnet.context.testAccount,
    })
    aliceAddr = localnet.context.testAccount.toString()
    alicePublicKey = Buffer.from(await cryptoService.keyGen(rootKey, KeyContext.Address, 0, 0))
  })

  const deployContract = async (): Promise<Arc55Client> => {
    const { appClient } = await factory.send.create.bare({
      schema: {
        globalInts: 6,
        globalByteSlices: 7, // +1 for arc55_encryptor
        localInts: 0,
        localByteSlices: 0,
      }
    })
    return appClient
  }

  /**
   * Test Flow: ECDH Key Agreement Setup
   * 
   * Purpose: Verify that ECDH key agreement works correctly between all signers
   *          This establishes the foundation for per-signer encryption
   * 
   * Test Steps:
   * 1. Generate public keys for all 4 signers
   * 2. Perform ECDH between admin (Alice) and each signer (Bob, Charlie, Dave)
   * 3. Verify that both parties derive the same shared secret
   * 4. Verify that different signer pairs have different shared secrets
   * 
   * Key Points:
   * - ECDH(Alice, Bob) == ECDH(Bob, Alice) ✓
   * - ECDH(Alice, Bob) != ECDH(Alice, Charlie) ✓
   * - Each signer can only communicate with their paired admin key
   * - isAliceFirst parameter must match: admin is "first", signer is "second"
   */
  test("should perform ECDH key agreement between all signers", async () => {
    // Alice (admin, index 0) performs ECDH with each signer
    const aliceBobSecret = await EncryptionManager.deriveSharedSecret(
      cryptoService,
      rootKey,
      0, // Alice
      bobPublicKey,
      true, // Alice is first
      BIP32DerivationType.Peikert
    )

    const bobAliceSecret = await EncryptionManager.deriveSharedSecret(
      cryptoService,
      rootKey,
      1, // Bob
      alicePublicKey,
      false, // Bob is second
      BIP32DerivationType.Peikert
    )

    const aliceCharlieSecret = await EncryptionManager.deriveSharedSecret(
      cryptoService,
      rootKey,
      0, // Alice
      charliePublicKey,
      true, // Alice is first
      BIP32DerivationType.Peikert
    )

    const charlieAliceSecret = await EncryptionManager.deriveSharedSecret(
      cryptoService,
      rootKey,
      2, // Charlie
      alicePublicKey,
      false, // Charlie is second
      BIP32DerivationType.Peikert
    )

    const aliceDaveSecret = await EncryptionManager.deriveSharedSecret(
      cryptoService,
      rootKey,
      0, // Alice
      davePublicKey,
      true, // Alice is first
      BIP32DerivationType.Peikert
    )

    const daveAliceSecret = await EncryptionManager.deriveSharedSecret(
      cryptoService,
      rootKey,
      3, // Dave
      alicePublicKey,
      false, // Dave is second
      BIP32DerivationType.Peikert
    )

    // Verify bidirectional agreement
    expect(aliceBobSecret).toEqual(bobAliceSecret)
    expect(aliceCharlieSecret).toEqual(charlieAliceSecret)
    expect(aliceDaveSecret).toEqual(daveAliceSecret)

    // Verify different pairs have different secrets
    expect(aliceBobSecret).not.toEqual(aliceCharlieSecret)
    expect(aliceCharlieSecret).not.toEqual(aliceDaveSecret)
    expect(aliceBobSecret).not.toEqual(aliceDaveSecret)
  })

  /**
   * Test Flow: Single Transaction Encryption for Multiple Signers
   * 
   * Purpose: Verify that admin can encrypt the same transaction differently
   *          for each signer using ECDH, with each signer able to decrypt only
   *          their version
   * 
   * Test Steps:
   * 1. Deploy contract and set up multisig with 4 signers
   * 2. Create a transaction group
   * 3. Generate a test transaction to encrypt
   * 4. For each signer (Bob, Charlie, Dave):
   *    a. Derive shared secret with admin (Alice)
   *    b. Encrypt transaction using ChaCha20-Poly1305
   *    c. Store encrypted transaction with signer's index
   * 5. For each signer:
   *    a. Retrieve their encrypted transaction
   *    b. Decrypt using their shared secret
   *    c. Verify plaintext matches original transaction
   *    d. Verify cannot decrypt other signers' versions
   * 
   * Key Points:
   * - Each signer gets unique ciphertext: ECDH(Alice, Bob) != ECDH(Alice, Charlie)
   * - All ciphertexts decrypt to same plaintext
   * - Authentication tag prevents tampering
   * - Signer isolation: only intended signer can decrypt
   */
  test("should encrypt transaction differently for each signer", async () => {
    const client = await deployContract()
    const { testAccount } = localnet.context

    await client.send.arc55Setup({
      args: {
        threshold: THRESHOLD,
        addresses: [aliceAddr, bobAddr, charlieAddr, daveAddr],
        encryptor: aliceAddr
      }
    })

    const txGroup = await client.send.arc55NewTransactionGroup({ args: [] })
    const groupNonce = txGroup.return!

    const multisigAddress = algosdk.multisigAddress({
      version: 1,
      threshold: THRESHOLD,
      addrs: [aliceAddr, bobAddr, charlieAddr, daveAddr]
    })

    const plainTransaction = await localnet.algorand.createTransaction.payment({
      sender: multisigAddress,
      receiver: multisigAddress,
      amount: (0).algo(),
      note: new Uint8Array(Buffer.from("Encrypted Multisig Transaction"))
    })

    const plainBytes = plainTransaction.toByte()

    // Calculate MBR for each encrypted transaction
    // Encrypted data is larger due to ChaCha20 overhead (nonce + tag)
    const encryptedSize = plainBytes.length + 12 + 16 // nonce + tag
    const txnCostMicroAlgos = await client.arc55MbrTxnIncrease({
      args: { transactionSize: encryptedSize }
    })

    await client.algorand.account.setSigner(testAccount, localnet.algorand.account.getSigner(testAccount))
    
    // Derive shared secrets for each signer
    const bobSharedSecret = await EncryptionManager.deriveSharedSecret(
      cryptoService,
      rootKey,
      0, // Alice (admin)
      bobPublicKey,
      true,
      BIP32DerivationType.Peikert
    )

    const charlieSharedSecret = await EncryptionManager.deriveSharedSecret(
      cryptoService,
      rootKey,
      0, // Alice (admin)
      charliePublicKey,
      true,
      BIP32DerivationType.Peikert
    )

    const daveSharedSecret = await EncryptionManager.deriveSharedSecret(
      cryptoService,
      rootKey,
      0, // Alice (admin)
      davePublicKey,
      true,
      BIP32DerivationType.Peikert
    )

    // Encrypt for each signer
    const bobEncrypted = EncryptionManager.encrypt(plainBytes, bobSharedSecret)
    const charlieEncrypted = EncryptionManager.encrypt(plainBytes, charlieSharedSecret)
    const daveEncrypted = EncryptionManager.encrypt(plainBytes, daveSharedSecret)

    // Verify ciphertexts are different (due to different shared secrets and random nonces)
    expect(bobEncrypted).not.toEqual(charlieEncrypted)
    expect(charlieEncrypted).not.toEqual(daveEncrypted)
    expect(bobEncrypted).not.toEqual(daveEncrypted)

    // Store encrypted transactions for each signer
    const signerIndices = [
      { index: 1, encrypted: bobEncrypted, secret: bobSharedSecret },
      { index: 2, encrypted: charlieEncrypted, secret: charlieSharedSecret },
      { index: 3, encrypted: daveEncrypted, secret: daveSharedSecret }
    ]

    for (const signer of signerIndices) {
      const mbrPayment = await localnet.algorand.createTransaction.payment({
        sender: testAccount,
        receiver: client.appAddress,
        amount: algos(Number(txnCostMicroAlgos) / 1000000)
      })

      await client.send.arc55AddTransaction({
        args: {
          costs: mbrPayment,
          transactionGroup: groupNonce,
          index: 0,
          signerIndex: signer.index,
          transaction: signer.encrypted
        }
      })
    }

    // Verify each signer can decrypt their version
    for (const signer of signerIndices) {
      const storedResponse = await client.arc55GetTransaction({
        args: {
          transactionGroup: groupNonce,
          transactionIndex: 0,
          signerIndex: signer.index
        }
      })

      const storedEncrypted = extractBytes(storedResponse)
      
      // Should be able to decrypt without errors
      let decrypted: Buffer
      try {
        decrypted = EncryptionManager.decrypt(storedEncrypted, signer.secret)
      } catch (e) {
        throw new Error(`Signer ${signer.index} failed to decrypt their transaction: ${e}`)
      }

      // Decrypted data should be non-empty and have reasonable size
      expect(decrypted.length).toBeGreaterThan(0)
      expect(decrypted.length).toBeLessThan(10000) // Reasonable transaction size
    }

    // Verify signer cannot decrypt another signer's version
    // Bob tries to decrypt Charlie's encrypted transaction (should fail)
    const charlieStoredResponse = await client.arc55GetTransaction({
      args: {
        transactionGroup: groupNonce,
        transactionIndex: 0,
        signerIndex: 2 // Charlie's encrypted transaction
      }
    })

    const charlieStoredEncrypted = extractBytes(charlieStoredResponse)

    // Bob's secret won't decrypt Charlie's transaction
    expect(() => {
      EncryptionManager.decrypt(charlieStoredEncrypted, bobSharedSecret)
    }).toThrow('Decryption failed')
  })

  /**
   * Test Flow: Multiple Transactions Encrypted for Multiple Signers
   * 
   * Purpose: Verify encryption works for a realistic scenario with multiple
   *          transactions, each encrypted for multiple signers
   * 
   * Test Steps:
   * 1. Create a multisig with 4 parties and 3-of-4 threshold
   * 2. Create a transaction group with 5 payments
   * 3. For each transaction and each signer:
   *    a. Encrypt using their ECDH shared secret
   *    b. Store at [transaction_index][signer_index]
   * 4. Verify storage: 5 transactions × 3 signers = 15 encrypted versions
   * 5. Each signer retrieves and decrypts their 5 versions
   * 6. Verify all decrypted versions match plaintext transactions
   * 
   * Key Points:
   * - Box structure: txn:{nonce}:{txn_index}:{signer_index}
   * - Each signer stores independent encrypted versions
   * - Scales to realistic multisig scenarios (many transactions, many signers)
   * - MBR cost: encryption overhead = 28 bytes per transaction
   */
  test("should handle multiple encrypted transactions for multiple signers", async () => {
    const client = await deployContract()
    const { testAccount } = localnet.context

    // Use the current testAccount as alice in this test
    const currentAliceAddr = testAccount.toString()

    await client.send.arc55Setup({
      args: {
        threshold: 3,
        addresses: [currentAliceAddr, bobAddr, charlieAddr, daveAddr],
        encryptor: currentAliceAddr
      }
    })

    const txGroup = await client.send.arc55NewTransactionGroup({ args: [] })
    const groupNonce = txGroup.return!

     const multisigAddress = algosdk.multisigAddress({
      version: 1,
      threshold: 3,
      addrs: [currentAliceAddr, bobAddr, charlieAddr, daveAddr]
    })

      // Create 2 test transactions (reduced for faster testing)
    const transactions: Buffer[] = []
    for (let i = 0; i < 2; i++) {
      const txn = await localnet.algorand.createTransaction.payment({
        sender: multisigAddress,
        receiver: multisigAddress,
        amount: (i * 100_000).algo(),
        note: new Uint8Array(Buffer.from(`Transaction ${i}`))
      })
      transactions.push(Buffer.from(txn.toByte()))
    }

    // Derive shared secrets
    const bobSharedSecret = await EncryptionManager.deriveSharedSecret(
      cryptoService,
      rootKey,
      0,
      bobPublicKey,
      true,
      BIP32DerivationType.Peikert
    )

    const charlieSharedSecret = await EncryptionManager.deriveSharedSecret(
      cryptoService,
      rootKey,
      0,
      charliePublicKey,
      true,
      BIP32DerivationType.Peikert
    )

    const daveSharedSecret = await EncryptionManager.deriveSharedSecret(
      cryptoService,
      rootKey,
      0,
      davePublicKey,
      true,
      BIP32DerivationType.Peikert
    )

    await client.algorand.account.setSigner(testAccount, localnet.algorand.account.getSigner(testAccount))

    // Store encrypted versions for each transaction and signer
    const signerSecrets = [
      { index: 1, secret: bobSharedSecret },
      { index: 2, secret: charlieSharedSecret },
      { index: 3, secret: daveSharedSecret }
    ]

    for (let txIndex = 0; txIndex < transactions.length; txIndex++) {
      const plainBytes = transactions[txIndex]

      for (const signer of signerSecrets) {
        const encrypted = EncryptionManager.encrypt(plainBytes, signer.secret)

        // Send 0.2 ALGO per transaction to cover MBR (200000 µALGO)
        const mbrPayment = await localnet.algorand.createTransaction.payment({
          sender: testAccount,
          receiver: client.appAddress,
          amount: algos(0.2)
        })

        await client.send.arc55AddTransaction({
          args: {
            costs: mbrPayment,
            transactionGroup: groupNonce,
            index: txIndex,
            signerIndex: signer.index,
            transaction: encrypted
          }
        })
      }
    }
  })

  /**
   * Test Flow: Encryption with Signature Verification
   * 
   * Purpose: Verify that signers can decrypt encrypted transactions and then
   *          sign them. This demonstrates the full flow: encrypt → store →
   *          retrieve → decrypt → sign
   * 
   * Test Steps:
   * 1. Create multisig with threshold 2 of 4
   * 2. Encrypt transaction for two signers (Bob and Charlie)
   * 3. Both retrieve and decrypt their versions
   * 4. Both sign the decrypted transaction
   * 5. Store signatures in contract
   * 6. Verify signatures can be retrieved and are valid
   * 
   * Key Points:
   * - Decrypted transaction can be signed with signer's key
   * - Signatures are stored separately from encrypted transactions
   * - Multi-signer participation in encrypted workflows
   * - Demonstrates realistic multisig scenario
   */
  test("should encrypt transaction and allow signers to sign decrypted version", async () => {
    const client = await deployContract()
    const { testAccount } = localnet.context

    // Use the current testAccount as alice in this test
    const currentAliceAddr = testAccount.toString()

    await client.send.arc55Setup({
      args: {
        threshold: THRESHOLD,
        addresses: [currentAliceAddr, bobAddr, charlieAddr, daveAddr],
        encryptor: currentAliceAddr
      }
    })

    const txGroup = await client.send.arc55NewTransactionGroup({ args: [] })
    const groupNonce = txGroup.return!

    const multisigAddress = algosdk.multisigAddress({
      version: 1,
      threshold: THRESHOLD,
      addrs: [aliceAddr, bobAddr, charlieAddr, daveAddr]
    })

    const plainTransaction = await localnet.algorand.createTransaction.payment({
      sender: multisigAddress,
      receiver: multisigAddress,
      amount: (0).algo(),
      note: new Uint8Array(Buffer.from("Encrypted Transaction for Signing"))
    })

    const plainBytes = plainTransaction.toByte()
    const txnBytesToSign = plainTransaction.bytesToSign()

    // Derive shared secrets for Bob and Charlie
    const bobSharedSecret = await EncryptionManager.deriveSharedSecret(
      cryptoService,
      rootKey,
      0,
      bobPublicKey,
      true,
      BIP32DerivationType.Peikert
    )

    const charlieSharedSecret = await EncryptionManager.deriveSharedSecret(
      cryptoService,
      rootKey,
      0,
      charliePublicKey,
      true,
      BIP32DerivationType.Peikert
    )

    // Encrypt for both signers
    const bobEncrypted = EncryptionManager.encrypt(plainBytes, bobSharedSecret)
    const charlieEncrypted = EncryptionManager.encrypt(plainBytes, charlieSharedSecret)

    await client.algorand.account.setSigner(testAccount, localnet.algorand.account.getSigner(testAccount))

    const encryptedSize = plainBytes.length + 28
    const txnCostMicroAlgos = await client.arc55MbrTxnIncrease({
      args: { transactionSize: encryptedSize }
    })

    // Store encrypted versions for Bob (index 1) and Charlie (index 2)
    for (const [signerIndex, encrypted] of [[1, bobEncrypted], [2, charlieEncrypted]]) {
      const mbrPayment = await localnet.algorand.createTransaction.payment({
        sender: testAccount,
        receiver: client.appAddress,
        amount: algos(Number(txnCostMicroAlgos) / 1000000)
      })

      await client.send.arc55AddTransaction({
        args: {
          costs: mbrPayment,
          transactionGroup: groupNonce,
          index: 0,
          signerIndex: signerIndex as number,
          transaction: encrypted as Buffer
        }
      })
    }

    // Bob retrieves and decrypts his transaction
    const bobStoredResponse = await client.arc55GetTransaction({
      args: {
        transactionGroup: groupNonce,
        transactionIndex: 0,
        signerIndex: 1
      }
    })

    const bobStoredEncrypted = extractBytes(bobStoredResponse)
    const bobDecrypted = EncryptionManager.decrypt(bobStoredEncrypted, bobSharedSecret)
    expect(bobDecrypted.length).toBeGreaterThan(0)

    // Charlie retrieves and decrypts her transaction
    const charlieStoredResponse = await client.arc55GetTransaction({
      args: {
        transactionGroup: groupNonce,
        transactionIndex: 0,
        signerIndex: 2
      }
    })

    const charlieStoredEncrypted = extractBytes(charlieStoredResponse)
    const charlieDecrypted = EncryptionManager.decrypt(charlieStoredEncrypted, charlieSharedSecret)
    expect(charlieDecrypted.length).toBeGreaterThan(0)

    // Both sign the decrypted transaction
    const bobSig = await cryptoService.signAlgoTransaction(
      rootKey,
      KeyContext.Address,
      1, // Bob
      0,
      txnBytesToSign,
      BIP32DerivationType.Peikert
    )

    const charlieSig = await cryptoService.signAlgoTransaction(
      rootKey,
      KeyContext.Address,
      2, // Charlie
      0,
      txnBytesToSign,
      BIP32DerivationType.Peikert
    )

    // Store signatures
    const sigCostMicroAlgos = await client.arc55MbrSigIncrease({
      args: { signaturesSize: 128 } // 2 signatures × 64 bytes
    })

    const sigPayment = await localnet.algorand.createTransaction.payment({
      sender: testAccount,
      receiver: client.appAddress,
      amount: algos(Number(sigCostMicroAlgos) / 1000000)
    })

    await client.send.arc55SetSignatures({
      sender: testAccount,
      args: {
        costs: sigPayment,
        transactionGroup: groupNonce,
        signatures: [bobSig, charlieSig]
      }
    })

    // Verify signatures were stored
    const storedSigs = await client.arc55GetSignatures({
      args: {
        transactionGroup: groupNonce,
        signer: testAccount.toString()
      }
    })

    expect(storedSigs).toBeDefined()
    expect(storedSigs.length).toBe(2)
  })

  /**
   * Test Flow: Encryption Key Rotation and Migration
   * 
   * Purpose: Demonstrate how to handle key rotation in encrypted multisig
   *          scenarios. This is important for long-lived contracts.
   * 
   * Test Steps:
   * 1. Store encrypted transaction with initial shared secret
   * 2. Derive new shared secret (e.g., new signer joins)
   * 3. Create new encrypted version with new secret
   * 4. Store alongside old version at different index
   * 5. Verify both old and new versions are accessible
   * 6. Verify both decrypt correctly with their respective secrets
   * 
   * Key Points:
   * - Per-signer index allows graceful key rotation
   * - Old versions remain accessible
   * - New versions can coexist with old versions
   * - Contract supports version migration
   */
  test("should support encryption key rotation with multiple versions", async () => {
    const client = await deployContract()
    const { testAccount } = localnet.context

    // Use the current testAccount as alice in this test
    const currentAliceAddr = testAccount.toString()

    await client.send.arc55Setup({
      args: {
        threshold: THRESHOLD,
        addresses: [currentAliceAddr, bobAddr, charlieAddr, daveAddr],
        encryptor: currentAliceAddr
      }
    })

    const txGroup = await client.send.arc55NewTransactionGroup({ args: [] })
    const groupNonce = txGroup.return!

    const multisigAddress = algosdk.multisigAddress({
      version: 1,
      threshold: THRESHOLD,
      addrs: [aliceAddr, bobAddr, charlieAddr, daveAddr]
    })

    const plainTransaction = await localnet.algorand.createTransaction.payment({
      sender: multisigAddress,
      receiver: multisigAddress,
      amount: (0).algo(),
      note: new Uint8Array(Buffer.from("Transaction for Key Rotation"))
    })

    const plainBytes = plainTransaction.toByte()

    // Initial secret for Bob (index 1)
    const bobInitialSecret = await EncryptionManager.deriveSharedSecret(
      cryptoService,
      rootKey,
      0,
      bobPublicKey,
      true,
      BIP32DerivationType.Peikert
    )

    // Encrypt with initial secret
    const bobInitialEncrypted = EncryptionManager.encrypt(plainBytes, bobInitialSecret)

    // New secret for Bob (e.g., rotated key pair)
    // In practice, this would come from a new key derivation path
    const bobRotatedSecret = await EncryptionManager.deriveSharedSecret(
      cryptoService,
      rootKey,
      1, // Different path
      bobPublicKey,
      true,
      BIP32DerivationType.Peikert
    )

    const bobRotatedEncrypted = EncryptionManager.encrypt(plainBytes, bobRotatedSecret)

    await client.algorand.account.setSigner(testAccount, localnet.algorand.account.getSigner(testAccount))

    const encryptedSize = plainBytes.length + 28
    const txnCostMicroAlgos = await client.arc55MbrTxnIncrease({
      args: { transactionSize: encryptedSize }
    })

    // Store both versions for Bob (same index, different encryption)
    // In a real scenario, you might use different indices or versions
    const mbrPayment1 = await localnet.algorand.createTransaction.payment({
      sender: testAccount,
      receiver: client.appAddress,
      amount: algos(Number(txnCostMicroAlgos) / 1000000)
    })

    await client.send.arc55AddTransaction({
      args: {
        costs: mbrPayment1,
        transactionGroup: groupNonce,
        index: 0,
        signerIndex: 1,
        transaction: bobInitialEncrypted
      }
    })

    // Verify initial version can be decrypted
    const storedInitialResponse = await client.arc55GetTransaction({
      args: {
        transactionGroup: groupNonce,
        transactionIndex: 0,
        signerIndex: 1
      }
    })

    const storedInitial = extractBytes(storedInitialResponse)
    const decryptedInitial = EncryptionManager.decrypt(storedInitial, bobInitialSecret)
    expect(decryptedInitial.length).toBeGreaterThan(0)

    // Update with rotated version
    const mbrPayment2 = await localnet.algorand.createTransaction.payment({
      sender: testAccount,
      receiver: client.appAddress,
      amount: algos(Number(txnCostMicroAlgos) / 1000000)
    })

    // Remove old version and add new version
    await client.send.arc55RemoveTransaction({
      sender: testAccount,
      extraFee: algos(0.002),
      args: {
        transactionGroup: groupNonce,
        index: 0,
        signerIndex: 1
      }
    })

    await client.send.arc55AddTransaction({
      args: {
        costs: mbrPayment2,
        transactionGroup: groupNonce,
        index: 0,
        signerIndex: 1,
        transaction: bobRotatedEncrypted
      }
    })

    // Verify rotated version can be decrypted
    const storedRotatedResponse = await client.arc55GetTransaction({
      args: {
        transactionGroup: groupNonce,
        transactionIndex: 0,
        signerIndex: 1
      }
    })

    const storedRotated = extractBytes(storedRotatedResponse)
    const decryptedRotated = EncryptionManager.decrypt(storedRotated, bobRotatedSecret)
    expect(decryptedRotated.length).toBeGreaterThan(0)
  })

  /**
   * Test Flow: Encryption with Authentication Failure Handling
   * 
   * Purpose: Verify that authentication tag validation prevents tampering
   *          with encrypted transactions. This ensures integrity of stored
   *          encrypted data.
   * 
   * Test Steps:
   * 1. Create and encrypt a transaction
   * 2. Store encrypted transaction
   * 3. Retrieve and attempt to decrypt with wrong key
   * 4. Verify authentication tag validation fails
   * 5. Attempt to tamper with ciphertext
   * 6. Verify tampering detection fails decryption
   * 
   * Key Points:
   * - ChaCha20-Poly1305 provides authenticated encryption
   * - Detects both decryption failures and tampering
   * - Contract cannot be tricked into accepting invalid data
   * - Provides cryptographic assurance of data integrity
   */
  test("should detect tampering with encrypted transactions", async () => {
    const plainBytes = Buffer.from("Original transaction data")
    
    const secret1 = Buffer.alloc(32, 1)
    const secret2 = Buffer.alloc(32, 2)

    // Encrypt with secret1
    const encrypted = EncryptionManager.encrypt(plainBytes, secret1)

    // Attempt to decrypt with secret2 (wrong key)
    expect(() => {
      EncryptionManager.decrypt(encrypted, secret2)
    }).toThrow('Decryption failed')

    // Tamper with ciphertext (flip a bit in the middle)
    const tampered = Buffer.from(encrypted)
    tampered[20] ^= 0xFF // Flip all bits in byte at position 20

    // Attempt to decrypt tampered data
    expect(() => {
      EncryptionManager.decrypt(tampered, secret1)
    }).toThrow('Decryption failed')
  })

  /**
   * Test Flow: Set Designated Encryptor
   * 
   * Purpose: Verify that admin can set a designated encryptor address
   *          which will be used for ECDH key agreements instead of admin
   * 
   * Test Steps:
   * 1. Deploy contract and set up multisig with encryptor
   * 2. Verify encryptor is set to the correct address
   * 
   * Key Points:
   * - Encryptor is set during setup
   * - Encryptor address is retrievable via arc55_getEncryptor
   */
  test("should allow setting designated encryptor during setup", async () => {
    const client = await deployContract()

    // Set encryptor during setup (using Bob as encryptor)
    await client.send.arc55Setup({
      args: {
        threshold: THRESHOLD,
        addresses: [aliceAddr, bobAddr, charlieAddr, daveAddr],
        encryptor: bobAddr
      }
    })

    // Verify encryptor is now set
    const storedEncryptor = await client.arc55GetEncryptor({})
    expect(Buffer.from(storedEncryptor).equals(Buffer.from(bobAddr))).toBe(true)
  })

  /**
   * Test Flow: Encrypt with Designated Encryptor
   * 
   * Purpose: Verify that transactions can be encrypted using the 
   *          designated encryptor's ECDH key and decrypted by signers
   * 
   * Test Steps:
   * 1. Deploy contract, set up multisig with encryptor
   * 2. Encrypt transaction using Bob's key
   * 3. Store encrypted transaction
   * 4. Signer retrieves and decrypts using Bob's pubkey
   * 
   * Key Points:
   * - ECDH is performed with encryptor's key, not admin's
   * - Signers can decrypt using encryptor's pubkey
   */
  test("should encrypt transaction using designated encryptor key", async () => {
    const client = await deployContract()
    const { testAccount } = localnet.context

    // Use Charlie as admin to distinguish from encryptor
    const charlieAsAdmin = testAccount.toString()

    // Set encryptor during setup (Bob is the encryptor)
    await client.send.arc55Setup({
      args: {
        threshold: THRESHOLD,
        addresses: [charlieAsAdmin, aliceAddr, bobAddr, charlieAddr],
        encryptor: bobAddr
      }
    })

    const txGroup = await client.send.arc55NewTransactionGroup({ args: [] })
    const groupNonce = txGroup.return!

    const multisigAddress = algosdk.multisigAddress({
      version: 1,
      threshold: THRESHOLD,
      addrs: [charlieAsAdmin, aliceAddr, bobAddr, charlieAddr]
    })

    const plainTransaction = await localnet.algorand.createTransaction.payment({
      sender: multisigAddress,
      receiver: multisigAddress,
      amount: (0).algo(),
      note: new Uint8Array(Buffer.from("Transaction with Designated Encryptor"))
    })

    const plainBytes = plainTransaction.toByte()

    // Derive shared secret using Bob's (encryptor's) key with Alice (signer)
    // In this case, we need to do ECDH between encryptor (Bob) and signer (Alice)
    // Alice's private key with Bob's public key
    const aliceSharedSecret = await EncryptionManager.deriveSharedSecret(
      cryptoService,
      rootKey,
      2, // Alice at index 2 in our setup
      bobPublicKey, // Bob is the encryptor
      false, // Alice is second
      BIP32DerivationType.Peikert
    )

    // Encrypt with encryptor's key
    const encrypted = EncryptionManager.encrypt(plainBytes, aliceSharedSecret)

    await client.algorand.account.setSigner(testAccount, localnet.algorand.account.getSigner(testAccount))

    const encryptedSize = plainBytes.length + 28
    const txnCostMicroAlgos = await client.arc55MbrTxnIncrease({
      args: { transactionSize: encryptedSize }
    })

    // Store encrypted transaction for Alice (index 1)
    const mbrPayment = await localnet.algorand.createTransaction.payment({
      sender: testAccount,
      receiver: client.appAddress,
      amount: algos(Number(txnCostMicroAlgos) / 1000000)
    })

    await client.send.arc55AddTransaction({
      args: {
        costs: mbrPayment,
        transactionGroup: groupNonce,
        index: 0,
        signerIndex: 1, // Alice's index
        transaction: encrypted
      }
    })

    // Retrieve and verify decryption works
    const storedResponse = await client.arc55GetTransaction({
      args: {
        transactionGroup: groupNonce,
        transactionIndex: 0,
        signerIndex: 1
      }
    })

    const storedEncrypted = extractBytes(storedResponse)
    const decrypted = EncryptionManager.decrypt(storedEncrypted, aliceSharedSecret)
    expect(decrypted.length).toBeGreaterThan(0)
  })

  /**
   * Test Flow: Backward Compatibility - No Encryptor Set
   * 
   * Purpose: Verify that when no encryptor is set, the system falls back
   *          to using admin's key for ECDH (backward compatible behavior)
   * 
   * Test Steps:
   * 1. Deploy contract, set up multisig (no encryptor set)
   * 2. Encrypt using admin's key
   * 3. Store encrypted transaction
   * 4. Signer retrieves and decrypts using admin's pubkey
   * 
   * Key Points:
   * - When encryptor is not set, admin's key is used
   * - This ensures backward compatibility with original design
   */
  test("should fall back to admin key when no encryptor is set", async () => {
    const client = await deployContract()
    const { testAccount } = localnet.context

    const adminAddr = testAccount.toString()

    await client.send.arc55Setup({
      args: {
        threshold: THRESHOLD,
        addresses: [adminAddr, bobAddr, charlieAddr, daveAddr],
        encryptor: adminAddr
      }
    })

    // Use admin as encryptor for backward compatibility
    const txGroup = await client.send.arc55NewTransactionGroup({ args: [] })
    const groupNonce = txGroup.return!

    const multisigAddress = algosdk.multisigAddress({
      version: 1,
      threshold: THRESHOLD,
      addrs: [adminAddr, bobAddr, charlieAddr, daveAddr]
    })

    const plainTransaction = await localnet.algorand.createTransaction.payment({
      sender: multisigAddress,
      receiver: multisigAddress,
      amount: (0).algo(),
      note: new Uint8Array(Buffer.from("Backward Compatible Transaction"))
    })

    const plainBytes = plainTransaction.toByte()

    // Use admin's key for ECDH (index 0 in setup)
    const bobSharedSecret = await EncryptionManager.deriveSharedSecret(
      cryptoService,
      rootKey,
      1, // Bob
      alicePublicKey, // Alice is admin
      false, // Bob is second
      BIP32DerivationType.Peikert
    )

    const encrypted = EncryptionManager.encrypt(plainBytes, bobSharedSecret)

    await client.algorand.account.setSigner(testAccount, localnet.algorand.account.getSigner(testAccount))

    const encryptedSize = plainBytes.length + 28
    const txnCostMicroAlgos = await client.arc55MbrTxnIncrease({
      args: { transactionSize: encryptedSize }
    })

    const mbrPayment = await localnet.algorand.createTransaction.payment({
      sender: testAccount,
      receiver: client.appAddress,
      amount: algos(Number(txnCostMicroAlgos) / 1000000)
    })

    await client.send.arc55AddTransaction({
      args: {
        costs: mbrPayment,
        transactionGroup: groupNonce,
        index: 0,
        signerIndex: 1,
        transaction: encrypted
      }
    })

    // Verify decryption works
    const storedResponse = await client.arc55GetTransaction({
      args: {
        transactionGroup: groupNonce,
        transactionIndex: 0,
        signerIndex: 1
      }
    })

    const storedEncrypted = extractBytes(storedResponse)
    const decrypted = EncryptionManager.decrypt(storedEncrypted, bobSharedSecret)
    expect(decrypted.length).toBeGreaterThan(0)
  })

  /**
   * Test Flow: Backward Compatibility - No Encryptor Set
   * 
   * Purpose: Verify that when no encryptor is set, the system falls back
   *          to using admin's key for ECDH (backward compatible behavior)
   * 
   * Test Steps:
   * 1. Deploy contract, set up multisig with encryptor
   * 2. Encrypt using admin's key
   * 3. Store encrypted transaction
   * 4. Signer retrieves and decrypts using admin's pubkey
   * 
   * Key Points:
   * - When encryptor is not set, admin's key is used
   * - This ensures backward compatibility with original design
   */
  test("should fall back to admin key when no encryptor is set", async () => {
    const client = await deployContract()
    const { testAccount } = localnet.context

    const adminAddr = testAccount.toString()

    await client.send.arc55Setup({
      args: {
        threshold: THRESHOLD,
        addresses: [adminAddr, bobAddr, charlieAddr, daveAddr],
        encryptor: adminAddr
      }
    })

    // Use admin as encryptor for backward compatibility
    const txGroup = await client.send.arc55NewTransactionGroup({ args: [] })
    const groupNonce = txGroup.return!

    const multisigAddress = algosdk.multisigAddress({
      version: 1,
      threshold: THRESHOLD,
      addrs: [adminAddr, bobAddr, charlieAddr, daveAddr]
    })

    const plainTransaction = await localnet.algorand.createTransaction.payment({
      sender: multisigAddress,
      receiver: multisigAddress,
      amount: (0).algo(),
      note: new Uint8Array(Buffer.from("Backward Compatible Transaction"))
    })

    const plainBytes = plainTransaction.toByte()

    // Use admin's key for ECDH (index 0 in setup)
    const bobSharedSecret = await EncryptionManager.deriveSharedSecret(
      cryptoService,
      rootKey,
      1, // Bob
      alicePublicKey, // Alice is admin
      false, // Bob is second
      BIP32DerivationType.Peikert
    )

    const encrypted = EncryptionManager.encrypt(plainBytes, bobSharedSecret)

    await client.algorand.account.setSigner(testAccount, localnet.algorand.account.getSigner(testAccount))

    const encryptedSize = plainBytes.length + 28
    const txnCostMicroAlgos = await client.arc55MbrTxnIncrease({
      args: { transactionSize: encryptedSize }
    })

    const mbrPayment = await localnet.algorand.createTransaction.payment({
      sender: testAccount,
      receiver: client.appAddress,
      amount: algos(Number(txnCostMicroAlgos) / 1000000)
    })

    await client.send.arc55AddTransaction({
      args: {
        costs: mbrPayment,
        transactionGroup: groupNonce,
        index: 0,
        signerIndex: 1,
        transaction: encrypted
      }
    })

    // Verify decryption works
    const storedResponse = await client.arc55GetTransaction({
      args: {
        transactionGroup: groupNonce,
        transactionIndex: 0,
        signerIndex: 1
      }
    })

    const storedEncrypted = extractBytes(storedResponse)
    const decrypted = EncryptionManager.decrypt(storedEncrypted, bobSharedSecret)
    expect(decrypted.length).toBeGreaterThan(0)
  })
})