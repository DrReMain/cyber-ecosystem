// Client-side ability vocabulary — the contract keeps abilities as open
// words (Test.ability comment: labels and colors are client vocabulary).
// Colors ride the six ability tokens (bg-ability-* utilities); copy comes
// from paraglide so both locales stay in one source.
import { m } from "#/paraglide/messages";
import type { Genre } from "./arranger";

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
        gradient: "bg-cta-fps",
      }
    : {
        cn: m.match_type_moba(),
        en: "MULTIPLAYER ONLINE BATTLE ARENA",
        gradient: "bg-cta-moba",
      };
}

export function genreLabel(genre: Genre): string {
  return genre === "fps" ? "FPS" : "MOBA";
}
