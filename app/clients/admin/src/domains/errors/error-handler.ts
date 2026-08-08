import { type AppError, normalize } from "./app-error";

export type ErrorSource = "render" | "window" | "rejection" | "query" | "mutation";

export type ErrorReporter = (error: AppError, source: ErrorSource) => void;
export type ErrorFeedback = (error: AppError, source: ErrorSource) => void;

export interface HandleOptions {
  report?: boolean;
  feedback?: boolean;
}

// Report-level dedup, INSIDE the handler (reporter implementations stay
// stateless). The same error object reaches multiple capture points: a failed
// query fires the cache onError once AND re-renders the route errorComponent
// several times (SSR + hydration + re-renders) — measured 5 reports for one
// failure. Dedup key is the ORIGINAL thrown object's identity: a fresh
// failure always throws a fresh object, so recurring issues still report;
// only re-statements of the same instance are silenced. Non-object throws
// (string rejections) cannot register in a WeakSet and report every time.
// Feedback is NOT deduped: every channel presents at most once by design.
const reported = new WeakSet<object>();

function defaultReporter(error: AppError, source: ErrorSource): void {
  console.log("[report]", source, error);
}

function defaultFeedback(error: AppError, source: ErrorSource): void {
  console.log("[feedback]", source, error);
}

let reporter: ErrorReporter = defaultReporter;
let feedback: ErrorFeedback = defaultFeedback;

export const errorHandler = {
  handle(error: unknown, source: ErrorSource, options?: HandleOptions): void {
    const appError = normalize(error);
    const { report = true, feedback: wantFeedback = true } = options ?? {};
    const identity = appError.original;
    if (report) {
      if (typeof identity !== "object" || identity === null) {
        reporter(appError, source);
      } else if (!reported.has(identity)) {
        reported.add(identity);
        reporter(appError, source);
      }
    }
    if (wantFeedback) feedback(appError, source);
  },
  register(next: { reporter?: ErrorReporter; feedback?: ErrorFeedback }): void {
    if (next.reporter) reporter = next.reporter;
    if (next.feedback) feedback = next.feedback;
  },
};
