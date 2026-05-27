import { env } from "#/env"

type CaptureException = typeof import("@sentry/browser").captureException

let captureException: CaptureException | null = null

export function initSentry(): void {
  if (typeof window === "undefined") return

  let dsn = ""
  if (!env.VITE_GLITCHTIP_DSN) {
    return
  } else {
    dsn = env.VITE_GLITCHTIP_DSN
  }

  import("@sentry/browser").then(
    ({
      init,
      captureException: capture,
      globalHandlersIntegration,
      makeFetchTransport,
    }) => {
      const projectId = new URL(dsn).pathname
        .replace(/^\//, "")
        .replace(/\/$/, "")

      init({
        dsn,
        transport: (options) => {
          const originalUrl = new URL(options.url)
          return makeFetchTransport({
            ...options,
            url: `/glitchtip/api/${projectId}/envelope/${originalUrl.search}`,
          })
        },
        environment: import.meta.env.MODE,
        attachStacktrace: true,
        defaultIntegrations: false,
        integrations: [globalHandlersIntegration()],
        beforeSend(event) {
          if (typeof window !== "undefined") {
            event.tags = {
              ...event.tags,
              route: window.location.pathname,
            }
            if (!event.tags.errorSource) {
              event.tags.errorSource = "global"
            }
          }
          return event
        },
      })
      captureException = capture
    },
  )
}

export function getCaptureException(): CaptureException | null {
  return captureException
}
