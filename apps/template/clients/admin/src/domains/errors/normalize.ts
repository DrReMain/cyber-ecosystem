export interface NormalizedError {
  message: string
  cause: unknown
}

export function fromError(error: unknown): NormalizedError {
  return {
    message: error instanceof Error ? error.message : String(error),
    cause: error,
  }
}
