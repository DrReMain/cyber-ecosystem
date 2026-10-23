import { describe, expect, it } from "vitest";
import {
  distance,
  drawType,
  isHit,
  outsidePaddedArea,
  randomCenter,
  TARGET_RECOGNITION_DEFAULTS,
} from "./logic";

const area = { width: 390, height: 700 };

describe("target-recognition logic", () => {
  it("draws the type lot over the 0.6/0.25/0.15 bands", () => {
    expect(drawType(() => 0)).toBe("enemy");
    expect(drawType(() => 0.59)).toBe("enemy");
    expect(drawType(() => 0.6)).toBe("friendly");
    expect(drawType(() => 0.84)).toBe("friendly");
    expect(drawType(() => 0.85)).toBe("neutral");
    expect(drawType(() => 1)).toBe("neutral");
  });

  it("keeps centers inside the 100px margin; collapses on a tiny area", () => {
    expect(randomCenter(area, 100, () => 0)).toEqual({ x: 100, y: 100 });
    expect(randomCenter(area, 100, () => 1)).toEqual({ x: 290, y: 600 });
    const tiny = { width: 150, height: 700 };
    expect(randomCenter(tiny, 100, () => 0)).toEqual({ x: 75, y: 350 });
  });

  it("hits within radius + slop − edge margin, misses past it", () => {
    // enemy: diameter 25 → boundary = 12.5 + 25 − 2 = 35.5
    expect(isHit(35.5, 25, TARGET_RECOGNITION_DEFAULTS)).toBe(true);
    expect(isHit(35.6, 25, TARGET_RECOGNITION_DEFAULTS)).toBe(false);
    expect(isHit(0, 25, TARGET_RECOGNITION_DEFAULTS)).toBe(true);
    // friendly: diameter 20 → boundary = 10 + 25 − 2 = 33
    expect(isHit(33, 20, TARGET_RECOGNITION_DEFAULTS)).toBe(true);
    expect(isHit(33.1, 20, TARGET_RECOGNITION_DEFAULTS)).toBe(false);
  });

  it("measures plain euclidean distance", () => {
    expect(distance({ x: 0, y: 0 }, { x: 3, y: 4 })).toBe(5);
  });

  it("flags a center stranded past the spawn padding after a shrink", () => {
    const padding = TARGET_RECOGNITION_DEFAULTS.spawnPaddingPx;
    const shrunk = { width: 400, height: 500 };
    // Exactly on the margin stays in-bounds; past it is stranded.
    expect(outsidePaddedArea({ x: padding, y: padding }, shrunk, padding)).toBe(false);
    expect(outsidePaddedArea({ x: padding - 0.1, y: padding }, shrunk, padding)).toBe(true);
    expect(outsidePaddedArea({ x: 380, y: 150 }, shrunk, padding)).toBe(true);
    expect(outsidePaddedArea({ x: 150, y: 190 }, shrunk, padding)).toBe(false);
  });
});
