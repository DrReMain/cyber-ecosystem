import type { ApiErrorDetail } from "./api-error"

export type ErrorHandler = (detail: ApiErrorDetail) => void | Promise<void>

const handlers = new Map<string, ErrorHandler>()

export function registerErrorHandler(
  reason: string,
  handler: ErrorHandler,
): void {
  handlers.set(reason, handler)
}

export function getErrorHandler(reason: string): ErrorHandler | undefined {
  return handlers.get(reason)
}
