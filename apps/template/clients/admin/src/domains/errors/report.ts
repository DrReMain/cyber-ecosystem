import { fromError } from "./normalize"
import type { ErrorReportContext } from "./report.types"
import { getCaptureException } from "./sentry"

export type { ErrorReportContext, ErrorSource } from "./report.types"

const reported = new WeakSet<object>()

export function reportError(
  error: Error,
  context?: Partial<ErrorReportContext>,
): void {
  if (reported.has(error)) return
  reported.add(error)

  const normalized = fromError(error)
  const capture = getCaptureException()
  if (capture) {
    const err =
      normalized.cause instanceof Error
        ? normalized.cause
        : new Error(normalized.message)
    capture(err, {
      tags: {
        errorSource: context?.source ?? "manual",
      },
    })
  }
}
