import { describe, expect, it } from "vitest";
import {
  distance,
  insideControlRect,
  isHit,
  joystickVector,
  outsidePaddedArea,
  randomCenter,
  randomCenterAvoidingControls,
  stepCrosshair,
  TRACKING_MOVING_DEFAULTS,
} from "./logic";

const area = { width: 390, height: 700 };
const padding = TRACKING_MOVING_DEFAULTS.spawnPaddingPx;

describe("tracking-moving logic", () => {
  it("hits within the shared threshold, misses past it", () => {
    expect(isHit(45, TRACKING_MOVING_DEFAULTS)).toBe(true);
    expect(isHit(45.1, TRACKING_MOVING_DEFAULTS)).toBe(false);
    expect(isHit(0, TRACKING_MOVING_DEFAULTS)).toBe(true);
  });

  it("draws dot centers inside the padded band; collapses on a tiny area", () => {
    expect(randomCenter(area, padding, () => 0)).toEqual({ x: padding, y: padding });
    expect(randomCenter(area, padding, () => 1)).toEqual({ x: 350, y: 660 });
    expect(randomCenter({ width: 60, height: 700 }, padding, () => 0)).toEqual({
      x: 30,
      y: 350,
    });
  });

  it("flags a dot stranded past the spawn padding after a shrink", () => {
    const shrunk = { width: 400, height: 500 };
    // Exactly on the margin stays in-bounds; past it is stranded.
    expect(outsidePaddedArea({ x: padding, y: padding }, shrunk, padding)).toBe(false);
    expect(outsidePaddedArea({ x: padding - 0.1, y: padding }, shrunk, padding)).toBe(true);
    expect(outsidePaddedArea({ x: 380, y: 150 }, shrunk, padding)).toBe(true);
    expect(outsidePaddedArea({ x: 150, y: 190 }, shrunk, padding)).toBe(false);
  });

  it("dead-zones the stub, clamps the rim, and scales by travel", () => {
    const radius = 48;
    expect(joystickVector({ x: 0, y: 0 }, radius)).toEqual({ x: 0, y: 0 });
    // Under a fifth of the radius is tremor, not intent.
    expect(joystickVector({ x: 9.5, y: 0 }, radius)).toEqual({ x: 0, y: 0 });
    const edge = joystickVector({ x: 10, y: 0 }, radius);
    expect(edge.x).toBeGreaterThan(0);
    expect(edge.x).toBeLessThanOrEqual(1);
    // A rim hold is full speed in that direction.
    expect(joystickVector({ x: 48, y: 0 }, radius)).toEqual({ x: 1, y: 0 });
    expect(joystickVector({ x: 0, y: -48 }, radius)).toEqual({ x: 0, y: -1 });
    // A diagonal past the rim clamps to a unit vector.
    const diag = joystickVector({ x: 100, y: 100 }, radius);
    expect(diag.x).toBeCloseTo(Math.SQRT1_2, 5);
    expect(diag.y).toBeCloseTo(Math.SQRT1_2, 5);
  });

  it("integrates the crosshair at the configured speed and pins it inside", () => {
    const speed = TRACKING_MOVING_DEFAULTS.crosshairSpeedPxS;
    expect(stepCrosshair({ x: 100, y: 100 }, { x: 1, y: 0 }, speed, 0.5, area)).toEqual({
      x: 300,
      y: 100,
    });
    expect(stepCrosshair({ x: 100, y: 100 }, { x: 0, y: -0.5 }, speed, 1, area)).toEqual({
      x: 100,
      y: 0,
    });
    // A full-speed sweep clamps at the stage edge, never past it.
    expect(stepCrosshair({ x: 380, y: 690 }, { x: 1, y: 1 }, speed, 1, area)).toEqual({
      x: 390,
      y: 700,
    });
    expect(stepCrosshair({ x: 55, y: 66 }, { x: 0, y: 0 }, speed, 2, area)).toEqual({
      x: 55,
      y: 66,
    });
  });

  it("measures plain euclidean distance", () => {
    expect(distance({ x: 0, y: 0 }, { x: 3, y: 4 })).toBe(5);
  });

  it("grows the control rect by the margin and judges membership strictly", () => {
    const rect = { x: 100, y: 500, width: 120, height: 160 };
    // Exactly on the grown edge is outside; a hair inside is owned.
    expect(insideControlRect({ x: 100 - 20, y: 600 }, rect, 20)).toBe(false);
    expect(insideControlRect({ x: 100 - 20 + 0.1, y: 600 }, rect, 20)).toBe(true);
    expect(insideControlRect({ x: 100 + 120 + 20, y: 600 }, rect, 20)).toBe(false);
    expect(insideControlRect({ x: 100 + 120 + 20 - 0.1, y: 600 }, rect, 20)).toBe(true);
    // Clear of the rect on all sides.
    expect(insideControlRect({ x: 300, y: 200 }, rect, 20)).toBe(false);
  });

  it("rejects spawns on a control and returns the first legal draw", () => {
    // A joystick-shaped zone in the bottom-left corner.
    const controls = [{ x: 0, y: 520, width: 120, height: 160 }];
    // Draws: (58.6, 598) lands on the control → rejected; (303.5, 164)
    // is clear → returned.
    let calls = 0;
    const draws = [0.06, 0.9, 0.85, 0.2];
    const rand = () => draws[calls++] ?? 0.5;
    expect(randomCenterAvoidingControls(area, padding, controls, 20, rand)).toEqual({
      x: 303.5,
      y: 164,
    });
    expect(calls).toBe(4);
  });

  it("falls back to the plain draw when controls crowd the whole stage", () => {
    const controls = [{ x: -50, y: -50, width: 600, height: 900 }];
    let calls = 0;
    const rand = () => {
      calls += 1;
      return 0.5;
    };
    // Every candidate is rejected 20 times, then the 21st (plain)
    // draw is taken rather than looping forever.
    expect(randomCenterAvoidingControls(area, padding, controls, 20, rand)).toEqual({
      x: 195,
      y: 350,
    });
    expect(calls).toBe(42);
  });
});
