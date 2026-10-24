import { describe, expect, it } from "vitest";
import {
  aabbOverlap,
  BAR_INSET_PX,
  barRect,
  clampBarPos,
  clampObstacle,
  DUAL_ZONE_DEFAULTS,
  nextSpawnDelay,
  outOfZone,
  passedBar,
  spawnObstacle,
  stepObstacle,
} from "./logic";

const zone = { width: 200, height: 400 };
const p = DUAL_ZONE_DEFAULTS;

describe("dual-zone logic", () => {
  it("lays the left bar horizontal near the bottom and the right vertical near the edge", () => {
    // Left: length 0.35×200=70 centered on pos, top anchored inset+24 above the floor.
    const left = barRect("left", zone, 100, p);
    expect(left).toEqual({
      left: 65,
      top: 400 - BAR_INSET_PX - 24,
      right: 135,
      bottom: 400 - BAR_INSET_PX,
    });
    // Right: length 0.35×400=140 centered on pos, against the right edge.
    const right = barRect("right", zone, 200, p);
    expect(right).toEqual({
      left: 200 - BAR_INSET_PX - 24,
      top: 130,
      right: 200 - BAR_INSET_PX,
      bottom: 270,
    });
  });

  it("clamps the bar's drag axis inside its zone, small zones included", () => {
    expect(clampBarPos("left", zone, 10, p)).toBe(35);
    expect(clampBarPos("left", zone, 190, p)).toBe(165);
    expect(clampBarPos("left", zone, 100, p)).toBe(100);
    // With bar_ratio below one the clamp band always fits a zone, even a
    // narrow one: 30 wide → [5.25, 24.75].
    expect(clampBarPos("left", { width: 30, height: 400 }, 0, p)).toBe(5.25);
    expect(clampBarPos("left", { width: 30, height: 400 }, 30, p)).toBe(24.75);
  });

  it("detects strict AABB overlap without counting touching edges", () => {
    const bar = { left: 50, top: 100, right: 150, bottom: 124 };
    expect(aabbOverlap({ x: 100, y: 112 }, 36, bar)).toBe(true);
    // Touching exactly (left edge of obstacle == right edge of bar) is not overlap.
    expect(aabbOverlap({ x: 168, y: 112 }, 36, bar)).toBe(false);
    expect(aabbOverlap({ x: 30, y: 112 }, 36, bar)).toBe(false);
    expect(aabbOverlap({ x: 100, y: 60 }, 36, bar)).toBe(false);
  });

  it("mints a dodge only once the obstacle fully crosses the bar plane", () => {
    const leftBar = barRect("left", zone, 100, p);
    // half=18; below the bar's bottom (388) needs center.y > 406.
    expect(passedBar("left", { x: 100, y: 406 }, 36, leftBar)).toBe(false);
    expect(passedBar("left", { x: 100, y: 407 }, 36, leftBar)).toBe(true);
    const rightBar = barRect("right", zone, 200, p);
    // Past the bar's right edge (188) needs center.x > 206.
    expect(passedBar("right", { x: 206, y: 200 }, 36, rightBar)).toBe(false);
    expect(passedBar("right", { x: 207, y: 200 }, 36, rightBar)).toBe(true);
  });

  it("spawns obstacles outside the near edge at a random cross position", () => {
    const left = spawnObstacle("left", zone, 36, () => 0.5);
    expect(left).toEqual({ x: 18 + 0.5 * (200 - 36), y: -18 });
    const right = spawnObstacle("right", zone, 36, () => 1);
    expect(right).toEqual({ x: -18, y: 18 + (400 - 36) });
  });

  it("moves each zone's obstacles along its travel axis only", () => {
    expect(stepObstacle("left", { x: 50, y: 10 }, 120, 0.5)).toEqual({ x: 50, y: 70 });
    expect(stepObstacle("right", { x: 10, y: 50 }, 120, 0.5)).toEqual({ x: 70, y: 50 });
  });

  it("drops obstacles from the stream once fully out the far side", () => {
    // half=18; out needs center.y − 18 > 400, i.e. y > 418.
    expect(outOfZone("left", { x: 50, y: 418 }, 36, zone)).toBe(false);
    expect(outOfZone("left", { x: 50, y: 418.1 }, 36, zone)).toBe(true);
    expect(outOfZone("right", { x: 219, y: 50 }, 36, zone)).toBe(true);
  });

  it("re-clamps only the cross axis on a shrink", () => {
    // Left zone shrunk to 150 wide: x clamps in, the travel axis y=-50 (a
    // still-entering obstacle) stays put.
    expect(clampObstacle("left", { x: 380, y: -50 }, { width: 150, height: 400 }, 36)).toEqual({
      x: 132,
      y: -50,
    });
    expect(clampObstacle("right", { x: -10, y: 900 }, { width: 200, height: 400 }, 36)).toEqual({
      x: -10,
      y: 382,
    });
  });

  it("spreads spawn delays over the configured min+span", () => {
    expect(nextSpawnDelay(p, () => 0)).toBe(800);
    expect(nextSpawnDelay(p, () => 1)).toBe(1500);
    expect(nextSpawnDelay(p, () => 0.5)).toBe(1150);
  });
});
