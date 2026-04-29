import type { ApiError } from "./api-error"

type ErrorEventType = "query" | "mutation"
type ErrorEventCallback = (error: ApiError, type: ErrorEventType) => void

const listeners = new Set<ErrorEventCallback>()

export function onApiError(cb: ErrorEventCallback): () => void {
  listeners.add(cb)
  return () => {
    listeners.delete(cb)
  }
}

export function emitApiError(error: ApiError, type: ErrorEventType): void {
  for (const cb of listeners) cb(error, type)
}
