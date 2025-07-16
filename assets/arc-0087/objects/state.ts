import { AlgorandClient } from "@algorandfoundation/algokit-utils";
import _ from "lodash";

import { PREFIX } from "./constants.js";
import { ALGORAND_CLIENT_REQUIRED, APPLICATION_ID_REQUIRED } from "./errors.js";

import type { KeyValueMap } from "./type";

/**
 * Assembles data from an Algorand application by retrieving and processing box names and their values.
 *
 * @example
 * ## Basic Usage
 * ```TypeScript
 * // Define an object
 * type Animal = {
 *   petName: string
 *   type: string
 * }
 * // Retrieve the type from the boxes, using the application ID
 * const animal = fromBoxes<Animal>(1001n)
 * ```
 *
 * @param {AlgorandClient} algorand - The Algorand client instance to interact with the blockchain.
 * @param {bigint} appId - The ID of the application to retrieve the data from.
 *
 * @throws {TypeError} If the provided `algorand` parameter is not an instance of `AlgorandClient`.
 * @throws {TypeError} If the provided `appId` is not of type `bigint`.
 */
export async function fromBoxes<T>(algorand: AlgorandClient, appId: bigint): Promise<T> {
  if (!(algorand instanceof AlgorandClient)) {
    throw new TypeError(ALGORAND_CLIENT_REQUIRED);
  }
  if (typeof appId !== "bigint") {
    throw new TypeError(APPLICATION_ID_REQUIRED);
  }

  const names = await algorand.app.getBoxNames(appId);
  const values = await algorand.app.getBoxValues(appId, names);
  const data: KeyValueMap = {};
  const decoder = new TextDecoder();

  for (let idx = 0; idx < names.length; idx++) {
    const bName = names[idx];
    const valueString = decoder.decode(values[idx]);
    _.set(data, bName.name.replace(PREFIX, ""), safeCastToNumber(valueString));
  }

  return Object.keys(data)
    .sort()
    .reduce((obj, key) => {
      obj[key] = data[key];
      return obj;
    }, {} as KeyValueMap) as T;
}

/**
 * Safely attempts to cast a string to a number.
 * If the conversion results in NaN, it returns the original string.
 *
 * @param {string} valueString - The string to be converted to a number.
 * @return {string | number} - The resulting number if conversion is successful; otherwise, the original string.
 *
 * @protected
 */
function safeCastToNumber(valueString: string): string | number {
  let value: string | number;
  value = parseInt(valueString);
  if (Number.isNaN(value)) {
    value = valueString;
  }
  return value;
}
