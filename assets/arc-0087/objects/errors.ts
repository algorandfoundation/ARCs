
/**
 * Represents an error that occurs when the size of an object `key` or `value` exceeds the maximum allowed size.
 */
export class MaxSizeError extends TypeError {
  constructor(message: string) {
    super(message)
    this.name = "MaxSizeError"
  }
}




export const APPLICATION_ID_REQUIRED = "Application ID is required"

export const ALGORAND_CLIENT_REQUIRED = "AlgorandClient is required"
