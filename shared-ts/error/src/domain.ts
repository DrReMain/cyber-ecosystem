import type {
  CopyResolver,
  ErrorPolicy,
  ErrorView,
  FeedbackSink,
  ReporterSink,
  SourceAdapter,
} from "./contracts";
import { type ErrorEnvelope, isEnvelope, makeEnvelope } from "./envelope";
import { type ErrorKind, KINDS, resolveKind } from "./kinds";

export interface DomainOptions {
  /** Source adapters, tried in order before the builtin local fallback. */
  sources?: readonly SourceAdapter[];
  /** UI notification sink; silent words never reach it. */
  feedback?: FeedbackSink;
  /** Telemetry sinks; fanned out to every sink for every word. */
  reporters?: readonly ReporterSink[];
  /** The single copy seam; defaults to raw-message passthrough. */
  copy?: CopyResolver;
}

export interface ErrorDomain {
  /** Pure classification — no sinks fire. Predicates and logic read here. */
  normalize(errorLike: unknown): ErrorEnvelope;
  /** Normalize, dedup, and dispatch to sinks under the policy ladder. */
  capture(errorLike: unknown, policy?: ErrorPolicy): ErrorEnvelope;
  /** Presentation answer — the only place copy is resolved. */
  view(errorLike: unknown): ErrorView;
}

const defaultCopy: CopyResolver = {
  title: (kind) => kind,
  message: (envelope) => envelope.message,
};

// Builtin chain tail: raw JS/DOM shapes. An AbortError name is a deliberate
// cancellation; TimeoutError (AbortSignal.timeout) is a network failure;
// everything else is unclassifiable.
function toLocalEnvelope(errorLike: unknown): ErrorEnvelope {
  const err = errorLike instanceof Error ? errorLike : undefined;
  const kind: ErrorKind =
    err?.name === "AbortError" ? "canceled" : err?.name === "TimeoutError" ? "network" : "unknown";
  return makeEnvelope(kind, err?.message ?? String(errorLike), {
    stack: err?.stack,
    cause: err?.cause,
  });
}

export function createErrorDomain(options: DomainOptions = {}): ErrorDomain {
  const sources = [...(options.sources ?? [])];
  const feedback = options.feedback;
  const reporters = [...(options.reporters ?? [])];
  const copy = options.copy ?? defaultCopy;
  const seen = new WeakSet<object>();

  const normalize = (errorLike: unknown): ErrorEnvelope => {
    try {
      if (isEnvelope(errorLike)) {
        return { ...errorLike, kind: resolveKind(errorLike.kind) };
      }
      for (const source of sources) {
        const envelope = source.recognize(errorLike);
        if (envelope !== undefined) {
          return { ...envelope, kind: resolveKind(envelope.kind) };
        }
      }
      return toLocalEnvelope(errorLike);
    } catch {
      // Hostile shapes (throwing getters, hostile toString) or a buggy
      // adapter must not leak into callers; degrade to a bare envelope.
      return makeEnvelope("unknown", "unclassifiable");
    }
  };

  const viewOf = (envelope: ErrorEnvelope): ErrorView => {
    const profile = KINDS[envelope.kind];
    return {
      kind: envelope.kind,
      silent: profile.silent,
      http: envelope.http ?? profile.http,
      title: copy.title(envelope.kind),
      message: copy.message(envelope),
      reason: envelope.reason,
    };
  };

  const capture = (errorLike: unknown, policy?: ErrorPolicy): ErrorEnvelope => {
    const envelope = normalize(errorLike);
    // Object-identity dedup: the same thrown instance reaching two sinks
    // (query cache + error boundary) dispatches once.
    if (typeof errorLike === "object" && errorLike !== null) {
      if (seen.has(errorLike)) {
        return envelope;
      }
      seen.add(errorLike);
    }
    // The word gate outranks the policy ladder: silent never feeds back.
    if (policy?.feedback !== false && !KINDS[envelope.kind].silent && feedback !== undefined) {
      try {
        feedback(viewOf(envelope));
      } catch {
        // A sink must never throw back into the pipeline.
      }
    }
    if (policy?.report !== false) {
      for (const reporter of reporters) {
        try {
          reporter(envelope);
        } catch {
          // A sink must never throw back into the pipeline.
        }
      }
    }
    return envelope;
  };

  const view = (errorLike: unknown): ErrorView => viewOf(normalize(errorLike));

  return { normalize, capture, view };
}
