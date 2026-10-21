import { describe, expect, it } from "vitest";
import {
  deviationOf,
  FLASH_AIMING_DEFAULTS,
  isCenterHit,
  randomCenter,
  randomLifetime,
  randomSize,
} from "./logic";

const area = { width: 390, height: 700 };

describe("flash-aiming logic", () => {
  it("draws diameters across [30, 80] over the rng range", () => {
    expect(randomSize(FLASH_AIMING_DEFAULTS, () => 0)).toBe(30);
    expect(randomSize(FLASH_AIMING_DEFAULTS, () => 1)).toBe(80);
  });

  it("keeps centers one diameter from the edges; collapses on a tiny area", () => {
    const size = 40;
    expect(randomCenter(area, size, () => 0)).toEqual({ x: 40, y: 40 });
    expect(randomCenter(area, size, () => 1)).toEqual({ x: 350, y: 660 });
    const tiny = { width: 50, height: 700 };
    expect(randomCenter(tiny, size, () => 0)).toEqual({ x: 25, y: 350 });
  });

  it("draws lifetimes across [500, 1500]", () => {
    expect(randomLifetime(FLASH_AIMING_DEFAULTS, () => 0)).toBe(500);
    expect(randomLifetime(FLASH_AIMING_DEFAULTS, () => 1)).toBe(1500);
  });

  it("center-hits at exactly the threshold and misses past it", () => {
    expect(isCenterHit(20, FLASH_AIMING_DEFAULTS)).toBe(true);
    expect(isCenterHit(20.1, FLASH_AIMING_DEFAULTS)).toBe(false);
    expect(isCenterHit(0, FLASH_AIMING_DEFAULTS)).toBe(true);
  });

  it("keeps deviation raw — no rounding law", () => {
    // √2 stays unrounded (red_dot would round to 1.4).
    expect(deviationOf({ x: 0, y: 0 }, { x: 1, y: 1 })).toBe(Math.SQRT2);
  });
});
