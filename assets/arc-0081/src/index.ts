import { decodeAddress as algosdkDecodeAddress, encodeUint64 as algosdkEncodeUint64 } from "algosdk";
import base32 from "hi-base32";
import sha512 from "js-sha512";
import base64 from "base64-js";

export interface PartKeyIntegrityHashArgs {
  account: string;
  selectionKeyB64: string;
  voteKeyB64: string;
  stateProofKeyB64: string;
  voteFirstValid?: number;
  voteLastValid: number;
  keyDilution: number;
}

export function getPartKeyIntegrityHash(args: PartKeyIntegrityHashArgs): string {
  const {
    account,
    selectionKeyB64,
    voteKeyB64,
    stateProofKeyB64,
    voteFirstValid = 0, // Important edge case: field could be missing if value is zero. msgpack omitempty issue
    voteLastValid,
    keyDilution,
  } = args;

  // decode everything to raw values, performing rudimentary validation
  const rawAccount = decodeAccount(account);

  const rawSelectionKey = decodeBase64(selectionKeyB64, 32, "selectionKeyB64");
  const rawVoteKey = decodeBase64(voteKeyB64, 32, "voteKeyB64");
  const rawStateProofKey = decodeBase64(stateProofKeyB64, 64, "stateProofKeyB64");

  const rawVoteFirstValid = encodeUint64(voteFirstValid, "voteFirstValid");
  const rawVoteLastValid = encodeUint64(voteLastValid, "voteLastValid");
  const rawKeyDilution = encodeUint64(keyDilution, "keyDilution");

  // concat raw values per ARC
  const raw = new Uint8Array([
    ...rawAccount,
    ...rawSelectionKey,
    ...rawVoteKey,
    ...rawStateProofKey,
    ...rawVoteFirstValid,
    ...rawVoteLastValid,
    ...rawKeyDilution,
  ]);

  // expected: 32 + 32 + 32 + 64 + 8 + 8 + 8
  if (raw.length !== 184) {
    throw new Error(`Expected concatenated buffer to be 184 bytes, found: ${raw.length}`);
  }

  // hash raw buffer to Uint8Array
  const fullHash = decodeHex(sha512.sha512_256(raw));

  // keep the first 8 bytes of the hash
  const partialHash = fullHash.slice(0, 8);

  // encode as base32 and trim trailing padding chars (=)
  return base32.encode(partialHash).replace(/=*$/, '');
}

function decodeAccount(account: string): Uint8Array {
  // Wrapping to prefix possible error messages
  try {
    return algosdkDecodeAddress(account).publicKey;
  } catch(e) {
    throw new Error(`While decoding account: ${(e as Error).message}`);
  }
}

const decodeBase64 = (b64: string, expectedLength: number, fieldName: string): Uint8Array => {
  // base64-js lib does not support unpadded base64 strings
  // calculate padding if needed
  const padding = b64.length % 4 === 0 ? "" : new Array(4 - b64.length % 4).fill("=").join("");

  let value;
  try {
    value = base64.toByteArray(b64+padding);
  } catch(e) {
    throw new Error(`Failed to decode field ${fieldName}: ${(e as Error).message}`);
  }

  // validate length matches expectation, throw if not
  if (value.length !== expectedLength) {
    throw new Error(`Field ${fieldName} was expected to have length ${expectedLength} but found length: ${value.length}`);
  }

  return value;
}

function encodeUint64(n: number, fieldName: string): Uint8Array {
  // this is just a sanity check
  // missing proper validations like non float, non negative, etc
  if (typeof n !== "number") {
    throw new Error(`Field ${fieldName} was expected to be numeric but found: ${typeof n}`);
  }

  return algosdkEncodeUint64(n);
}

function decodeHex(hexString: string): Uint8Array {
  return Uint8Array.from(hexString.match(/.{1,2}/g)!.map((byte) => parseInt(byte, 16)));
}
