import { registerErrorHandler } from "./error-handlers"

export function registerErrorHandlers(): void {
  registerErrorHandler("FLOW_ERROR_RATE_LIMITED", (_) => {
    if (typeof window === "undefined") return
    // biome-ignore lint/suspicious/noConsole: <demo>
    console.info("implement me")
  })
}
