import { describe, expect, it } from "vitest";
import { type AttemptRecord, EngineParams } from "./contract";
import { deriveLocalResults } from "./results";

function recordOf(
  events: Record<string, unknown>[],
  stats: Record<string, number> = {},
  exit: AttemptRecord["exit"] = "completed",
): AttemptRecord {
  return { events, backgroundSpans: [], stats, startedAtMs: null, endedAtMs: null, exit };
}

describe("local result derivation", () => {
  it("red_dot: taps, hits, mean reaction", () => {
    const record = recordOf([
      { reactionTimeMs: 300, isHit: true },
      { reactionTimeMs: 500, isHit: false },
      { reactionTimeMs: 400, isHit: true },
    ]);
    expect(deriveLocalResults("red_dot", record, new EngineParams({}))).toEqual([
      { key: "taps", value: "3" },
      { key: "hits", value: "2" },
      { key: "avg_reaction", value: "400ms" },
    ]);
  });

  it("flash_aiming: center hits honor the engine-param threshold", () => {
    const events = [
      { deviation: 10, aimTimeMs: 500 },
      { deviation: 25, aimTimeMs: 700 },
      { deviation: 20, aimTimeMs: 300 },
    ];
    const defaults = deriveLocalResults(
      "flash_aiming",
      recordOf(events, { misses: 1 }),
      new EngineParams({}),
    );
    expect(defaults).toEqual([
      { key: "hits", value: "3" },
      { key: "misses", value: "1" },
      { key: "center_hits", value: "2" }, // 10 and 20 ≤ 20
      { key: "hit_rate", value: "75.0%" }, // 3 / (3+1)
      { key: "avg_aim", value: "500ms" },
    ]);
    const tightened = deriveLocalResults(
      "flash_aiming",
      recordOf(events, { misses: 1 }),
      new EngineParams({ center_hit_deviation_px: "5" }),
    );
    expect(tightened[2]).toEqual({ key: "center_hits", value: "0" });
  });

  it("multi_target: tier counts, value sum, stats passthrough", () => {
    const record = recordOf(
      [
        { tier: "high", elapsedMs: 600 },
        { tier: "trap", elapsedMs: 800 },
        { tier: "high", elapsedMs: 1000 },
      ],
      { missedHigh: 2, incorrectClicks: 4 },
    );
    expect(deriveLocalResults("multi_target", record, new EngineParams({}))).toEqual([
      { key: "taps", value: "3" },
      { key: "score", value: "150" }, // 100 − 50 + 100
      { key: "high_hits", value: "2" },
      { key: "trap_hits", value: "1" },
      { key: "missed_high", value: "2" },
      { key: "incorrect_clicks", value: "4" },
      { key: "avg_decision", value: "800ms" },
    ]);
  });

  it("target_recognition: enemy rate against spawned, misclicks, blanks", () => {
    const record = recordOf(
      [
        { targetType: "enemy" },
        { targetType: "enemy" },
        { targetType: "friendly" },
        { targetType: "neutral" },
      ],
      { spawned_enemy: 5, spawned_friendly: 2, spawned_neutral: 1, blankClicks: 3 },
    );
    expect(deriveLocalResults("target_recognition", record, new EngineParams({}))).toEqual([
      { key: "enemy_hits", value: "2" },
      { key: "hit_rate", value: "40.0%" }, // 2 / 5
      { key: "enemy_missed", value: "3" },
      { key: "friendly_misclicks", value: "1" },
      { key: "neutral_misclicks", value: "1" },
      { key: "blank_clicks", value: "3" },
    ]);
  });

  it("unknown codes fall back to the generic tap/hit pair", () => {
    const record = recordOf([{ isHit: true }, { isHit: false }]);
    expect(deriveLocalResults("economic_sim", record, new EngineParams({}))).toEqual([
      { key: "taps", value: "2" },
      { key: "hits", value: "1" },
    ]);
  });
});
