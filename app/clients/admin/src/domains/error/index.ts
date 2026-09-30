import { createErrorDomain } from "@cyber-ecosystem/shared-error";
import { connectSource } from "./connect-source";
import { copy } from "./copy";
import { initSentryReporter, sentryReporter } from "./sentry-reporter";
import { sonnerFeedback } from "./sonner-feedback";
import { viewTransitionSource } from "./view-transition-source";
import { installWindowCapture } from "./window-capture";

export { errorMutationCache, errorQueryCache, shouldRetry, shouldThrow } from "./react-query";

export const errorDomain = createErrorDomain({
  sources: [connectSource, viewTransitionSource],
  feedback: sonnerFeedback,
  reporters: [sentryReporter],
  copy,
});

export const { capture, normalize, view } = errorDomain;

export function setupErrorDomain(): void {
  initSentryReporter();
  installWindowCapture(errorDomain);
}
