export const normalizeError = (err: unknown) => {
  if (err instanceof Error) {
    return err as Error
  } else {
    return new Error(String(err))
  }
}
