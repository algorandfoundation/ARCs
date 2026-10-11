import { CryptoKX, KeyPair, crypto_kx_client_session_keys, crypto_kx_server_session_keys, crypto_scalarmult, crypto_scalarmult_ed25519_base_noclamp, crypto_secretbox_easy, crypto_secretbox_open_easy, crypto_sign_ed25519_pk_to_curve25519, crypto_sign_ed25519_sk_to_curve25519, crypto_sign_keypair, ready, to_base64 } from "libsodium-wrappers-sumo"
import * as bip39 from "bip39"
import { randomBytes } from "crypto"
import { BIP32DerivationType, ContextualCryptoApi, ERROR_TAGS_FOUND, Encoding, KeyContext, SignMetadata, harden } from "./contextual.api.crypto"
import * as msgpack from "algo-msgpack-with-bigint"
import { deriveChildNodePrivate, deriveChildNodePublic, fromSeed } from "./bip32-ed25519"
import { sha512_256 } from "js-sha512"
import base32 from "hi-base32"
import { JSONSchemaType } from "ajv"
import { readFileSync } from "fs"
import path from "path"
import nacl from "tweetnacl"
const libBip32Ed25519 = require('bip32-ed25519')

function encodeAddress(publicKey: Buffer): string {
    const keyHash: string = sha512_256.create().update(publicKey).hex()

    // last 4 bytes of the hash
    const checksum: string = keyHash.slice(-8)

    return base32.encode(ConcatArrays(publicKey, Buffer.from(checksum, "hex"))).slice(0, 58)
}

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

