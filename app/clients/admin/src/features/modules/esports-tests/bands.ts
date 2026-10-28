import { m } from "#/paraglide/messages";

// Open vocabularies rendered locally — the six-ability and category word
// sets are seed data, not contract enums, so the page maps what it sees and
// falls back to the raw word for anything unmapped.
export const abilityLabel = (raw: string | null | undefined): string => {
  const labels: Record<string, string> = {
    reactionSpeed: m.esports_tests_ability_reaction(),
    aimingAccuracy: m.esports_tests_ability_aiming(),
    trackingStability: m.esports_tests_ability_tracking(),
    skillReleaseAccuracy: m.esports_tests_ability_skill(),
    resourceAllocation: m.esports_tests_ability_resource(),
    decisionEfficiency: m.esports_tests_ability_decision(),
  };
  return labels[raw ?? ""] ?? (raw || "-");
};

export const categoryLabel = (raw: string | null | undefined): string => {
  const labels: Record<string, string> = {
    fps: m.esports_tests_category_fps(),
    moba: m.esports_tests_category_moba(),
    universal: m.esports_tests_category_universal(),
  };
  return labels[raw ?? ""] ?? (raw || "-");
};
