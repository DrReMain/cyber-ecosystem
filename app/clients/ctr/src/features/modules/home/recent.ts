import { abilityWord, genreLabel, narrowGenre } from "#/features/modules/match/vocabulary";
import { type WireDigest, wireDigestOf } from "#/features/modules/report/digest";
import { tierWord } from "#/features/modules/report/vocabulary";

// The recent card's framework-free derivations — the session report's
// flattened wire shape plus the strip and chip shapes the card renders,
// kept pure so the page carries only the query wiring (resume.ts's own
// split).

// The read gates: a card renders only when its read answered — the
// helpers keep the page's wiring flat and the absence honest (a guest or
// a failed read renders no card, never a zero-filled one).
export function authedDigestOf(
  authed: boolean,
  query: { isSuccess: boolean; data: unknown },
): WireDigest | null {
  return authed && query.isSuccess ? wireDigestOf(query.data) : null;
}

export function recentWireOf(data: unknown): RecentWire | null {
  return data !== undefined ? (data as unknown as RecentWire) : null;
}

// The transport's flattened shape: scalars for the WKT wrappers, a
// value-name string for the tier enum, an RFC 3339 string for the stamp.
export interface RecentWire {
  anchorRecordId: string | undefined;
  compositePercentile: number | undefined;
  tier: string | undefined;
  genre: string | undefined;
  startedAt: string | undefined;
  abilities: Array<{ ability: string | undefined; percentile: number | undefined }>;
}

// The first four measured abilities as percentile cells — a cohort-less
// line never joins the strip (there is no number to show, never a zero).
export function recentItemsOf(
  recent: RecentWire | null,
): Array<{ abilityCn: string; pct: number }> {
  if (recent === null) {
    return [];
  }
  return (recent.abilities ?? [])
    .filter((a) => typeof a.percentile === "number")
    .slice(0, 4)
    .map((a) => ({ abilityCn: abilityWord(a.ability ?? "").cn, pct: a.percentile ?? 0 }));
}

// The composite's chip letter — UNSPECIFIED (the thin-cohort no-verdict)
// renders no chip rather than a made-up grade.
export function recentTierLetterOf(recent: RecentWire | null): string | undefined {
  if (recent === null) {
    return undefined;
  }
  const tier = tierWord(String(recent.tier ?? ""));
  return tier.kind === "none" ? undefined : tier.letter;
}

// The session's genre word this UI speaks; an unknown word renders empty.
export function recentGenreWordOf(recent: RecentWire | null): string {
  const genre = narrowGenre(recent?.genre);
  return genre !== null ? genreLabel(genre) : "";
}
