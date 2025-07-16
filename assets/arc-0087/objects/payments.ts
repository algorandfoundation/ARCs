import _ from "lodash";
import {
  MAX_BOX_SIZE,
  MAX_KEY_SIZE,
  MIN_BALANCE,
  PER_BOX,
  PER_UNIT,
  PREFIX,
} from "./constants.js";
import { toPaths } from "./paths.js";
import { MaxSizeError } from "./errors.js";
import { assemble, diff } from "./state.js";

const { get } = _;

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
 * ## Basic Usage
 * ```typescript
 * // Define an object
 * const obj = {
 *   key: "value"
 * }
 * // Calculate the MBR
 * const mbr = toMBR(obj)
 * ```
 *
 * ## Cached Paths
 * Using the {@link paths} parameter
 * ```typescript
 * // Define an object
 * const obj = {
 *  key: "value"
 * }
 * // Calculate the paths
 * const paths = toPaths(obj)
 * // Calculate the MBR
 * const mbr = toMBR(obj, paths)
 * ```
 */
export function toMBR(
  obj: unknown,
  pathsCache?: (string | undefined)[],
): bigint {
  if (typeof obj === "undefined") {
    throw new TypeError("Object is required");
  }
  let paths;
  if (typeof pathsCache === "undefined") {
    paths = toPaths(obj).filter((path) => typeof path !== "undefined");
  } else {
    paths = pathsCache.filter((path) => typeof path !== "undefined");
  }
  const encoder = new TextEncoder();
  return paths.reduce((acc, path) => {
    // Key size is the length of the path plus bytes for the prefix
    const keySize = BigInt(path.length) + BigInt(PREFIX.length);
    if (keySize > MAX_KEY_SIZE) {
      throw new MaxSizeError(`Key size exceeds maximum of ${MAX_KEY_SIZE}`);
    }

    const boxSize = BigInt(encoder.encode(get(obj, path || "")).length);
    if (boxSize > MAX_BOX_SIZE) {
      throw new MaxSizeError(`Box size exceeds maximum of ${MAX_BOX_SIZE}`);
    }
    return acc + (PER_BOX + PER_UNIT * (keySize + boxSize));
  }, BigInt(MIN_BALANCE));
}
export async function getMBRDifference(a: unknown, b: unknown) {
  return toMBR(assemble(diff(a, b)));
}
