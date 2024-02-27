import { createHash, createHmac } from "crypto";
import {
    crypto_core_ed25519_add,
  crypto_scalarmult_ed25519_base_noclamp
} from "libsodium-wrappers-sumo";
var BN = require("bn.js");

/**
 *
 * Reference of BIP32-Ed25519 Hierarchical Deterministic Keys over a Non-linear Keyspace (https://acrobat.adobe.com/id/urn:aaid:sc:EU:04fe29b0-ea1a-478b-a886-9bb558a5242a)
 *
 * @see section V. BIP32-Ed25519: Specification;
 * 
 * A) Root keys
 * 
 * @param seed - 256 bite seed generated from BIP39 Mnemonic
 * @returns - Extended root key (kL, kR, c) where kL is the left 32 bytes of the root key, kR is the right 32 bytes of the root key, and c is the chain code. Total 96 bytes
 */
export function fromSeed(seed: Buffer): Uint8Array {
  // k = H512(seed)
  let k: Buffer = createHash("sha512").update(seed).digest();
  let kL: Buffer = k.subarray(0, 32);
  let kR: Buffer = k.subarray(32, 64);

  // While the third highest bit of the last byte of kL is not zero
  while ((kL[31] & 0b00100000) !== 0) {
    k = createHmac("sha512", kL).update(kR).digest();
    kL = k.subarray(0, 32);
    kR = k.subarray(32, 64);
  }

  // clamp
  //Set the bits in kL as follows:
  // little Endianess
  kL[0] &= 0b11_11_10_00; // the lowest 3 bits of the first byte of kL are cleared
  kL[31] &= 0b01_11_11_11; // the highest bit of the last byte is cleared
  kL[31] |= 0b01_00_00_00; // the second highest bit of the last byte is set

  // chain root code
  // SHA256(0x01||k)
  const c: Buffer = createHash("sha256").update(Buffer.concat([new Uint8Array([0x01]), seed])).digest();
  return new Uint8Array(Buffer.concat([kL, kR, c]));
}

/**
 * @see section V. BIP32-Ed25519: Specification;
 * 
 * subsections:
 * 
 * B) Child Keys
 * and
 * C) Private Child Key Derivation
 * 
 * @param extendedKey - extended key (kL, kR, c) where kL is the left 32 bytes of the root key the scalar (pvtKey). kR is the right 32 bytes of the root key, and c is the chain code. Total 96 bytes
 * @param index - index of the child key
 * @returns - (kL, kR, c) where kL is the left 32 bytes of the child key (the new scalar), kR is the right 32 bytes of the child key, and c is the chain code. Total 96 bytes
 */
export function deriveChildNodePrivate(
  extendedKey: Uint8Array,
  index: number
): Uint8Array {
  const kl: Buffer = Buffer.from(extendedKey.subarray(0, 32));
  const kr: Buffer = Buffer.from(extendedKey.subarray(32, 64));
  const cc: Uint8Array = extendedKey.subarray(64, 96);

  const { z, childChainCode } = index < 0x80000000 ? derivedNonHardened(kl, cc, index) : deriveHardened(kl, kr, cc, index);

  const chainCode = childChainCode.subarray(32, 64);
  const zl = z.subarray(0, 32);
  const zr = z.subarray(32, 64);

  // left = kl + 8 * trunc28(zl)
  // right = zr + kr

  const left = new BN(kl, 16, "le").add(new BN(zl.subarray(0, 28), 16, "le").mul(new BN(8))).toArrayLike(Buffer, "le", 32);
  let right = new BN(kr, 16, "le").add(new BN(zr, 16, "le")).toArrayLike(Buffer, "le").slice(0, 32);


  const rightBuffer = Buffer.alloc(32);
  Buffer.from(right).copy(rightBuffer, 0, 0, right.length)


  // just padding
  // if (right.length !== 32) {
  //   right = Buffer.from(right.toString("hex") + "00", "hex");
  // }

  return Buffer.concat([left, rightBuffer, chainCode]);
}

