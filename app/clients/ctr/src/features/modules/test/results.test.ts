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
      // Differential law (server scorer's rebuildable form): mean of
      // pick-to-pick deltas [600−0, 800−600, 1000−800] = 1000/3.
      { key: "avg_decision", value: "333ms" },
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

  it("moving_target: std dev, hold rate, and legacy's loss-counter law", () => {
    // Two not-lost → lost transitions (40 and 35 after held samples); the
    // consecutive miss (45) mints nothing.
    const record = recordOf([
      { deviation: 10, hit: true },
      { deviation: 25, hit: true },
      { deviation: 40, hit: false },
      { deviation: 45, hit: false },
      { deviation: 28, hit: true },
      { deviation: 35, hit: false },
      { deviation: 8, hit: true },
    ]);
    expect(deriveLocalResults("moving_target", record, new EngineParams({}))).toEqual([
      { key: "samples", value: "7" },
      { key: "avg_deviation", value: "27.3px" },
      { key: "std_dev", value: "13.2px" },
      { key: "hold_rate", value: "57.1%" },
      { key: "loss_count", value: "2" },
    ]);
    // An empty stream (a never-touched run) stays zero across the board —
    // no fake loss, no NaN deviation.
    expect(deriveLocalResults("moving_target", recordOf([]), new EngineParams({}))).toEqual([
      { key: "samples", value: "0" },
      { key: "avg_deviation", value: "0.0px" },
      { key: "std_dev", value: "0.0px" },
      { key: "hold_rate", value: "0.0%" },
      { key: "loss_count", value: "0" },
    ]);
    // A run that opens already lost mints nothing (legacy inits the
    // tracking state to "lost").
    const opensLost = recordOf([
      { deviation: 40, hit: false },
      { deviation: 12, hit: true },
    ]);
    const rows = deriveLocalResults("moving_target", opensLost, new EngineParams({}));
    expect(rows[4]).toEqual({ key: "loss_count", value: "0" });
  });

  it("unknown codes fall back to the generic tap/hit pair", () => {
    const record = recordOf([{ isHit: true }, { isHit: false }]);
    expect(deriveLocalResults("no_such_game", record, new EngineParams({}))).toEqual([
      { key: "taps", value: "2" },
      { key: "hits", value: "1" },
    ]);
  });
});
