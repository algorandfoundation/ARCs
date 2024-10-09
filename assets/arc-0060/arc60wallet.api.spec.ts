import { randomBytes, createHash } from "crypto"
import { Arc60WalletApi, CAIP122, ERROR_BAD_JSON, ERROR_FAILED_DECODING, ERROR_FAILED_DOMAIN_AUTH, ERROR_INVALID_SCOPE, ERROR_INVALID_SIGNER, ERROR_MISSING_AUTHENTICATION_DATA, ERROR_MISSING_DOMAIN, FIDO2ClientData, ScopeType, StdSigData, StdSigDataResponse } from "./arc60wallet.api"
import { crypto_sign_verify_detached, ready } from "libsodium-wrappers-sumo"
const { AlgorandEncoder } = require('@algorandfoundation/algo-models')

jest.setTimeout(20000)

describe('ARC60 TEST SUITE', () => {

    let arc60wallet: Arc60WalletApi
    let seed: Uint8Array

    beforeEach(() => {
        seed = new Uint8Array(Buffer.from("b12e7cb8127d8fd07b03893f1aaa743bb737cff749ebac7f9af62b376f4494cc", 'hex'))
        arc60wallet = new Arc60WalletApi(seed);
    })

    // describe group for rawSign
    describe('rawSign', () => {
        it('(OK) should sign data correctly', async () => {
            const data = new Uint8Array(32).fill(2);
            const rawSign = (arc60wallet as any).rawSign.bind(arc60wallet);
            const signature = await rawSign(seed, data);

            expect(signature).toBeInstanceOf(Uint8Array);
            expect(signature.length).toBe(64); // Ed25519 signature length
        });

        it('(FAILS) should throw error for shorter incorrect length seed', async () => {
            const data = new Uint8Array(32).fill(2);
            const badSeed = new Uint8Array(31); // Incorrect shorter length seed
            const rawSign = (arc60wallet as any).rawSign.bind(arc60wallet);

            try {
                await rawSign(badSeed, data);
            } catch (error) {
                expect(error).toBeDefined();
            }
        });
        it('(FAILS) should throw error for longer incorrect length seed', async () => {
            const data = new Uint8Array(32).fill(2);
            const badSeed = new Uint8Array(33); // Incorrect longer length seed
            const rawSign = (arc60wallet as any).rawSign.bind(arc60wallet);

            try {
                await rawSign(badSeed, data);
            } catch (error) {
                expect(error).toBeDefined();
            }
        });
    });

    // describe group for getPublicKey
    describe('getPublicKey', () => {
        it('(OK) should return the correct public key', async () => {
            const publicKey = await Arc60WalletApi.getPublicKey(seed);

            expect(publicKey).toBeInstanceOf(Uint8Array);
            expect(publicKey.length).toBe(32); // Ed25519 public key length
        });

        it('(FAILS) should throw error for shorter incorrect length seed', async () => {
            const badSeed = new Uint8Array(31); // Incorrect length seed

            try {
                await Arc60WalletApi.getPublicKey(badSeed);
            } catch (error) {
                expect(error).toBeDefined();
            }
        });
        it('(FAILS) should throw error for longer incorrect length seed', async () => {
            const badSeed = new Uint8Array(33); // Incorrect length seed

            try {
                await Arc60WalletApi.getPublicKey(badSeed);
            } catch (error) {
                expect(error).toBeDefined();
            }
        });
    })

    // describe bad scope
    describe('SCOPE == INVALID', () => {
        it('\(FAILS) Tries to sign with invalid scope', async () => {
            const challenge: Uint8Array = new Uint8Array(randomBytes(32))
            const publicKey: Uint8Array = await Arc60WalletApi.getPublicKey(seed)

            const signData: StdSigData = {
                data: Buffer.from(challenge).toString('base64'),
                signer: publicKey,
                domain: "arc60.io",
                // random unique id, to help RP / Client match requests
                requestId: Buffer.from(randomBytes(32)).toString('base64'),
                authenticationData: new Uint8Array(createHash('sha256').update("arc60.io").digest())
            }

            // bad scope
            expect(arc60wallet.signData(signData, { scope: ScopeType.UNKNOWN, encoding: 'base64' })).rejects.toThrow(ERROR_INVALID_SCOPE)
        })
    })

    describe(`AUTH sign request`, () => {

        describe('CAIP-122', () => {
            it('(OK) Signing CAIP-122 requests', async () => {
                const publicKey: Uint8Array = await Arc60WalletApi.getPublicKey(seed)

                const caip122Request: CAIP122 = {
                    domain: "arc60.io",
                    chain_id: "283",
                    account_address: new AlgorandEncoder().encodeAddress(publicKey), // 
                    type: "ed25519",
                    statement: "We are requesting you to sign this message to authenticate to arc60.io",
                    uri: "https://arc60.io",
                    version: "1",
                    nonce: Buffer.from(randomBytes(32)).toString('base64'),
                    resources: ["auth", "sign"],
                    "expiration-time": "2022-12-31T23:59:59Z",
                    "not-before": "2021-12-31T23:59:59Z",
                    "issued-at": "2021-12-31T23:59:59Z",
                }

                // Disply message title according EIP-4361
                const msgTitle: string = `Sign this message to authenticate to ${caip122Request.domain} with account ${caip122Request.account_address}`

                // Display message body according EIP-4361
                const msgBodyPlaceHolders: string = `URI: ${caip122Request.uri}\n` +
                    `Chain ID: ${caip122Request.chain_id}\n` +
                    `Type: ${caip122Request.type}\n` +
                    `Nonce: ${caip122Request.nonce}\n` +
                    `Statement: ${caip122Request.statement}\n` +
                    `Expiration Time: ${caip122Request["expiration-time"]}\n` +
                    `Not Before: ${caip122Request["not-before"]}\n` +
                    `Issued At: ${caip122Request["issued-at"]}\n` +
                    `Resources: ${(caip122Request.resources ?? []).join(' , \n')}\n`
                    
                // Display message according EIP-4361
                const msg: string = `${msgTitle}\n\n${msgBodyPlaceHolders}`
                console.log(msg)

                // authenticationData
                const authenticationData: Uint8Array = new Uint8Array(createHash('sha256').update(caip122Request.domain).digest())

                const signData: StdSigData = {
                    data: Buffer.from(JSON.stringify(caip122Request)).toString('base64'),
                    signer: publicKey,
                    domain: caip122Request.domain, // should be same as origin / authenticationData
                    // random unique id, to help RP / Client match requests
                    requestId: Buffer.from(randomBytes(32)).toString('base64'),
                    authenticationData: authenticationData
                }

                const signResponse: StdSigDataResponse = await arc60wallet.signData(signData, { scope: ScopeType.AUTH, encoding: 'base64' })
                expect(signResponse).toBeDefined()

                // hash of clientDataJson
                const clientDataJsonHash: Buffer = createHash('sha256').update(JSON.stringify(caip122Request)).digest();
                const authenticatorDataHash: Buffer = createHash('sha256').update(authenticationData).digest();

                // payload to sign concatenation of clientDataJsonHash || authenticationData
                const payloadToSign: Buffer = Buffer.concat([clientDataJsonHash, authenticatorDataHash])

                // verify signature
                await ready //libsodium
                expect(crypto_sign_verify_detached(signResponse.signature, payloadToSign, publicKey)).toBeTruthy()
            })
        })

        describe("FIDO2 / WebAuthn", () => {
            it ('(OK) Signing FIDO2 requests', async () => {
                const seed: Uint8Array = new Uint8Array(Buffer.from("F20xGg5EBNo9ykzwP4B8/HX+woSoBHnUbFcxQNOVURg=", 'base64'))

                // get public key 
                const publicKey: Uint8Array = await Arc60WalletApi.getPublicKey(seed)

                // create arc60 wallet for specific seed
                const arc60wallet: Arc60WalletApi = new Arc60WalletApi(seed);

                const rpId: string = "webauthn.io"

                // FIDO2 request
                const fido2Request: FIDO2ClientData = {
                    origin: "https://webauthn.io",
                    rpId: rpId,
                    challenge: "g8OebU4sWOCGljYnKXw4WUFNDszbeWfBJJKwmrTHuvc"
                }

                const rpHash: Buffer = createHash('sha256').update(rpId).digest()

                // Set the flag for behavior
                const up = true
                const uv = true
                const be = true
                const bs = true
                var flags: number = 0
                if (up) {
                    flags = flags | 0x01
                }
                if (uv) {
                    flags = flags | 0x04
                }
                if (be) {
                    flags = flags | 0x08
                }
                if (bs) {
                    flags = flags | 0x10
                }

                const authData: Uint8Array = new Uint8Array(Buffer.concat([rpHash, Buffer.from([flags]), Buffer.from([0, 0, 0, 0])]))

                const signData: StdSigData = {
                    data: Buffer.from(JSON.stringify(fido2Request)).toString('base64'),
                    signer: publicKey,
                    domain: "webauthn.io", // should be same as origin / authenticationData
                    // random unique id, to help RP / Client match requests
                    requestId: Buffer.from(randomBytes(32)).toString('base64'),
                    authenticationData: authData
                }

                const signResponse: StdSigDataResponse = await arc60wallet.signData(signData, { scope: ScopeType.AUTH, encoding: 'base64' })
                expect(signResponse).toBeDefined()

                // To Verify

                // hash of clientDataJson
                const clientDataJsonHash: Buffer = createHash('sha256').update(JSON.stringify(fido2Request)).digest();
                const authenticatorDataHash: Buffer = createHash('sha256').update(authData).digest();

                // payload to sign concatenation of clientDataJsonHash || authData
                const payloadToSign: Buffer = Buffer.concat([clientDataJsonHash, authenticatorDataHash])

                // verify signature
                await ready //libsodium
                expect(crypto_sign_verify_detached(signResponse.signature, payloadToSign, publicKey)).toBeTruthy()
            })
        })

        it('(OK) Signing AUTH requests', async () => {
            const challenge: Uint8Array = new Uint8Array(randomBytes(32))
            const authenticationData: Uint8Array = new Uint8Array(createHash('sha256').update("arc60.io").digest()) 

            const clientDataJson = {
                "type": "arc60.create",
                "challenge": Buffer.from(challenge).toString('base64'),
                "origin": "https://arc60.io"
            }

            const publicKey: Uint8Array = await Arc60WalletApi.getPublicKey(seed)

            const signData: StdSigData = {
                data: Buffer.from(JSON.stringify(clientDataJson)).toString('base64'),
                signer: publicKey,
                domain: "arc60.io", // should be same as origin / authenticationData
                // random unique id, to help RP / Client match requests
                requestId: Buffer.from(randomBytes(32)).toString('base64'),
                authenticationData: authenticationData
            }

            const signResponse: StdSigDataResponse = await arc60wallet.signData(signData, { scope: ScopeType.AUTH, encoding: 'base64' })
            expect(signResponse).toBeDefined()

            // hash of clientDataJson
            const clientDataJsonHash: Buffer = createHash('sha256').update(JSON.stringify(clientDataJson)).digest();
            const authenticatorDataHash: Buffer = createHash('sha256').update(authenticationData).digest();

            // payload to sign concatenation of clientDataJsonHash || authenticationData
            const payloadToSign: Buffer = Buffer.concat([clientDataJsonHash, authenticatorDataHash])

            // verify signature 
            await ready //libsodium
            expect(crypto_sign_verify_detached(signResponse.signature, payloadToSign, publicKey)).toBeTruthy()
        })

        it('(FAILS) Tries to sign with bad json', async () => {
            const authenticationData: Uint8Array = new Uint8Array(createHash('sha256').update("arc60.io").digest()) 

            const clientDataJson = "{ bad json"
            const publicKey: Uint8Array = await Arc60WalletApi.getPublicKey(seed)

            const signData: StdSigData = {
                data: Buffer.from(clientDataJson).toString('base64'),
                signer: publicKey,
                domain: "arc60.io", // should be same as origin / authenticationData
                // random unique id, to help RP / Client match requests
                requestId: Buffer.from(randomBytes(32)).toString('base64'),
                authenticationData: authenticationData
            }

            expect(arc60wallet.signData(signData, { scope: ScopeType.AUTH, encoding: 'base64' })).rejects.toThrow(ERROR_BAD_JSON)
        })

        it('(FAILS) Tries to sign with bad json schema', async () => {
            const challenge: Uint8Array = new Uint8Array(randomBytes(32))
            const authenticationData: Uint8Array = new Uint8Array(createHash('sha256').update("arc60.io").digest()) 

            const clientDataJson = {
                "type": "arc60.create",
                "challenge": Buffer.from(challenge).toString('base64'),
                "origin": "https://arc60.io"
            }

            const publicKey: Uint8Array = await Arc60WalletApi.getPublicKey(seed)

            const signData: StdSigData = {
                data: Buffer.from(JSON.stringify(clientDataJson)).toString('base64'),
                signer: publicKey,
                domain: "<bad domain>", // should be same as origin / authenticationData
                // random unique id, to help RP / Client match requests
                requestId: Buffer.from(randomBytes(32)).toString('base64'),
                authenticationData: authenticationData
            }

            expect(arc60wallet.signData(signData, { scope: ScopeType.AUTH, encoding: 'base64' })).rejects.toThrow(ERROR_FAILED_DOMAIN_AUTH)
        })

        it('(FAILS) Is missing domain', async () => {
            const challenge: Uint8Array = new Uint8Array(randomBytes(32))
            const authenticationData: Uint8Array = new Uint8Array(createHash('sha256').update("arc60.io").digest()) 

            const clientDataJson = {
                "type": "arc60.create",
                "challenge": Buffer.from(challenge).toString('base64'),
                "origin": "https://arc60.io"
            }

            const publicKey: Uint8Array = await Arc60WalletApi.getPublicKey(seed)

            const signData = {
                data: Buffer.from(JSON.stringify(clientDataJson)).toString('base64'),
                signer: publicKey,
                // random unique id, to help RP / Client match requests
                requestId: Buffer.from(randomBytes(32)).toString('base64'),
                authenticationData: authenticationData
            }

            expect(arc60wallet.signData(signData as StdSigData, { scope: ScopeType.AUTH, encoding: 'base64' })).rejects.toThrow(ERROR_MISSING_DOMAIN)
        })

        it('(FAILS) Is missing authenticationData', async () => {
            const challenge: Uint8Array = new Uint8Array(randomBytes(32))
            const clientDataJson = {
                "type": "arc60.create",
                "challenge": Buffer.from(challenge).toString('base64'),
                "origin": "https://arc60.io"
            }

            const publicKey: Uint8Array = await Arc60WalletApi.getPublicKey(seed)

            const signData = {
                data: Buffer.from(JSON.stringify(clientDataJson)).toString('base64'),
                signer: publicKey,
                domain: "arc60.io",
            }

            expect(arc60wallet.signData(signData as StdSigData, { scope: ScopeType.AUTH, encoding: 'base64' })).rejects.toThrow(ERROR_MISSING_AUTHENTICATION_DATA)
        })
    })

    // bad signer
    describe('Invalid or Unkown Signer', () => {
        it('(FAILS) Tries to sign with bad signer', async () => {
            const challenge: Uint8Array = new Uint8Array(randomBytes(32))

            const signData: StdSigData = {
                data: Buffer.from(challenge).toString('base64'),
                signer: new Uint8Array(31), // Bad signer,
                domain: "arc60.io",
                // random unique id, to help RP / Client match requests
                requestId: Buffer.from(randomBytes(32)).toString('base64'),
                authenticationData: new Uint8Array(createHash('sha256').update("arc60.io").digest())
            }

            expect(arc60wallet.signData(signData, { scope: ScopeType.AUTH, encoding: 'base64' })).rejects.toThrow(ERROR_INVALID_SIGNER)
        })
    })

    // unkown encoding
    describe('Unknown Encoding', () => {
        it('(FAILS) Tries to sign with unknown encoding', async () => {
            const challenge: Uint8Array = new Uint8Array(randomBytes(32))
            const publicKey: Uint8Array = await Arc60WalletApi.getPublicKey(seed)

            const signData: StdSigData = {
                data: Buffer.from(challenge).toString('base64'),
                signer: publicKey,
                domain: "arc60.io",
                // random unique id, to help RP / Client match requests
                requestId: Buffer.from(randomBytes(32)).toString('base64'),
                authenticationData: new Uint8Array(createHash('sha256').update("arc60.io").digest())
            }

            expect(arc60wallet.signData(signData, { scope: ScopeType.AUTH, encoding: 'unknown' })).rejects.toThrow(ERROR_FAILED_DECODING)
        })
    })
})