import { describe, expect, it } from "vitest";
import { CC_ROSTER, COUNTER_RELATIONS, counterScore, drawEnemyLineup, hitPairs } from "./logic";

describe("the relation mirror", () => {
  it("carries twelve rows over the roster's roles", () => {
    expect(COUNTER_RELATIONS).toHaveLength(12);
    for (const relation of COUNTER_RELATIONS) {
      expect(CC_ROSTER).toContain(relation.counter);
      expect(CC_ROSTER).toContain(relation.countered);
      expect(relation.score).toBe(20);
    }
  });

  it("keeps every role named on at least one side", () => {
    const named = new Set<string>();
    for (const relation of COUNTER_RELATIONS) {
      named.add(relation.counter);
      named.add(relation.countered);
    }
    for (const role of CC_ROSTER) {
      expect(named.has(role)).toBe(true);
    }
  });
});

describe("drawEnemyLineup", () => {
  it("draws the lineup size as distinct roster roles", () => {
    for (let run = 0; run < 50; run += 1) {
      const lineup = drawEnemyLineup(5, Math.random);
      expect(lineup).toHaveLength(5);
      expect(new Set(lineup).size).toBe(5);
      for (const role of lineup) {
        expect(CC_ROSTER).toContain(role);
      }
    }
  });
});

describe("hitPairs and counterScore", () => {
  it("pairs every selected counter against a countered enemy", () => {
    // Tank counters Fighter and Assassin; Support counters Assassin only
    // among these enemies — three pairs, sixty points.
    const hits = hitPairs(["Tank", "Support"], ["Fighter", "Assassin"]);
    expect(counterScore(hits)).toBe(60);
    expect(hits).toHaveLength(3);
  });

  it("scores zero when nothing counters the enemy", () => {
    expect(counterScore(hitPairs(["Support"], ["Marksman"]))).toBe(0);
  });

  it("mirrors the legacy both-way pairs (Fighter and Marksman)", () => {
    expect(counterScore(hitPairs(["Fighter"], ["Marksman"]))).toBe(20);
    expect(counterScore(hitPairs(["Marksman"], ["Fighter"]))).toBe(20);
  });
});
