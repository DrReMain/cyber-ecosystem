// red_dot pure logic — the legacy semantic anchors, verbatim: hit threshold
// at twice the visual radius, deviation rounded to one decimal, safe-edge
// spawn padding. All functions are pure (rng injectable) so the boundaries
// are pinned by logic.test.ts.

export interface Point {
  readonly x: number;
  readonly y: number;
}

export interface RedDotParams {
  /** Visual radius in px (legacy TARGET_SIZE: 20). */
  readonly targetRadiusPx: number;
  /** Target lifetime in ms before an auto-swap (legacy: 1500). */
  readonly targetLifetimeMs: number;
  /** Safe-edge spawn padding in px (legacy: 50). */
  readonly spawnPaddingPx: number;
}

export const RED_DOT_DEFAULTS: RedDotParams = {
  targetRadiusPx: 20,
  targetLifetimeMs: 1500,
  spawnPaddingPx: 50,
};

export interface SpawnArea {
  readonly width: number;
  readonly height: number;
}

export function distance(a: Point, b: Point): number {
  return Math.sqrt((a.x - b.x) ** 2 + (a.y - b.y) ** 2);
}

/** Legacy rounding: one decimal — the single rounding law for this game. */
export function round1(value: number): number {
  return Math.round(value * 10) / 10;
}

/** Deviation from tap to target center, at one decimal (legacy law). */
export function deviationOf(target: Point, click: Point): number {
  return round1(distance(target, click));
}

/** Hit threshold = 2× the visual radius (legacy: deviation <= TARGET_SIZE*2). */
export function isHit(deviation: number, params: RedDotParams): boolean {
  return deviation <= params.targetRadiusPx * 2;
}

/** Random position inside the padded rect; an area too small collapses to center. */
export function randomPosition(
  area: SpawnArea,
  params: RedDotParams,
  rng: () => number = Math.random,
): Point {
  const spanX = area.width - params.spawnPaddingPx * 2;
  const spanY = area.height - params.spawnPaddingPx * 2;
  if (spanX <= 0 || spanY <= 0) {
    return { x: area.width / 2, y: area.height / 2 };
  }
  return {
    x: params.spawnPaddingPx + rng() * spanX,
    y: params.spawnPaddingPx + rng() * spanY,
  };
}

/** Whether a point still sits inside the padded rect (post-resize check). */
export function outsidePaddedArea(point: Point, area: SpawnArea, params: RedDotParams): boolean {
  return (
    point.x < params.spawnPaddingPx ||
    point.x > area.width - params.spawnPaddingPx ||
    point.y < params.spawnPaddingPx ||
    point.y > area.height - params.spawnPaddingPx
  );
}
