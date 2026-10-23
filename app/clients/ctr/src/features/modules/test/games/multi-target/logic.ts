// multi_target pure logic — the legacy semantic anchors, verbatim: four
// value tiers drawn by weighted lot (high 100 ×2 / medium 50 ×3 / low 25 ×3
// / trap −50 ×2), every tier living 2500 ms, at most eight targets on
// screen with the oldest crowded out, and a spawn cadence drawn once per
// run. All functions are pure (rng injectable) so the boundaries are pinned
// by logic.test.ts.

export interface Point {
  readonly x: number;
  readonly y: number;
}

export type Tier = "high" | "medium" | "low" | "trap";

export interface TierSpec {
  readonly tier: Tier;
  /** Score value (legacy: 100 / 50 / 25 / −50). */
  readonly value: number;
  /** Visual radius in px (legacy size: 30 / 40 / 50 / 45 — width = 2×). */
  readonly radiusPx: number;
  /** Draw weight (legacy: 2 / 3 / 3 / 2). */
  readonly weight: number;
}

export const TIER_SPECS: readonly TierSpec[] = [
  { tier: "high", value: 100, radiusPx: 30, weight: 2 },
  { tier: "medium", value: 50, radiusPx: 40, weight: 3 },
  { tier: "low", value: 25, radiusPx: 50, weight: 3 },
  { tier: "trap", value: -50, radiusPx: 45, weight: 2 },
];

export interface MultiTargetParams {
  /** Spawn cadence lower bound in ms (legacy: 500, drawn once per run). */
  readonly spawnIntervalMinMs: number;
  /** Spawn cadence random span in ms (legacy: 400 → 500–900). */
  readonly spawnIntervalSpanMs: number;
  /** Every tier's lifetime in ms (legacy: 2500). */
  readonly targetLifetimeMs: number;
  /** Max simultaneous targets; the oldest is crowded out beyond it (legacy: 8). */
  readonly maxOnScreen: number;
}

export const MULTI_TARGET_DEFAULTS: MultiTargetParams = {
  spawnIntervalMinMs: 500,
  spawnIntervalSpanMs: 400,
  targetLifetimeMs: 2500,
  maxOnScreen: 8,
};

export interface SpawnArea {
  readonly width: number;
  readonly height: number;
}

/** Spawn cadence drawn once per run (legacy: 500 + Math.random() * 400). */
export function spawnInterval(params: MultiTargetParams, rng: () => number = Math.random): number {
  return params.spawnIntervalMinMs + rng() * params.spawnIntervalSpanMs;
}

/** Weighted tier draw over the specs (legacy weight lot). */
export function drawTier(
  rng: () => number = Math.random,
  specs: readonly TierSpec[] = TIER_SPECS,
): TierSpec {
  const totalWeight = specs.reduce((sum, spec) => sum + spec.weight, 0);
  const roll = rng() * totalWeight;
  let cumulative = 0;
  for (const spec of specs) {
    cumulative += spec.weight;
    if (roll < cumulative) {
      return spec;
    }
  }
  return specs[specs.length - 1] as TierSpec;
}

/**
 * Center position inside the play rect, exactly inscribed (legacy:
 * x = size + rand·(width − 2·size), size = radius). An area too small
 * collapses to center.
 */
export function randomCenter(
  area: SpawnArea,
  radius: number,
  rng: () => number = Math.random,
): Point {
  const spanX = area.width - radius * 2;
  const spanY = area.height - radius * 2;
  if (spanX <= 0 || spanY <= 0) {
    return { x: area.width / 2, y: area.height / 2 };
  }
  return {
    x: radius + rng() * spanX,
    y: radius + rng() * spanY,
  };
}

/** Whether a center still sits its radius from the edges (the spawn margin). */
export function outsideSafeArea(center: Point, area: SpawnArea, radiusPx: number): boolean {
  return (
    center.x < radiusPx ||
    center.x > area.width - radiusPx ||
    center.y < radiusPx ||
    center.y > area.height - radiusPx
  );
}
