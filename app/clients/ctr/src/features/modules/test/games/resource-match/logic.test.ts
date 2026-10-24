import { describe, expect, it } from "vitest";
import {
  clampPoint,
  distance,
  RESOURCE_MATCH_DEFAULTS,
  spawnPoint,
  stepEnemy,
  stepToward,
} from "./logic";

describe("stepToward", () => {
  it("snaps to the target inside one step", () => {
    const moved = stepToward({ x: 0, y: 0 }, { x: 5, y: 0 }, 180, 1);
    expect(moved.point).toEqual({ x: 5, y: 0 });
    expect(moved.arrived).toBe(true);
  });

  it("moves exactly the speed budget along the line", () => {
    const moved = stepToward({ x: 0, y: 0 }, { x: 300, y: 400 }, 180, 0.5);
    expect(moved.arrived).toBe(false);
    // 3-4-5 direction: half a second at 180px/s is 90px = 54/72.
    expect(moved.point.x).toBeCloseTo(54);
    expect(moved.point.y).toBeCloseTo(72);
  });

  it("holds still on a zero-length leg", () => {
    const moved = stepToward({ x: 7, y: 7 }, { x: 7, y: 7 }, 180, 1);
    expect(moved.point).toEqual({ x: 7, y: 7 });
    expect(moved.arrived).toBe(true);
  });
});

describe("stepEnemy", () => {
  const area = { width: 300, height: 300 };

  it("reflects off the right wall and stays in bounds", () => {
    const stepped = stepEnemy({ x: 296, y: 150 }, 0, 100, 0.2, area, 20);
    expect(stepped.point.x).toBeLessThanOrEqual(280);
    // A rightward heading (0) reflects to leftward (π) modulo float noise.
    expect(Math.abs(Math.abs(stepped.dir) - Math.PI)).toBeLessThan(1e-9);
  });

  it("reflects off the top wall vertically", () => {
    const stepped = stepEnemy({ x: 150, y: 22 }, -Math.PI / 2, 100, 0.2, area, 20);
    expect(stepped.point.y).toBeGreaterThanOrEqual(20);
    // An upward heading flips to downward.
    expect(stepped.dir).toBeCloseTo(Math.PI / 2);
  });

  it("crosses freely inside the field", () => {
    const stepped = stepEnemy({ x: 150, y: 150 }, Math.PI, 100, 0.1, area, 20);
    expect(stepped.point.x).toBeCloseTo(140);
    expect(stepped.dir).toBeCloseTo(Math.PI);
  });
});

describe("clampPoint", () => {
  it("clamps both axes into the padded band", () => {
    const clamped = clampPoint({ x: -5, y: 999 }, { width: 200, height: 100 }, 20);
    expect(clamped).toEqual({ x: 20, y: 80 });
  });
});

describe("spawnPoint", () => {
  const area = { width: 400, height: 500 };
  const gap = RESOURCE_MATCH_DEFAULTS.spawnMinGapPx;

  it("honors the min gap from every occupied point when the field has room", () => {
    const occupied = [
      { x: 50, y: 50 },
      { x: 350, y: 450 },
      { x: 200, y: 250 },
    ];
    for (let i = 0; i < 50; i += 1) {
      const point = spawnPoint(area, occupied, gap, 15, Math.random);
      for (const other of occupied) {
        expect(distance(point, other)).toBeGreaterThanOrEqual(gap - 1e-6);
      }
      expect(point.x).toBeGreaterThanOrEqual(15);
      expect(point.x).toBeLessThanOrEqual(385);
      expect(point.y).toBeGreaterThanOrEqual(15);
      expect(point.y).toBeLessThanOrEqual(485);
    }
  });

  it("terminates on a degenerate field by returning the farthest candidate", () => {
    // A grid denser than the gap can admit: the fallback arm must still
    // return an in-bounds point.
    const crowded: { x: number; y: number }[] = [];
    for (let x = 30; x <= 370; x += 60) {
      for (let y = 30; y <= 470; y += 60) {
        crowded.push({ x, y });
      }
    }
    const point = spawnPoint(area, crowded, gap, 15, Math.random);
    expect(point.x).toBeGreaterThanOrEqual(15);
    expect(point.x).toBeLessThanOrEqual(385);
    expect(point.y).toBeGreaterThanOrEqual(15);
    expect(point.y).toBeLessThanOrEqual(485);
  });
});
