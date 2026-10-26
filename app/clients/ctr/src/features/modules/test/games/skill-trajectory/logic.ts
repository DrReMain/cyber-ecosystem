// Pure board law for skill_trajectory — the seeded board, the placement
// matching, and the polygon geometry the module renders with. The seed is
// the board's only input: each target draws its position (with spacing
// retries), then its shape, then its color, all from the PRNG, so the
// server can reproduce the board from random_seed later.

export interface Point {
  readonly x: number;
  readonly y: number;
}

export interface SpawnArea {
  readonly width: number;
  readonly height: number;
}

// The contract's shape vocabulary in fixed draw order.
export const SHAPE_ORDER = ["circle", "square", "triangle", "pentagon", "hexagon", "star"] as const;
export type StShape = (typeof SHAPE_ORDER)[number];

// Semantic color tokens — facts must not drift with the locale, so the
// round's color_name carries a stable token and the view maps it.
export const COLOR_TOKENS = ["red", "orange", "green", "blue", "purple"] as const;
export type StColor = (typeof COLOR_TOKENS)[number];

export const SKILL_TRAJECTORY_DEFAULTS = {
  totalRounds: 5,
  displayTimeMs: 3000,
  targetSizePx: 60,
  paletteSizePx: 40,
  matchDistancePx: 40,
};

// The unmatched round's deviation sentinel (the contract's law: no
// placement pairs with 999).
export const MISS_DEV_SENTINEL = 999;

