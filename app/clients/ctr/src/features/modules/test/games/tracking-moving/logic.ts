// tracking_moving — fire at a dot that holds its ground until shot down.
// Legacy semantic anchors: one green dot spawns at a random padded spot
// and stays put; the crosshair rides a bottom-left joystick at a fixed
// top speed; a bottom-right button fires — distance is crosshair to dot
// center, a hit sits within the shared threshold (the server re-derives
// the verdict from the same key); a killed dot respawns after a short
// random delay. Shot frames carry the dot's spawn-to-kill lifetime,
// movement frames carry only the player polyline — hit presence is the
// contract's frame discriminator, so the two streams never mix shape.

export interface Point {
  readonly x: number;
  readonly y: number;
}

export interface SpawnArea {
  readonly width: number;
  readonly height: number;
}

export interface TrackingMovingParams {
  readonly hitThresholdPx: number;
  readonly dotSizePx: number;
  readonly spawnPaddingPx: number;
  readonly crosshairSpeedPxS: number;
  readonly movementSampleIntervalMs: number;
  readonly respawnDelayMaxMs: number;
}

export const TRACKING_MOVING_DEFAULTS: TrackingMovingParams = {
  hitThresholdPx: 45,
  dotSizePx: 40,
  spawnPaddingPx: 40,
  crosshairSpeedPxS: 400,
  movementSampleIntervalMs: 100,
  respawnDelayMaxMs: 1000,
};

export function distance(a: Point, b: Point): number {
  return Math.hypot(a.x - b.x, a.y - b.y);
}

export function isHit(deviation: number, params: TrackingMovingParams): boolean {
  return deviation <= params.hitThresholdPx;
}

/** Uniform draw inside the padded band; a tiny stage collapses to center. */
export function randomCenter(area: SpawnArea, paddingPx: number, rand: () => number): Point {
  if (area.width <= paddingPx * 2 || area.height <= paddingPx * 2) {
    return { x: area.width / 2, y: area.height / 2 };
  }
  return {
    x: paddingPx + rand() * (area.width - paddingPx * 2),
    y: paddingPx + rand() * (area.height - paddingPx * 2),
  };
}

export function outsidePaddedArea(center: Point, area: SpawnArea, paddingPx: number): boolean {
  return (
    center.x < paddingPx ||
    center.y < paddingPx ||
    center.x > area.width - paddingPx ||
    center.y > area.height - paddingPx
  );
}

/** Knob travel under a fifth of the base radius is finger tremor, not intent. */
export const JOYSTICK_DEAD_ZONE_RATIO = 0.2;

/**
 * Knob offset (px from the base center) → a steering vector of length
 * ≤ 1: dead-zoned at the stub, clamped at the rim, scaled by travel so
 * a rim hold is full speed and a gentle push is a slow glide.
 */
export function joystickVector(input: Point, baseRadiusPx: number): Point {
  const len = Math.hypot(input.x, input.y);
  const dead = baseRadiusPx * JOYSTICK_DEAD_ZONE_RATIO;
  if (len < dead || len === 0) {
    return { x: 0, y: 0 };
  }
  const scale = Math.min(len, baseRadiusPx) / len / baseRadiusPx;
  return { x: input.x * scale, y: input.y * scale };
}

/** Integrate one frame of crosshair motion, pinned inside the stage. */
export function stepCrosshair(
  pos: Point,
  vector: Point,
  speedPxS: number,
  dtSec: number,
  area: SpawnArea,
): Point {
  const x = Math.min(Math.max(pos.x + vector.x * speedPxS * dtSec, 0), area.width);
  const y = Math.min(Math.max(pos.y + vector.y * speedPxS * dtSec, 0), area.height);
  return { x, y };
}

/** A stage-space box a control owns (the joystick, the fire button). */
export interface Rect {
  readonly x: number;
  readonly y: number;
  readonly width: number;
  readonly height: number;
}

/** True inside the rect grown by the margin — the zone the control
 * claims for its glow and the player's thumb. */
export function insideControlRect(p: Point, rect: Rect, marginPx: number): boolean {
  return (
    p.x > rect.x - marginPx &&
    p.x < rect.x + rect.width + marginPx &&
    p.y > rect.y - marginPx &&
    p.y < rect.y + rect.height + marginPx
  );
}

/**
 * Uniform draw in the padded band that never lands on a control's
 * rect — a dot under the joystick or the fire button is unaimable.
 * Rejection sampling over the plain draw; when the stage is too small
 * to satisfy both laws it gives up after 20 tries and takes the plain
 * draw — a cramped stage prefers a playable dot over an impossible
 * one. (Deliberate divergence from the legacy uniform-over-the-stage
 * spawn, which had no control avoidance — walkthrough ruling
 * 2026-10-27.)
 */
export function randomCenterAvoidingControls(
  area: SpawnArea,
  paddingPx: number,
  controls: readonly Rect[],
  marginPx: number,
  rand: () => number,
): Point {
  const MAX_TRIES = 20;
  for (let i = 0; i < MAX_TRIES; i++) {
    const candidate = randomCenter(area, paddingPx, rand);
    if (!controls.some((rect) => insideControlRect(candidate, rect, marginPx))) {
      return candidate;
    }
  }
  return randomCenter(area, paddingPx, rand);
}

export function round1(value: number): number {
  return Math.round(value * 10) / 10;
}
