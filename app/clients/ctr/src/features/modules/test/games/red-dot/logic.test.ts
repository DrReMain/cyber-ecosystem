import { describe, expect, it } from "vitest";
import { deviationOf, isHit, outsidePaddedArea, RED_DOT_DEFAULTS, randomPosition } from "./logic";

const area = { width: 390, height: 700 };

describe("red-dot logic", () => {
  it("hits at exactly 2× the radius and misses past it", () => {
    expect(isHit(40.0, RED_DOT_DEFAULTS)).toBe(true);
    expect(isHit(40.1, RED_DOT_DEFAULTS)).toBe(false);
    expect(isHit(0, RED_DOT_DEFAULTS)).toBe(true);
  });

  it("rounds deviation to one decimal", () => {
    // 3-4-5 triangle: exact distance 5 from (0,0) to (3,4).
    expect(deviationOf({ x: 0, y: 0 }, { x: 3, y: 4 })).toBe(5);
    // 1.4142… → 1.4
    expect(deviationOf({ x: 0, y: 0 }, { x: 1, y: 1 })).toBe(1.4);
  });

  it("spawns inside the padded rect across the rng range", () => {
    const at0 = randomPosition(area, RED_DOT_DEFAULTS, () => 0);
    const at1 = randomPosition(area, RED_DOT_DEFAULTS, () => 1);
    expect(at0).toEqual({ x: 50, y: 50 });
    expect(at1).toEqual({ x: 340, y: 650 });
    expect(outsidePaddedArea(at0, area, RED_DOT_DEFAULTS)).toBe(false);
    expect(outsidePaddedArea(at1, area, RED_DOT_DEFAULTS)).toBe(false);
  });

  it("collapses to center when the area cannot hold the padding", () => {
    const tiny = { width: 80, height: 80 };
    expect(randomPosition(tiny, RED_DOT_DEFAULTS, () => 0)).toEqual({ x: 40, y: 40 });
  });

  it("flags points outside the padded rect after a shrink", () => {
    const shrunk = { width: 200, height: 700 };
    expect(outsidePaddedArea({ x: 300, y: 100 }, shrunk, RED_DOT_DEFAULTS)).toBe(true);
    expect(outsidePaddedArea({ x: 100, y: 100 }, shrunk, RED_DOT_DEFAULTS)).toBe(false);
  });
});
