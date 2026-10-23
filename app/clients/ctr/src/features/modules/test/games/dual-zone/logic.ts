// Pure law for dual_zone — bar geometry, obstacle motion, the collision
// AABB, and the pass-plane that mints a dodge. Axes de-mirrored from the
// legacy page: left-zone obstacles fall down toward a horizontal bar
// dragged along x; right-zone obstacles fly right toward a vertical bar
// dragged along y.

export interface Point {
  readonly x: number;
  readonly y: number;
}

export interface ZoneSize {
  readonly width: number;
  readonly height: number;
}

export interface Rect {
  readonly left: number;
  readonly top: number;
  readonly right: number;
  readonly bottom: number;
}

export type Zone = "left" | "right";

export const DUAL_ZONE_DEFAULTS = {
  shapeSizePx: 36,
  barThicknessPx: 24,
  barRatio: 0.35,
  obstacleSpeedPxS: 120,
  spawnIntervalMinMs: 800,
  spawnIntervalSpanMs: 700,
};

// The bar's fixed offset from its zone's far edge (visual breathing room
// between the bar and the boundary).
export const BAR_INSET_PX = 12;

export function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), Math.max(min, max));
}

// The dodge bar: the left zone's lies horizontal near its bottom edge
// (dragged along x), the right zone's vertical near its right edge
// (dragged along y); each length is bar_ratio of its cross extent.
export function barRect(
  zone: Zone,
  size: ZoneSize,
  pos: number,
  params: typeof DUAL_ZONE_DEFAULTS,
): Rect {
  if (zone === "left") {
    const length = size.width * params.barRatio;
    const top = size.height - BAR_INSET_PX - params.barThicknessPx;
    return {
      left: pos - length / 2,
      top,
      right: pos + length / 2,
      bottom: top + params.barThicknessPx,
    };
  }
  const length = size.height * params.barRatio;
  const left = size.width - BAR_INSET_PX - params.barThicknessPx;
  return {
    left,
    top: pos - length / 2,
    right: left + params.barThicknessPx,
    bottom: pos + length / 2,
  };
}

// The bar's drag axis clamps inside its zone; a zone narrower than the
// bar parks the bar centered.
export function clampBarPos(
  zone: Zone,
  size: ZoneSize,
  pos: number,
  params: typeof DUAL_ZONE_DEFAULTS,
): number {
  const extent = zone === "left" ? size.width : size.height;
  const half = (extent * params.barRatio) / 2;
  return clamp(pos, half, extent - half);
}

// Axis-aligned overlap between the obstacle's square bounds and the bar.
export function aabbOverlap(center: Point, sizePx: number, bar: Rect): boolean {
  const half = sizePx / 2;
  return (
    center.x + half > bar.left &&
    center.x - half < bar.right &&
    center.y + half > bar.top &&
    center.y - half < bar.bottom
  );
}

// The dodge plane: a left-zone obstacle dodges once fully below the bar,
// a right-zone obstacle once fully past its right edge.
export function passedBar(zone: Zone, center: Point, sizePx: number, bar: Rect): boolean {
  const half = sizePx / 2;
  return zone === "left" ? center.y - half > bar.bottom : center.x - half > bar.right;
}

// Obstacles enter from outside the near edge: the left zone from above
// the top at a random x, the right zone from left of the edge at a
// random y.
export function spawnObstacle(
  zone: Zone,
  size: ZoneSize,
  sizePx: number,
  rand: () => number,
): Point {
  const half = sizePx / 2;
  const span = Math.max(0, (zone === "left" ? size.width : size.height) - sizePx);
  return zone === "left"
    ? { x: half + rand() * span, y: -half }
    : { x: -half, y: half + rand() * span };
}

export function stepObstacle(zone: Zone, center: Point, speedPxS: number, dtSec: number): Point {
  return zone === "left"
    ? { x: center.x, y: center.y + speedPxS * dtSec }
    : { x: center.x + speedPxS * dtSec, y: center.y };
}

// Fully out the far side — a resolved obstacle stops existing once it
// clears the zone.
export function outOfZone(zone: Zone, center: Point, sizePx: number, size: ZoneSize): boolean {
  const half = sizePx / 2;
  return zone === "left" ? center.y - half > size.height : center.x - half > size.width;
}

// A shrunk zone can strand an obstacle beyond the bar's reach on its
// cross axis — re-clamp that one axis only; the travel axis keeps its
// position (entry offsets above the top are legal out-of-bounds).
export function clampObstacle(zone: Zone, center: Point, size: ZoneSize, sizePx: number): Point {
  const half = sizePx / 2;
  if (zone === "left") {
    return { x: clamp(center.x, half, Math.max(half, size.width - half)), y: center.y };
  }
  return { x: center.x, y: clamp(center.y, half, Math.max(half, size.height - half)) };
}

export function round1(value: number): number {
  return Math.round(value * 10) / 10;
}

export function nextSpawnDelay(
  params: { spawnIntervalMinMs: number; spawnIntervalSpanMs: number },
  rand: () => number,
): number {
  return params.spawnIntervalMinMs + rand() * params.spawnIntervalSpanMs;
}
