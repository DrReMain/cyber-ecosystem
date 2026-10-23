// The submit seam's pure half — one attempt record mapped onto the unified
// attempt contract. Facts up, verdicts down: only event facts, the three
// contract counters, background spans, and the wall anchors cross the wire;
// every aggregate is recomputed server-side. Games stay untouched — they
// already emit gen-shaped camelCase events, and this adapter is the single
// place that knows how those records become the oneof raw_detail.
import { create, type MessageInitShape } from "@bufbuild/protobuf";
import { timestampFromMs } from "@bufbuild/protobuf/wkt";
import {
  type SubmitTestAttemptRequest,
  SubmitTestAttemptRequestSchema,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test_pb";
import type { AttemptRecord } from "./contract";

export interface SubmitAttemptInput {
  readonly code: string;
  readonly record: AttemptRecord;
  /** Present binds the attempt to a match session; absent = local practice.
   * The session flow feeds this when it lands — the seam is shaped for it. */
  readonly matchSessionId?: string;
}

type RawDetailInit = MessageInitShape<typeof SubmitTestAttemptRequestSchema>["rawDetail"];

type RedDotInit = Extract<NonNullable<RawDetailInit>, { case: "redDot" }>;
type FlashAimingInit = Extract<NonNullable<RawDetailInit>, { case: "flashAiming" }>;
type MultiTargetInit = Extract<NonNullable<RawDetailInit>, { case: "multiTarget" }>;
type TargetRecognitionInit = Extract<NonNullable<RawDetailInit>, { case: "targetRecognition" }>;

// Wire-stability constant: the ordinal is server law (session-scoped count
// plus one); the field rides at 1 until the client submitter drops it.
const PRACTICE_ATTEMPT_NO = 1;

function num(value: unknown): number {
  return typeof value === "number" && Number.isFinite(value) ? value : 0;
}

function intMs(value: unknown): number {
  return Math.max(0, Math.round(num(value)));
}

function point(value: unknown): { x: number; y: number } {
  const raw = (value ?? {}) as { x?: unknown; y?: unknown };
  return { x: num(raw.x), y: num(raw.y) };
}

function stringId(value: unknown): string {
  return typeof value === "string" ? value : "";
}

// Each builder mirrors one frozen Detail message field-for-field — the
// mapping is the review anchor against the contract, pinned by tests.
function redDotDetail(record: AttemptRecord): RedDotInit {
  return {
    case: "redDot",
    value: {
      events: record.events.map((event) => ({
        target: point(event.target),
        click: point(event.click),
        reactionTimeMs: intMs(event.reactionTimeMs),
        deviation: num(event.deviation),
        isHit: event.isHit === true,
      })),
    },
  };
}

function flashAimingDetail(record: AttemptRecord): FlashAimingInit {
  return {
    case: "flashAiming",
    value: {
      shots: record.events.map((event) => ({
        targetId: stringId(event.targetId),
        target: point(event.target),
        targetSize: num(event.targetSize),
        click: point(event.click),
        deviation: num(event.deviation),
        aimTimeMs: intMs(event.aimTimeMs),
      })),
      misses: Math.max(0, Math.round(record.stats.misses ?? 0)),
    },
  };
}

function multiTargetDetail(record: AttemptRecord): MultiTargetInit {
  return {
    case: "multiTarget",
    value: {
      picks: record.events.map((event) => ({
        targetId: stringId(event.targetId),
        tier: stringId(event.tier),
        isHit: event.isHit === true,
        elapsedMs: intMs(event.elapsedMs),
      })),
      missedHigh: Math.max(0, Math.round(record.stats.missedHigh ?? 0)),
    },
  };
}

function targetRecognitionDetail(record: AttemptRecord): TargetRecognitionInit {
  return {
    case: "targetRecognition",
    value: {
      clicks: record.events.map((event) => ({
        targetId: stringId(event.targetId),
        targetType: stringId(event.targetType),
        reactionTimeMs: intMs(event.reactionTimeMs),
        elapsedMs: intMs(event.elapsedMs),
      })),
      // The engine stats one spawned counter per target type; the contract
      // carries only the enemy denominator (the hit-rate denominator).
      spawnedEnemy: Math.max(0, Math.round(record.stats["spawned_enemy"] ?? 0)),
    },
  };
}

const DETAIL_BUILDERS: Record<string, (record: AttemptRecord) => RawDetailInit> = {
  red_dot: redDotDetail,
  flash_aiming: flashAimingDetail,
  multi_target: multiTargetDetail,
  target_recognition: targetRecognitionDetail,
};

/**
 * Build the submit request from one finished attempt, or null when the
 * attempt must not cross the seam: a code whose detail mapping has not
 * landed (later engine batches), a non-completed exit (an aborted run
 * records nothing — the exit-confirm copy promises exactly that), or
 * missing wall anchors (defensive: the host sets both on every natural
 * end).
 */
export function buildSubmitAttempt({
  code,
  record,
  matchSessionId,
}: Readonly<SubmitAttemptInput>): SubmitTestAttemptRequest | null {
  const builder = DETAIL_BUILDERS[code];
  if (builder === undefined || record.exit !== "completed") {
    return null;
  }
  if (record.startedAtMs === null || record.endedAtMs === null) {
    return null;
  }
  return create(SubmitTestAttemptRequestSchema, {
    testCode: code,
    attemptNo: PRACTICE_ATTEMPT_NO,
    ...(matchSessionId !== undefined ? { matchSessionId } : {}),
    startedAt: timestampFromMs(record.startedAtMs),
    endedAt: timestampFromMs(record.endedAtMs),
    backgrounds: record.backgroundSpans.map((span) => ({
      elapsedMs: intMs(span.elapsedMs),
      durationMs: intMs(span.durationMs),
    })),
    rawDetail: builder(record),
  });
}
