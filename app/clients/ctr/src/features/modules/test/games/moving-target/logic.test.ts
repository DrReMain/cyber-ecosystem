import { describe, expect, it } from "vitest";
import {
  clampInside,
  distance,
  holdingBand,
  MOVING_TARGET_DEFAULTS,
  PATTERN_ORDER,
  patternIndexAt,
  patternOffset,
  patternPhase,
  randomWalkStep,
} from "./logic";

const area = { width: 390, height: 700 };
// margin 100 → half-band 95px wide, 250px tall.
const margin = 100;

describe("moving-target logic", () => {
  it("rotates the six patterns five seconds each and wraps", () => {
    expect(PATTERN_ORDER[patternIndexAt(0, 5000)]).toBe("straight");
    expect(PATTERN_ORDER[patternIndexAt(4999, 5000)]).toBe("straight");
    expect(PATTERN_ORDER[patternIndexAt(5000, 5000)]).toBe("curve");
    expect(PATTERN_ORDER[patternIndexAt(29999, 5000)]).toBe("variable");
    expect(PATTERN_ORDER[patternIndexAt(30000, 5000)]).toBe("straight");
    // A non-positive duration or pre-run time parks on the first pattern.
    expect(patternIndexAt(99999, 0)).toBe(0);
    expect(patternIndexAt(-5, 5000)).toBe(0);
  });

  it("converts pattern age to the legacy 50 ms tick base", () => {
    expect(patternPhase(0)).toBe(0);
    expect(patternPhase(100)).toBe(2);
  });

  it("keeps every deterministic pattern inside its amplitude band", () => {
    // straight/circle at phase 0: center row, full amplitude sideways.
    expect(patternOffset("straight", 0, area, margin)).toEqual({ x: 0, y: 0 });
    const circle = patternOffset("circle", 0, area, margin);
    expect(circle.x).toBeCloseTo(95, 5);
    expect(circle.y).toBeCloseTo(0, 5);
    // straight peaks when sin(phase·0.02) = 1, i.e. phase = 25π.
    const straight = patternOffset("straight", 25 * Math.PI, area, margin);
    expect(straight.x).toBeCloseTo(95, 5);
    // The curve pattern peaks on the short axis at sin(phase·0.05) = 1,
    // i.e. phase = 10π, where the vertical amplitude clamps at 200.
    const curve = patternOffset("curve", 10 * Math.PI, area, margin);
    expect(curve.y).toBeCloseTo(200, 5);
    // A stage without a padded band pins everything at zero offset.
    expect(patternOffset("straight", 100, { width: 150, height: 150 }, margin)).toEqual({
      x: 0,
      y: 0,
    });
  });

  it("clamps stray points in-bounds and collapses on a tiny stage", () => {
    expect(clampInside({ x: 150, y: 350 }, area, margin)).toEqual({ x: 150, y: 350 });
    expect(clampInside({ x: -30, y: 800 }, area, margin)).toEqual({ x: 100, y: 600 });
    expect(clampInside({ x: 10, y: 350 }, { width: 150, height: 700 }, margin)).toEqual({
      x: 75,
      y: 350,
    });
  });

  it("walks one random step per jitter pair and clamps to the margin", () => {
    const stepped = randomWalkStep({ x: 150, y: 350 }, area, margin, () => 0.5);
    expect(stepped).toEqual({ x: 155, y: 355 });
    // A −10 jump from 595 stays in-bounds; the same jump from 95 clamps.
    const inward = randomWalkStep({ x: 95, y: 595 }, area, margin, () => -1);
    expect(inward).toEqual({ x: 100, y: 585 });
    const outward = randomWalkStep({ x: 380, y: 590 }, area, margin, () => 1);
    expect(outward).toEqual({ x: 290, y: 600 });
  });

  it("bands deviations by the server's thresholds", () => {
    expect(holdingBand(0, MOVING_TARGET_DEFAULTS)).toBe("perfect");
    expect(holdingBand(20, MOVING_TARGET_DEFAULTS)).toBe("perfect");
    expect(holdingBand(20.1, MOVING_TARGET_DEFAULTS)).toBe("good");
    expect(holdingBand(30, MOVING_TARGET_DEFAULTS)).toBe("good");
    expect(holdingBand(30.1, MOVING_TARGET_DEFAULTS)).toBe("lost");
  });

  it("measures plain euclidean distance", () => {
    expect(distance({ x: 0, y: 0 }, { x: 3, y: 4 })).toBe(5);
  });
});
