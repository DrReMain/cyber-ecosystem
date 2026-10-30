import { describe, expect, it } from "vitest";
import { ascending, drawNumbers, spawnNumberPoints } from "./logic";

describe("drawNumbers", () => {
  it("draws the requested count as unique numbers inside the pool", () => {
    for (let run = 0; run < 50; run += 1) {
      const drawn = drawNumbers(1, 20, 5, Math.random);
      expect(drawn).toHaveLength(5);
      expect(new Set(drawn).size).toBe(5);
      for (const n of drawn) {
        expect(n).toBeGreaterThanOrEqual(1);
        expect(n).toBeLessThanOrEqual(20);
      }
    }
  });

  it("clamps to the pool size when the ask exceeds it", () => {
    const drawn = drawNumbers(1, 3, 5, Math.random);
    expect(drawn).toHaveLength(3);
    expect(new Set(drawn).size).toBe(3);
  });
});

describe("spawnNumberPoints", () => {
  it("returns nothing on a degenerate area", () => {
    expect(spawnNumberPoints({ width: 0, height: 0 }, 5, 60, 80, Math.random)).toEqual([]);
  });

  it("keeps every point inside the padded bounds", () => {
    const area = { width: 390, height: 520 };
    for (let run = 0; run < 50; run += 1) {
      for (const point of spawnNumberPoints(area, 5, 60, 80, Math.random)) {
        expect(point.x).toBeGreaterThanOrEqual(60);
        expect(point.x).toBeLessThanOrEqual(330);
        expect(point.y).toBeGreaterThanOrEqual(60);
        expect(point.y).toBeLessThanOrEqual(460);
      }
    }
  });

  it("keeps the min-gap law on a field that fits five targets", () => {
    const area = { width: 390, height: 520 };
    for (let run = 0; run < 50; run += 1) {
      const points = spawnNumberPoints(area, 5, 60, 80, Math.random);
      expect(points).toHaveLength(5);
      for (let i = 0; i < points.length; i += 1) {
        for (let j = i + 1; j < points.length; j += 1) {
          const a = points[i];
          const b = points[j];
          if (a === undefined || b === undefined) {
            throw new Error("unreachable");
          }
          expect(Math.hypot(a.x - b.x, a.y - b.y)).toBeGreaterThanOrEqual(80);
        }
      }
    }
  });
});

describe("ascending", () => {
  it("sorts a copy and leaves the source order alone", () => {
    const source = [9, 3, 7];
    expect(ascending(source)).toEqual([3, 7, 9]);
    expect(source).toEqual([9, 3, 7]);
  });
});
