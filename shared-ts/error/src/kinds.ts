/** The closed vocabulary: one word per behavior profile. */
export type ErrorKind =
  | "unknown"
  | "network"
  | "server"
  | "canceled"
  | "business"
  | "auth"
  | "forbidden";

export interface KindProfile {
  /** Not a fault: never feeds back, renders neutral, never throws pages into boundaries. */
  readonly silent: boolean;
  /** HTTP semantics for rendering when no adapter override exists. */
  readonly http: number;
}

// The word table: word = behavior. Adding a word is a breaking change for
// every kind-switching consumer — only a genuinely new behavior profile
// justifies one; finer distinctions travel on `reason`/`code` instead.
export const KINDS = {
  // Cannot be classified: a crash, an uncaught throw, or a shape no source
  // adapter recognizes.
  unknown: { silent: false, http: 500 },
  // Cannot reach the service: offline, timeout, unreachable.
  network: { silent: false, http: 502 },
  // The service itself failed.
  server: { silent: false, http: 500 },
  // Deliberate cancellation — not a fault.
  canceled: { silent: true, http: 499 },
  // A rule refused the input, server-side or client-side.
  business: { silent: false, http: 422 },
  // Identity rejected: sign-in required again.
  auth: { silent: false, http: 401 },
  // Permission denied — a 403 state, not a fault.
  forbidden: { silent: false, http: 403 },
} as const satisfies Record<ErrorKind, KindProfile>;

/** Defensive for kinds arriving as raw wire strings. */
export function resolveKind(kind: string): ErrorKind {
  return kind in KINDS ? (kind as ErrorKind) : "unknown";
}
