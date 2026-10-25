// moving_target — sustained tracking of one roaming target. Legacy
// semantic anchors: six movement patterns in fixed order, five seconds
// each (straight → curve → zigzag → circle → random → variable); the
// motion formulas are phase functions where phase is pattern age in
// 50 ms ticks (the legacy 20 Hz stepper's time base, kept so the
// trajectories match); the player holds the pointer on the target
// while sampling records the crosshair-to-target deviation. The
// perfect/good classification is the server's judgment — this side
// only renders the same thresholds as feedback rings and stamps each
// sample's hit as crosshair-inside-the-disc.

export type TrackPattern = "straight" | "curve" | "zigzag" | "circle" | "random" | "variable";

export const PATTERN_ORDER: readonly TrackPattern[] = [
  "straight",
  "curve",
  "zigzag",
  "circle",
  "random",
  "variable",
];

export interface Point {
  readonly x: number;
  readonly y: number;
}

export interface SpawnArea {
  readonly width: number;
  readonly height: number;
}

export interface MovingTargetParams {
  readonly patternDurationMs: number;
  readonly sampleIntervalMs: number;
  readonly targetSizePx: number;
  readonly marginPx: number;
  readonly perfectDeviationPx: number;
  readonly goodDeviationPx: number;
}

export const MOVING_TARGET_DEFAULTS: MovingTargetParams = {
  patternDurationMs: 5000,
  sampleIntervalMs: 50,
  targetSizePx: 60,
  marginPx: 100,
  perfectDeviationPx: 20,
  goodDeviationPx: 30,
};

/** One legacy stepper tick — the formulas' shared time base. */
export const PATTERN_TICK_MS = 50;

export function patternIndexAt(elapsedMs: number, patternDurationMs: number): number {
  if (patternDurationMs <= 0 || elapsedMs < 0) {
    return 0;
  }
  return Math.floor(elapsedMs / patternDurationMs) % PATTERN_ORDER.length;
}

export function patternPhase(patternAgeMs: number): number {
  return patternAgeMs / PATTERN_TICK_MS;
}

/**
 * The pattern at a rotation index, normalized in-range. The fallback is
 * dead by construction (the index is taken mod the order's length) — it
 * only satisfies the unchecked-index checker.
 */
export function patternAt(index: number): TrackPattern {
  const normalized = ((index % PATTERN_ORDER.length) + PATTERN_ORDER.length) % PATTERN_ORDER.length;
  return PATTERN_ORDER[normalized] ?? "straight";
}

/** The random walk is stateful, so it lives in the module; the rest are pure. */
export type DeterministicPattern = Exclude<TrackPattern, "random">;

/** Center-relative offset of a deterministic pattern at a phase. */
export function patternOffset(
  pattern: DeterministicPattern,
  phase: number,
  area: SpawnArea,
  marginPx: number,
): Point {
  const halfWidth = Math.max(0, area.width / 2 - marginPx);
  const halfHeight = Math.max(0, area.height / 2 - marginPx);
  switch (pattern) {
    case "straight":
      return { x: Math.sin(phase * 0.02) * Math.min(400, halfWidth), y: 0 };
    case "curve":
      return {
        x: Math.sin(phase * 0.03) * Math.min(500, halfWidth),
        y: Math.sin(phase * 0.05) * Math.min(200, halfHeight),
      };
    case "zigzag": {
      const leg = Math.floor(phase / 30);
      const maxMove = Math.min(300, Math.min(halfWidth, halfHeight));
      const subX = Math.min((phase % 30) * 10, maxMove);
      const subY = Math.min((phase % 30) * 5, maxMove / 2);
      return {
        x: (leg % 2 === 0 ? 1 : -1) * subX,
        y: (phase % 60 < 30 ? 1 : -1) * subY,
      };
    }
    case "circle": {
      const angle = phase * 0.03;
      const radius = Math.min(300, Math.min(halfWidth, halfHeight));
      return { x: Math.cos(angle) * radius, y: Math.sin(angle) * radius };
    }
    case "variable": {
      const speedFactor = 1 + Math.sin(phase * 0.01) * 0.5;
      const angle = phase * 0.02 * speedFactor;
      return {
        x: Math.cos(angle) * Math.min(400, halfWidth),
        y: Math.sin(angle) * Math.min(250, halfHeight),
      };
    }
  }
}

/**
 * One random-walk step: jitter in [−1, 1) per axis scaled to the legacy
 * 20 px jump, then clamped. Each call draws one axis.
 */
export function randomWalkStep(
  current: Point,
  area: SpawnArea,
  marginPx: number,
  jitter: () => number,
): Point {
  return clampInside(
    { x: current.x + jitter() * 20, y: current.y + jitter() * 20 },
    area,
    marginPx,
  );
}

export function clampInside(p: Point, area: SpawnArea, marginPx: number): Point {
  // A stage without a full padded band has no in-bounds strip; collapse
  // to the center (the resize-guard family's tiny-area law).
  if (area.width <= marginPx * 2 || area.height <= marginPx * 2) {
    return { x: area.width / 2, y: area.height / 2 };
  }
  return {
    x: Math.max(marginPx, Math.min(area.width - marginPx, p.x)),
    y: Math.max(marginPx, Math.min(area.height - marginPx, p.y)),
  };
}

export function distance(a: Point, b: Point): number {
  return Math.hypot(a.x - b.x, a.y - b.y);
}

export type HoldBand = "perfect" | "good" | "lost";

export function holdingBand(deviation: number, params: MovingTargetParams): HoldBand {
  if (deviation <= params.perfectDeviationPx) {
    return "perfect";
  }
  if (deviation <= params.goodDeviationPx) {
    return "good";
  }
  return "lost";
}

export function round1(value: number): number {
  return Math.round(value * 10) / 10;
}
