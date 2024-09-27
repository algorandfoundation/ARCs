import {
    crypto_core_ed25519_scalar_add,
    crypto_core_ed25519_scalar_mul,
    crypto_core_ed25519_scalar_reduce,
    crypto_hash_sha512,
    crypto_scalarmult_ed25519_base_noclamp,
    ready
} from 'libsodium-wrappers-sumo'

import * as crypto from 'crypto'

/**
 * Helper types
 */
export interface CAIP122 {
    domain: string; //RFC 4501 dnsauthority that is requesting the signing.
    account_address: string; //	Blockchain address performing the signing, expressed as the account_address segment of a CAIP-10 address; should NOT include CAIP-2 chain_id.
    uri: string; // RFC 3986 URI referring to the resource that is the subject of the signing i.e. the subject of the claim.
    version: string; //Current version of the message.
    statement?: string; // 	Human-readable ASCII assertion that the user will sign. It MUST NOT contain \n.
    nonce?: string; // Randomized token to prevent signature replay attacks.
    "issued-at"?: string; // 	RFC 3339 date-time that indicates the issuance time.
    "expiration-time"?: string; // RFC 3339 date-time that indicates when the signed authentication message is no longer valid.
    "not-before"?: string; // RFC 3339 date-time that indicates when the signed authentication message becomes valid.
    "request-id"?: string; // Unique identifier for the request.
    chain_id: string; // CAIP-2 chain_id of the blockchain where the account_address is valid.
    resources?: string[]; // 	List of information or references to information the user wishes to have resolved as part of the authentication by the relying party; express as RFC 3986 URIs and separated by \n.
    signature?: Uint8Array; // 	signature of the message. 
    type: string; // Type of the signature to be generated, as defined in the namespaces for this CAIP.
}


/**
type: Always set to "webauthn.create" for registration or "webauthn.authenticate" for authentication.
origin: The origin of the request, typically a URL.
challenge: A base64url-encoded challenge string.
rpId: The identifier of the Relying Party (RP).
userId: Optional user ID.
extensions: An object containing extension-defined data, if any. Extension identifiers are used as keys, and the corresponding values are the extension-defined data.

 */
export interface FIDO2ClientData {
    type?: string;
    origin: string;
    challenge: string;
    rpId?: string;
    userId?: string;
    extensions?: any
}


export interface HDWalletMetadata {
    /**
    * HD Wallet purpose value. First derivation path level. 
    * Hardened derivation is used.
    */
    purpose: number,

    /**
    * HD Wallet coin type value. Second derivation path level.
    * Hardened derivation is used.
    */
    coinType: number,

    /**
    * HD Wallet account number. Third derivation path level.
    * Hardened derivation is used.
    */
    account: number,

    /**
    * HD Wallet change value. Fourth derivation path level.
    * Soft derivation is used.
    */
    change: number,

    /**
    * HD Wallet address index value. Fifth derivation path level.
    * Soft derivation is used.
    */
    addrIdx: number,
}

// StdSigData type
export interface StdSigData {
    data: string;
    signer: Uint8Array;
    domain: string;
    authenticationData: Uint8Array;
    requestId: string;
    hdPath?: HDWalletMetadata;
    signature?: Uint8Array;
}

export interface StdSigDataResponse extends StdSigData {
    signature: Uint8Array;
}

// ScopeType type
export enum ScopeType {
    UNKNOWN = -1,
    AUTH = 1
}

// StdSignMetadata type
export interface StdSignMetadata {
    scope: ScopeType;
    encoding: string;
}

// StdSignature type 64 bytes array
export type StdSignature = Uint8Array;

export class SignDataError extends Error {
    constructor(public readonly code: number, message: string, data?: any) {
        super(message);
    }
}

// Error Codes & Messages
export const ERROR_INVALID_SCOPE: SignDataError = new SignDataError(4600, 'Invalid Scope');
export const ERROR_FAILED_DECODING: SignDataError = new SignDataError(4602, 'Failed decoding');
export const ERROR_INVALID_SIGNER: SignDataError = new SignDataError(4603, 'Invalid Signer');
export const ERROR_MISSING_DOMAIN: SignDataError = new SignDataError(4607, 'Missing Domain');
export const ERROR_MISSING_AUTHENTICATION_DATA: SignDataError = new SignDataError(4608, 'Missing Authentication Data');
export const ERROR_BAD_JSON: SignDataError = new SignDataError(4609, 'Bad JSON');
export const ERROR_FAILED_DOMAIN_AUTH: SignDataError = new SignDataError(4610, 'Failed Domain Auth');

