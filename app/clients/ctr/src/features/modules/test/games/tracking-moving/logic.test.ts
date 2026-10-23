import { describe, expect, it } from "vitest";
import {
  distance,
  isHit,
  joystickVector,
  outsidePaddedArea,
  randomCenter,
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
});
