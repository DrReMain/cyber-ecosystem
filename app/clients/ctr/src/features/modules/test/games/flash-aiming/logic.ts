// flash_aiming pure logic — the legacy semantic anchors, verbatim: one
// target visible at a time, diameter drawn from [30, 80] px, center kept a
// full diameter from the edges, lifetime drawn from [500, 1500] ms, center
// hit at ≤ 20 px from the target center. All functions are pure (rng
// injectable) so the boundaries are pinned by logic.test.ts.

export interface Point {
  readonly x: number;
  readonly y: number;
}

export interface FlashAimingParams {
  /** Spawn cadence in ms — the only target source (legacy: 800). */
  readonly spawnIntervalMs: number;
  /** Target diameter lower bound in px (legacy: 30). */
  readonly sizeMinPx: number;
  /** Target diameter random span in px (legacy: 50 → diameters 30–80). */
  readonly sizeSpanPx: number;
  /** Target lifetime lower bound in ms (legacy: 500). */
  readonly lifetimeMinMs: number;
  /** Target lifetime random span in ms (legacy: 1000 → lifetimes 500–1500). */
  readonly lifetimeSpanMs: number;
  /** Center-hit threshold in px from the target center (legacy: 20). */
  readonly centerHitDeviationPx: number;
}

export const FLASH_AIMING_DEFAULTS: FlashAimingParams = {
  spawnIntervalMs: 800,
  sizeMinPx: 30,
  sizeSpanPx: 50,
  lifetimeMinMs: 500,
  lifetimeSpanMs: 1000,
  centerHitDeviationPx: 20,
};

export interface SpawnArea {
  readonly width: number;
  readonly height: number;
}

/** Target diameter: min + rng·span (legacy: 30 + Math.random() * 50). */
export function randomSize(params: FlashAimingParams, rng: () => number = Math.random): number {
  return params.sizeMinPx + rng() * params.sizeSpanPx;
}

/**
 * Center position inside the play rect. Legacy law: a full diameter of
 * margin — x = size + rand·(width − 2·size) — so a target never clips an
 * edge however large it draws. An area too small collapses to center.
 */
export function randomCenter(
  area: SpawnArea,
  size: number,
  rng: () => number = Math.random,
): Point {
  const spanX = area.width - size * 2;
  const spanY = area.height - size * 2;
  if (spanX <= 0 || spanY <= 0) {
    return { x: area.width / 2, y: area.height / 2 };
  }
  return {
    x: size + rng() * spanX,
    y: size + rng() * spanY,
  };
}

/** Lifetime before an unclicked target expires (legacy: 500 + rand·1000). */
export function randomLifetime(params: FlashAimingParams, rng: () => number = Math.random): number {
  return params.lifetimeMinMs + rng() * params.lifetimeSpanMs;
}

export function distance(a: Point, b: Point): number {
  return Math.sqrt((a.x - b.x) ** 2 + (a.y - b.y) ** 2);
}

/**
 * Deviation from tap to target center, raw — legacy flash stored the
 * unrounded distance (no red-dot-style one-decimal law here).
 */
export function deviationOf(target: Point, click: Point): number {
  return distance(target, click);
}

/** Center hit at ≤ the threshold px (legacy: deviation <= 20). */
export function isCenterHit(deviation: number, params: FlashAimingParams): boolean {
  return deviation <= params.centerHitDeviationPx;
}
