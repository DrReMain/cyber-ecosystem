import { describe, expect, it } from "vitest";
import {
  bestPair,
  clampInside,
  colorAt,
  decimatePath,
  distance,
  generateBoard,
  MISS_DEV_SENTINEL,
  matchPairs,
  mulberry32,
  SKILL_TRAJECTORY_DEFAULTS,
  shapeAt,
  shapeVertices,
} from "./logic";

const area = { width: 390, height: 640 };

describe("skill-trajectory logic", () => {
  it("draws one deterministic sequence per seed inside [0, 1)", () => {
    const first = mulberry32(42);
    const second = mulberry32(42);
    const other = mulberry32(43);
    for (let i = 0; i < 20; i += 1) {
      const a = first();
      const b = second();
      expect(a).toBe(b);
      expect(a).toBeGreaterThanOrEqual(0);
      expect(a).toBeLessThan(1);
    }
    expect(other()).not.toBe(first());
  });

  it("generates a three-target board with vocabulary shapes and spacing", () => {
    const board = generateBoard(mulberry32(7), area, SKILL_TRAJECTORY_DEFAULTS);
    expect(board).toHaveLength(3);
    const minSpacing = SKILL_TRAJECTORY_DEFAULTS.targetSizePx * 2.5;
    for (const target of board) {
      expect(["circle", "square", "triangle", "pentagon", "hexagon", "star"]).toContain(
        target.shape,
      );
      expect(["red", "orange", "green", "blue", "purple"]).toContain(target.color);
      expect(target.center.x).toBeGreaterThanOrEqual(SKILL_TRAJECTORY_DEFAULTS.targetSizePx);
      expect(target.center.x).toBeLessThanOrEqual(
        area.width - SKILL_TRAJECTORY_DEFAULTS.targetSizePx,
      );
    }
    for (let i = 0; i < board.length; i += 1) {
      for (let j = i + 1; j < board.length; j += 1) {
        const a = board[i];
        const b = board[j];
        if (a !== undefined && b !== undefined) {
          expect(distance(a.center, b.center)).toBeGreaterThanOrEqual(minSpacing);
        }
      }
    }
    // The same seed redraws the same board; a different seed moves it.
    expect(generateBoard(mulberry32(7), area, SKILL_TRAJECTORY_DEFAULTS)).toEqual(board);
    expect(generateBoard(mulberry32(8), area, SKILL_TRAJECTORY_DEFAULTS)).not.toEqual(board);
  });

  it("normalizes shape and color indices with a dead fallback", () => {
    expect(shapeAt(0)).toBe("circle");
    expect(shapeAt(6)).toBe("circle");
    expect(shapeAt(-1)).toBe("star");
    expect(colorAt(0)).toBe("red");
    expect(colorAt(5)).toBe("red");
    expect(colorAt(-1)).toBe("purple");
  });

  it("matches the closest same-shape same-color pair within the cap", () => {
    const target = { shape: "circle" as const, color: "blue" as const, center: { x: 100, y: 100 } };
    const close = { shape: "circle" as const, color: "blue" as const, center: { x: 125, y: 100 } };
    const nearer = { shape: "circle" as const, color: "blue" as const, center: { x: 110, y: 100 } };
    const wrongColor = {
      shape: "circle" as const,
      color: "red" as const,
      center: { x: 100, y: 100 },
    };
    const beyond = { shape: "circle" as const, color: "blue" as const, center: { x: 145, y: 100 } };
    const best = bestPair([target], [close, nearer, wrongColor], 40);
    expect(best?.placement).toBe(nearer);
    expect(best?.deviation).toBe(10);
    expect(bestPair([target], [beyond], 40)).toBeNull();
    expect(bestPair([target], [wrongColor], 40)).toBeNull();
  });

  it("pairs every match greedily nearest-first and leaves the rest unpaired", () => {
    const t1 = { shape: "circle" as const, color: "blue" as const, center: { x: 100, y: 100 } };
    const t2 = { shape: "circle" as const, color: "blue" as const, center: { x: 300, y: 100 } };
    const t3 = { shape: "square" as const, color: "red" as const, center: { x: 200, y: 300 } };
    // p1 sits between both circle targets but nearer t1; p2 claims t2;
    // p3 is a square just past the cap; nothing claims t3.
    const p1 = { shape: "circle" as const, color: "blue" as const, center: { x: 120, y: 100 } };
    const p2 = { shape: "circle" as const, color: "blue" as const, center: { x: 310, y: 100 } };
    const p3 = { shape: "square" as const, color: "red" as const, center: { x: 250, y: 300 } };
    const pairs = matchPairs([t1, t2, t3], [p1, p2, p3], 40);
    expect(pairs).toHaveLength(2);
    // Nearest-first ordering: the representative head equals bestPair.
    expect(pairs[0]?.target).toBe(t2);
    expect(pairs[0]?.placement).toBe(p2);
    expect(pairs[0]?.deviation).toBe(10);
    expect(pairs[1]?.target).toBe(t1);
    expect(pairs[1]?.placement).toBe(p1);
    expect(pairs[1]?.deviation).toBe(20);
    expect(bestPair([t1, t2, t3], [p1, p2, p3], 40)?.target).toBe(t2);
    // The wrong-color and beyond-cap placements never pair.
    expect(matchPairs([t3], [p3], 40)).toEqual([]);
    expect(matchPairs([], [], 40)).toEqual([]);
  });

  it("renders each shape's polygon geometry", () => {
    const c = { x: 0, y: 0 };
    expect(shapeVertices("circle", c, 30)).toEqual([]);
    // Axis-aligned square at circumradius 30 → half-side ≈ 21.21.
    const square = shapeVertices("square", c, 30);
    expect(square[0]?.x).toBeCloseTo(30 / Math.SQRT2, 5);
    expect(square).toHaveLength(4);
    expect(shapeVertices("triangle", c, 30)).toHaveLength(3);
    // The star alternates outer and inner radius from the top; the first
    // inner vertex sits 36° around on the inner ring.
    const star = shapeVertices("star", c, 30);
    expect(star[0]?.y).toBeCloseTo(-30, 5);
    expect(star[1]?.y).toBeCloseTo(-12 * Math.sin((54 * Math.PI) / 180), 5);
    expect(star).toHaveLength(10);
    expect(shapeVertices("hexagon", c, 30)).toHaveLength(6);
  });

  it("clamps into the padded band and collapses a too-small axis", () => {
    expect(clampInside({ x: 150, y: 320 }, area, 30)).toEqual({ x: 150, y: 320 });
    expect(clampInside({ x: -30, y: 800 }, area, 30)).toEqual({ x: 30, y: 610 });
    expect(clampInside({ x: 10, y: 320 }, { width: 40, height: 640 }, 30)).toEqual({
      x: 20,
      y: 320,
    });
  });

  it("thins drag paths keeping both endpoints", () => {
    const long = Array.from({ length: 100 }, (_, i) => ({ x: i, y: 0 }));
    const thinned = decimatePath(long, 8);
    expect(thinned).toHaveLength(8);
    expect(thinned[0]).toEqual({ x: 0, y: 0 });
    expect(thinned[7]).toEqual({ x: 99, y: 0 });
    expect(decimatePath(long, 1)).toEqual([{ x: 99, y: 0 }]);
    expect(
      decimatePath(
        [
          { x: 1, y: 1 },
          { x: 2, y: 2 },
        ],
        8,
      ),
    ).toHaveLength(2);
  });

  it("keeps the sentinel at 999 and measures euclidean distance", () => {
    expect(MISS_DEV_SENTINEL).toBe(999);
    expect(distance({ x: 0, y: 0 }, { x: 3, y: 4 })).toBe(5);
  });
});
