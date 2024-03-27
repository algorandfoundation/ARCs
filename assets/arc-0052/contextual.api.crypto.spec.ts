import { CryptoKX, KeyPair, crypto_kx_client_session_keys, crypto_kx_server_session_keys, crypto_scalarmult, crypto_scalarmult_ed25519_base_noclamp, crypto_secretbox_easy, crypto_secretbox_open_easy, crypto_sign_ed25519_pk_to_curve25519, crypto_sign_ed25519_sk_to_curve25519, crypto_sign_keypair, ready, to_base64 } from "libsodium-wrappers-sumo"
import * as bip39 from "bip39"
import { randomBytes } from "crypto"
import { ContextualCryptoApi, ERROR_BAD_DATA, ERROR_TAGS_FOUND, Encoding, KeyContext, SignMetadata, harden } from "./contextual.api.crypto"
import * as msgpack from "algo-msgpack-with-bigint"
import { fromSeed } from "./bip32-ed25519"
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

    describe("\(JS Library) Reference Implementation alignment with known BIP32-Ed25519 JS LIB", () => {
        it("\(OK) BIP32-Ed25519 derive key m'/44'/283'/0'/0/0", async () => {
            await ready
            const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 0, 0)

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
            const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 0, 1)

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
            const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 1, 1)

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
            const key: Uint8Array = await cryptoService.keyGen(KeyContext.Identity, 1, 2)

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
                        expect(key).toEqual(new Uint8Array(Buffer.from("62fe832b7ad10544be8337a670435e5064ae4a66e77bd78909765b46b576a6f3", "hex")))
                    })
            
                    it("\(OK) Derive m'/44'/283'/0'/0/1 Algorand Address Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 0, 1)
                        expect(key).toEqual(new Uint8Array(Buffer.from("530461002eaccec0c7b5795925aa104a7fb45f85ef0aa95bbb5be93b6f8537ad", "hex")))
                    })
            
                    it("\(OK) Derive m'/44'/283'/0'/0/2 Algorand Address Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 0, 2)
                        expect(key).toEqual(new Uint8Array(Buffer.from("2281c81bee04ee039fa482c283541c6ab06c8324db6f1cc59c68252e1d58bcb3", "hex")))
                    })

                })

                describe("Hard Derivations", () => {
                    it("\(OK) Derive m'/44'/283'/1'/0/0 Algorand Address Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 1, 0)
                        expect(key).toEqual(new Uint8Array(Buffer.from("9e12643f6c0068dcf53b04daced6f8c1a90ad21c954a66df4140d79303166a67", "hex")))
                    })
        
                    it("\(OK) Derive m'/44'/283'/2'/0/1 Algorand Address Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 2, 1)
                        expect(key).toEqual(new Uint8Array(Buffer.from("8a5ddf62d51a2c50e51dbad4634356cc72314a81edd917ac91da96477a9fb5b0", "hex")))
                    })

                    it("\(OK) Derive m'/44'/283'/3'/0/0 Algorand Address Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 3, 0)
                        expect(key).toEqual(new Uint8Array(Buffer.from("2358e0f2b465ab3e8f55139d8316654d4be39ebb22367d36409fd02a20b0e017", "hex")))
                    })
                })
            })

            describe("Identities", () => {
                describe("Soft Derivations", () => {
                    it("\(OK) Derive m'/44'/0'/0'/0/0 Identity Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Identity, 0, 0)
                        expect(key).toEqual(new Uint8Array(Buffer.from("b6d7eea5af0ad83edf4340659e72f0ea2b4566de1fc3b63a40a425aabebe5e49", "hex")))
                    })
            
                    it("\(OK) Derive m'/44'/0'/0'/0/1 Identity Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Identity, 0, 1)
                        expect(key).toEqual(new Uint8Array(Buffer.from("b5cec676c5a2129ed1be4223a2702439bbb2462fd77b43f27e2f79fd194a30a2", "hex")))
                    })
            
                    it("\(OK) Derive m'/44'/0'/0'/0/2 Identity Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Identity, 0, 2)
                        expect(key).toEqual(new Uint8Array(Buffer.from("435e5e3446431d462572abee1b8badb88608906a6af27b8497bccfd503edb6fe", "hex")))
                    })
                })

                describe("Hard Derivations", () => {
                    it("\(OK) Derive m'/44'/0'/1'/0/0 Identity Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Identity, 1, 0)
                        expect(key).toEqual(new Uint8Array(Buffer.from("bf63be83fff9bc9d0aebc231d50342110e5220247e50de376b47e154b5d32a3e", "hex")))
                    })
        
                    it("\(OK) Derive m'/44'/0'/2'/0/1 Identity Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Identity, 2, 1)              
                        expect(key).toEqual(new Uint8Array(Buffer.from("edb10fff24a4745df52f1a0ab1ae71b3752d019c8c2437d46ab8c8e634a74cd4", "hex")))
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
                const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 0, 0)
                // this transaction wes successfully submitted to the network https://testnet.explorer.perawallet.app/tx/UJG3NVCSCW5A63KPV35BPAABLXMXTTEM2CVUKNS4EML3H3EYGMCQ/
                const prefixEncodedTx = new Uint8Array(Buffer.from('VFiJo2FtdM0D6KNmZWXNA+iiZnbOAkeSd6NnZW6sdGVzdG5ldC12MS4womdoxCBIY7UYpLPITsgQ8i1PEIHLD3HwWaesIN7GL39w5Qk6IqJsds4CR5Zfo3JjdsQgYv6DK3rRBUS+gzemcENeUGSuSmbne9eJCXZbRrV2pvOjc25kxCBi/oMretEFRL6DN6ZwQ15QZK5KZud714kJdltGtXam86R0eXBlo3BheQ==', 'base64'))
                const sig = await cryptoService.signAlgoTransaction(KeyContext.Address, 0, 0, prefixEncodedTx)
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
                expect(cipherText).toEqual(new Uint8Array([251,7,48,58,57,22,135,152,150,116,242,138,26,155,136,252,163,209,7,34,125,135,218,222,102,45,250,55,34]))
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
})
