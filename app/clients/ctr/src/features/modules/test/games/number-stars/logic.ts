// Pure law for number_stars — the round draw and the board layout. A
// round draws its numbers as five unique picks from the 1-20 pool (the
// pool shuffle is not seeded: the server judges only the click stream,
// so board reproducibility has no consumer), and lays them out with a
// physical min-gap so no two tap targets ever collide on a phone field.
// The tap law itself — a click is correct when it names the smallest
// not-yet-tapped number — lives here as the ascending expectation the
// engine walks with a progress pointer.

export interface NsPoint {
  readonly x: number;
  readonly y: number;
}

export interface NsArea {
  readonly width: number;
  readonly height: number;
}

export const NUMBER_STARS_DEFAULTS = {
  numbersPerRound: 5,
  minNumber: 1,
  maxNumber: 20,
  spawnPaddingPx: 60,
  spawnMinGapPx: 80,
};

/** Five unique numbers from the pool — partial Fisher-Yates, in draw order. */
export function drawNumbers(
  minNumber: number,
  maxNumber: number,
  count: number,
  rand: () => number,
): number[] {
  const pool: number[] = [];
  for (let n = minNumber; n <= maxNumber; n += 1) {
    pool.push(n);
  }
  const take = Math.min(count, pool.length);
  for (let i = 0; i < take; i += 1) {
    const j = i + Math.floor(rand() * (pool.length - i));
    const drawn = pool[j] ?? 0;
    pool[j] = pool[i] ?? 0;
    pool[i] = drawn;
  }
  return pool.slice(0, take);
}

function distance(a: NsPoint, b: NsPoint): number {
  return Math.hypot(a.x - b.x, a.y - b.y);
}

function nearestGap(points: readonly NsPoint[], candidate: NsPoint): number {
  let nearest = Number.POSITIVE_INFINITY;
  for (const point of points) {
    nearest = Math.min(nearest, distance(point, candidate));
  }
  return nearest;
}

/** Spread `count` tap targets inside the padded area, keeping the min-gap
 * law: each point draws candidates and takes the first legal one, or the
 * best (farthest-from-nearest) candidate when a small field makes the gap
 * unreachable — the engine's wrong-tap restart guarantees a retry anyway. */
export function spawnNumberPoints(
  area: NsArea,
  count: number,
  paddingPx: number,
  minGapPx: number,
  rand: () => number,
): NsPoint[] {
  const width = area.width - paddingPx * 2;
  const height = area.height - paddingPx * 2;
  if (width <= 0 || height <= 0) {
    return [];
  }
  const points: NsPoint[] = [];
  for (let i = 0; i < count; i += 1) {
    const best = drawPoint(points, width, height, paddingPx, minGapPx, rand);
    if (best !== null) {
      points.push(best);
    }
  }
  return points;
}

function drawPoint(
  points: readonly NsPoint[],
  width: number,
  height: number,
  paddingPx: number,
  minGapPx: number,
  rand: () => number,
): NsPoint | null {
  let best: NsPoint | null = null;
  let bestNearest = -1;
  for (let attempt = 0; attempt < 128; attempt += 1) {
    const candidate: NsPoint = {
      x: paddingPx + rand() * width,
      y: paddingPx + rand() * height,
    };
    const nearest = nearestGap(points, candidate);
    if (nearest >= minGapPx) {
      return candidate;
    }
    if (nearest > bestNearest) {
      bestNearest = nearest;
      best = candidate;
    }
  }
  return best;
}

/** The ascending expectation, as a sorted copy (the click stream's truth
 * stays the engine's progress pointer, this only orders it). */
export function ascending(numbers: readonly number[]): number[] {
  return [...numbers].sort((a, b) => a - b);
}
