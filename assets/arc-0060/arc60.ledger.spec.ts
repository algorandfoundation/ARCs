import { randomBytes, createHash } from "crypto"
import HID from "node-hid";
import { crypto_sign_verify_detached, ready } from "libsodium-wrappers-sumo"
const truestamp = require("@truestamp/canonify")
const { AlgorandEncoder } = require('@algorandfoundation/algo-models')
import { Arc60WalletApi, CAIP122, ERROR_BAD_JSON, ERROR_FAILED_DECODING, ERROR_FAILED_DOMAIN_AUTH, ERROR_INVALID_SCOPE, ERROR_INVALID_SIGNER, ERROR_MISSING_AUTHENTICATION_DATA, ERROR_MISSING_DOMAIN, FIDO2ClientData, ScopeType, StdSigData, StdSigDataResponse } from "./arc60wallet.api";
import TransportNodeHid from "@ledgerhq/hw-transport-node-hid";
import { AlgorandApp, ResponseAddress } from '@zondax/ledger-algorand'

jest.setTimeout(60000)

describe('ARC60 TEST SUITE', () => {
    let ledgerApp: AlgorandApp
    let transport: TransportNodeHid

    beforeEach(async () => {
        transport = await TransportNodeHid.open(null)
        ledgerApp = new (AlgorandApp)(transport)
    })

    afterEach(async () => {
        await transport.close()
    })

    describe('SCOPE == INVALID', () => {
        it('(FAILS) Tries to sign with invalid scope', async () => {
            const challenge: Uint8Array = new Uint8Array(randomBytes(32))
            const response: ResponseAddress = await ledgerApp.getAddressAndPubKey()
            const pubBuf = response.publicKey

            const signData: StdSigData = {
                data: Buffer.from(challenge).toString('base64'),
                signer: pubBuf,
                domain: "arc60.io",
                requestId: Buffer.from(randomBytes(32)).toString('base64'),
                authenticationData: new Uint8Array(createHash('sha256').update("arc60.io").digest())
            }

            expect(ledgerApp.signData(signData, { scope: ScopeType.UNKNOWN, encoding: 'base64' })).rejects.toThrow(ERROR_INVALID_SCOPE)
        })
    })

    describe(`AUTH sign request`, () => {

        describe('CAIP-122', () => {
            it('(OK) Signing CAIP-122 requests', async () => {

                const response: ResponseAddress = await ledgerApp.getAddressAndPubKey()
                const pubBuf = response.publicKey

                const caip122Request: CAIP122 = {
                    domain: "arc60.io",
                    chain_id: "283",
                    account_address: response.address.toString(),
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

                const msgTitle: string = `Sign this message to authenticate to ${caip122Request.domain} with account ${caip122Request.account_address}`
                const msgBodyPlaceHolders: string = `URI: ${caip122Request.uri}\n` +
                    `Chain ID: ${caip122Request.chain_id}\n` +
                    `Type: ${caip122Request.type}\n` +
                    `Nonce: ${caip122Request.nonce}\n` +
                    `Statement: ${caip122Request.statement}\n` +
                    `Expiration Time: ${caip122Request["expiration-time"]}\n` +
                    `Not Before: ${caip122Request["not-before"]}\n` +
                    `Issued At: ${caip122Request["issued-at"]}\n` +
                    `Resources: ${(caip122Request.resources ?? []).join(' , \n')}\n`
                    
                const msg: string = `${msgTitle}\n\n${msgBodyPlaceHolders}`
                console.log(msg)

                const authenticationData: Uint8Array = new Uint8Array(createHash('sha256').update(caip122Request.domain).digest())

                const signData: StdSigData = {
                    data: Buffer.from(truestamp.canonify(caip122Request) || '').toString('base64'),
                    signer: pubBuf,
                    domain: caip122Request.domain,
                    requestId: Buffer.from(randomBytes(32)).toString('base64'),
                    authenticationData: authenticationData,
                    hdPath: "m/44'/283'/0'/0/0"
                }
                
                const signResponse = await ledgerApp.signData(signData, { scope: ScopeType.AUTH, encoding: 'base64' })
                const publicKey = new Uint8Array(Buffer.from(pubBuf.toString(), 'hex'))

                expect(signResponse).toBeDefined()

                const clientDataJsonHash: Buffer = createHash('sha256').update(truestamp.canonify(caip122Request) || '').digest();
                const authenticatorDataHash: Buffer = createHash('sha256').update(authenticationData).digest();

                const payloadToSign: Buffer = Buffer.concat([clientDataJsonHash, authenticatorDataHash])

                await ready
                expect(crypto_sign_verify_detached(signResponse.signature, payloadToSign, publicKey)).toBeTruthy()
            })
        })

        describe("FIDO2 / WebAuthn", () => {
            it('(OK) Signing FIDO2 requests', async () => {
                const seed: Uint8Array = new Uint8Array(Buffer.from("F20xGg5EBNo9ykzwP4B8/HX+woSoBHnUbFcxQNOVURg=", 'base64'))

                let publicKey: Uint8Array = await Arc60WalletApi.getPublicKey(seed)
                const arc60wallet: Arc60WalletApi = new Arc60WalletApi(seed);

                const rpId: string = "webauthn.io"

                const fido2Request: FIDO2ClientData = {
                    origin: "https://webauthn.io",
                    rpId: rpId,
                    challenge: "g8OebU4sWOCGljYnKXw4WUFNDszbeWfBJJKwmrTHuvc"
                }

                const rpHash: Buffer = createHash('sha256').update(rpId).digest()

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
                    data: Buffer.from(truestamp.canonify(fido2Request) || '').toString('base64'),
                    signer: publicKey,
                    domain: "webauthn.io",
                    requestId: Buffer.from(randomBytes(32)).toString('base64'),
                    authenticationData: authData
                }

                const response: ResponseAddress = await ledgerApp.getAddressAndPubKey()
                const pubBuf = response.publicKey
                
                const signResponse = await ledgerApp.signData({ ...signData, signer: pubBuf }, { scope: ScopeType.AUTH, encoding: 'base64' })
                publicKey = new Uint8Array(Buffer.from(pubBuf.toString(), 'hex'))

                expect(signResponse).toBeDefined()

                const clientDataJsonHash: Buffer = createHash('sha256').update(truestamp.canonify(fido2Request) || '').digest();
                const authenticatorDataHash: Buffer = createHash('sha256').update(authData).digest();

                const payloadToSign: Buffer = Buffer.concat([clientDataJsonHash, authenticatorDataHash])

                await ready
                expect(crypto_sign_verify_detached(signResponse.signature, payloadToSign, publicKey)).toBeTruthy()
            })
        })

        it('(OK) Signing AUTH requests', async () => {
            const challenge: Uint8Array = new Uint8Array(randomBytes(32))
            const authenticationData: Uint8Array = new Uint8Array(createHash('sha256').update("arc60.io").digest()) 

            const response: ResponseAddress = await ledgerApp.getAddressAndPubKey()
            const pubBuf = response.publicKey

            const clientDataJson = {
                "type": "arc60.create",
                "challenge": Buffer.from(challenge).toString('base64'),
                "origin": "https://arc60.io"
            }

            const signData: StdSigData = {
                data: Buffer.from(truestamp.canonify(clientDataJson) || '').toString('base64'),
                signer: pubBuf,
                domain: "arc60.io",
                requestId: Buffer.from(randomBytes(32)).toString('base64'),
                authenticationData: authenticationData
            }
                
            const signResponse = await ledgerApp.signData(signData, { scope: ScopeType.AUTH, encoding: 'base64' })
            const publicKey = new Uint8Array(Buffer.from(pubBuf.toString(), 'hex'))

            expect(signResponse).toBeDefined()

            const clientDataJsonHash: Buffer = createHash('sha256').update(truestamp.canonify(clientDataJson) || '').digest();
            const authenticatorDataHash: Buffer = createHash('sha256').update(authenticationData).digest();

            const payloadToSign: Buffer = Buffer.concat([clientDataJsonHash, authenticatorDataHash])

            await ready
            expect(crypto_sign_verify_detached(signResponse.signature, payloadToSign, publicKey)).toBeTruthy()
        })

        it('(OK) Signing AUTH requests without requestId', async () => {
            const challenge: Uint8Array = new Uint8Array(randomBytes(32))
            const authenticationData: Uint8Array = new Uint8Array(createHash('sha256').update("arc60.io").digest()) 

            const response: ResponseAddress = await ledgerApp.getAddressAndPubKey()
            const pubBuf = response.publicKey

            const clientDataJson = {
                "type": "arc60.create",
                "challenge": Buffer.from(challenge).toString('base64'),
                "origin": "https://arc60.io"
            }


            const signData: StdSigData = {
                data: Buffer.from(truestamp.canonify(clientDataJson) || '').toString('base64'),
                signer: pubBuf,
                domain: "arc60.io",
                authenticationData: authenticationData
            }
                
            const signResponse = await ledgerApp.signData(signData, { scope: ScopeType.AUTH, encoding: 'base64' })
            const publicKey = new Uint8Array(Buffer.from(pubBuf.toString(), 'hex'))

            expect(signResponse).toBeDefined()

            const clientDataJsonHash: Buffer = createHash('sha256').update(truestamp.canonify(clientDataJson) || '').digest();
            const authenticatorDataHash: Buffer = createHash('sha256').update(authenticationData).digest();

            const payloadToSign: Buffer = Buffer.concat([clientDataJsonHash, authenticatorDataHash])

            await ready
            expect(crypto_sign_verify_detached(signResponse.signature, payloadToSign, publicKey)).toBeTruthy()
        })

        it('(FAILS) Tries to sign with bad json', async () => {
            const challenge: Uint8Array = new Uint8Array(randomBytes(32))
            const authenticationData: Uint8Array = new Uint8Array(createHash('sha256').update("arc60.io").digest()) 

            const response: ResponseAddress = await ledgerApp.getAddressAndPubKey()
            const pubBuf = response.publicKey

            const clientDataJson = "{ bad json"

            const signData: StdSigData = {
                data: Buffer.from(clientDataJson).toString('base64'),
                signer: pubBuf,
                domain: "arc60.io",
                requestId: Buffer.from(randomBytes(32)).toString('base64'),
                authenticationData: authenticationData
            }

            expect(ledgerApp.signData(signData, { scope: ScopeType.AUTH, encoding: 'base64' })).rejects.toThrow(ERROR_BAD_JSON)
        })

        it('(FAILS) Tries to sign with bad json schema', async () => {
            const challenge: Uint8Array = new Uint8Array(randomBytes(32))
            const authenticationData: Uint8Array = new Uint8Array(createHash('sha256').update("arc60.io").digest()) 

            const response: ResponseAddress = await ledgerApp.getAddressAndPubKey()
            const pubBuf = response.publicKey

            const clientDataJson = {
                "type": "arc60.create",
                "challenge": Buffer.from(challenge).toString('base64'),
                "origin": "https://arc60.io"
            }

            const signData: StdSigData = {
                data: Buffer.from(truestamp.canonify(clientDataJson) || '').toString('base64'),
                signer: pubBuf,
                domain: "<bad domain>",
                requestId: Buffer.from(randomBytes(32)).toString('base64'),
                authenticationData: authenticationData
            }
                
            expect(ledgerApp.signData(signData, { scope: ScopeType.AUTH, encoding: 'base64' })).rejects.toThrow(ERROR_FAILED_DOMAIN_AUTH)
        })

        it('(FAILS) Is missing domain', async () => {
            const challenge: Uint8Array = new Uint8Array(randomBytes(32))
            const authenticationData: Uint8Array = new Uint8Array(createHash('sha256').update("arc60.io").digest()) 

            const clientDataJson = {
                "type": "arc60.create",
                "challenge": Buffer.from(challenge).toString('base64'),
                "origin": "https://arc60.io"
            }

            const response: ResponseAddress = await ledgerApp.getAddressAndPubKey()
            const pubBuf = response.publicKey

            const signData = {
                data: Buffer.from(truestamp.canonify(clientDataJson) || '').toString('base64'),
                requestId: Buffer.from(randomBytes(32)).toString('base64'),
                authenticationData: authenticationData
            }

                
            expect(ledgerApp.signData({ ...signData as StdSigData, signer: pubBuf }, { scope: ScopeType.AUTH, encoding: 'base64' })).rejects.toThrow(ERROR_MISSING_DOMAIN)
        })

        it('(FAILS) Is missing authenticationData', async () => {
            const challenge: Uint8Array = new Uint8Array(randomBytes(32))
            const clientDataJson = {
                "type": "arc60.create",
                "challenge": Buffer.from(challenge).toString('base64'),
                "origin": "https://arc60.io"
            }

            const response: ResponseAddress = await ledgerApp.getAddressAndPubKey()
            const pubBuf = response.publicKey

            const signData = {
                data: Buffer.from(truestamp.canonify(clientDataJson) || '').toString('base64'),
                domain: "arc60.io",
            }

            expect(ledgerApp.signData({ ...signData as StdSigData, signer: pubBuf }, { scope: ScopeType.AUTH, encoding: 'base64' })).rejects.toThrow(ERROR_MISSING_AUTHENTICATION_DATA)
        })
    })

    describe('Invalid or Unkown Signer', () => {
        it('(FAILS) Tries to sign with bad signer', async () => {
            const challenge: Uint8Array = new Uint8Array(randomBytes(32))

            const signData: StdSigData = {
                data: Buffer.from(challenge).toString('base64'),
                signer: new Uint8Array(31),
                domain: "arc60.io",
                requestId: Buffer.from(randomBytes(32)).toString('base64'),
                authenticationData: new Uint8Array(createHash('sha256').update("arc60.io").digest())
            }

            expect(ledgerApp.signData({ ...signData as StdSigData }, { scope: ScopeType.AUTH, encoding: 'base64' })).rejects.toThrow(ERROR_INVALID_SIGNER)
        })
    })

    describe('Unknown Encoding', () => {
        it('(FAILS) Tries to sign with unknown encoding', async () => {
            const challenge: Uint8Array = new Uint8Array(randomBytes(32))

            const response: ResponseAddress = await ledgerApp.getAddressAndPubKey()
            const pubBuf = response.publicKey

            const signData: StdSigData = {
                data: Buffer.from(challenge).toString('base64'),
                domain: "arc60.io",
                signer: pubBuf,
                requestId: Buffer.from(randomBytes(32)).toString('base64'),
                authenticationData: new Uint8Array(createHash('sha256').update("arc60.io").digest())
            }

            expect(ledgerApp.signData({ ...signData as StdSigData, signer: pubBuf }, { scope: ScopeType.AUTH, encoding: 'unknown' })).rejects.toThrow(ERROR_FAILED_DECODING)
        })
    })
})
