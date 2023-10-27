import { CryptoKX, KeyPair, crypto_kx_client_session_keys, crypto_kx_server_session_keys, crypto_scalarmult, crypto_scalarmult_ed25519_base_noclamp, crypto_secretbox_NONCEBYTES, crypto_secretbox_easy, crypto_secretbox_open_easy, crypto_sign_ed25519_pk_to_curve25519, crypto_sign_ed25519_sk_to_curve25519, crypto_sign_keypair, crypto_sign_seed_keypair, ready, to_base64 } from "libsodium-wrappers-sumo"
import * as bip39 from "bip39"
import { sha512_256 } from "js-sha512"
import base32 from "hi-base32"
import { randomBytes } from "crypto"
import { ContextualCryptoApi, Encoding, KeyContext, SignMetadata } from "./contextual.api.crypto"

/**
 * @param publicKey 
 * @returns 
 */
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

    describe("\(Derivations) Context", () => {
            describe("Addresses", () => {
                describe("Soft Derivations", () => {
                    it("\(OK) Derive m'/44'/283'/0'/0/0 Algorand Address Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 0, 0)
                        expect(key).toEqual(new Uint8Array(Buffer.from("9183c50768a724c5b77a4ddf56d7da2f204d56da68ced53626222ace25b25cd8", "hex")))
                    })
            
                    it("\(OK) Derive m'/44'/283'/0'/0/1 Algorand Address Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 0, 1)
                        expect(key).toEqual(new Uint8Array(Buffer.from("15965b321e69bfb6386a1dc31925ee608f8fdc9f272442e4a57e47508390a714", "hex")))
                    })
            
                    it("\(OK) Derive m'/44'/283'/0'/0/2 Algorand Address Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 0, 2)
                        expect(key).toEqual(new Uint8Array(Buffer.from("af1bb9cba94b7a7569c186ecffe881b01bd2cbded23418d2fb51c7ffe675502c", "hex")))
                    })

                    it("\(OK) Derive m'/44'/283'/3'/0/0 Algorand Address Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 3, 0)
                        expect(key).toEqual(new Uint8Array(Buffer.from("d81624d97ef08248724eb2432a514efb857d2705783dcb012ede40fcaaeadcd0", "hex")))
                    })
                })

                describe("Hard Derivations", () => {
                    it("\(OK) Derive m'/44'/283'/1'/0/0 Algorand Address Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 1, 0)
                        expect(key).toEqual(new Uint8Array(Buffer.from("ccca8fd695161983b306bb0756f674d0b437d77b4bdf2b546ab2ecf2890270dd", "hex")))
                    })
        
                    it("\(OK) Derive m'/44'/283'/2'/0/1 Algorand Address Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 2, 1)
                        expect(key).toEqual(new Uint8Array(Buffer.from("d9675a7ce320937ce756617f60a73a72d30de183277bc8ecf8f923787f4f9875", "hex")))
                    })

                    describe("Ledger Addresses test Vectors", ()=> {
                        it("\(OK) Derive m'/44'/283'/0'/0/0 Algorand Address", async () => {
                            const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 0, 0)
                            const algorandAddress: string = encodeAddress(Buffer.from(key))
                            expect(algorandAddress).toEqual("SGB4KB3IU4SMLN32JXPVNV62F4QE2VW2NDHNKNRGEIVM4JNSLTMJV53P6I")
                        })
            
                        it("\(OK) Derive m'/44'/283'/1'/0/0 Algorand Address", async () => {
                            const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 1, 0)
                            const algorandAddress: string = encodeAddress(Buffer.from(key))
                            expect(algorandAddress).toEqual("ZTFI7VUVCYMYHMYGXMDVN5TU2C2DPV33JPPSWVDKWLWPFCICODO6HWLPLI")
                        })

                        it("\(OK) Derive m'/44'/283'/2'/0/0 Algorand Address", async () => {
                            const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 2, 0)
                            const algorandAddress: string = encodeAddress(Buffer.from(key))
                            expect(algorandAddress).toEqual("PSVABV5X6EGQZIHKZWVUUISWIZZKZ6RFEQ6O3IOOHPAJBSVHNNQENRMGQY")
                        })

                        it("\(OK) Derive m'/44'/283'/3'/0/0 Algorand Address", async () => {
                            const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 3, 0)
                            const algorandAddress: string = encodeAddress(Buffer.from(key))
                            expect(algorandAddress).toEqual("3ALCJWL66CBEQ4SOWJBSUUKO7OCX2JYFPA64WAJO3ZAPZKXK3TIPC4J3VQ")
                        })

                        it("\(OK) Derive m'/44'/283'/4'/0/0 Algorand Address", async () => {
                            const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 4, 0)
                            const algorandAddress: string = encodeAddress(Buffer.from(key))
                            expect(algorandAddress).toEqual("42ASNPEG4P73PI4DYJDILSX6CKOQ275N3IX6GDLVUMDRIGMBURCFNVXNEI")
                        })

                        it("\(OK) Derive m'/44'/283'/5'/0/0 Algorand Address", async () => {
                            const key: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 5, 0)
                            const algorandAddress: string = encodeAddress(Buffer.from(key))
                            expect(algorandAddress).toEqual("XGLSRHKE2RD4VNOKWIHBR4U2SEGX2H2AN75ZLJJ7VLLXWIZPJPCB4MZUFY")
                        })
                    })
                })
            })

            describe("Identities", () => {
                describe("Soft Derivations", () => {
                    it("\(OK) Derive m'/44'/0'/0'/0/0 Identity Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Identity, 0, 0)
                        expect(key).toEqual(new Uint8Array(Buffer.from("19dc07a51e5be0ce009d751c6d5fbd3a14d3dc4d0fc9c3c6f13e3581a08e4322", "hex")))
                    })
            
                    it("\(OK) Derive m'/44'/0'/0'/0/1 Identity Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Identity, 0, 1)
                        expect(key).toEqual(new Uint8Array(Buffer.from("270779072e644fd8575925a49084ecbdec3907ec876ebc9081d11278dea24165", "hex")))
                    })
            
                    it("\(OK) Derive m'/44'/0'/0'/0/2 Identity Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Identity, 0, 2)
                        expect(key).toEqual(new Uint8Array(Buffer.from("618d1493f75752fe398f3d7a361d3003d5652ae3446804256a63bbea8e11a179", "hex")))
                    })
                })

                describe("Hard Derivations", () => {
                    it("\(OK) Derive m'/44'/0'/1'/0/0 Identity Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Identity, 1, 0)
                        expect(key).toEqual(new Uint8Array(Buffer.from("9afd78d1456265bbdea1b82e2bc1d31ba64589f53325d69f222375fcd09f02c6", "hex")))
                    })
        
                    it("\(OK) Derive m'/44'/0'/2'/0/1 Identity Key", async () => {
                        const key: Uint8Array = await cryptoService.keyGen(KeyContext.Identity, 2, 1)
                        expect(key).toEqual(new Uint8Array(Buffer.from("c3364323ba03483057b08a52087c4d6bae2e26c6dbbb3adce6aa8db627eba2b1", "hex")))
                    })
                })
            })

        describe("Signing Typed Data", () => {
            it("\(OK) Sign Arbitrary Message against Schem", async () => {
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

            it("\(FAIL) Signing attempt fails because of invalid data against Schema", async () => {    
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
        })


        it("\(OK) ECDH", async () => {
            const aliceKey: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 0, 0)
            const bobKey: Uint8Array = await cryptoService.keyGen(KeyContext.Address, 0, 1)

            const aliceSharedSecret: Uint8Array = await cryptoService.ECDH(KeyContext.Address, 0, 0, bobKey)
            const bobSharedSecret: Uint8Array = await cryptoService.ECDH(KeyContext.Address, 0, 1, aliceKey)

            expect(aliceSharedSecret).toEqual(bobSharedSecret)
        })

        it("\(OK) ECDH, Encrypt and Decrypt", async () => {
            const aliceKey: Uint8Array = await cryptoService.keyGen(KeyContext.Identity, 0, 0)
            const bobKey: Uint8Array = await cryptoService.keyGen(KeyContext.Identity, 0, 1)

            const aliceSharedSecret: Uint8Array = await cryptoService.ECDH(KeyContext.Identity, 0, 0, bobKey)
            const bobSharedSecret: Uint8Array = await cryptoService.ECDH(KeyContext.Identity, 0, 1, aliceKey)

            expect(aliceSharedSecret).toEqual(bobSharedSecret)

            const message: Uint8Array = new Uint8Array(Buffer.from("Hello World"))
            const nonce: Uint8Array = randomBytes(crypto_secretbox_NONCEBYTES)

            // encrypt
            const cipherText: Uint8Array = crypto_secretbox_easy(message, nonce, aliceSharedSecret)

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
