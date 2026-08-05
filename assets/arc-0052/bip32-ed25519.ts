import { createHash, createHmac } from "crypto";
import { read } from "fs";
import {
    crypto_core_ed25519_add,
  crypto_scalarmult_ed25519_base_noclamp,
  ready
} from "libsodium-wrappers-sumo";
var BN = require("bn.js");
import * as util from 'util'

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
 * This function takes an array of up to 256 bits and sets the last g trailing bits to zero
 * 
 * @param array - An array of up to 256 bits
 * @param g - The number of bits to zero
 * @returns - The array with the last g bits set to zero
 */
export function trunc_256_minus_g_bits(array: Uint8Array, g: number): Uint8Array {
  if (g < 0 || g > 256) {
    throw new Error("Number of bits to zero must be between 0 and 256.");
  }

  // make a copy of array
  const truncated = new Uint8Array(array);

  let remainingBits = g;

  // Start from the last byte and move backward
  for (let i = truncated.length - 1; i >= 0 && remainingBits > 0; i--) {
      if (remainingBits >= 8) {
          // If more than 8 bits remain to be zeroed, zero the entire byte
          truncated[i] = 0;
          remainingBits -= 8;
      } else {
          // Zero out the most significant bits
          truncated[i] &= (0xFF >> remainingBits)
          break;
      }
  }

  return truncated
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
 * @param g - Defines how many bits to zero in the left 32 bytes of the child key. Standard BIP32-ed25519 derivations use 32 bits. 
 * @returns - (kL, kR, c) where kL is the left 32 bytes of the child key (the new scalar), kR is the right 32 bytes of the child key, and c is the chain code. Total 96 bytes
 */
export async function deriveChildNodePrivate(
  extendedKey: Uint8Array,
  index: number, 
  g: number = 9
): Promise<Uint8Array> {
  await ready // wait for libsodium to be ready

  const kL: Buffer = Buffer.from(extendedKey.subarray(0, 32));
  const kR: Buffer = Buffer.from(extendedKey.subarray(32, 64));
  const cc: Uint8Array = extendedKey.subarray(64, 96);

  // Steps 1 & 3: Produce Z and child chain code, in accordance with hardening branching logic
  const { z, childChainCode } = index < 0x80000000 ? derivedNonHardened(kL, cc, index) : deriveHardened(kL, kR, cc, index);

  // Step 2: compute child private key
  const zLeft = z.subarray(0, 32); // 32 bytes
  const zRight = z.subarray(32, 64);

  // ######################################
  // Standard BIP32-ed25519 derivation
  // #######################################
  // zL = kl + 8 * trunc_keep_28_bytes (z_left_hand_side) 
  // zR = zr + kr

  // ######################################
  // Chris Peikert's ammendment to BIP32-ed25519 derivation
  // #######################################
  // zL = kl + 8 * trunc_256_minus_g_bits (z_left_hand_side, g) 
  // Needs to satisfy g >= d + 6
  //
  // D = 2 ^ d , D is the maximum levels of BIP32 derivations to ensure a more secure key derivation
  

  // Picking g == 9 && d == 3
  // 256 - 9 == 247 bits (30 bytes + leftover)
  // D = 2 ^ 3 == 8 Max Levels of derivations (Although we only need 5 due to BIP44)

  // making sure 
  // g == 9 >= 3 + 6

  const zL: Uint8Array = trunc_256_minus_g_bits(zLeft, g);

  // zL = kL + 8 * truncated(z_left_hand_side)
  // Big Integers + little Endianess
  const klBigNum = new BN(kL, 16, 'le')
  const big8 = new BN(8);
  const zlBigNum = new BN(zL, 16, 'le')

  const zlBigNumMul8 = klBigNum.add(zlBigNum.mul(big8))

  // check if zlBigNumMul8 is equal or larger than 2^255
  if (zlBigNumMul8.cmp(new BN(2).pow(new BN(255))) >= 0) { 
    console.log(util.inspect(zlBigNumMul8), { colors: true, depth: null })
    throw new Error('zL * 8 is larger than 2^255, which is not safe')
  }

  const left = klBigNum.add(zlBigNum.mul(big8)).toArrayLike(Buffer, 'le', 32);

  let right = new BN(kR, 16, 'le').add(new BN(zRight, 16, 'le')).toArrayLike(Buffer, 'le').slice(0, 32);
  
  const rightBuffer = Buffer.alloc(32);
  Buffer.from(right).copy(rightBuffer, 0, 0, right.length) // padding with zeros if needed

  // return (kL, kR, c)
  return new Uint8Array(Buffer.concat([left, rightBuffer, childChainCode]))
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
 * @param g - Defines how many bits to zero in the left 32 bytes of the child key. Standard BIP32-ed25519 derivations use 32 bits. 
 * @returns - 64 bytes, being the 32 bytes of the child key (the new public key) followed by the 32 bytes of the chain code
 */
export async function deriveChildNodePublic(extendedKey: Uint8Array, index: number, g: number = 9): Promise<Uint8Array> {
    if (index > 0x80000000) throw new Error('can not derive public key with harden')

    const pk: Buffer = Buffer.from(extendedKey.subarray(0, 32))
    const cc: Buffer = Buffer.from(extendedKey.subarray(32, 64))

    const data: Buffer = Buffer.allocUnsafe(1 + 32 + 4);
    data.writeUInt32LE(index, 1 + 32);

    pk.copy(data, 1);

    // Step 1: Compute Z
    data[0] = 0x02;
    const z: Buffer = createHmac("sha512", cc).update(data).digest();
    
    // Step 2: Compute child public key
    const zL: Uint8Array = trunc_256_minus_g_bits(z.subarray(0, 32), g)

    // ######################################
    // Standard BIP32-ed25519 derivation
    // #######################################
    // zL = 8 * 28bytesOf(z_left_hand_side)
    
    // ######################################
    // Chris Peikert's ammendment to BIP32-ed25519 derivation
    // #######################################
    // zL = 8 * trunc_256_minus_g_bits (z_left_hand_side, g)

    const left = new BN(zL, 16, 'le').mul(new BN(8)).toArrayLike(Buffer, 'le', 32);
    const p: Uint8Array = crypto_scalarmult_ed25519_base_noclamp(left);

    // Step 3: Compute child chain code
    data[0] = 0x03;
    const fullChildChainCode: Buffer = createHmac("sha512", cc).update(data).digest();
    const childChainCode: Buffer = fullChildChainCode.subarray(32, 64);

    return new Uint8Array(Buffer.concat([crypto_core_ed25519_add(p, pk), childChainCode]))
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
  const fullChildChainCode: Buffer = createHmac("sha512", cc).update(data).digest();
  const childChainCode: Buffer = fullChildChainCode.subarray(32, 64);

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
  const fullChildChainCode: Buffer = createHmac("sha512", cc).update(data).digest();
  const childChainCode: Buffer = fullChildChainCode.subarray(32, 64);

  return { z, childChainCode };
}
