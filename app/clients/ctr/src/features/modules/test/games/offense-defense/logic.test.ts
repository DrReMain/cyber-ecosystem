import { describe, expect, it } from "vitest";
import { pickScenarios } from "./logic";
import { OFFENSE_SCENARIOS } from "./scenarios";

describe("pickScenarios", () => {
  it("draws a fixed lineup for a fixed rand", () => {
    let state = 42;
    const rand = () => {
      state = (state * 1103515245 + 12345) % 2147483648;
      return state / 2147483648;
    };
    const first = pickScenarios(OFFENSE_SCENARIOS, 10, rand);
    state = 42;
    const second = pickScenarios(OFFENSE_SCENARIOS, 10, rand);
    expect(first.map((s) => s.code)).toEqual(second.map((s) => s.code));
  });

  it("never repeats a scenario inside one lineup", () => {
    for (let run = 0; run < 20; run += 1) {
      const picked = pickScenarios(OFFENSE_SCENARIOS, 10, Math.random);
      const codes = new Set(picked.map((s) => s.code));
      expect(codes.size).toBe(picked.length);
    }
  });

  it("clamps an oversized ask to the whole bank", () => {
    const picked = pickScenarios(OFFENSE_SCENARIOS, 99, Math.random);
    expect(picked.length).toBe(OFFENSE_SCENARIOS.length);
    expect(new Set(picked.map((s) => s.code)).size).toBe(OFFENSE_SCENARIOS.length);
  });
});

describe("scenario bank mirror", () => {
  it("carries three distinct stances and one optimal per scenario", () => {
    for (const scenario of OFFENSE_SCENARIOS) {
      const choices = scenario.choices.map((entry) => entry.choice);
      expect(new Set(choices).size).toBe(3);
      expect(choices).toContain(scenario.optimal);
    }
  });
});