// Deterministic 32-bit PRNG — one seed draws one board sequence.
export function mulberry32(seed: number): () => number {
  let a = seed >>> 0;
  return () => {
    a = (a + 0x6d2b79f5) >>> 0;
    let t = a;
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

export function shapeAt(index: number): StShape {
  const order = SHAPE_ORDER.length;
  return SHAPE_ORDER[((index % order) + order) % order] ?? "circle";
}

export function colorAt(index: number): StColor {
  const order = COLOR_TOKENS.length;
  return COLOR_TOKENS[((index % order) + order) % order] ?? "blue";
}

export function round1(value: number): number {
  return Math.round(value * 10) / 10;
}

export function distance(a: Point, b: Point): number {
  return Math.hypot(a.x - b.x, a.y - b.y);
}

export function randomCenter(area: SpawnArea, padding: number, rand: () => number): Point {
  const w = Math.max(0, area.width - padding * 2);
  const h = Math.max(0, area.height - padding * 2);
  if (w === 0 || h === 0) {
    return { x: area.width / 2, y: area.height / 2 };
  }
  return { x: padding + rand() * w, y: padding + rand() * h };
}

function clampAxis(value: number, size: number, padding: number): number {
  if (size <= padding * 2) {
    return size / 2;
  }
  return Math.min(Math.max(value, padding), size - padding);
}

// Clamp into the padded band; an axis too small for its padding collapses
// to its center.
export function clampInside(p: Point, area: SpawnArea, padding: number): Point {
  return {
    x: clampAxis(p.x, area.width, padding),
    y: clampAxis(p.y, area.height, padding),
  };
}

export interface StTarget {
  readonly shape: StShape;
  readonly color: StColor;
  readonly center: Point;
}

export interface StPlacement {
  readonly shape: StShape;
  readonly color: StColor;
  readonly center: Point;
}

// Three targets per board, positions rejected against the min-spacing
// rule (2.5× the target size, the legacy law) with a 100-attempt cap —
// past the cap the last draw wins, matching the legacy settle.
export function generateBoard(
  rng: () => number,
  area: SpawnArea,
  params: typeof SKILL_TRAJECTORY_DEFAULTS,
): StTarget[] {
  const minSpacing = params.targetSizePx * 2.5;
  const targets: StTarget[] = [];
  for (let i = 0; i < 3; i += 1) {
    let center = randomCenter(area, params.targetSizePx, rng);
    let attempts = 0;
    while (
      attempts < 100 &&
      targets.some((target) => distance(target.center, center) < minSpacing)
    ) {
      center = randomCenter(area, params.targetSizePx, rng);
      attempts += 1;
    }
    const shape = shapeAt(Math.floor(rng() * SHAPE_ORDER.length));
    const color = colorAt(Math.floor(rng() * COLOR_TOKENS.length));
    targets.push({ shape, color, center });
  }
  return targets;
}

// The round's representative: the closest same-shape same-color pair
// within the match distance. The reported original/placement/deviation
// triple is exact, so a board-reproducing server can verify it as-is.
// Generic so callers keep their concrete entity types (the module's
// pieces carry the drag path the fact's trajectory reads).
// One candidate pair's deviation, or null when the shape/color gate or the
// distance cap refuses it.
function openCandidate<T extends StTarget, P extends StPlacement>(
  target: T,
  placement: P,
  maxDistancePx: number,
): number | null {
  if (placement.shape !== target.shape || placement.color !== target.color) {
    return null;
  }
  const deviation = distance(target.center, placement.center);
  return deviation > maxDistancePx ? null : deviation;
}

// The closest still-open same-shape same-color pair within the cap — the
// shared search under both the fact's representative and the feedback
// face's exhaustive pairing.
function bestOpenPair<T extends StTarget, P extends StPlacement>(
  targets: readonly T[],
  placements: readonly P[],
  takenTargets: ReadonlySet<T>,
  takenPlacements: ReadonlySet<P>,
  maxDistancePx: number,
): { target: T; placement: P; deviation: number } | null {
  let best: { target: T; placement: P; deviation: number } | null = null;
  for (const target of targets) {
    if (takenTargets.has(target)) {
      continue;
    }
    for (const placement of placements) {
      if (takenPlacements.has(placement)) {
        continue;
      }
      const deviation = openCandidate(target, placement, maxDistancePx);
      if (deviation !== null && (best === null || deviation < best.deviation)) {
        best = { target, placement, deviation };
      }
    }
  }
  return best;
}

// The round's representative: the closest same-shape same-color pair
// within the match distance. The reported original/placement/deviation
// triple is exact, so a board-reproducing server can verify it as-is.
// Generic so callers keep their concrete entity types (the module's
// pieces carry the drag path the fact's trajectory reads).
export function bestPair<T extends StTarget, P extends StPlacement>(
  targets: readonly T[],
  placements: readonly P[],
  maxDistancePx: number,
): { target: T; placement: P; deviation: number } | null {
  return bestOpenPair(targets, placements, new Set(), new Set(), maxDistancePx);
}

// Every pair the feedback face shows: the same shape/color gate and
// distance cap as the fact's representative, but the globally closest
// remaining pair is taken repeatedly until nothing pairs (the legacy
// feedback screen's greedy display law). Pairs are ordered nearest-first,
// so the representative (bestPair) is always the head when both match.
export function matchPairs<T extends StTarget, P extends StPlacement>(
  targets: readonly T[],
  placements: readonly P[],
  maxDistancePx: number,
): { target: T; placement: P; deviation: number }[] {
  const pairs: { target: T; placement: P; deviation: number }[] = [];
  const takenTargets = new Set<T>();
  const takenPlacements = new Set<P>();
  for (;;) {
    const best = bestOpenPair(targets, placements, takenTargets, takenPlacements, maxDistancePx);
    if (best === null) {
      return pairs;
    }
    takenTargets.add(best.target);
    takenPlacements.add(best.placement);
    pairs.push(best);
  }
}

// Polygon vertices for the non-circle shapes (circle renders as an SVG
// circle); angles are degrees clockwise from up.
export function shapeVertices(shape: StShape, center: Point, radius: number): Point[] {
  const at = (deg: number, r: number): Point => {
    const rad = ((deg - 90) * Math.PI) / 180;
    return { x: center.x + r * Math.cos(rad), y: center.y + r * Math.sin(rad) };
  };
  switch (shape) {
    case "circle":
      return [];
    case "square":
      return [at(45, radius), at(135, radius), at(225, radius), at(315, radius)];
    case "triangle":
      return [at(0, radius), at(120, radius), at(240, radius)];
    case "pentagon":
      return [0, 72, 144, 216, 288].map((deg) => at(deg, radius));
    case "hexagon":
      return [0, 60, 120, 180, 240, 300].map((deg) => at(deg, radius));
    case "star": {
      const vertices: Point[] = [];
      for (let i = 0; i < 10; i += 1) {
        vertices.push(at(i * 36, i % 2 === 0 ? radius : radius * 0.4));
      }
      return vertices;
    }
  }
}

// Thin a drag polyline to at most maxPoints keeping both endpoints — the
// trajectory rides the contract without bloating it.
export function decimatePath(points: readonly Point[], maxPoints: number): Point[] {
  const count = points.length;
  if (count === 0) {
    return [];
  }
  const last = points[count - 1];
  if (last === undefined || maxPoints <= 1) {
    return last === undefined ? [] : [last];
  }
  if (count <= maxPoints) {
    return points.slice();
  }
  const stride = (count - 1) / (maxPoints - 1);
  const out: Point[] = [];
  for (let i = 0; i < maxPoints - 1; i += 1) {
    const p = points[Math.round(i * stride)];
    if (p !== undefined) {
      out.push(p);
    }
  }
  if (last !== undefined) {
    out.push(last);
  }
  return out;
}
