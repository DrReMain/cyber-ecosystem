// Local result derivation — the local result face's only math. Facts and
// local aggregates only: no authoritative score, no grade, no rating —
// those are the server's dual-track verdicts. Rows are per-game,
// derived from the attempt record's event stream plus the game-local
// stats the fact stream cannot express; formatting is decided here so the
// component stays markup-only and the derivations stay pinnable.
import type { AttemptRecord, EngineParams } from "./contract";
import { FLASH_AIMING_DEFAULTS } from "./games/flash-aiming/logic";
import { TIER_SPECS } from "./games/multi-target/logic";

export interface ResultRow {
  /** Copy-key suffix under test_result_ (e.g. "hits" → m.test_result_hits). */
  readonly key: string;
  /** Preformatted display value ("12", "63.4%", "412ms"). */
  readonly value: string;
}

function pct(part: number, whole: number): string {
  if (whole <= 0) {
    return "0.0%";
  }
  return `${((part / whole) * 100).toFixed(1)}%`;
}

function ms(value: number): string {
  return `${Math.round(value)}ms`;
}

function mean(values: readonly number[]): number {
  if (values.length === 0) {
    return 0;
  }
  return values.reduce((sum, value) => sum + value, 0) / values.length;
}

function num(value: unknown): number {
  return typeof value === "number" && Number.isFinite(value) ? value : 0;
}

function deriveRedDot(record: AttemptRecord): readonly ResultRow[] {
  const events = record.events;
  const hits = events.filter((event) => event.isHit === true).length;
  return [
    { key: "taps", value: String(events.length) },
    { key: "hits", value: String(hits) },
    { key: "avg_reaction", value: ms(mean(events.map((event) => num(event.reactionTimeMs)))) },
  ];
}

function deriveFlashAiming(record: AttemptRecord, params: EngineParams): readonly ResultRow[] {
  const threshold = params.num(
    "center_hit_deviation_px",
    FLASH_AIMING_DEFAULTS.centerHitDeviationPx,
  );
  const shots = record.events;
  const centerHits = shots.filter((event) => num(event.deviation) <= threshold).length;
  const misses = record.stats.misses ?? 0;
  return [
    { key: "hits", value: String(shots.length) },
    { key: "misses", value: String(misses) },
    { key: "center_hits", value: String(centerHits) },
    { key: "hit_rate", value: pct(shots.length, shots.length + misses) },
    { key: "avg_aim", value: ms(mean(shots.map((event) => num(event.aimTimeMs)))) },
  ];
}

function deriveMultiTarget(record: AttemptRecord): readonly ResultRow[] {
  const picks = record.events;
  const valueByTier = new Map<string, number>(
    TIER_SPECS.map((spec) => [spec.tier, spec.value] as const),
  );
  const countByTier = new Map<string, number>();
  let score = 0;
  for (const pick of picks) {
    const tier = typeof pick.tier === "string" ? pick.tier : "";
    countByTier.set(tier, (countByTier.get(tier) ?? 0) + 1);
    score += valueByTier.get(tier) ?? 0;
  }
  return [
    { key: "taps", value: String(picks.length) },
    { key: "score", value: String(score) },
    { key: "high_hits", value: String(countByTier.get("high") ?? 0) },
    { key: "trap_hits", value: String(countByTier.get("trap") ?? 0) },
    { key: "missed_high", value: String(record.stats.missedHigh ?? 0) },
    { key: "incorrect_clicks", value: String(record.stats.incorrectClicks ?? 0) },
    { key: "avg_decision", value: ms(mean(picks.map((pick) => num(pick.elapsedMs)))) },
  ];
}

function deriveTargetRecognition(record: AttemptRecord): readonly ResultRow[] {
  const clicks = record.events;
  const count = (type: string) => clicks.filter((click) => click.targetType === type).length;
  const enemyHits = count("enemy");
  const spawned = record.stats.spawned_enemy ?? 0;
  return [
    { key: "enemy_hits", value: String(enemyHits) },
    { key: "hit_rate", value: pct(enemyHits, spawned) },
    { key: "enemy_missed", value: String(Math.max(0, spawned - enemyHits)) },
    { key: "friendly_misclicks", value: String(count("friendly")) },
    { key: "neutral_misclicks", value: String(count("neutral")) },
    { key: "blank_clicks", value: String(record.stats.blankClicks ?? 0) },
  ];
}

