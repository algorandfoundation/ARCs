import {
    crypto_core_ed25519_scalar_add,
    crypto_core_ed25519_scalar_mul,
    crypto_core_ed25519_scalar_reduce,
    crypto_hash_sha512,
    crypto_scalarmult_ed25519_base_noclamp,
    crypto_sign_verify_detached,
    ready,
    crypto_sign_ed25519_pk_to_curve25519,
    crypto_scalarmult,
    crypto_sign_detached,
    crypto_sign,
    crypto_sign_SECRETKEYBYTES,
    crypto_sign_ed25519_sk_to_pk,
    crypto_scalarmult_ed25519_base
} from 'libsodium-wrappers-sumo';
import * as msgpack from "algo-msgpack-with-bigint"
import Ajv from "ajv"
import { deriveChildNodePrivate, fromSeed } from './bip32-ed25519';
import { randomBytes } from 'crypto';


/**
 * 
 */
export enum KeyContext {
    Address = 0,
    Identity = 1,
    Cardano = 2,
    TESTVECTOR_1 = 3,
    TESTVECTOR_2 = 4,
    TESTVECTOR_3 = 5
}

export interface ChannelKeys {
    tx: Uint8Array
    rx: Uint8Array
}

export enum Encoding {
    CBOR = "cbor",
    MSGPACK = "msgpack",
    BASE64 = "base64",
    NONE = "none"
}

export interface SignMetadata {
    encoding: Encoding 
    schema: Object
}

export const harden = (num: number): number => 0x80_00_00_00 + num;

function GetBIP44PathFromContext(context: KeyContext, account:number, key_index: number): number[] {
    switch (context) {
        case KeyContext.Address:
            return [harden(44), harden(283), harden(account), 0, key_index]
        case KeyContext.Identity:
            return [harden(44), harden(0), harden(account), 0, key_index]
        default:
            throw new Error("Invalid context")
    }
}

export const ERROR_BAD_DATA: Error = new Error("Invalid Data")
export const ERROR_TAGS_FOUND: Error = new Error("Transactions tags found")

export class ContextualCryptoApi {

    // Only for testing, seed shouldn't be persisted 
    constructor(private readonly seed: Buffer) {

    }


    /**
     * Derives a child key from the root key based on BIP44 path
     * 
     * @param rootKey - root key in extended format (kL, kR, c). It should be 96 bytes long
     * @param bip44Path - BIP44 path (m / purpose' / coin_type' / account' / change / address_index). The ' indicates that the value is hardened
     * @param isPrivate  - if true, return the private key, otherwise return the public key
     * @returns - The public key of 32 bytes. If isPrivate is true, returns the private key instead.
     */
    private async deriveKey(rootKey: Uint8Array, bip44Path: number[], isPrivate: boolean = true): Promise<Uint8Array> {
        let derived: Uint8Array = deriveChildNodePrivate(Buffer.from(rootKey), bip44Path[0])
            derived = deriveChildNodePrivate(derived, bip44Path[1])
            derived = deriveChildNodePrivate(derived, bip44Path[2])
            derived = deriveChildNodePrivate(derived, bip44Path[3])

            // Public Key SOFT derivations are possible without using the private key of the parent node
            // Could be an implementation choice. 
            // Example: 
            // const nodeScalar: Uint8Array = derived.subarray(0, 32)
            // const nodePublic: Uint8Array = crypto_scalarmult_ed25519_base_noclamp(nodeScalar)
            // const nodeCC: Uint8Array = derived.subarray(64, 96)

            // // [Public][ChainCode]
            // const extPub: Uint8Array = new Uint8Array(Buffer.concat([nodePublic, nodeCC]))
            // const publicKey: Uint8Array = deriveChildNodePublic(extPub, bip44Path[4]).subarray(0, 32)

            derived = deriveChildNodePrivate(derived, bip44Path[4])

        // const scalar = derived.subarray(0, 32) // scalar == pvtKey
        return isPrivate ? derived : crypto_scalarmult_ed25519_base_noclamp(derived.subarray(0, 32))
    }

    /**
     * 
     * 
     * @param context - context of the key (i.e Address, Identity)
     * @param account - account number. This value will be hardened as part of BIP44
     * @param keyIndex - key index. This value will be a SOFT derivation as part of BIP44.
     * @returns - public key 32 bytes
     */
    async keyGen(context: KeyContext, account:number, keyIndex: number): Promise<Uint8Array> {
        await ready // libsodium

        const rootKey: Uint8Array = fromSeed(this.seed)
        const bip44Path: number[] = GetBIP44PathFromContext(context, account, keyIndex)

        return await this.deriveKey(rootKey, bip44Path, false)
    }

