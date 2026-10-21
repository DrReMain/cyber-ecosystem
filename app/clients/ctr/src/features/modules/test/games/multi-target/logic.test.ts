import { describe, expect, it } from "vitest";
import { drawTier, MULTI_TARGET_DEFAULTS, randomCenter, spawnInterval, TIER_SPECS } from "./logic";

const area = { width: 390, height: 700 };

describe("multi-target logic", () => {
  it("draws the weighted tiers in their exact band shares (2/3/3/2)", () => {
    expect(drawTier(() => 0).tier).toBe("high");
    expect(drawTier(() => 0.19).tier).toBe("high");
    expect(drawTier(() => 0.2).tier).toBe("medium");
    expect(drawTier(() => 0.49).tier).toBe("medium");
    expect(drawTier(() => 0.5).tier).toBe("low");
    expect(drawTier(() => 0.79).tier).toBe("low");
    expect(drawTier(() => 0.8).tier).toBe("trap");
    expect(drawTier(() => 1).tier).toBe("trap");
  });

  it("carries the legacy value and radius per tier", () => {
    const byTier = new Map(TIER_SPECS.map((spec) => [spec.tier, spec]));
    expect(byTier.get("high")).toMatchObject({ value: 100, radiusPx: 30 });
    expect(byTier.get("medium")).toMatchObject({ value: 50, radiusPx: 40 });
    expect(byTier.get("low")).toMatchObject({ value: 25, radiusPx: 50 });
    expect(byTier.get("trap")).toMatchObject({ value: -50, radiusPx: 45 });
  });

  it("draws the spawn cadence across [500, 900]", () => {
    expect(spawnInterval(MULTI_TARGET_DEFAULTS, () => 0)).toBe(500);
    expect(spawnInterval(MULTI_TARGET_DEFAULTS, () => 1)).toBe(900);
  });

  it("inscribes centers exactly; collapses on a tiny area", () => {
    expect(randomCenter(area, 40, () => 0)).toEqual({ x: 40, y: 40 });
    expect(randomCenter(area, 40, () => 1)).toEqual({ x: 350, y: 660 });
    const tiny = { width: 50, height: 700 };
    expect(randomCenter(tiny, 40, () => 0)).toEqual({ x: 25, y: 350 });
  });
});
