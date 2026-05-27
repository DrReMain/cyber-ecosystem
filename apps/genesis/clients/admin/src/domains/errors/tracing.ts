import { context, trace } from "@opentelemetry/api"
import { ZoneContextManager } from "@opentelemetry/context-zone"
import { OTLPTraceExporter } from "@opentelemetry/exporter-trace-otlp-http"
import { registerInstrumentations } from "@opentelemetry/instrumentation"
import { FetchInstrumentation } from "@opentelemetry/instrumentation-fetch"
import { LongTaskInstrumentation } from "@opentelemetry/instrumentation-long-task"
import { resourceFromAttributes } from "@opentelemetry/resources"
import { BatchSpanProcessor } from "@opentelemetry/sdk-trace-base"
import { WebTracerProvider } from "@opentelemetry/sdk-trace-web"
import { env } from "#/env"

export const initTracing = () => {
  if (!env.VITE_OTEL_URL) return
  if (typeof window !== "undefined") {
    const exporter = new OTLPTraceExporter({
      url: env.VITE_OTEL_URL,
    })

    const provider = new WebTracerProvider({
      resource: resourceFromAttributes({
        "service.name": "genesis_client_admin",
        "service.namespace": "genesis",
        "deployment.environment": import.meta.env.DEV
          ? "development"
          : "production",
      }),
      spanProcessors: [new BatchSpanProcessor(exporter)],
    })

    provider.register({
      contextManager: new ZoneContextManager(),
    })

    const ignoreUrls = [/\/otel\/.*/, /\/glitchtip\/.*/, /\/_serverFn\/.*/]

    registerInstrumentations({
      instrumentations: [
        new FetchInstrumentation({
          propagateTraceHeaderCorsUrls: /.*/,
          ignoreUrls,
        }),
        new LongTaskInstrumentation(),
      ],
    })
  }
}

export { context, trace }
