import { describe, expect, it } from "vitest";
import {
  confirmsHit,
  ECONOMIC_SIM_DEFAULTS,
  goldOf,
  healthAt,
  nextMinionType,
  windowOf,
} from "./logic";

const D = ECONOMIC_SIM_DEFAULTS;

describe("healthAt — the stepped decay", () => {
  it("walks the melee minion down in whole tens", () => {
    expect(healthAt(0, "melee", D)).toBe(100);
    expect(healthAt(100, "melee", D)).toBe(90);
    expect(healthAt(900, "melee", D)).toBe(10);
    expect(healthAt(1000, "melee", D)).toBe(0);
    expect(healthAt(5000, "melee", D)).toBe(0);
  });

  it("clamps a pre-spawn negative to full", () => {
    expect(healthAt(-50, "melee", D)).toBe(100);
  });

  it("lands the ranged minion on four — the game's only perfect value", () => {
    // 100 - 12k: ..., 16, 4 — perfect (≤5); a further step dies.
    expect(healthAt(800, "ranged", D)).toBe(4);
    expect(healthAt(900, "ranged", D)).toBe(0);
  });

  it("never lands a melee or cannon minion inside its perfect window", () => {
    // The legacy geometry: melee steps in tens (best is good at 10),
    // cannon in fifteens (best is good at 10) — no perfect taps exist.
    for (let t = 0; t <= 700; t += 25) {
      const melee = healthAt(t, "melee", D);
      const cannon = healthAt(t, "cannon", D);
      if (melee > 0) {
        expect(windowOf(melee, "melee", D)).not.toBe("perfect");
      }
      if (cannon > 0) {
        expect(windowOf(cannon, "cannon", D)).not.toBe("perfect");
      }
    }
  });
});

describe("windowOf — the four windows", () => {
  it("bounds the melee windows at 5 / 10 / 30", () => {
    expect(windowOf(5, "melee", D)).toBe("perfect");
    expect(windowOf(6, "melee", D)).toBe("good");
    expect(windowOf(10, "melee", D)).toBe("good");
    expect(windowOf(11, "melee", D)).toBe("early");
    expect(windowOf(30, "melee", D)).toBe("early");
    expect(windowOf(31, "melee", D)).toBe("too_early");
  });

  it("bounds the cannon windows at 8 / 15 / 30", () => {
    expect(windowOf(8, "cannon", D)).toBe("perfect");
    expect(windowOf(9, "cannon", D)).toBe("good");
    expect(windowOf(15, "cannon", D)).toBe("good");
    expect(windowOf(16, "cannon", D)).toBe("early");
    expect(windowOf(31, "cannon", D)).toBe("too_early");
  });

  it("confirms only perfect and good as last hits", () => {
    expect(confirmsHit("perfect")).toBe(true);
    expect(confirmsHit("good")).toBe(true);
    expect(confirmsHit("early")).toBe(false);
    expect(confirmsHit("too_early")).toBe(false);
  });
});

describe("goldOf — floored shares", () => {
  it("pays the minion bases with floored shares", () => {
    expect(goldOf("perfect", "melee", D)).toBe(21);
    expect(goldOf("good", "melee", D)).toBe(12);
    expect(goldOf("early", "melee", D)).toBe(6);
    expect(goldOf("too_early", "melee", D)).toBe(0);
    expect(goldOf("perfect", "cannon", D)).toBe(90);
    expect(goldOf("good", "cannon", D)).toBe(54);
    expect(goldOf("early", "cannon", D)).toBe(27);
  });
});

describe("nextMinionType — the legacy weights", () => {
  it("draws cannon above 0.9, ranged above 0.5, melee below", () => {
    expect(nextMinionType(() => 0.95)).toBe("cannon");
    expect(nextMinionType(() => 0.6)).toBe("ranged");
    expect(nextMinionType(() => 0.3)).toBe("melee");
  });
});
