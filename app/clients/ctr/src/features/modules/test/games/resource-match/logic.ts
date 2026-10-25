// resource_match pure logic — the movement and economy physics of the
// collecting field. Two laws carry the scorer's registered obligations:
// spawn spacing keeps any two pickups physically ≥ the 80 ms collect
// floor apart (two overlapping spawns would mint a fabricated density a
// truthful player cannot produce), and the per-kind totals cap at the
// spawn economy's budget (the server rejects a count past its cap, so
// the field may never mint more than the budget holds).

export type RmKind = "minion" | "monster" | "dragon";
export const RM_KINDS: readonly RmKind[] = ["minion", "monster", "dragon"];

export const RESOURCE_MATCH_DEFAULTS = {
  minionCount: 25,
  monsterCount: 12,
  dragonCount: 5,
  collisionCap: 10,
  playerSizePx: 40,
  enemySizePx: 40,
  resourceSizePx: 30,
  collectRadiusPx: 25,
  playerSpeedPxS: 180,
  enemySpeedPxS: 100,
  enemyCount: 5,
  onFieldCount: 6,
  spawnMinGapPx: 100,
  trajectoryIntervalMs: 200,
} as const;

export interface Point {
  x: number;
  y: number;
}

export interface Area {
  width: number;
  height: number;
}

export function distance(a: Point, b: Point): number {
  return Math.hypot(a.x - b.x, a.y - b.y);
}

export function clampPoint(p: Point, area: Area, halfSizePx: number): Point {
  return {
    x: Math.min(Math.max(p.x, halfSizePx), Math.max(halfSizePx, area.width - halfSizePx)),
    y: Math.min(Math.max(p.y, halfSizePx), Math.max(halfSizePx, area.height - halfSizePx)),
  };
}

// Move toward a target at a fixed speed; arriving inside one step snaps
// to the target so the player never orbits it.
export function stepToward(
  from: Point,
  target: Point,
  speedPxS: number,
  dtSec: number,
): { point: Point; arrived: boolean } {
  const dist = distance(from, target);
  const step = speedPxS * dtSec;
  if (dist <= step || dist === 0) {
    return { point: { ...target }, arrived: true };
  }
  const scale = step / dist;
  return {
    point: { x: from.x + (target.x - from.x) * scale, y: from.y + (target.y - from.y) * scale },
    arrived: false,
  };
}

// Advance a wanderer and reflect off the field's walls; the position is
// clamped in-bounds so a reflection never carries it outside.
export function stepEnemy(
  from: Point,
  dirRad: number,
  speedPxS: number,
  dtSec: number,
  area: Area,
  halfSizePx: number,
): { point: Point; dir: number } {
  let dir = dirRad;
  const next: Point = {
    x: from.x + Math.cos(dir) * speedPxS * dtSec,
    y: from.y + Math.sin(dir) * speedPxS * dtSec,
  };
  if (next.x < halfSizePx || next.x > area.width - halfSizePx) {
    dir = Math.PI - dir;
  }
  if (next.y < halfSizePx || next.y > area.height - halfSizePx) {
    dir = -dir;
  }
  return { point: clampPoint(next, area, halfSizePx), dir };
}

// Draw one point at least minGapPx from every occupied point, inside
// the padded bounds. Rejection sampling runs a bounded budget; when the
// field is too crowded to satisfy the gap, the best candidate seen (the
// farthest from its nearest neighbor) is returned so placement always
// terminates — the crowd itself is bounded by the on-field cap, so the
// fallback arm is a degenerate-field guard, not a live path.
export function spawnPoint(
  area: Area,
  occupied: readonly Point[],
  minGapPx: number,
  halfSizePx: number,
  rand: () => number,
): Point {
  let best: Point = { x: area.width / 2, y: area.height / 2 };
  let bestGap = -1;
  for (let attempt = 0; attempt < 512; attempt += 1) {
    const candidate = {
      x: halfSizePx + rand() * Math.max(0, area.width - halfSizePx * 2),
      y: halfSizePx + rand() * Math.max(0, area.height - halfSizePx * 2),
    };
    let gap = Number.POSITIVE_INFINITY;
    for (const other of occupied) {
      gap = Math.min(gap, distance(candidate, other));
    }
    if (gap >= minGapPx) {
      return candidate;
    }
    if (gap > bestGap) {
      bestGap = gap;
      best = candidate;
    }
  }
  return best;
}

export function round1(value: number): number {
  return Math.round(value * 10) / 10;
}
