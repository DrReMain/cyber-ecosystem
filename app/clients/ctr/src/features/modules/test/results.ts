// Local result derivation — the local result face's only math. Facts and
// local aggregates only: no authoritative score, no grade, no rating —
// those are the server's dual-track verdicts. Rows are per-game,
// derived from the attempt record's event stream plus the game-local
// stats the fact stream cannot express; formatting is decided here so the
// component stays markup-only and the derivations stay pinnable.
import type { AttemptRecord, EngineParams } from "./contract";
import { hitPairs } from "./games/character-counter/logic";
import { FLASH_AIMING_DEFAULTS } from "./games/flash-aiming/logic";
import { TIER_SPECS } from "./games/multi-target/logic";
import { ascending } from "./games/number-stars/logic";
import { OFFENSE_SCENARIOS } from "./games/offense-defense/scenarios";

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
  // Pick-to-pick switching cost, mirroring the server scorer's rebuildable
  // form: the frozen pick carries no spawn stamp, so the differential of
  // consecutive run-relative elapsed values (first against zero) is the
  // decision interval. The raw-elapsed mean this row used to show was a
  // run-position average, not a decision time.
  let prev = 0;
  const deltas: number[] = [];
  for (const pick of picks) {
    const elapsed = num(pick.elapsedMs);
    deltas.push(elapsed - prev);
    prev = elapsed;
  }
  return [
    { key: "taps", value: String(picks.length) },
    { key: "score", value: String(score) },
    { key: "high_hits", value: String(countByTier.get("high") ?? 0) },
    { key: "trap_hits", value: String(countByTier.get("trap") ?? 0) },
    { key: "missed_high", value: String(record.stats.missedHigh ?? 0) },
    { key: "incorrect_clicks", value: String(record.stats.incorrectClicks ?? 0) },
    { key: "avg_decision", value: ms(mean(deltas)) },
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
  const deviations = shots.map((event) => num(event.deviation));
  const avg = mean(deviations);
  // Population std dev — legacy's deviationStdDev, the stability input.
  const stdDev = deviations.length
    ? Math.sqrt(mean(deviations.map((d) => (d - avg) ** 2)))
    : 0;
  // Legacy's loss counter (its lastTrackingState law): the state starts
  // "lost", and only a not-lost → lost transition counts — a run that
  // begins off-target mints nothing, and staying lost mints once.
  let losses = 0;
  let lost = true;
  for (const event of shots) {
    if (event.hit === true) {
      lost = false;
    } else if (!lost) {
      losses += 1;
      lost = true;
    }
  }
  return [
    { key: "samples", value: String(shots.length) },
    { key: "avg_deviation", value: `${avg.toFixed(1)}px` },
    { key: "std_dev", value: `${stdDev.toFixed(1)}px` },
    { key: "hold_rate", value: pct(held, shots.length) },
    { key: "loss_count", value: String(losses) },
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

function deriveOffenseDefense(record: AttemptRecord): readonly ResultRow[] {
  const decisions = record.events;
  // Correctness here reads the bank mirror for the local face only —
  // the server re-judges every pick against the seeded key.
  const optimalByCode = new Map(OFFENSE_SCENARIOS.map((s) => [s.code, s.optimal] as const));
  const correct = decisions.filter((decision) => {
    const optimal = optimalByCode.get(decision.scenarioId as string);
    return optimal !== undefined && decision.choice === optimal;
  }).length;
  const timeouts = decisions.filter((decision) => decision.timeout === true).length;
  return [
    { key: "rounds", value: String(decisions.length) },
    { key: "correct", value: String(correct) },
    { key: "timeouts", value: String(timeouts) },
    {
      key: "avg_decision",
      value: ms(mean(decisions.map((decision) => num(decision.decisionTimeMs)))),
    },
  ];
}

function deriveResourceMatch(record: AttemptRecord): readonly ResultRow[] {
  const byType = (type: string) => record.events.filter((event) => event.type === type);
  const collects = byType("collect");
  const collisions = byType("collision");
  const countKind = (kind: string) => collects.filter((collect) => collect.kind === kind).length;
  // Survival reads the last fact's stamp across both streams: a death
  // ends on its fatal collision, a clean run coasts to its final event.
  const last = record.events.reduce((max, event) => Math.max(max, num(event.elapsedMs)), 0);
  return [
    { key: "collected", value: String(collects.length) },
    { key: "minions", value: String(countKind("minion")) },
    { key: "monsters", value: String(countKind("monster")) },
    { key: "dragons", value: String(countKind("dragon")) },
    { key: "collisions", value: String(collisions.length) },
    { key: "survived", value: `${(last / 1000).toFixed(1)}s` },
  ];
}

// The same walk the server runs, on the local mirror of the stream:
// per round an ascending progress pointer decides each click, and the
// decision intervals ride the correct taps within the round only.
interface NsRoundScore {
  readonly correct: number;
  readonly deltas: number[];
  readonly done: boolean;
}

function scoreOneNsRound(
  numbers: number[],
  clicks: readonly Record<string, unknown>[],
): NsRoundScore {
  const expected = ascending(numbers);
  const deltas: number[] = [];
  let progress = 0;
  let prevCorrectMs: number | null = null;
  for (const click of clicks) {
    if (click.number === expected[progress]) {
      const elapsed = num(click.elapsedMs);
      if (prevCorrectMs !== null) {
        deltas.push(elapsed - prevCorrectMs);
      }
      prevCorrectMs = elapsed;
      progress += 1;
    }
  }
  return { correct: progress, deltas, done: progress >= expected.length };
}

function deriveNumberStars(record: AttemptRecord): readonly ResultRow[] {
  const groups: { numbers: number[]; clicks: Record<string, unknown>[] }[] = [];
  for (const event of record.events) {
    if (event.type === "round") {
      groups.push({ numbers: event.numbers as number[], clicks: [] });
    } else if (event.type === "click" && groups.length > 0) {
      groups[groups.length - 1]?.clicks.push(event);
    }
  }
  let correct = 0;
  let completed = 0;
  let clicks = 0;
  const deltas: number[] = [];
  for (const group of groups) {
    const score = scoreOneNsRound(group.numbers, group.clicks);
    correct += score.correct;
    clicks += group.clicks.length;
    deltas.push(...score.deltas);
    if (score.done) {
      completed += 1;
    }
  }
  return [
    { key: "rounds", value: String(completed) },
    { key: "taps", value: String(clicks) },
    { key: "correct", value: String(correct) },
    { key: "incorrect", value: String(clicks - correct) },
    { key: "avg_interval", value: ms(mean(deltas)) },
  ];
}

function deriveCharacterCounter(record: AttemptRecord): readonly ResultRow[] {
  const rounds = record.events;
  let pairs = 0;
  for (const round of rounds) {
    pairs += hitPairs(round.selectedLineup as string[], round.enemyLineup as string[]).length;
  }
  return [
    { key: "rounds", value: String(rounds.length) },
    { key: "counter_hits", value: String(pairs) },
    { key: "avg_time", value: ms(mean(rounds.map((round) => num(round.timeSpentMs)))) },
  ];
}

function deriveEconomicSim(record: AttemptRecord): readonly ResultRow[] {
  const hits = record.events.filter((event) => event.type === "hit");
  const cs = hits.filter((hit) => hit.isLastHit === true).length;
  const purchases = record.events.filter((event) => event.type === "purchase");
  const goldTotal = hits.reduce((sum, hit) => sum + num(hit.goldEarned), 0);
  return [
    { key: "cs", value: String(cs) },
    { key: "perfect", value: String(record.stats.perfect_cs ?? 0) },
    { key: "gold", value: String(goldTotal) },
    { key: "items", value: String(purchases.length) },
    { key: "final_health", value: String(record.stats.final_health_hp ?? 0) },
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
    case "offense_defense":
      return deriveOffenseDefense(record);
    case "resource_match":
      return deriveResourceMatch(record);
    case "number_stars":
      return deriveNumberStars(record);
    case "character_counter":
      return deriveCharacterCounter(record);
    case "economic_sim":
      return deriveEconomicSim(record);
    default: {
      const hits = record.events.filter((event) => event.isHit === true).length;
      return [
        { key: "taps", value: String(record.events.length) },
        { key: "hits", value: String(hits) },
      ];
    }
  }
}
