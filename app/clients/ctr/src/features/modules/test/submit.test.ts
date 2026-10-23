import { describe, expect, it } from "vitest";
import type { AttemptRecord } from "./contract";
import { buildSubmitAttempt } from "./submit";

// Records are hand-built in the exact shape the four engines emit (the
// games' collect calls), so these tests pin the adapter against both sides:
// the engine fact shapes and the frozen Detail contract fields.
function record(overrides: Partial<AttemptRecord> = {}): AttemptRecord {
  return {
    events: [],
    backgroundSpans: [{ elapsedMs: 1000, durationMs: 500 }],
    stats: {},
    exit: "completed",
    startedAtMs: 1_728_000_000_000,
    endedAtMs: 1_728_000_030_000,
    ...overrides,
  };
}

describe("buildSubmitAttempt", () => {
  it("maps a red_dot record onto the contract field-for-field", () => {
    const req = buildSubmitAttempt({
      code: "red_dot",
      record: record({
        events: [
          {
            target: { x: 10.2, y: 20 },
            click: { x: 11, y: 21.8 },
            reactionTimeMs: 312,
            deviation: 1.41,
            isHit: true,
          },
        ],
      }),
    });
    expect(req).not.toBeNull();
    expect(req?.testCode).toBe("red_dot");
    // Wire-stability constant — the server owns the true ordinal.
    expect(req?.attemptNo).toBe(1);
    expect(req?.matchSessionId).toBeUndefined();
    expect(req?.startedAt?.seconds).toBe(1_728_000_000n);
    expect(req?.endedAt?.nanos).toBe(0);
    // create() brands nested messages ($typeName) — compare plain fields.
    expect(
      req?.backgrounds.map((span) => ({ elapsedMs: span.elapsedMs, durationMs: span.durationMs })),
    ).toEqual([{ elapsedMs: 1000, durationMs: 500 }]);
    expect(req?.rawDetail.case).toBe("redDot");
    if (req?.rawDetail.case === "redDot") {
      expect(req.rawDetail.value.events.length).toBe(1);
      expect(req.rawDetail.value.events[0]?.target?.x).toBe(10.2);
      expect(req.rawDetail.value.events[0]?.target?.y).toBe(20);
      expect(req.rawDetail.value.events[0]?.isHit).toBe(true);
      expect(req.rawDetail.value.events[0]?.reactionTimeMs).toBe(312);
    }
  });

  it("carries flash_aiming shots plus the misses counter the stream cannot rebuild", () => {
    const req = buildSubmitAttempt({
      code: "flash_aiming",
      record: record({
        events: [
          {
            targetId: "t1",
            target: { x: 1, y: 2 },
            targetSize: 44,
            click: { x: 3, y: 4 },
            deviation: 2.5,
            aimTimeMs: 401,
          },
        ],
        stats: { misses: 3 },
      }),
    });
    expect(req?.rawDetail.case).toBe("flashAiming");
    if (req?.rawDetail.case === "flashAiming") {
      expect(req.rawDetail.value.misses).toBe(3);
      expect(req.rawDetail.value.shots[0]?.targetId).toBe("t1");
      expect(req.rawDetail.value.shots[0]?.aimTimeMs).toBe(401);
    }
  });

  it("carries multi_target picks plus missed_high", () => {
    const req = buildSubmitAttempt({
      code: "multi_target",
      record: record({
        events: [{ targetId: "t2", tier: "high", isHit: true, elapsedMs: 900 }],
        stats: { missedHigh: 2, incorrectClicks: 7 },
      }),
    });
    expect(req?.rawDetail.case).toBe("multiTarget");
    if (req?.rawDetail.case === "multiTarget") {
      expect(req.rawDetail.value.missedHigh).toBe(2);
      // incorrect_clicks has no upstream carrier — it must not leak in.
      expect("incorrectClicks" in req.rawDetail.value).toBe(false);
      expect(req.rawDetail.value.picks[0]?.tier).toBe("high");
    }
  });

  it("carries target_recognition clicks plus the spawned_enemy denominator", () => {
    const req = buildSubmitAttempt({
      code: "target_recognition",
      record: record({
        events: [
          {
            targetId: "e1",
            targetType: "enemy",
            reactionTimeMs: 288,
            elapsedMs: 300,
          },
        ],
        stats: { spawned_enemy: 12, spawned_friendly: 5, blankClicks: 1 },
      }),
    });
    expect(req?.rawDetail.case).toBe("targetRecognition");
    if (req?.rawDetail.case === "targetRecognition") {
      expect(req.rawDetail.value.spawnedEnemy).toBe(12);
      expect(req.rawDetail.value.clicks[0]?.targetType).toBe("enemy");
      // Only the enemy denominator crosses; sibling counts stay local.
      expect("spawnedFriendly" in req.rawDetail.value).toBe(false);
    }
  });

  it("refuses codes without a landed detail mapping", () => {
    expect(buildSubmitAttempt({ code: "tracking_moving", record: record() })).toBeNull();
  });

  it("refuses non-completed exits — an aborted run records nothing", () => {
    expect(buildSubmitAttempt({ code: "red_dot", record: record({ exit: "aborted" }) })).toBeNull();
    expect(buildSubmitAttempt({ code: "red_dot", record: record({ exit: "failed" }) })).toBeNull();
  });

  it("refuses records missing wall anchors", () => {
    expect(
      buildSubmitAttempt({ code: "red_dot", record: record({ startedAtMs: null }) }),
    ).toBeNull();
    expect(buildSubmitAttempt({ code: "red_dot", record: record({ endedAtMs: null }) })).toBeNull();
  });

  it("binds the session id when present — the seam's E4 shape", () => {
    const req = buildSubmitAttempt({
      code: "red_dot",
      record: record(),
      matchSessionId: "sess-1",
    });
    expect(req?.matchSessionId).toBe("sess-1");
  });

  it("clamps negative and fractional millisecond facts at the boundary", () => {
    const req = buildSubmitAttempt({
      code: "red_dot",
      record: record({
        events: [{ reactionTimeMs: -5, deviation: Number.NaN }],
      }),
    });
    expect(req?.rawDetail.case).toBe("redDot");
    if (req?.rawDetail.case === "redDot") {
      expect(req.rawDetail.value.events[0]?.reactionTimeMs).toBe(0);
      expect(req.rawDetail.value.events[0]?.deviation).toBe(0);
    }
  });
});
