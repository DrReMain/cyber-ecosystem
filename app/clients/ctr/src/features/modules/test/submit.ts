import { create } from "@bufbuild/protobuf";
import { timestampFromDate } from "@bufbuild/protobuf/wkt";
import type {
  DualZoneDetail,
  FlashAimingDetail,
  MovingTargetDetail,
  MultiTargetDetail,
  OffenseDefenseDetail,
  RedDotDetail,
  ResourceMatchDetail,
  SkillTrajectoryDetail,
  SubmitTestAttemptRequest,
  TargetRecognitionDetail,
  TrackingMovingDetail,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test_pb";
import { SubmitTestAttemptRequestSchema } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test_pb";
import type { AttemptRecord } from "./contract";

// The submit adapter: one run record → one SubmitTestAttemptRequest. The
// engine's event stream already carries proto-Detail item shapes, so the
// per-game table only names the wrapping oneof arm and folds the
// contract's top-level counters out of the game-local stats the event
// stream cannot express. A code without an entry here has no engine and
// never reaches a run.
type DetailBuilder = (record: AttemptRecord) => SubmitTestAttemptRequest["rawDetail"];

const DETAIL_BUILDERS: Readonly<Record<string, DetailBuilder>> = {
  red_dot: (r) => ({
    case: "redDot",
    value: { events: r.events } as RedDotDetail,
  }),
  flash_aiming: (r) => ({
    case: "flashAiming",
    value: { shots: r.events, misses: r.stats.misses ?? 0 } as FlashAimingDetail,
  }),
  moving_target: (r) => ({
    case: "movingTarget",
    value: { shots: r.events } as MovingTargetDetail,
  }),
  multi_target: (r) => ({
    case: "multiTarget",
    value: { picks: r.events, missedHigh: r.stats.missedHigh ?? 0 } as MultiTargetDetail,
  }),
  target_recognition: (r) => ({
    case: "targetRecognition",
    value: {
      clicks: r.events,
      spawnedEnemy: r.stats.spawned_enemy ?? 0,
    } as TargetRecognitionDetail,
  }),
  tracking_moving: (r) => ({
    case: "trackingMoving",
    value: { frames: r.events } as TrackingMovingDetail,
  }),
  skill_trajectory: (r) => ({
    case: "skillTrajectory",
    // int64 rides the wire as bigint; the game banks the board seed in
    // its stats so the server can reproduce boards from it.
    value: {
      randomSeed: BigInt(Math.trunc(r.stats.random_seed ?? 0)),
      rounds: r.events,
    } as SkillTrajectoryDetail,
  }),
  dual_zone: (r) => ({
    case: "dualZone",
    value: { behaviors: r.events } as DualZoneDetail,
  }),
  offense_defense: (r) => ({
    case: "offenseDefense",
    // A timeout decision carries no choice on the wire — absence is the
    // contract's "unanswered", never an empty string.
    value: { decisions: r.events } as OffenseDefenseDetail,
  }),
  resource_match: (r) => {
    // The engine tags its mixed stream; the three detail arrays are the
    // partition of that stream by tag. Items stay plain records — the
    // request's create() walk converts them into messages, the same
    // deep conversion every game's plain event shapes already ride.
    const trajectory: Record<string, unknown>[] = [];
    const collects: Record<string, unknown>[] = [];
    const collisions: Record<string, unknown>[] = [];
    for (const event of r.events) {
      if (event.type === "collect") {
        collects.push({ kind: event.kind as string, elapsedMs: event.elapsedMs as number });
      } else if (event.type === "collision") {
        collisions.push({ elapsedMs: event.elapsedMs as number });
      } else if (event.type === "trajectory") {
        trajectory.push({
          position: { x: event.x as number, y: event.y as number },
          elapsedMs: event.elapsedMs as number,
        });
      }
    }
    return {
      case: "resourceMatch",
      value: { trajectory, collects, collisions } as ResourceMatchDetail,
    };
  },
};

export interface SubmitInput {
  readonly code: string;
  readonly record: AttemptRecord;
  readonly matchSessionId: string | undefined;
}

export function buildSubmitRequest(input: SubmitInput): SubmitTestAttemptRequest | null {
  const build = DETAIL_BUILDERS[input.code];
  if (build === undefined) {
    return null;
  }
  // A completed run always carries both wall anchors; the fallback only
  // guards the impossible unstarted shape.
  const endedAtMs = input.record.endedAtMs ?? Date.now();
  const startedAtMs = input.record.startedAtMs ?? endedAtMs;
  return create(SubmitTestAttemptRequestSchema, {
    testCode: input.code,
    ...(input.matchSessionId !== undefined ? { matchSessionId: input.matchSessionId } : {}),
    startedAt: timestampFromDate(new Date(startedAtMs)),
    endedAt: timestampFromDate(new Date(endedAtMs)),
    backgrounds: input.record.backgroundSpans.map((span) => ({
      elapsedMs: span.elapsedMs,
      durationMs: span.durationMs,
    })),
    rawDetail: build(input.record),
  });
}

// The server verdict the end face renders: dual-track score and grade on a
// valid attempt, the ops-vocabulary reason on an invalid one, and the
// session's retake receipt — zero reads as "budget spent", and the whole
// retest face hides on practice attempts (no session binding).
export type SubmitVerdict =
  | { phase: "idle" }
  | { phase: "submitting" }
  | { phase: "error" }
  | {
      phase: "done";
      valid: boolean;
      invalidReason: string | undefined;
      authoritativeScore: number | undefined;
      grade: string | undefined;
      retestLeft: number;
      inSession: boolean;
    };