    /**
     * Ref: https://datatracker.ietf.org/doc/html/rfc8032#section-5.1.6
     * 
     *  Edwards-Curve Digital Signature Algorithm (EdDSA)
     * 
     * @param context - context of the key (i.e Address, Identity)
     * @param account - account number. This value will be hardened as part of BIP44
     * @param keyIndex - key index. This value will be a SOFT derivation as part of BIP44.
     * @param data - data to be signed in raw bytes
     * @param metadata - metadata object that describes how `data` was encoded and what schema to use to validate against
     * 
     * @returns - signature holding R and S, totally 64 bytes
     * */ 
    async signData(context: KeyContext, account: number, keyIndex: number, data: Uint8Array, metadata: SignMetadata): Promise<Uint8Array> {
        // validate data
        const result: boolean | Error = this.validateData(data, metadata)
        
        if (result instanceof Error) { // decoding errors
            throw result
        }

        if (!result) { // failed schema validation
            throw ERROR_BAD_DATA
        }
        
        await ready // libsodium

        const rootKey: Uint8Array = fromSeed(this.seed)
        const bip44Path: number[] = GetBIP44PathFromContext(context, account, keyIndex)
        const raw: Uint8Array = await this.deriveKey(rootKey, bip44Path, true)

        const scalar = raw.slice(0, 32);
        const c = raw.slice(32, 64);

        // \(1): pubKey = scalar * G (base point, no clamp)
        const publicKey = crypto_scalarmult_ed25519_base_noclamp(scalar);

        // \(2): h = hash(c + msg) mod q
        const hash: bigint = Buffer.from(crypto_hash_sha512(Buffer.concat([c, data]))).readBigInt64LE()
        
        // \(3):  r = hash(hash(privKey) + msg) mod q 
        const q: bigint = BigInt(2n ** 252n + 27742317777372353535851937790883648493n);
        const rBigInt = hash % q
        const rBString = rBigInt.toString(16) // convert to hex string

        // fill 32 bytes of r
        // convert to Uint8Array
        const r = new Uint8Array(32) 
        for (let i = 0; i < rBString.length; i += 2) {
            r[i / 2] = parseInt(rBString.substring(i, i + 2), 16);
        }

        // \(4):  R = r * G (base point, no clamp)
        const R = crypto_scalarmult_ed25519_base_noclamp(r)

        let h = crypto_hash_sha512(Buffer.concat([R, publicKey, data]));
        h = crypto_core_ed25519_scalar_reduce(h);

        // \(5): S = (r + h * k) mod q
        const S = crypto_core_ed25519_scalar_add(r, crypto_core_ed25519_scalar_mul(h, scalar))

        return Buffer.concat([R, S]);
    }


    /**
     * SAMPLE IMPLEMENTATION to show how to validate data with encoding and schema, using base64 as an example 
     * 
     * @param message 
     * @param metadata 
     * @returns 
     */
    private validateData(message: Uint8Array, metadata: SignMetadata): boolean | Error {

        // Check that decoded doesn't include the following prefixes: TX, MX, progData, Program
        // These prefixes are reserved for the protocol
        if (this.hasAlgorandTags(message)) {
            return ERROR_TAGS_FOUND
        }

        let decoded: Uint8Array
        switch (metadata.encoding) {
            case Encoding.BASE64:
                decoded = new Uint8Array(Buffer.from(Buffer.from(message).toString(), 'base64'))
                break
            case Encoding.MSGPACK:
                decoded = msgpack.decode<Uint8Array>(message) as Uint8Array
                break

            case Encoding.NONE:
                decoded = message
                break
            default:
                throw new Error("Invalid encoding")
        }
        
        // Check after decoding too
        // Some one might try to encode a regular transaction with the protocol reserved prefixes
        if (this.hasAlgorandTags(decoded)) {
            return ERROR_TAGS_FOUND
        }

        // validate with schema
        const ajv = new Ajv()
		const validate = ajv.compile(metadata.schema)

        const valid = validate(decoded)

        if (!valid) console.log(ajv.errors)

        return valid
    }

    /**
     * Detect if the message has Algorand protocol specific tags
     * 
     * @param message - raw bytes of the message
     * @returns - true if message has Algorand protocol specific tags, false otherwise
     */
    private hasAlgorandTags(message: Uint8Array): boolean {

        // Check that decoded doesn't include the following prefixes: TX, MX, progData, Program
        if (Buffer.from(message.subarray(0, 2)).toString("ascii") === "TX" || 
            Buffer.from(message.subarray(0, 2)).toString("ascii") === "MX" || 
            Buffer.from(message.subarray(0, 8)).toString("ascii") === "progData" || 
            Buffer.from(message.subarray(0, 7)).toString("ascii") === "Program") {
            return true
        }

        return false
    }


    /**
     * Wrapper around libsodium basica signature verification
     * 
     * Any lib or system that can verify EdDSA signatures can be used
     * 
     * @param signature - raw 64 bytes signature (R, S)
     * @param message - raw bytes of the message
     * @param publicKey - raw 32 bytes public key (x,y)
     * @returns true if signature is valid, false otherwise
     */
    async verifyWithPublicKey(signature: Uint8Array, message: Uint8Array, publicKey: Uint8Array): Promise<boolean> {
        return crypto_sign_verify_detached(signature, message, publicKey)
    }


    /**
     * Function to perform ECDH against a provided public key
     * 
     * ECDH reference link: https://en.wikipedia.org/wiki/Elliptic-curve_Diffie%E2%80%93Hellman
     * 
     * It creates a shared secret between two parties. Each party only needs to be aware of the other's public key.
     * This symmetric secret can be used to derive a symmetric key for encryption and decryption. Creating a private channel between the two parties.
     * 
     * @param context - context of the key (i.e Address, Identity)
     * @param account - account number. This value will be hardened as part of BIP44
     * @param keyIndex - key index. This value will be a SOFT derivation as part of BIP44.
     * @param otherPartyPub - raw 32 bytes public key of the other party
     * @returns - raw 32 bytes shared secret
     */
    async ECDH(context: KeyContext, account: number, keyIndex: number, otherPartyPub: Uint8Array): Promise<Uint8Array> {
        await ready

        const rootKey: Uint8Array = fromSeed(this.seed)
        
        const bip44Path: number[] = GetBIP44PathFromContext(context, account, keyIndex)
        const childKey: Uint8Array = await this.deriveKey(rootKey, bip44Path, true)

        const scalar: Uint8Array = childKey.slice(0, 32)

        return crypto_scalarmult(scalar, crypto_sign_ed25519_pk_to_curve25519(otherPartyPub))
    }
}