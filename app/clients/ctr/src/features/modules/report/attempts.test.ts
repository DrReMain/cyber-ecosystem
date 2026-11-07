import { describe, expect, it } from "vitest";
import { elapsedSecondsOf, epochOf, type WireAttempt } from "./attempts";

// The session-done receipt's library seam: attempt stamps travel
// dual-shaped on the wire (Q4 law) and the play clock is the span the
// attempts themselves define — pure reads the page renders verbatim.

describe("epochOf", () => {
  it("reads an RFC 3339 string and a raw { seconds } object to the same epoch", () => {
    expect(epochOf("1970-01-01T00:00:01Z")).toBe(1000);
    expect(epochOf({ seconds: 1 })).toBe(1000);
  });

  it("degrades malformed shapes to undefined, never NaN", () => {
    expect(epochOf("not-a-stamp")).toBeUndefined();
    expect(epochOf({ seconds: "x" })).toBeUndefined();
    expect(epochOf(undefined)).toBeUndefined();
    expect(epochOf(42)).toBeUndefined();
  });
});

describe("elapsedSecondsOf", () => {
  it("spans first submit to last submit in submit order", () => {
    const attempts: WireAttempt[] = [
      { ability: "a", at: "2026-11-07T02:00:00Z" },
      { ability: "b", at: "2026-11-07T03:00:00Z" },
    ];
    expect(elapsedSecondsOf(attempts)).toBe(3600);
  });

  it("spans across mixed wire shapes", () => {
    expect(
      elapsedSecondsOf([
        { ability: "a", at: { seconds: 0 } },
        { ability: "b", at: { seconds: 90 } },
      ]),
    ).toBe(90);
  });

  it("reads a single attempt as a zero span and no attempts as absent", () => {
    expect(elapsedSecondsOf([{ ability: "a", at: "2026-11-07T02:00:00Z" }])).toBe(0);
    expect(elapsedSecondsOf([])).toBeUndefined();
  });

  it("stays undefined when either endpoint stamp is unreadable", () => {
    expect(
      elapsedSecondsOf([
        { ability: "a", at: "not-a-stamp" },
        { ability: "b", at: "2026-11-07T02:00:01Z" },
      ]),
    ).toBeUndefined();
  });
});
