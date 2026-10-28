import { abilityWord, GENRE_ORDER, genreLabel } from "#/features/modules/match/vocabulary";
import {
  type WireDigest,
  type WireDigestSession,
  wireDateOf,
} from "#/features/modules/report/digest";
import { m } from "#/paraglide/messages";

// The overview line's framework-free derivations: the ability-overview
// rows and the per-genre history slots the profile page composes over
// the digest's flattened wire shape.

/** One ability-overview row: the six-word vocabulary's identity plus the
 * digest's cross-session best line; an absent best is the untested row
 * (undefined fields, never zeros). */
export interface AbilityOverviewRow {
  ability: string;
  abilityCn: string;
  best: number | undefined;
  grade: string | undefined;
  gradient: string;
  hint: string | null;
  pct: number | undefined;
}

// The six-word vocabulary's presentation order — the prototype's own row
// order. Tested rows rise by their cross-session percentile (a
// cohort-less best keeps its seat but never outranks a measured one),
// untested rows sink with the way in named.
export function overviewRowsOf(digest: WireDigest): AbilityOverviewRow[] {
  const order = [
    "reactionSpeed",
    "aimingAccuracy",
    "trackingStability",
    "skillReleaseAccuracy",
    "resourceAllocation",
    "decisionEfficiency",
  ];
  const byAbility = new Map(digest.abilities.map((a) => [a.ability, a]));
  const rows = order.map((ability) => {
    const line = byAbility.get(ability);
    const word = abilityWord(ability);
    return {
      ability,
      abilityCn: word.cn,
      best: line?.bestScore,
      grade: line?.bestGrade,
      gradient: word.gradient,
      hint: line === undefined ? untestedHintOf(ability) : null,
      pct: line?.bestPercentile,
    };
  });
  const tested = (row: AbilityOverviewRow) => row.best !== undefined;
  return [...rows].sort((a, b) => {
    if (!tested(a)) return 1;
    if (!tested(b)) return -1;
    return (b.pct ?? -1) - (a.pct ?? -1);
  });
}

// The untested row's way in: the genre order that covers the ability —
// a single covering genre names itself (the prototype's "test MOBA to
// unlock" shape generalized), a both-genres ability stays generic (the
// member simply has not tested yet).
function untestedHintOf(ability: string): string {
  const covers = (["fps", "moba"] as const).filter((g) => GENRE_ORDER[g].includes(ability));
  const only = covers.length === 1 ? covers[0] : undefined;
  return only !== undefined
    ? m.profile_ability_hint_one({ genre: genreLabel(only) })
    : m.profile_ability_hint_any();
}

/** One history slot per seeded genre, the digest's latest session line
 * for it — undefined keeps the "never played" row honest. */
export interface HistorySlot {
  genre: "fps" | "moba";
  line: WireDigestSession | undefined;
}

export function historySlotsOf(digest: WireDigest): HistorySlot[] {
  return (["fps", "moba"] as const).map((genre) => ({
    genre,
    line: digest.sessions.find((s) => s.genre === genre),
  }));
}

// The row's progress words: never played states it; a finished session
// carries its local date (an absent stamp keeps only the words); an
// unfinished one its honest progress counts.
export function historyDescOf(slot: HistorySlot): string {
  const line = slot.line;
  if (line === undefined) {
    return m.profile_hist_none();
  }
  if (line.finished) {
    const date = wireDateOf(line.createdAt);
    return date !== "" ? `${date} · ${m.profile_hist_done()}` : m.profile_hist_done();
  }
  return m.profile_hist_prog({ done: line.doneCount, total: line.stepCount });
}

// The four-stat strip's derivations. The composite rides the latest
// scored session's own read (the digest's lines carry no verdict by
// design); the best is the abilities' highest cross-session score. The
// improvement count stays a dash: it needs the per-session composite
// sequence, a read face the contract does not carry yet — unknown, never
// a zero.
export function compTopOf(compositePct: number | undefined): number | null {
  return typeof compositePct === "number" ? Math.round(100 - compositePct) : null;
}

export function bestScoreOf(digest: WireDigest): number | null {
  return digest.abilities.length > 0 ? Math.max(...digest.abilities.map((a) => a.bestScore)) : null;
}
