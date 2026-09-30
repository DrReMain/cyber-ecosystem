import { type ErrorKind, resolveKind } from "./kinds";

/** Single diagnostic frame — governance never recurses into diagnostics. */
export interface ErrorCause {
  readonly name?: string;
  readonly message: string;
  readonly stack?: string;
}

/**
 * The canonical form: a classified, serializable fact. Copy-free — message
 * is the raw text; presentation resolves wording from {@link ErrorView}.
 */
export interface ErrorEnvelope {
  readonly brand: "error-envelope";
  readonly kind: ErrorKind;
  /** Declared by the recognizing adapter; "local" is the builtin fallback. */
  readonly source: string;
  /** Raw message, before any copy resolution. */
  readonly message: string;
  /** Machine-readable detail — the primary channel for logic consumers. */
  readonly reason?: string;
  /** Transport code in wire form ("permission_denied"). */
  readonly code?: string;
  /** Adapter override; the kind profile provides the default. */
  readonly http?: number;
  readonly stack?: string;
  readonly cause?: ErrorCause;
}

export function isEnvelope(error: unknown): error is ErrorEnvelope {
  return (
    typeof error === "object" &&
    error !== null &&
    (error as ErrorEnvelope).brand === "error-envelope"
  );
}

export function makeEnvelope(
  kind: ErrorKind,
  message: string,
  extras: {
    source?: string;
    reason?: string;
    code?: string;
    http?: number;
    stack?: string;
    /** Raw cause: an Error or a message string becomes one diagnostic frame. */
    cause?: unknown;
  } = {},
): ErrorEnvelope {
  const rawCause = extras.cause;
  const cause =
    rawCause instanceof Error
      ? { name: rawCause.name, message: rawCause.message, stack: rawCause.stack }
      : typeof rawCause === "string" && rawCause !== ""
        ? { message: rawCause }
        : undefined;
  return {
    brand: "error-envelope",
    kind: resolveKind(kind),
    source: extras.source ?? "local",
    message,
    reason: extras.reason,
    code: extras.code,
    http: extras.http,
    stack: extras.stack,
    cause,
  };
}
