import { describe, expect, it } from "vitest";
import { pendingDetailsOf, type WireDetail } from "./end-face";

// The curve-pending read-back's runtime data shape, verbatim off a live
// zero-interaction number_stars run: proto3 implicit presence omits a
// zero-valued raw score from the parsed message entirely (the wire JSON
// carries rawScore:0; the runtime object does not), while non-default
// fields travel. The consumer must read a missing raw as the zero it is.
const RUNTIME_ZERO_RAW: WireDetail = {
  ability: "reactionSpeed",
  gameCode: "number_stars",
  attemptNo: 1,
  at: "2026-11-02T07:49:35.041320Z",
  rawScore: undefined,
  authoritativeScore: undefined,
  grade: undefined,
} as WireDetail;

const RUNTIME_SCORED: WireDetail = {
  gameCode: "flash_aiming",
  attemptNo: 2,
  rawScore: 220,
  authoritativeScore: undefined,
  grade: undefined,
};

describe("the curve-pending read-back data shape", () => {
  it("a zero raw score reads as zero, never NaN — the implicit-presence law", () => {
    const pending = pendingDetailsOf([RUNTIME_ZERO_RAW], "number_stars");
    expect(pending).toHaveLength(1);
    expect(Math.max(...pending.map((d) => d.rawScore ?? 0))).toBe(0);
  });

  it("the filter narrows to the visited game and the max rides the raw track", () => {
    const all = [RUNTIME_ZERO_RAW, RUNTIME_SCORED];
    expect(pendingDetailsOf(all, "flash_aiming")).toHaveLength(1);
    expect(Math.max(...pendingDetailsOf(all, "flash_aiming").map((d) => d.rawScore ?? 0))).toBe(220);
    expect(pendingDetailsOf(all, "other_game")).toHaveLength(0);
  });
});