export class Arc60WalletApi {

    /**
     * Constructor for Arc60WalletApi
     * 
     * @param k - is the seed value as part of Ed25519 key generation.
     * 
     * The following link has a visual explanation of the key gen and signing process: 
     * 
     * https://private-user-images.githubusercontent.com/1436105/316953159-aba6b82f-b558-41b9-abcb-57f682026f96.png?jwt=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJnaXRodWIuY29tIiwiYXVkIjoicmF3LmdpdGh1YnVzZXJjb250ZW50LmNvbSIsImtleSI6ImtleTUiLCJleHAiOjE3MTU4Mzk0MjMsIm5iZiI6MTcxNTgzOTEyMywicGF0aCI6Ii8xNDM2MTA1LzMxNjk1MzE1OS1hYmE2YjgyZi1iNTU4LTQxYjktYWJjYi01N2Y2ODIwMjZmOTYucG5nP1gtQW16LUFsZ29yaXRobT1BV1M0LUhNQUMtU0hBMjU2JlgtQW16LUNyZWRlbnRpYWw9QUtJQVZDT0RZTFNBNTNQUUs0WkElMkYyMDI0MDUxNiUyRnVzLWVhc3QtMSUyRnMzJTJGYXdzNF9yZXF1ZXN0JlgtQW16LURhdGU9MjAyNDA1MTZUMDU1ODQzWiZYLUFtei1FeHBpcmVzPTMwMCZYLUFtei1TaWduYXR1cmU9NjE3NGUxMDkzMzkxY2RkMzU3OTVhOWY0MzJkNzBmOGJhN2JlNDY4OTQzYzBjY2QwNzEyOTg1NzcwNDAzM2EzMSZYLUFtei1TaWduZWRIZWFkZXJzPWhvc3QmYWN0b3JfaWQ9MCZrZXlfaWQ9MCZyZXBvX2lkPTAifQ.YBMeEI0SRaxjeRRZRiZi0_58wEeCk0hgl5gIyPXluas
     * 
     */
    constructor(private readonly k: Uint8Array) {
        this.k = k;
    }

    /**
     * Arbitrary data signing function. 
     * Based on the provided scope and encoding, it decodes the data and signs it.
     * 
     * 
     * @param signingData - includes the data to be signed and the signer's public key
     * @param metadata - includes the scope and encoding of the data
     * @returns - signature of the data from the signer.
     * 
     * @throws - Error 4600 - Invalid Scope - if the scope is not supported
     * @throws - Error 4602 - Failed decoding - if the encoding is not supported
     * @throws - Error 4603 - Invalid Signer - if the signer is not the same as the signer's public key
     * @throws - Error 4607 - Missing Domain - if the domain is not provided
     * @throws - Error 4608 - Missing Authentication Data - if the authentication data is not provided
     * @throws - Error 4609 - Bad JSON - if the data is not a valid JSON
     * @throws - Error 4610 - Failed Domain Auth - if the domain is not the same as the domain in the authentication data
     */
    async signData(signingData: StdSigData, metadata: StdSignMetadata): Promise<StdSigDataResponse> {
        // decode signing data with chosen metadata.encoding
        let decodedData: Uint8Array
        let toSign: Uint8Array

        // check signer
        if (!Arc60WalletApi.isEqual(signingData.signer,(await Arc60WalletApi.getPublicKey(this.k)))) {
            throw ERROR_INVALID_SIGNER;
        }

        // decode data
        switch(metadata.encoding) {
            case 'base64':
                decodedData = Buffer.from(signingData.data, 'base64');
                break;
            default:
                throw ERROR_FAILED_DECODING;
        }
        
        // validate against schema
        switch(metadata.scope) {

            case ScopeType.AUTH:
                // Expects 2 parameters
                // clientDataJson and domain

                // validate clientDataJson is a valid JSON
                let clientDataJson: any;
                try {
                    clientDataJson = JSON.parse(decodedData.toString());
                } catch (e) {
                    throw ERROR_BAD_JSON;
                }

                const domain: string = signingData.domain ?? (() => { throw ERROR_MISSING_DOMAIN })()
                const authenticatorData: Uint8Array = signingData.authenticationData ?? (() => { throw ERROR_MISSING_AUTHENTICATION_DATA })()

                // Craft authenticatorData from domain
                // sha256
                const rp_id_hash: Buffer = crypto.createHash('sha256').update(domain).digest();

                // attestedCredentialData = aaguid (16 bytes) || credential_id_length (2 bytes) || credential_id || credential_public_key || extensions
                // authenticator_data = rp_id_hash (32 bytes) || flags (1 byte) || sign_count (4 bytes) || attested_credential_data

                // check that the first 32 bytes of authenticatorDataHash are the same as the sha256 of domain
                if(Buffer.compare(authenticatorData.slice(0, 32), rp_id_hash) !== 0) {
                    throw ERROR_FAILED_DOMAIN_AUTH;
                }

                const clientDataJsonHash: Buffer = crypto.createHash('sha256').update(JSON.stringify(clientDataJson)).digest();
                const authenticatorDataHash: Buffer = crypto.createHash('sha256').update(authenticatorData).digest();

                // Concatenate clientDataJsonHash and authenticatorData
                toSign = Buffer.concat([clientDataJsonHash, authenticatorDataHash]);
                break;

            default:
                throw ERROR_INVALID_SCOPE;
        }

        // perform signature using libsodium
        const signature: Uint8Array =  await this.rawSign(this.k, toSign);

        // craft response
        return {
            ...signingData,
            signature: signature
        }
    }