/**
 *  * @see section V. BIP32-Ed25519: Specification;
 * 
 * subsections:
 * 
 * D) Public Child key
 * 
 * @param extendedKey - extend public key (p, c) where p is the public key and c is the chain code. Total 64 bytes
 * @param index - unharden index (i < 2^31) of the child key
 * @returns - 64 bytes, being the 32 bytes of the child key (the new public key) followed by the 32 bytes of the chain code
 */
export function deriveChildNodePublic(extendedKey: Uint8Array, index: number): Uint8Array {
    if (index > 0x80000000) throw new Error('can not derive public key with harden')

    const pk: Buffer = Buffer.from(extendedKey.subarray(0, 32))
    const cc: Buffer = Buffer.from(extendedKey.subarray(32, 64))

    const data: Buffer = Buffer.allocUnsafe(1 + 32 + 4);
    data.writeUInt32LE(index, 1 + 32);

    pk.copy(data, 1);
    data[0] = 0x02;
        
    const z: Buffer = createHmac("sha512", cc).update(data).digest();
    data[0] = 0x03;
    
    const i: Buffer = createHmac("sha512", cc).update(data).digest();

    // Section V. BIP32-Ed25519: Specification; subsection D) Public Child Key Derivation
    const chainCode: Buffer = i.subarray(32, 64);
    const zl: Buffer = z.subarray(0, 32);

    // left = 8 * 28bytesOf(zl)
    const left = new BN(zl.subarray(0, 28), 16, 'le').mul(new BN(8)).toArrayLike(Buffer, 'le', 32);

    const p: Uint8Array = crypto_scalarmult_ed25519_base_noclamp(left);
    return Buffer.concat([crypto_core_ed25519_add(p, pk), chainCode]);
}

/**
 *
 * @see section V. BIP32-Ed25519: Specification
 * 
 * @param kl - The scalar
 * @param cc - chain code
 * @param index - non-hardened ( < 2^31 ) index
 * @returns - (z, c) where z is the 64-byte child key and c is the chain code
 */
function derivedNonHardened(
  kl: Uint8Array,
  cc: Uint8Array,
  index: number
): { z: Uint8Array; childChainCode: Uint8Array } {
  const data: Buffer = Buffer.allocUnsafe(1 + 32 + 4);
  data.writeUInt32LE(index, 1 + 32);

  var pk = Buffer.from(crypto_scalarmult_ed25519_base_noclamp(kl));
  pk.copy(data, 1);

  data[0] = 0x02;
  const z: Buffer = createHmac("sha512", cc).update(data).digest();

  data[0] = 0x03;
  const childChainCode: Buffer = createHmac("sha512", cc).update(data).digest();

  return { z, childChainCode };
}

/**
 *
 * @see section V. BIP32-Ed25519: Specification
 * 
 * @param kl - The scalar (a.k.a private key)
 * @param kr - the right 32 bytes of the root key
 * @param cc - chain code
 * @param index - hardened ( >= 2^31 ) index
 * @returns - (z, c) where z is the 64-byte child key and c is the chain code
 */
function deriveHardened(
  kl: Uint8Array,
  kr: Uint8Array,
  cc: Uint8Array,
  index: number
): { z: Uint8Array; childChainCode: Uint8Array } {
  const data: Buffer = Buffer.allocUnsafe(1 + 64 + 4);
  data.writeUInt32LE(index, 1 + 64);
  Buffer.from(kl).copy(data, 1);
  Buffer.from(kr).copy(data, 1 + 32);

  data[0] = 0x00;
  const z: Buffer = createHmac("sha512", cc).update(data).digest();
  data[0] = 0x01;
  const childChainCode: Buffer = createHmac("sha512", cc).update(data).digest();

  return { z, childChainCode };
}