describe("Contextual Derivation & Signing", () => {

    let cryptoService: ContextualCryptoApi
    let bip39Mnemonic: string = "salon zoo engage submit smile frost later decide wing sight chaos renew lizard rely canal coral scene hobby scare step bus leaf tobacco slice"
    let seed: Buffer

    beforeAll(() => {
        seed = bip39.mnemonicToSeedSync(bip39Mnemonic, "")
    })
    
	beforeEach(() => {
        cryptoService = new ContextualCryptoApi(seed)
    })

	afterEach(() => {})

    /**
     * Testing against other known bip32-ed25519 lib that complies with the BIP32-Ed25519 specification
     * 
     * @see BIP32-ed25519 Hierarchical Deterministic Keys over a Non-linear Keyspace (https://acrobat.adobe.com/id/urn:aaid:sc:EU:04fe29b0-ea1a-478b-a886-9bb558a5242a)
     * 
     * We call the traditional derivation Khovratovich
     */
    describe("\(JS Library) Reference Implementation alignment with known BIP32-Ed25519 JS LIB", () => {
        it("\(OK) BIP32-Ed25519 derive key m'/44'/283'/0'/0/0", async () => {
            await ready
            const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 0, 0, BIP32DerivationType.Khovratovich)

            const rooKey: Uint8Array = fromSeed(seed)
  
            let derivedKey: Uint8Array = libBip32Ed25519.derivePrivate(Buffer.from(rooKey), harden(44))
            derivedKey = libBip32Ed25519.derivePrivate(derivedKey, harden(283))
            derivedKey = libBip32Ed25519.derivePrivate(derivedKey, harden(0))
            derivedKey = libBip32Ed25519.derivePrivate(derivedKey, 0)
            derivedKey = libBip32Ed25519.derivePrivate(derivedKey, 0)

            const scalar = derivedKey.subarray(0, 32) // scalar == pvtKey
            const derivedPub: Uint8Array = crypto_scalarmult_ed25519_base_noclamp(scalar) // calculate public key
            expect(derivedPub).toEqual(key)
        })

        it("\(OK) BIP32-Ed25519 derive key m'/44'/283'/0'/0/1", async () => {
            await ready
            const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 0, 1, BIP32DerivationType.Khovratovich)

            const rooKey: Uint8Array = fromSeed(seed)
  
            let derivedKey: Uint8Array = libBip32Ed25519.derivePrivate(Buffer.from(rooKey), harden(44))
            derivedKey = libBip32Ed25519.derivePrivate(derivedKey, harden(283))
            derivedKey = libBip32Ed25519.derivePrivate(derivedKey, harden(0))
            derivedKey = libBip32Ed25519.derivePrivate(derivedKey, 0)
            derivedKey = libBip32Ed25519.derivePrivate(derivedKey, 1)

            const scalar = derivedKey.subarray(0, 32) // scalar == pvtKey
            const derivedPub: Uint8Array = crypto_scalarmult_ed25519_base_noclamp(scalar) // calculate public key
            expect(derivedPub).toEqual(key)
        })

        it("\(OK) BIP32-Ed25519 derive PUBLIC key m'/44'/283'/1'/0/1", async () => {
            await ready
            const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 1, 1, BIP32DerivationType.Khovratovich)

            const rooKey: Uint8Array = fromSeed(seed)
  
            let derivedKey: Uint8Array = libBip32Ed25519.derivePrivate(Buffer.from(rooKey), harden(44))
            derivedKey = libBip32Ed25519.derivePrivate(derivedKey, harden(283))
            derivedKey = libBip32Ed25519.derivePrivate(derivedKey, harden(1))

            // ext private => ext public format!
            const nodeScalar: Uint8Array = derivedKey.subarray(0, 32)
            const nodePublic: Uint8Array = crypto_scalarmult_ed25519_base_noclamp(nodeScalar)
            const nodeCC: Uint8Array = derivedKey.subarray(64, 96)

            // [Public][ChainCode]
            const extPub: Buffer = Buffer.concat([nodePublic, nodeCC])
            
            derivedKey = libBip32Ed25519.derivePublic(extPub, 0)
            derivedKey = libBip32Ed25519.derivePublic(derivedKey, 1)

            const derivedPub = new Uint8Array(derivedKey.subarray(0, 32)) // public key from extended format
            expect(derivedPub).toEqual(key)
        })

        it("\(OK) BIP32-Ed25519 derive PUBLIC key m'/44'/0'/1'/0/2", async () => {
            await ready
            const key: Uint8Array = await cryptoService.keyGen(KeyContext.Identity, 1, 2, BIP32DerivationType.Khovratovich)

            const rooKey: Uint8Array = fromSeed(seed)
  
            let derivedKey: Uint8Array = libBip32Ed25519.derivePrivate(Buffer.from(rooKey), harden(44))
            derivedKey = libBip32Ed25519.derivePrivate(derivedKey, harden(0))
            derivedKey = libBip32Ed25519.derivePrivate(derivedKey, harden(1))

            // ext private => ext public format!
            const nodeScalar: Uint8Array = derivedKey.subarray(0, 32)
            const nodePublic: Uint8Array = crypto_scalarmult_ed25519_base_noclamp(nodeScalar)
            const nodeCC: Uint8Array = derivedKey.subarray(64, 96)

            // [Public][ChainCode]
            const extPub: Buffer = Buffer.concat([nodePublic, nodeCC])
            
            derivedKey = libBip32Ed25519.derivePublic(extPub, 0)
            derivedKey = libBip32Ed25519.derivePublic(derivedKey, 2)

            const derivedPub = new Uint8Array(derivedKey.subarray(0, 32)) // public key from extended format
            expect(derivedPub).toEqual(key)
        })
    })

    it("\(OK) Root Key", async () => {
        const rootKey: Uint8Array = fromSeed(seed)
        expect(rootKey.length).toBe(96)
        expect(Buffer.from(rootKey)).toEqual(Buffer.from("a8ba80028922d9fcfa055c78aede55b5c575bcd8d5a53168edf45f36d9ec8f4694592b4bc892907583e22669ecdf1b0409a9f3bd5549f2dd751b51360909cd05796b9206ec30e142e94b790a98805bf999042b55046963174ee6cee2d0375946", "hex"))
    })

    describe("\(Derivations) Context", () => {
            describe("Addresses", () => {
                describe("Soft Derivations", () => {
                    it("\(OK) Derive m'/44'/283'/0'/0/0 Algorand Address Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 0, 0)
                        expect(key).toEqual(new Uint8Array(Buffer.from("7bda7ac12627b2c259f1df6875d30c10b35f55b33ad2cc8ea2736eaa3ebcfab9", "hex")))
                    })
                    
                    it("\(OK) Derive m'/44'/283'/0'/0/1 Algorand Address Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 0, 1)
                        expect(key).toEqual(new Uint8Array(Buffer.from("5bae8828f111064637ac5061bd63bc4fcfe4a833252305f25eeab9c64ecdf519", "hex")))
                    })
                    
                    it("\(OK) Derive m'/44'/283'/0'/0/2 Algorand Address Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 0, 2)
                        expect(key).toEqual(new Uint8Array(Buffer.from("00a72635e97cba966529e9bfb4baf4a32d7b8cd2fcd8e2476ce5be1177848cb3", "hex")))
                    })
                    
                })
                
                describe("Hard Derivations", () => {
                    it("\(OK) Derive m'/44'/283'/1'/0/0 Algorand Address Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 1, 0)
                        expect(key).toEqual(new Uint8Array(Buffer.from("358d8c4382992849a764438e02b1c45c2ca4e86bbcfe10fd5b963f3610012bc9", "hex")))
                    })
                    
                    it("\(OK) Derive m'/44'/283'/2'/0/1 Algorand Address Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 2, 1)
                        expect(key).toEqual(new Uint8Array(Buffer.from("1f0f75fbbca12b22523973191061b2f96522740e139a3420c730717ac5b0dfc0", "hex")))
                    })
                    
                    it("\(OK) Derive m'/44'/283'/3'/0/0 Algorand Address Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 3, 0)
                        expect(key).toEqual(new Uint8Array(Buffer.from("f035316f915b342ea5fe78dccb59d907b93805732219d436a1bd8488ff4e5b1b", "hex")))
                    })
                })
            })
            
            describe("Identities", () => {
                describe("Soft Derivations", () => {
                    it("\(OK) Derive m'/44'/0'/0'/0/0 Identity Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Identity, 0, 0)
                        expect(key).toEqual(new Uint8Array(Buffer.from("ff8b1863ef5e40d0a48c245f26a6dbdf5da94dc75a1851f51d8a04e547bd5f5a", "hex")))
                    })
                    
                    it("\(OK) Derive m'/44'/0'/0'/0/1 Identity Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Identity, 0, 1)
                        expect(key).toEqual(new Uint8Array(Buffer.from("2b46c2af0890493e486049d456509a0199e565b41a5fb622f0ea4b9337bd2b97", "hex")))
                    })
                    
                    it("\(OK) Derive m'/44'/0'/0'/0/2 Identity Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Identity, 0, 2)
                        expect(key).toEqual(new Uint8Array(Buffer.from("2713f135f19ef3dcfca73cb536b1e077b1165cd0b7bedbef709447319ff0016d", "hex")))
                    })
                })
                
                describe("Hard Derivations", () => {
                    it("\(OK) Derive m'/44'/0'/1'/0/0 Identity Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Identity, 1, 0)
                        expect(key).toEqual(new Uint8Array(Buffer.from("232847ae1bb95babcaa50c8033fab98f59e4b4ad1d89ac523a90c830e4ceee4a", "hex")))
                    })
                    
                    it("\(OK) Derive m'/44'/0'/2'/0/1 Identity Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Identity, 2, 1)
                        expect(key).toEqual(new Uint8Array(Buffer.from("8f68b6572860d84e8a41e38db1c8c692ded5eb291846f2e5bbfde774a9c6d16e", "hex")))
                    })
                })
            })

        describe("Signing Typed Data", () => {
            it("\(OK) Sign authentication challenge of 32 bytes, encoded base64", async () => {
                const challenge: Uint8Array = new Uint8Array(randomBytes(32))
                
                // read auth schema file for authentication. 32 bytes challenge to sign
                const authSchema: JSONSchemaType<any> = JSON.parse(readFileSync(path.resolve(__dirname, "schemas/auth.request.json"), "utf8"))
                const metadata: SignMetadata = { encoding: Encoding.BASE64, schema: authSchema }                
                const base64Challenge: string = Buffer.from(challenge).toString("base64")

                const encoded: Uint8Array = new Uint8Array(Buffer.from(base64Challenge))

                const signature: Uint8Array = await cryptoService.signData(KeyContext.Address,0, 0, encoded,  metadata)
                expect(signature).toHaveLength(64)

                const isValid: boolean = await cryptoService.verifyWithPublicKey(signature, encoded, await cryptoService.keyGen(KeyContext.Address, 0, 0))
                expect(isValid).toBe(true)
            })

            it("\(OK) Sign authentication challenge of 32 bytes, encoded msgpack", async () => {
                const challenge: Uint8Array = new Uint8Array(randomBytes(32))

                // read auth schema file for authentication. 32 bytes challenge to sign
                const authSchema: JSONSchemaType<any> = JSON.parse(readFileSync(path.resolve(__dirname, "schemas/auth.request.json"), "utf8"))
                const metadata: SignMetadata = { encoding: Encoding.MSGPACK, schema: authSchema }
                const encoded: Uint8Array = msgpack.encode(challenge)

                const signature: Uint8Array = await cryptoService.signData(KeyContext.Address,0, 0, encoded,  metadata)
                expect(signature).toHaveLength(64)

                const isValid: boolean = await cryptoService.verifyWithPublicKey(signature, encoded, await cryptoService.keyGen(KeyContext.Address, 0, 0))
                expect(isValid).toBe(true)
            })

            it ("\(OK) Sign authentication challenge of 32 bytes, no encoding", async () => {
                const challenge: Uint8Array = new Uint8Array(randomBytes(32))

                // read auth schema file for authentication. 32 bytes challenge to sign
                const authSchema: JSONSchemaType<any> = JSON.parse(readFileSync(path.resolve(__dirname, "schemas/auth.request.json"), "utf8"))
                const metadata: SignMetadata = { encoding: Encoding.NONE, schema: authSchema }

                const signature: Uint8Array = await cryptoService.signData(KeyContext.Address,0, 0, challenge,  metadata)
                expect(signature).toHaveLength(64)

                const isValid: boolean = await cryptoService.verifyWithPublicKey(signature, challenge, await cryptoService.keyGen(KeyContext.Address, 0, 0))
                expect(isValid).toBe(true)
            })

            it("\(OK) Sign Arbitrary Message against Schema, encoded base64", async () => {
                const firstKey: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 0, 0)
    
                const message = {
                    letter: "Hello World"
                }

                const encoded: Buffer = Buffer.from(to_base64(JSON.stringify(message)))

                // Schema of what we are signing
                const jsonSchema = {
                    type: "object",
                    properties: {
                        letter: {
                            type: "string"
                        }
                    }
                }

                const metadata: SignMetadata = { encoding: Encoding.BASE64, schema: jsonSchema } 

                const signature: Uint8Array = await cryptoService.signData(KeyContext.Address,0, 0, encoded,  metadata)
                expect(signature).toHaveLength(64)
    
                const isValid: boolean = await cryptoService.verifyWithPublicKey(signature, encoded, firstKey)
                expect(isValid).toBe(true)
            })

            it("\(OK) Sign Arbitrary Message against Schema, encoded msgpack", async () => {
                const firstKey: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 0, 0)

                const message = {
                    letter: "Hello World"
                }

                const encoded: Buffer = Buffer.from(msgpack.encode(message))

                // Schema of what we are signing
                const jsonSchema = {
                    type: "object",
                    properties: {
                        letter: {
                            type: "string"
                        }
                    }
                }

                const metadata: SignMetadata = { encoding: Encoding.MSGPACK, schema: jsonSchema }

                const signature: Uint8Array = await cryptoService.signData(KeyContext.Address,0, 0, encoded,  metadata)
                expect(signature).toHaveLength(64)

                const isValid: boolean = await cryptoService.verifyWithPublicKey(signature, encoded, firstKey)
                expect(isValid).toBe(true)
            })

            it("\(FAIL) Signing attempt fails because of invalid data against Schema, encoded base64", async () => {
                const message = {
                    letter: "Hello World"
                }

                const encoded: Buffer = Buffer.from(to_base64(JSON.stringify(message)))

                // Schema of what we are signing
                const jsonSchema = {
                    type: "string"
                }

                const metadata: SignMetadata = { encoding: Encoding.BASE64, schema: jsonSchema } 
                expect(cryptoService.signData(KeyContext.Identity,0, 0, encoded,  metadata)).rejects.toThrowError()
            })

            it("\(FAIL) Signing attempt fails because of invalid data against Schema, encoded msgpack", async () => {
                const message = {
                    letter: "Hello World"
                }

                const encoded: Buffer = Buffer.from(msgpack.encode(message))

                // Schema of what we are signing
                const jsonSchema = {
                    type: "string"
                }

                const metadata: SignMetadata = { encoding: Encoding.MSGPACK, schema: jsonSchema }
                expect(cryptoService.signData(KeyContext.Identity,0, 0, encoded,  metadata)).rejects.toThrowError()
            })

            describe("Reject Regular Transaction Signing. IF TAG Prexies are present signing must fail", () => {
                it("\(FAIL) [TX] Tag", async () => {
                    const transaction: Buffer = Buffer.concat([Buffer.from("TX"), msgpack.encode(randomBytes(64))])
                    const metadata: SignMetadata = { encoding: Encoding.MSGPACK, schema: {} }
                    expect(cryptoService.signData(KeyContext.Identity,0, 0, transaction,  metadata)).rejects.toThrowError(ERROR_TAGS_FOUND)
                })

                it("\(FAIL) [MX] Tag", async () => {
                    const transaction: Buffer = Buffer.concat([Buffer.from("MX"), msgpack.encode(randomBytes(64))])
                    const metadata: SignMetadata = { encoding: Encoding.MSGPACK, schema: {} }
                    expect(cryptoService.signData(KeyContext.Identity,0, 0, transaction,  metadata)).rejects.toThrowError(ERROR_TAGS_FOUND)
                })

                it("\(FAIL) [Program] Tag", async () => {
                    const transaction: Buffer = Buffer.concat([Buffer.from("Program"), msgpack.encode(randomBytes(64))])
                    const metadata: SignMetadata = { encoding: Encoding.MSGPACK, schema: {} }
                    expect(cryptoService.signData(KeyContext.Identity,0, 0, transaction,  metadata)).rejects.toThrowError(ERROR_TAGS_FOUND)
                })

                it("\(FAIL) [progData] Tag", async () => {
                    const transaction: Buffer = Buffer.concat([Buffer.from("ProgData"), msgpack.encode(randomBytes(64))])
                    const metadata: SignMetadata = { encoding: Encoding.MSGPACK, schema: {} }
                    expect(cryptoService.signData(KeyContext.Identity,0, 0, transaction,  metadata)).rejects.toThrowError(ERROR_TAGS_FOUND)
                })
            })
        })

        describe("signing transactions", () => {
            it("\(OK) Sign Transaction", async () => {
                const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 0, 0, BIP32DerivationType.Khovratovich)
                // this transaction wes successfully submitted to the network https://testnet.explorer.perawallet.app/tx/UJG3NVCSCW5A63KPV35BPAABLXMXTTEM2CVUKNS4EML3H3EYGMCQ/
                const prefixEncodedTx = new Uint8Array(Buffer.from('VFiJo2FtdM0D6KNmZWXNA+iiZnbOAkeSd6NnZW6sdGVzdG5ldC12MS4womdoxCBIY7UYpLPITsgQ8i1PEIHLD3HwWaesIN7GL39w5Qk6IqJsds4CR5Zfo3JjdsQgYv6DK3rRBUS+gzemcENeUGSuSmbne9eJCXZbRrV2pvOjc25kxCBi/oMretEFRL6DN6ZwQ15QZK5KZud714kJdltGtXam86R0eXBlo3BheQ==', 'base64'))
                const sig = await cryptoService.signAlgoTransaction(KeyContext.Address, 0, 0, prefixEncodedTx, BIP32DerivationType.Khovratovich)
                expect(encodeAddress(Buffer.from(key))).toEqual("ML7IGK322ECUJPUDG6THAQ26KBSK4STG4555PCIJOZNUNNLWU3Z3ZFXITA")
                expect(nacl.sign.detached.verify(prefixEncodedTx, sig, key)).toBe(true)
            })
        })

        describe("ECDH cases", () => {
            // Making sure Alice & Bob Have different root keys 
            let aliceCryptoService: ContextualCryptoApi
            let bobCryptoService: ContextualCryptoApi
            beforeEach(() => {
                aliceCryptoService = new ContextualCryptoApi(bip39.mnemonicToSeedSync("exact remain north lesson program series excess lava material second riot error boss planet brick rotate scrap army riot banner adult fashion casino bamboo", ""))
                bobCryptoService = new ContextualCryptoApi(bip39.mnemonicToSeedSync("identify length ranch make silver fog much puzzle borrow relax occur drum blue oval book pledge reunion coral grace lamp recall fever route carbon", ""))
            })

            it("\(OK) ECDH", async () => {
                const aliceKey: Uint8Array = await aliceCryptoService.keyGen(KeyContext.Identity, 0, 0)
                const bobKey: Uint8Array = await bobCryptoService.keyGen(KeyContext.Identity, 0, 0)

                const aliceSharedSecret: Uint8Array = await aliceCryptoService.ECDH(KeyContext.Identity, 0, 0, bobKey, true)
                const bobSharedSecret: Uint8Array = await bobCryptoService.ECDH(KeyContext.Identity, 0, 0, aliceKey, false)
                expect(aliceSharedSecret).toEqual(bobSharedSecret)

                const aliceSharedSecret2: Uint8Array = await aliceCryptoService.ECDH(KeyContext.Identity, 0, 0, bobKey, false)
                const bobSharedSecret2: Uint8Array = await bobCryptoService.ECDH(KeyContext.Identity, 0, 0, aliceKey, true)
                expect(aliceSharedSecret2).toEqual(bobSharedSecret2)
                expect(aliceSharedSecret2).not.toEqual(aliceSharedSecret)

            })
    
            it("\(OK) ECDH, Encrypt and Decrypt", async () => {
                const aliceKey: Uint8Array = await aliceCryptoService.keyGen(KeyContext.Identity, 0, 0)
                const bobKey: Uint8Array = await bobCryptoService.keyGen(KeyContext.Identity, 0, 0)
    
                const aliceSharedSecret: Uint8Array = await aliceCryptoService.ECDH(KeyContext.Identity, 0, 0, bobKey, true)
                const bobSharedSecret: Uint8Array = await bobCryptoService.ECDH(KeyContext.Identity, 0, 0, aliceKey, false)

                expect(aliceSharedSecret).toEqual(bobSharedSecret)
    
                const message: Uint8Array = new Uint8Array(Buffer.from("Hello, World!"))
                const nonce: Uint8Array = new Uint8Array([16,197,142,8,174,91,118,244,202,136,43,200,97,242,104,99,42,154,191,32,67,30,6,123])
    
                // encrypt
                const cipherText: Uint8Array = crypto_secretbox_easy(message, nonce, aliceSharedSecret)

                // log cipherText uint8array
                console.log("cipherText", cipherText)

                expect(cipherText).toEqual(new Uint8Array([20, 107, 126, 154, 152, 197, 252, 227, 148,  39, 245, 136, 233,  10, 199,  20, 219,   3,  53,   2, 113, 6, 190,  21, 193, 119,  43,  44, 230]))
                // decrypt
                const plainText: Uint8Array = crypto_secretbox_open_easy(cipherText, nonce, bobSharedSecret)
                expect(plainText).toEqual(message)
            })
    
            it("Libsodium example ECDH", async () => {
                await ready
                // keypair
                const alice: KeyPair = crypto_sign_keypair()
    
                const alicePvtKey: Uint8Array = alice.privateKey
                const alicePubKey: Uint8Array = alice.publicKey
    
                const aliceXPvt: Uint8Array = crypto_sign_ed25519_sk_to_curve25519(alicePvtKey)
                const aliceXPub: Uint8Array = crypto_sign_ed25519_pk_to_curve25519(alicePubKey)
        
                // bob
                const bob: KeyPair = crypto_sign_keypair()
                
                const bobPvtKey: Uint8Array = bob.privateKey
                const bobPubKey: Uint8Array = bob.publicKey
    
                const bobXPvt: Uint8Array = crypto_sign_ed25519_sk_to_curve25519(bobPvtKey)
                const bobXPub: Uint8Array = crypto_sign_ed25519_pk_to_curve25519(bobPubKey)
    
                // shared secret
                const aliceSecret: Uint8Array = crypto_scalarmult(aliceXPvt, bobXPub)
                const bobSecret: Uint8Array = crypto_scalarmult(bobXPvt, aliceXPub)
                expect(aliceSecret).toEqual(bobSecret)
    
                const aliceSharedSecret: CryptoKX = crypto_kx_client_session_keys(aliceXPub, aliceXPvt, bobXPub)
                const bobSharedSecret: CryptoKX = crypto_kx_server_session_keys(bobXPub, bobXPvt, aliceXPub)
    
                // bilateral encryption channels
                expect(aliceSharedSecret.sharedRx).toEqual(bobSharedSecret.sharedTx)
                expect(bobSharedSecret.sharedTx).toEqual(aliceSharedSecret.sharedRx)
            })
        })
    })

    describe("\(deriveNodePrivate)", () => {
        it("\(FAIL) Should fail if during derivation scalar >= 2^255", async () => {
            await ready // libsodium ready

            const rootKey = fromSeed(seed)
            // 44'/283'/0'/0
            const bip44Path = [harden(44), harden(283), harden(0), 0, 0]

            await ready // libsodium

            // Pick `g`, which is amount of bits zeroed from each derived node
            const g: number = 9

            // 44'
            let derivationNode: Uint8Array = await deriveChildNodePrivate(rootKey, bip44Path[0], g)
            // 283'
            derivationNode = await deriveChildNodePrivate(derivationNode, bip44Path[1], g)
            // 0'
            derivationNode = await deriveChildNodePrivate(derivationNode, bip44Path[2], g)
            // 0
            derivationNode = await deriveChildNodePrivate(derivationNode, bip44Path[3], g)
            // 0
            derivationNode = await deriveChildNodePrivate(derivationNode, bip44Path[4], g)
    
            for (let i = 0; i < 19; i++) {
                derivationNode = await deriveChildNodePrivate(derivationNode, 0, g)
            }

            // for the seed in this test, we know where the scalar breaks (>= 2 ^ 255)
            // expect error at the 20th level for this known seed
            expect(deriveChildNodePrivate(derivationNode, 0, g)).rejects.toThrow("zL * 8 is larger than 2^255, which is not safe")
        })
    })

    describe("\(deriveNodePublic", () => {
        it("\(OK) From m'/44'/283'/0'/0 root level (excluding address index) derive N keys with only public information", async () => {
            
            // wallet level m'/44'/283'/0'/0 root; node derivation before address_index 
            const walletRoot: Uint8Array = await cryptoService.deriveKey(fromSeed(seed), [harden(44), harden(283), harden(0), 0], false, BIP32DerivationType.Peikert)

            // should be able to derive all public keys from this root without knowing private information
            // since these are SOFTLY derived
            const numPublicKeysToDerive: number = 10
            for (let i = 0; i < numPublicKeysToDerive; i++) {
                // assuming in a third party that only has public information
                // I'm provided with the wallet level m'/44'/283'/0'/0 root [public, chaincode]
                // no private information is shared
                // i can SOFTLY derive N public keys / addresses from this root
                const derivedKey: Uint8Array = new Uint8Array(await deriveChildNodePublic(walletRoot, i, BIP32DerivationType.Peikert)) // g == 9
            
                // Deriving from my own wallet where i DO have private information
                const myKey: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 0, i, BIP32DerivationType.Peikert)
                
                // they should match 
                // derivedKey.subarray(0, 32) ==  public key (excluding chaincode)
                expect(derivedKey.slice(0, 32)).toEqual(myKey)
            }
        })

        it("\(FAIL) From m'/44'/283'/0'/0' root level (excluding address index) should not be able to derive correct addresses from a hardened derivation", async () => {
            
            // wallet level m'/44'/283'/0'/0' root; node derivation before address_index 
            const walletRoot: Uint8Array = await cryptoService.deriveKey(fromSeed(seed), [harden(44), harden(283), harden(0), harden(0)], false, BIP32DerivationType.Peikert)


            const numPublicKeysToDerive: number = 10
            for (let i = 0; i < numPublicKeysToDerive; i++) {                
                const derivedKey: Uint8Array = new Uint8Array(await deriveChildNodePublic(walletRoot, i, BIP32DerivationType.Peikert)) // g == 9
            
                // Deriving from my own wallet where i DO have private information
                const myKey: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 0, i, BIP32DerivationType.Peikert)
                
                // they should NOT match  since the `change` level (as part of BIP44) was hardened
                // derivedKey.subarray(0, 32) ==  public key (excluding chaincode)
                expect(derivedKey.slice(0, 32)).not.toEqual(myKey)
            }
        })
    })
})
