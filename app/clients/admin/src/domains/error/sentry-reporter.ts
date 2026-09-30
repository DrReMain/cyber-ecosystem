import type { ReporterSink } from "@cyber-ecosystem/shared-error";
import * as Sentry from "@sentry/react";
import { env } from "#/env";

let initialized = false;

export function initSentryReporter(): void {
  if (import.meta.env.SSR || !env.VITE_GLITCHTIP_DSN || initialized) {
    return;
  }
  initialized = true;
  Sentry.init({
    dsn: env.VITE_GLITCHTIP_DSN,
    environment: import.meta.env.MODE,
    tracesSampleRate: 0,
  });
}

export const sentryReporter: ReporterSink = (envelope) => {
  if (!initialized) {
    return;
  }
  const error = new Error(
    envelope.reason ?? (envelope.message !== "" ? envelope.message : envelope.kind),
    envelope.cause ? { cause: new Error(envelope.cause.message) } : undefined,
  );
  error.stack = envelope.stack || error.stack;
  Sentry.captureException(error, {
    fingerprint: [envelope.kind, envelope.reason ?? envelope.code ?? envelope.source],
    tags: {
      "error.kind": envelope.kind,
      "error.source": envelope.source,
      ...(envelope.reason ? { "error.reason": envelope.reason } : {}),
      ...(envelope.code ? { "error.code": envelope.code } : {}),
    },
    extra: {
      kind: envelope.kind,
      source: envelope.source,
      code: envelope.code,
      reason: envelope.reason,
      url: typeof location !== "undefined" ? location.href : undefined,
    },
  });
};
