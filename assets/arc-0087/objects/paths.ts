import { KeyValueMap } from "./type";

/**
 * Recursively generates the paths for all keys in an object or array, formatted as dot notation.
 *
 * @param {any} obj The object or array for which to generate key paths.
 * @param {string} parentKey The key or path of the parent object, used to build the complete key paths.
 * @return {string[]} An array of strings representing the paths of the keys in dot notation.
 */
export function toPaths(obj: unknown, parentKey?: string): string[] {
  let result: string[];
  let isObjectKey = false;
  if (Array.isArray(obj)) {
    // Map the Array to paths
    result = obj.flatMap((obj, idx) =>
      toPaths(obj, (parentKey || "") + "[" + idx++ + "]"),
    );
  }
  // TODO: better object detection
  else if (Object.prototype.toString.call(obj) === "[object Object]") {
    isObjectKey = true;
    // Map the Object Keys to paths
    result = Object.keys(obj as object).flatMap((key) =>
      toPaths((obj as KeyValueMap)[key], key).map(
        (subkey: string) => (parentKey ? parentKey + "." : "") + subkey,
      ),
    );
  }
  // If the object is not an array or object, return an empty array
  else {
    result = [];
  }

  // Return the result array with the parent key appended
  if (parentKey && !isObjectKey) return [...result, parentKey];
  else return result;
}
