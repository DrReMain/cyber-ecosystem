import { createQueryOptions } from "@connectrpc/connect-query";
import { getAbilityDigest } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/report-ReportService_connectquery";
import type { QueryClient } from "@tanstack/react-query";
import { connectTransport } from "#/services/connect";

// The cross-session digest's client seam. The key built here plus this
// module's invalidation reaches every cache entry the UI can create; the
// test module invalidates it beside the match session and the session
// report after a submit — a new attempt moves every aggregate this face
// carries (counts, bests, the latest lines). The wire narrowing lives
// once here for the three consumer faces (home's recent and stat cards,
// profile's overview line, the report page's history-best bar).

export function abilityDigestKey() {
  return createQueryOptions(getAbilityDigest, {}, { transport: connectTransport }).queryKey;
}

export function invalidateAbilityDigest(queryClient: QueryClient): void {
  void queryClient.invalidateQueries({ queryKey: abilityDigestKey() });
}

/** The transport flattens the WKT wrappers on the digest's lines: the
 * optional numbers arrive as plain scalars-or-undefined and the stamps as
 * RFC 3339 strings — narrowed once here, the same wire-stamp idiom the
 * resume card set. */
export interface WireDigestAbility {
  ability: string;
  bestScore: number;
  bestGrade: string | undefined;
  bestPercentile: number | undefined;
  sessionsTested: number;
}

export interface WireDigestSession {
  sessionId: string;
  genre: string;
  finished: boolean;
  doneCount: number;
  stepCount: number;
  createdAt: string | undefined;
  anchorRecordId: string | undefined;
}

export interface WireDigest {
  sessionsTotal: number;
  sessionsFps: number;
  sessionsMoba: number;
  abilities: WireDigestAbility[];
  sessions: WireDigestSession[];
}

// The single narrowing point. The generated type claims the WKT message
// shapes the JSON transport never delivers — and protojson drops an empty
// repeated field's key entirely, so an absent abilities/sessions stays
// undefined on the wire (the digest of a fresh member renders both empty;
// a bare cast would hand the UI undefined where a list was promised).
export function wireDigestOf(data: unknown): WireDigest {
  if (data === null || data === undefined) {
    return {
      sessionsTotal: 0,
      sessionsFps: 0,
      sessionsMoba: 0,
      abilities: [],
      sessions: [],
    };
  }
  const d = data as Partial<WireDigest>;
  return {
    sessionsTotal: d.sessionsTotal ?? 0,
    sessionsFps: d.sessionsFps ?? 0,
    sessionsMoba: d.sessionsMoba ?? 0,
    abilities: d.abilities ?? [],
    sessions: d.sessions ?? [],
  };
}

/** The latest scored session across genres — the recent card's subject
 * and the profile composite's. RFC 3339 stamps compare lexicographically,
 * so the newest wins without parsing; a line without an anchor never
 * qualifies (nothing scored means no verdict to read). */
export function latestScoredOf(digest: WireDigest): WireDigestSession | undefined {
  return digest.sessions
    .filter((s) => s.anchorRecordId !== undefined)
    .sort((a, b) => (b.createdAt ?? "").localeCompare(a.createdAt ?? ""))[0];
}

/** The ability→cross-session-best map — the report page's history-best
 * bar reads it. The best includes the current attempt, so an equal mark
 * on the bar means this attempt set it. */
export function bestMapOf(digest: WireDigest): Map<string, number> {
  return new Map(digest.abilities.map((a) => [a.ability, a.bestScore] as const));
}

// A wire stamp → the local date string the history and recent faces show;
// an absent or malformed stamp renders as an empty string so the row's
// desc keeps only the progress words, never a NaN date.
export function wireDateOf(stamp: string | undefined): string {
  if (typeof stamp !== "string") {
    return "";
  }
  const ms = Date.parse(stamp);
  if (Number.isNaN(ms)) {
    return "";
  }
  const d = new Date(ms);
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}
