/**
 * Represents an error that occurs when the size of an object `key` or `value` exceeds the maximum allowed size.
 *
 * @protected
 */
export class MaxSizeError extends TypeError {
  constructor(message: string) {
    super(message);
    this.name = "MaxSizeError";
  }
}

/**
 * A constant string that holds the error message indicating that an application ID is required.
 * This is typically used for validation when an application ID is missing or not provided.
 *
 * @protected
 */
export const APPLICATION_ID_REQUIRED = "Application ID is required";

/**
 * A constant string that specifies the error message to be displayed
 * when an Algorand client instance is required but not provided.
 *
 * This variable is typically used to enforce the presence of an
 * Algorand client instance in operations or functions that depend on it.
 *
 * @protected
 */
export const ALGORAND_CLIENT_REQUIRED = "AlgorandClient is required";