function deriveMovingTarget(record: AttemptRecord): readonly ResultRow[] {
  const shots = record.events;
  const held = shots.filter((event) => event.hit === true).length;
  return [
    { key: "samples", value: String(shots.length) },
    {
      key: "avg_deviation",
      value: `${mean(shots.map((event) => num(event.deviation))).toFixed(1)}px`,
    },
    { key: "hold_rate", value: pct(held, shots.length) },
  ];
}

function deriveSkillTrajectory(record: AttemptRecord): readonly ResultRow[] {
  const rounds = record.events;
  const matched = rounds.filter((round) => round.playerPlacement !== undefined);
  return [
    { key: "rounds", value: String(rounds.length) },
    { key: "matches", value: String(matched.length) },
    {
      key: "avg_deviation",
      value: `${mean(matched.map((round) => num(round.deviation))).toFixed(1)}px`,
    },
    { key: "avg_place", value: ms(mean(matched.map((round) => num(round.placementTimeMs)))) },
  ];
}

function deriveDualZone(record: AttemptRecord): readonly ResultRow[] {
  const behaviors = record.events;
  const dodges = behaviors.filter((behavior) => behavior.event === "dodge").length;
  // Survival reads the last fact's stamp: a collision run ends on its
  // fatal event, a clean run coasts to its final dodge.
  const last = behaviors.reduce((max, behavior) => Math.max(max, num(behavior.elapsedMs)), 0);
  return [
    { key: "dodges", value: String(dodges) },
    { key: "survived", value: `${(last / 1000).toFixed(1)}s` },
  ];
}

function deriveTrackingMoving(record: AttemptRecord): readonly ResultRow[] {
  const frames = record.events;
  const shotFrames = frames.filter((event) => event.hit !== undefined);
  const hits = shotFrames.filter((event) => event.hit === true).length;
  // The polyline over every frame's player point: a shot frame repeats
  // the crosshair it fired from, so consecutive duplicates add nothing.
  let movement = 0;
  let prev: { x: number; y: number } | null = null;
  for (const frame of frames) {
    const player = frame.player as { x: number; y: number } | undefined;
    if (player !== undefined && prev !== null) {
      movement += Math.hypot(player.x - prev.x, player.y - prev.y);
    }
    prev = player ?? prev;
  }
  const lifetimes = shotFrames
    .filter((event) => event.hit === true)
    .map((event) => num(event.dotLifetimeMs));
  return [
    { key: "shots", value: String(shotFrames.length) },
    { key: "hits", value: String(hits) },
    { key: "hit_rate", value: pct(hits, shotFrames.length) },
    { key: "movement", value: `${Math.round(movement)}px` },
    { key: "avg_acquire", value: ms(mean(lifetimes)) },
  ];
}

/** Per-game local rows; unknown codes fall back to the generic tap/hit pair. */
export function deriveLocalResults(
  code: string,
  record: AttemptRecord,
  params: EngineParams,
): readonly ResultRow[] {
  switch (code) {
    case "red_dot":
      return deriveRedDot(record);
    case "flash_aiming":
      return deriveFlashAiming(record, params);
    case "moving_target":
      return deriveMovingTarget(record);
    case "multi_target":
      return deriveMultiTarget(record);
    case "target_recognition":
      return deriveTargetRecognition(record);
    case "tracking_moving":
      return deriveTrackingMoving(record);
    case "skill_trajectory":
      return deriveSkillTrajectory(record);
    case "dual_zone":
      return deriveDualZone(record);
    default: {
      const hits = record.events.filter((event) => event.isHit === true).length;
      return [
        { key: "taps", value: String(record.events.length) },
        { key: "hits", value: String(hits) },
      ];
    }
  }
}
