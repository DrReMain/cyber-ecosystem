import { getErrorHandler } from "#/domains/errors"
import { getLocale } from "#/paraglide/runtime"
import { client } from "#/services/openapi/client.gen"

client.interceptors.request.use((request) => {
  request.headers.set("Accept-Language", getLocale())
  return request
})

client.interceptors.error.use((error) => {
  if (
    typeof error === "object" &&
    error !== null &&
    "reason" in error &&
    "message" in error
  ) {
    const e = error as Record<string, unknown>
    const reason = e.reason as string
    const handler = getErrorHandler(reason)
    if (handler) {
      handler({
        code: (e.code as number) ?? 500,
        reason,
        message: (e.message as string) ?? String(error),
        metadata: (e.metadata as Record<string, unknown>) ?? {},
      })
    }
  }
  throw error
})

export { client }
