// WKT Timestamp on the wire is dual-shaped (Q4 law): an ISO string off
// toJson, or the raw { seconds } object — both read, epoch out.
export interface WireAttempt {
  ability: string;
  at?: unknown;
}

export const epochOf = (value: unknown): number | undefined => {
  if (typeof value === "string") {
    const t = Date.parse(value);
    return Number.isNaN(t) ? undefined : t;
  }
  if (typeof value === "object" && value !== null && "seconds" in value) {
    const s = Number((value as { seconds: unknown }).seconds);
    return Number.isFinite(s) ? s * 1000 : undefined;
  }
  return undefined;
};

// First submit → last submit — the sitting's play span. The session's
// own started_at is the mint stamp (a match session can sit for days
// before its lineup completes), not the play clock.
export const elapsedSecondsOf = (attempts: WireAttempt[]): number | undefined => {
  const first = attempts[0];
  const last = attempts[attempts.length - 1];
  if (first === undefined || last === undefined) return undefined;
  const from = epochOf(first.at);
  const to = epochOf(last.at);
  if (from === undefined || to === undefined) return undefined;
  return Math.max(0, Math.round((to - from) / 1000));
};
