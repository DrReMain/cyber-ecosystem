// Client-side ability vocabulary — the contract keeps abilities as open
// words (Test.ability comment: labels and colors are client vocabulary).
// Colors ride the six ability tokens (bg-ability-* utilities); copy comes
// from paraglide so both locales stay in one source.
import { m } from "#/paraglide/messages";

export interface AbilityWord {
  cn: string;
  en: string;
  axis: string;
  gradient: string;
}

export function abilityWord(ability: string): AbilityWord {
  const table: Record<string, AbilityWord> = {
    reactionSpeed: {
      cn: m.match_ability_reaction(),
      en: "REACTION SPEED",
      axis: "RS",
      gradient: "bg-ability-reaction",
    },
    aimingAccuracy: {
      cn: m.match_ability_aiming(),
      en: "AIMING ACCURACY",
      axis: "AA",
      gradient: "bg-ability-aiming",
    },
    trackingStability: {
      cn: m.match_ability_tracking(),
      en: "TRACKING STABILITY",
      axis: "TS",
      gradient: "bg-ability-tracking",
    },
    skillReleaseAccuracy: {
      cn: m.match_ability_skill(),
      en: "SKILL ACCURACY",
      axis: "SRA",
      gradient: "bg-ability-skill",
    },
    resourceAllocation: {
      cn: m.match_ability_resource(),
      en: "RESOURCE ALLOCATION",
      axis: "RA",
      gradient: "bg-ability-resource",
    },
    decisionEfficiency: {
      cn: m.match_ability_decision(),
      en: "DECISION EFFICIENCY",
      axis: "DE",
      gradient: "bg-ability-decision",
    },
  };
  return table[ability] ?? { cn: ability, en: ability.toUpperCase(), axis: "?", gradient: "" };
}

export interface GenreMeta {
  cn: string;
  en: string;
  gradient: string;
}

export function genreMeta(genre: Genre): GenreMeta {
  return genre === "fps"
    ? {
        cn: m.match_type_fps(),
        en: "FIRST-PERSON SHOOTER",
        gradient: "bg-genre-fps",
      }
    : {
        cn: m.match_type_moba(),
        en: "MULTIPLAYER ONLINE BATTLE ARENA",
        gradient: "bg-genre-moba",
      };
}

export function genreLabel(genre: Genre): string {
  return genre === "fps" ? "FPS" : "MOBA";
}

// The genre word this UI speaks — the two seeded genres. The session's
// genre arrives as an open string from the server; faces narrow it here.
export type Genre = "fps" | "moba";

// Narrows the session's open genre word to this UI's vocabulary — an
// unknown word narrows to null and renders no genre face.
export function narrowGenre(word: string | null | undefined): Genre | null {
  return word === "moba" ? "moba" : word === "fps" ? "fps" : null;
}

// The per-genre ability sequence shown on the pick page's preview cards.
// The session's authoritative order is the server's render_order; this
// frozen copy is presentation-only and hydrates nowhere else.
export const GENRE_ORDER: Readonly<Record<Genre, readonly string[]>> = {
  fps: ["reactionSpeed", "aimingAccuracy", "trackingStability", "skillReleaseAccuracy"],
  moba: ["resourceAllocation", "decisionEfficiency", "reactionSpeed", "skillReleaseAccuracy"],
};

// The preference-majority vote the server runs at session creation,
// mirrored for the pick page's recommended badge — the session itself
// always votes server-side; this never seeds or overrides a draw.
export function defaultGenre(
  preferenceCodes: readonly string[],
  games: readonly { id?: string | null; category?: string | null }[],
): Genre | null {
  const categoryOf = new Map(games.map((g) => [g.id ?? "", g.category ?? ""]));
  let fps = 0;
  let moba = 0;
  for (const code of preferenceCodes) {
    const category = categoryOf.get(code);
    if (category === "fps") fps++;
    if (category === "moba") moba++;
  }
  if (fps === 0 && moba === 0) return null;
  return fps === moba ? null : fps > moba ? "fps" : "moba";
}