    /**
     * Raw Signing function called by signData and signTransaction
     *
     * Ref: https://datatracker.ietf.org/doc/html/rfc8032#section-5.1.6
     *
     * Edwards-Curve Digital Signature Algorithm (EdDSA)
     *
     * @param k - seed value for Ed25519 key generation
     * @param data
     * - data to be signed in raw bytes
     *
     * @returns
     * - signature holding R and S, totally 64 bytes
     */
    private async rawSign(k: Uint8Array, data: Uint8Array): Promise<Uint8Array> {
        await ready // libsodium

        // SHA512 hash of the seed value using nodejs crypto
        const raw: Uint8Array = new Uint8Array(crypto.createHash('sha512')
            .update(k)
            .digest())

        const scalar: Uint8Array = raw.slice(0, 32);
        const rH: Uint8Array = raw.slice(32, 64);

        // clamp scalar
        //Set the bits in kL as follows:
        // little Endianess
        scalar[0] &= 0b11_11_10_00; // the lowest 3 bits of the first byte of kL are cleared
        scalar[31] &= 0b01_11_11_11; // the highest bit of the last byte is cleared
        scalar[31] |= 0b01_00_00_00; // the second highest bit of the last byte is set

        // \(1): pubKey = scalar * G (base point, no clamp)
        const publicKey = crypto_scalarmult_ed25519_base_noclamp(scalar);

        // \(2): h = hash(c || msg) mod q
        const r = crypto_core_ed25519_scalar_reduce(crypto_hash_sha512(Buffer.concat([rH, data])))

        // \(4):  R = r * G (base point, no clamp)
        const R = crypto_scalarmult_ed25519_base_noclamp(r)

        // h = hash(R || pubKey || msg) mod q
        let h = crypto_core_ed25519_scalar_reduce(crypto_hash_sha512(Buffer.concat([R, publicKey, data])));

        // \(5): S = (r + h * k) mod q
        const S = crypto_core_ed25519_scalar_add(r, crypto_core_ed25519_scalar_mul(h, scalar))

        return new Uint8Array(Buffer.concat([R, S]))
    }

    private static isEqual(a: Uint8Array, b: Uint8Array): boolean {
        if (a.byteLength !== b.byteLength) {
          return false;
        }
        return a.every((value, index) => value === b[index]);
      }

    /**
     * 
     * @param k - seed
     * @returns - public key
     */
    static async getPublicKey(k: Uint8Array): Promise<Uint8Array> {
        await ready // libsodium

        // SHA512 hash of the seed value using nodejs crypto
        const raw: Uint8Array = new Uint8Array(crypto.createHash('sha512')
            .update(k)
            .digest())

        const scalar: Uint8Array = raw.slice(0, 32);

        // clamp scalar
        //Set the bits in kL as follows:
        // little Endianess
        scalar[0] &= 0b11_11_10_00; // the lowest 3 bits of the first byte of kL are cleared
        scalar[31] &= 0b01_11_11_11; // the highest bit of the last byte is cleared
        scalar[31] |= 0b01_00_00_00; // the second highest bit of the last byte is set

        // \(1): pubKey = scalar * G (base point, no clamp)
        return crypto_scalarmult_ed25519_base_noclamp(scalar);
    }
}