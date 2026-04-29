import { Code, ConnectError } from "@connectrpc/connect"

export interface ApiErrorDetail {
  code: number
  reason: string
  message: string
  metadata: Record<string, unknown>
}

export class ApiError extends Error {
  readonly code: number
  readonly reason: string
  readonly detail: ApiErrorDetail

  constructor(detail: ApiErrorDetail) {
    super(detail.message)
    this.name = "ApiError"
    this.code = detail.code
    this.reason = detail.reason
    this.detail = detail
  }
}

export const grpcCodeToHttpStatus: Record<number, number> = {
  [Code.Canceled]: 499,
  [Code.Unknown]: 500,
  [Code.InvalidArgument]: 400,
  [Code.DeadlineExceeded]: 504,
  [Code.NotFound]: 404,
  [Code.AlreadyExists]: 409,
  [Code.PermissionDenied]: 403,
  [Code.ResourceExhausted]: 429,
  [Code.FailedPrecondition]: 400,
  [Code.Aborted]: 409,
  [Code.OutOfRange]: 400,
  [Code.Unimplemented]: 501,
  [Code.Internal]: 500,
  [Code.Unavailable]: 503,
  [Code.DataLoss]: 500,
  [Code.Unauthenticated]: 401,
}

export function extractReason(error: ConnectError): string {
  for (const detail of error.details) {
    if (
      "debug" in detail &&
      typeof detail.debug === "object" &&
      detail.debug !== null &&
      "reason" in detail.debug
    ) {
      return String(detail.debug.reason)
    }
  }
  return error.rawMessage
}

function isHttpApiError(
  error: unknown,
): error is Record<"code" | "reason" | "message" | "metadata", unknown> {
  return (
    typeof error === "object" &&
    error !== null &&
    "reason" in error &&
    "message" in error
  )
}

export function toApiError(error: unknown): ApiError {
  if (error instanceof ApiError) return error
  if (error instanceof ConnectError) {
    return new ApiError({
      code: grpcCodeToHttpStatus[error.code] ?? 500,
      reason: extractReason(error),
      message: error.rawMessage,
      metadata: Object.fromEntries(error.metadata.entries()),
    })
  }
  if (isHttpApiError(error)) {
    return new ApiError({
      code: (error.code as number) ?? 500,
      reason: (error.reason as string) ?? "INTERNAL_ERROR",
      message: (error.message as string) ?? String(error),
      metadata: (error.metadata as Record<string, unknown>) ?? {},
    })
  }
  return new ApiError({
    code: 500,
    reason: "INTERNAL_ERROR",
    message: error instanceof Error ? error.message : String(error),
    metadata: {},
  })
}
