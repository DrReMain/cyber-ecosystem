// target_recognition pure logic — the legacy semantic anchors, verbatim:
// one target on screen drawn by type lot (60% enemy / 25% friendly / 15%
// neutral), a 100 px spawn margin, taps hit within the target radius plus a
// detection slop minus an edge margin, and each type owning its value and
// lifespan. Correctness is never judged here — the click carries the type,
// the server judges. All functions are pure (rng injectable) so the
// boundaries are pinned by logic.test.ts.

export interface Point {
  readonly x: number;
  readonly y: number;
}

export type TargetType = "enemy" | "friendly" | "neutral";

export interface TypeSpec {
  readonly type: TargetType;
  /** Score value (legacy: enemy +10, friendly −15, neutral −10). */
  readonly value: number;
  /** Visual diameter in px (legacy size: 25 / 20 / 20). */
  readonly diameterPx: number;
  /** Declared lifetime in ms (legacy: 2000 / 2500 / 2200 — the 700 ms
   * cadence replaces first, the lifespan is the safety net). */
  readonly lifespanMs: number;
}

export const TYPE_SPECS: readonly TypeSpec[] = [
  { type: "enemy", value: 10, diameterPx: 25, lifespanMs: 2000 },
  { type: "friendly", value: -15, diameterPx: 20, lifespanMs: 2500 },
  { type: "neutral", value: -10, diameterPx: 20, lifespanMs: 2200 },
];

export interface TargetRecognitionParams {
  /** Spawn cadence in ms (legacy: 700). */
  readonly spawnIntervalMs: number;
  /** Spawn margin from the edges in px (legacy: 100). */
  readonly spawnPaddingPx: number;
  /** Tap slop beyond the target radius in px (legacy: 25). */
  readonly clickDetectionRangePx: number;
  /** Edge-grazing band inside the slop that does NOT count as a hit (legacy: 2). */
  readonly invalidClickMarginPx: number;
}

export const TARGET_RECOGNITION_DEFAULTS: TargetRecognitionParams = {
  spawnIntervalMs: 700,
  spawnPaddingPx: 100,
  clickDetectionRangePx: 25,
  invalidClickMarginPx: 2,
};

export interface SpawnArea {
  readonly width: number;
  readonly height: number;
}

/** Type lot over the legacy band shares (0.6 / 0.25 / 0.15). */
export function drawType(rng: () => number = Math.random): TargetType {
  const roll = rng();
  if (roll < 0.6) {
    return "enemy";
  }
  if (roll < 0.85) {
    return "friendly";
  }
  return "neutral";
}

/**
 * Center position inside the padded rect (legacy: 100 + rand·(w − 200)).
 * An area too small collapses to center.
 */
export function randomCenter(
  area: SpawnArea,
  padding: number,
  rng: () => number = Math.random,
): Point {
  const spanX = area.width - padding * 2;
  const spanY = area.height - padding * 2;
  if (spanX <= 0 || spanY <= 0) {
    return { x: area.width / 2, y: area.height / 2 };
  }
  return {
    x: padding + rng() * spanX,
    y: padding + rng() * spanY,
  };
}

/** Whether a center still sits inside the padded rect (the spawn margin). */
export function outsidePaddedArea(center: Point, area: SpawnArea, paddingPx: number): boolean {
  return (
    center.x < paddingPx ||
    center.x > area.width - paddingPx ||
    center.y < paddingPx ||
    center.y > area.height - paddingPx
  );
}

export function distance(a: Point, b: Point): number {
  return Math.sqrt((a.x - b.x) ** 2 + (a.y - b.y) ** 2);
}

/**
 * Hit test — tap distance from the target center within radius + slop −
 * edge margin (legacy: distance ≤ 25 with 2 px kept off the boundary).
 */
export function isHit(
  tapDistance: number,
  diameter: number,
  params: TargetRecognitionParams,
): boolean {
  return tapDistance <= diameter / 2 + params.clickDetectionRangePx - params.invalidClickMarginPx;
}
