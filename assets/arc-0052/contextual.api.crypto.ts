import { createHash, createHmac } from 'crypto';
import {
    crypto_core_ed25519_scalar_add,
    crypto_core_ed25519_scalar_mul,
    crypto_core_ed25519_scalar_reduce,
    crypto_hash_sha512,
    crypto_scalarmult_ed25519_base_noclamp,
    crypto_sign_verify_detached,
    ready,
    crypto_sign_ed25519_pk_to_curve25519,
    crypto_scalarmult
} from 'libsodium-wrappers-sumo';

import Ajv, { JSONSchemaType } from "ajv"
const bip32ed25519 = require("bip32-ed25519");


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
    BASE64 = "base64"
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

export class ContextualCryptoApi {

    // Only for testing, seed shouldn't be persisted 
    constructor(private readonly seed: Buffer) {

    }

    /**
     * 
     * Reference of BIP32-Ed25519 Hierarchical Deterministic Keys over a Non-linear Keyspace (https://acrobat.adobe.com/id/urn:aaid:sc:EU:04fe29b0-ea1a-478b-a886-9bb558a5242a)
     * 
     * @param seed - 256 bite seed generated from BIP39 Mnemonic 
     * @returns - Extended root key (kL, kR, c) where kL is the left 32 bytes of the root key, kR is the right 32 bytes of the root key, and c is the chain code. Total 96 bytes
     */
    private async rootKey(seed: Buffer): Promise<Uint8Array> {
        // SLIP-0010
        // We should have been using [BIP32-Ed25519 HierarchicalDeterministicKeysoveraNon-linear Keyspace] instead. Which would mean SHA512(seed)
        // As in the [Section V].A Root keys. 
        const c: Buffer = createHmac('sha256', "ed25519 seed").update(Buffer.concat([new Uint8Array([0x1]), seed])).digest()
        let I: Buffer = createHmac('sha512', "ed25519 seed").update(seed).digest()

        // split into KL and KR.
        // (KL, KR) is the extended private key
        let kL = I.subarray(0, 32) 
        let kR = I.subarray(32, 64)

        // Specific to our Algorand app implementation (Taken from Ledger reference implementation: https://github.com/LedgerHQ/orakolo/blob/0b2d5e669ec61df9a824df9fa1a363060116b490/src/python/orakolo/HDEd25519.py#L130)
        // Seems to try to find a rootKey in which the last bits are cleared
        // Shouldn't be necessary has keys are expected to be clamped as in the next step
        while ((kL[31] & 0b00_10_00_00) != 0) {
            I = createHmac('sha512', "ed25519 seed").update(I).digest()
            kL = I.subarray(0, 32)
            kR = I.subarray(32, 64)
        }

        // clamping
        // This bit is "compliant" with [BIP32-Ed25519 Hierarchical Deterministic Keys over a Non-linear Keyspace]
        //Set the bits in kL as follows:
        // little Endianess 
        kL[0] &= 0b11_11_10_00; // the lowest 3 bits of the first byte of kL are cleared
        kL[31] &= 0b01_11_11_11; // the highest bit of the last byte is cleared
        kL[31] |= 0b01_00_00_00; // the second highest bit of the last byte is set

        return new Uint8Array(Buffer.concat([kL, kR, c]))
    }


    /**
     * Derives a child key from the root key based on BIP44 path
     * 
     * @param rootKey - root key in extended format (kL, kR, c). It should be 96 bytes long
     * @param bip44Path - BIP44 path (m / purpose' / coin_type' / account' / change / address_index). The ' indicates that the value is hardened
     * @param isPrivate  - if true, return the private key, otherwise return the public key
     * @returns 
     */
    private async deriveKey(rootKey: Uint8Array, bip44Path: number[], isPrivate: boolean = true): Promise<Uint8Array> {
        let derived = bip32ed25519.derivePrivate(Buffer.from(rootKey), bip44Path[0])
            derived = bip32ed25519.derivePrivate(derived, bip44Path[1])
            derived = bip32ed25519.derivePrivate(derived, bip44Path[2])
            derived = bip32ed25519.derivePrivate(derived, bip44Path[3])
            derived = bip32ed25519.derivePrivate(derived, bip44Path[4])

            const derivedKl = derived.subarray(0, 32)
            const xpvt = createHash('sha512').update(derivedKl).digest()

            // Keys clamped again
            // This is specific to our algorand app implementation
            // Taken from reference code: https://github.com/Zondax/ledger-algorand/blob/test/tests_zemu/tests/key_derivation.ts#L84
            // But not part of Ledger's reference implementation:https://github.com/LedgerHQ/orakolo/blob/0b2d5e669ec61df9a824df9fa1a363060116b490/src/python/orakolo/HDEd25519.py#L156
            // And not part of [BIP32-Ed25519 Hierarchical Deterministic Keys over a Non-linear Keyspace] derivation spec
            xpvt[0] &= 0b11_11_10_00
            xpvt[31] &= 0b01_11_11_11
            xpvt[31] |= 0b01_00_00_00 // This bit set is not in the BIP32-Ed25519 derivation sepc

        const scalar: Uint8Array = xpvt.subarray(0, 32)
        return isPrivate ? xpvt : crypto_scalarmult_ed25519_base_noclamp(scalar)
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

        const rootKey: Uint8Array = await this.rootKey(this.seed)
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
        if (!this.validateData(data, metadata)) {
            throw new Error("Invalid data")
        }
        
        await ready // libsodium

        const rootKey: Uint8Array = await this.rootKey(this.seed)
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
    private validateData(message: Uint8Array, metadata: SignMetadata): boolean {
        let decoded: Buffer
        switch (metadata.encoding) {
            case Encoding.BASE64:
                decoded = Buffer.from(message.toString(), 'base64')
                break
            default:
                throw new Error("Invalid encoding")
        }

        // validate with schema
        const ajv = new Ajv()
		const validate = ajv.compile(metadata.schema)
        return validate(decoded)
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

        const rootKey: Uint8Array = await this.rootKey(this.seed)
        
        const bip44Path: number[] = GetBIP44PathFromContext(context, account, keyIndex)
        const childKey: Uint8Array = await this.deriveKey(rootKey, bip44Path, true)

        const scalar: Uint8Array = childKey.slice(0, 32)

        return crypto_scalarmult(scalar, crypto_sign_ed25519_pk_to_curve25519(otherPartyPub))
    }
}