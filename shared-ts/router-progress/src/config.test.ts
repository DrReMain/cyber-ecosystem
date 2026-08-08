import { describe, expect, it } from "vitest";
import { resolveRouterProgressConfig } from "./config";

describe("resolveRouterProgressConfig", () => {
  it("normalizes invalid timing and percentage values", () => {
    const config = resolveRouterProgressConfig({
      height: -1,
      initialMin: 90,
      initialMax: 10,
      trickleCeiling: 5,
      trickleRate: Number.POSITIVE_INFINITY,
      finishDelay: -1,
      transitionDuration: -1,
      cycleDuration: -1,
      color: " ",
    });

    expect(config.height).toBe(0);
    expect(config.initialMax).toBe(90);
    expect(config.trickleCeiling).toBe(90);
    expect(config.finishDelay).toBe(0);
    expect(config.transitionDuration).toBe(0);
    expect(config.cycleDuration).toBe(0);
    expect(config.color).not.toBe(" ");
  });
});
