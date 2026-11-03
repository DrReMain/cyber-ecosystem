import { describe, expect, it } from "vitest";
import {
  bestMapOf,
  latestScoredOf,
  type WireDigest,
  type WireDigestSession,
  wireDateOf,
  wireDigestOf,
} from "./digest";

// The digest's library seam: the transport flattens the WKT wrappers to
// scalars and the stamps to RFC 3339 strings — the shapes the UI actually
// consumes, the class typecheck alone cannot pin down.

const sessionOf = (over: Partial<WireDigestSession>): WireDigestSession => ({
  sessionId: "s",
  genre: "fps",
  finished: true,
  doneCount: 1,
  stepCount: 1,
  createdAt: undefined,
  anchorRecordId: undefined,
  ...over,
});

const digestOf = (over: Partial<WireDigest>): WireDigest => ({
  ...wireDigestOf(undefined),
  ...over,
});

describe("wireDigestOf", () => {
  it("narrows an absent read to the empty digest, never undefined fields", () => {
    const empty = wireDigestOf(undefined);
    expect(empty.sessionsTotal).toBe(0);
    expect(empty.abilities).toEqual([]);
    expect(empty.sessions).toEqual([]);
  });

  it("normalizes a wire shape whose empty repeated keys protojson dropped", () => {
    // A fresh member's digest: both lists empty on the server, and the
    // connect protocol's protojson omits the keys outright — a bare cast
    // would hand the UI undefined where a list was promised.
    const fresh = wireDigestOf({ sessionsTotal: 0, sessionsFps: 0, sessionsMoba: 0 });
    expect(fresh.abilities).toEqual([]);
    expect(fresh.sessions).toEqual([]);
    expect(latestScoredOf(fresh)).toBeUndefined();
  });

  it("narrows absent or partial verdict objects to the closed state", () => {
    // protojson drops an empty message field's key entirely, so a digest
    // whose verdict objects traveled keyless must read as closed — the
    // entry cards' tags never render off an undefined access.
    const dropped = wireDigestOf({ sessionsTotal: 2 });
    expect(dropped.reportAccess).toEqual({ allowed: false, reason: undefined });
    expect(dropped.trendAdvancedAccess).toEqual({ allowed: false, reason: undefined });
    const carried = wireDigestOf({
      reportAccess: { allowed: true, reason: "REPORT_ACCESS_REASON_MEMBERSHIP_ACTIVE" },
      trendAdvancedAccess: { allowed: false, reason: "REPORT_ACCESS_REASON_NO_ENTITLEMENT" },
    });
    expect(carried.reportAccess.allowed).toBe(true);
    expect(carried.trendAdvancedAccess).toEqual({
      allowed: false,
      reason: "REPORT_ACCESS_REASON_NO_ENTITLEMENT",
    });
  });
});

describe("latestScoredOf", () => {
  it("picks the newest scored line across genres by its RFC 3339 stamp", () => {
    const fps = sessionOf({
      sessionId: "fps-latest",
      genre: "fps",
      createdAt: "2026-10-27T02:00:00Z",
      anchorRecordId: "rec-fps",
    });
    const moba = sessionOf({
      sessionId: "moba-latest",
      genre: "moba",
      createdAt: "2026-10-28T02:00:00Z",
      anchorRecordId: "rec-moba",
    });
    expect(latestScoredOf(digestOf({ sessions: [fps, moba] }))?.sessionId).toBe("moba-latest");
  });

  it("never picks an anchor-less line, however new it is", () => {
    const newerUnscored = sessionOf({ createdAt: "2026-10-29T00:00:00Z" });
    const olderScored = sessionOf({
      sessionId: "scored",
      createdAt: "2026-10-20T00:00:00Z",
      anchorRecordId: "rec",
    });
    expect(latestScoredOf(digestOf({ sessions: [newerUnscored, olderScored] }))?.sessionId).toBe(
      "scored",
    );
  });

  it("returns undefined when nothing scored", () => {
    expect(latestScoredOf(digestOf({ sessions: [sessionOf({})] }))).toBeUndefined();
  });
});

describe("bestMapOf", () => {
  it("maps every measured ability to its flattened cross-session best", () => {
    const digest = digestOf({
      abilities: [
        {
          ability: "reactionSpeed",
          bestScore: 78,
          bestGrade: "B",
          bestPercentile: 62,
          sessionsTested: 2,
        },
        {
          ability: "aimingAccuracy",
          bestScore: 81,
          bestGrade: "A",
          bestPercentile: 88,
          sessionsTested: 1,
        },
      ],
    });
    expect(bestMapOf(digest).get("reactionSpeed")).toBe(78);
    expect(bestMapOf(digest).get("aimingAccuracy")).toBe(81);
    expect(bestMapOf(digest).get("unknown")).toBeUndefined();
  });
});

describe("wireDateOf", () => {
  it("renders an RFC 3339 stamp as a local date string", () => {
    expect(wireDateOf("2026-10-28T01:02:03Z")).toMatch(/^\d{4}-\d{2}-\d{2}$/);
  });
  it("degrades an absent or malformed stamp to the empty string", () => {
    expect(wireDateOf(undefined)).toBe("");
    expect(wireDateOf("not-a-stamp")).toBe("");
  });
});
