import _ from "lodash";
import {
  MAX_BOX_SIZE,
  MAX_KEY_SIZE,
  MIN_BALANCE,
  PER_BOX,
  PER_UNIT,
  PREFIX,
} from "./constants.js";
import { MaxSizeError } from "./errors.js";
import { toPaths } from "./paths.js";

/**
 * Calculates the minimum balance requirement (MBR) for an object.
 *
 * @param {any} obj The object for which to calculate the MBR.
 * @param {(string|undefined)[]} pathsCache Precomputed paths of the object.
 *
 * @throws {TypeError} Throws a `TypeError` if the object is undefined.
 * @throws {MaxSizeError} Throws a {@link MaxSizeError} if the object key or value exceeds the maximum size.
 *
 * @example
 * ```TypeScript
 * // Define an object
 * const animal = {
 *   type: "dog",
 *   petName: "Sadie"
 * }
 * // Calculate the MBR
 * const mbr = toMBR(animal)
 * ```
 */
export function toMBR(obj: unknown, pathsCache?: string[]): bigint {
  if (typeof obj === "undefined") {
    throw new TypeError("Object is required");
  }
  const paths = typeof pathsCache === "undefined" ? toPaths(obj) : pathsCache;

  const encoder = new TextEncoder();
  return paths.reduce((acc, path) => {
    // Key size is the length of the path plus bytes for the prefix
    const keySize = BigInt(path.length) + BigInt(PREFIX.length);
    if (keySize > MAX_KEY_SIZE) {
      throw new MaxSizeError(`Key size exceeds maximum of ${MAX_KEY_SIZE}`);
    }

    const boxSize = BigInt(encoder.encode(_.get(obj, path || "")).length);
    if (boxSize > MAX_BOX_SIZE) {
      throw new MaxSizeError(`Box size exceeds maximum of ${MAX_BOX_SIZE}`);
    }
    return acc + (PER_BOX + PER_UNIT * (keySize + boxSize));
  }, BigInt(MIN_BALANCE));
}
