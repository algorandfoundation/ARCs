import {GROUP_SIZE, MAX_ABI_SIZE} from "./constants.js";
import {toPaths} from "./paths.js";


export function toChunks(obj: any){
  return toPaths(obj)
    .reduce((accumulator, currentValue, index) => {
      const chunkIndex = Math.floor(index / GROUP_SIZE);
      if (!accumulator[chunkIndex]) {
        accumulator[chunkIndex] = []; // Start a new chunk
      }

      accumulator[chunkIndex].push(currentValue);

      return accumulator;
    }, [] as (string | undefined)[][])
}


/**
 * Splits a given string into chunks of specified size.
 *
 * @param {string} value - The string to be divided into chunks.
 * @param {number} size - The size of each chunk.
 * @return {string[]} An array containing the string chunks.
 * @TODO: refactor to bytes
 */
export function chunkValue(value: string, size: number): string[]{
  if(typeof value !== 'string'){
    throw new TypeError("Value must be a string")
  }
  if(typeof size !== 'number'){
    throw new TypeError("Chunk size must be a number")
  }
  const chunks: string[] = []
  for(let i = 0; i < value.length; i += size){
    chunks.push(value.slice(i, i + size))
  }
  return chunks
}


/**
 * Splits a given value into chunks based on the size limitation of an ABI interface.
 *
 * The key is used to calculate the chunk size based on the limitation of the ABI interface.
 * The value is then split into chunks based on the calculated chunk size.
 *
 * @param {string} key - A string representing the key whose length is used to calculate the chunk size.
 * @param {string} value - The string value to be chunked.
 * @return {string[]} An array of string chunks derived from the provided value.
 * @TODO: refactor to bytes
 */
export function chunkKeyValue(key: string, value: string): string[]{
  if(typeof key !== 'string'){
    throw new TypeError("Key must be a string")
  }
  if(typeof value !== 'string'){
    throw new TypeError("Value must be a string")
  }
  return chunkValue(value, Number(MAX_ABI_SIZE) - key.length)
}
