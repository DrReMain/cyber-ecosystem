import { useNavigate } from "@tanstack/react-router";
import { ChevronLeft } from "lucide-react";
import { Button } from "react-aria-components";
import { abilityWord } from "#/features/modules/match/vocabulary";
import { m } from "#/paraglide/messages";
import { type AttemptRecord, EngineParams, type StepContext, type StepRef } from "../contract";
import { getGameModule } from "../registry";
import { deriveLocalResults } from "../results";
import type { SubmitVerdict } from "../submit";
import { ProgStrip } from "./prog-strip";

interface EndFaceProps {
  readonly code: string;
  readonly record: AttemptRecord;
  /** The attempt's boot params — the derivations re-read the same thresholds the game ran with. */
  readonly engineParams: Readonly<Record<string, string>>;
  readonly verdict: SubmitVerdict;
  /** The live session binding — null after a practice run. */
  readonly step: StepContext | null;
  readonly testName: string;
  readonly onResubmit: () => void;
  readonly onRetry: () => void;
  readonly onExit: () => void;
}

// Row key → label copy (every derivation key has its bilingual home here).
const ROW_LABELS: Record<string, () => string> = {
  taps: m.test_result_taps,
  hits: m.test_result_hits,
  avg_reaction: m.test_result_avg_reaction,
  misses: m.test_result_misses,
  center_hits: m.test_result_center_hits,
  hit_rate: m.test_result_hit_rate,
  avg_aim: m.test_result_avg_aim,
  score: m.test_result_score,
  high_hits: m.test_result_high_hits,
  trap_hits: m.test_result_trap_hits,
  missed_high: m.test_result_missed_high,
  incorrect_clicks: m.test_result_incorrect_clicks,
  avg_decision: m.test_result_avg_decision,
  enemy_hits: m.test_result_enemy_hits,
  enemy_missed: m.test_result_enemy_missed,
  friendly_misclicks: m.test_result_friendly_misclicks,
  neutral_misclicks: m.test_result_neutral_misclicks,
  blank_clicks: m.test_result_blank_clicks,
  samples: m.test_result_samples,
  avg_deviation: m.test_result_avg_deviation,
  std_dev: m.test_result_std_dev,
  hold_rate: m.test_result_hold_rate,
  loss_count: m.test_result_loss_count,
  shots: m.test_result_shots,
  movement: m.test_result_movement,
  avg_acquire: m.test_result_avg_acquire,
  rounds: m.test_result_rounds,
  matches: m.test_result_matches,
  avg_place: m.test_result_avg_place,
  dodges: m.test_result_dodges,
  survived: m.test_result_survived,
  correct: m.test_result_correct,
  timeouts: m.test_result_timeouts,
  collected: m.test_result_collected,
  minions: m.test_result_minions,
  monsters: m.test_result_monsters,
  dragons: m.test_result_dragons,
  collisions: m.test_result_collisions,
  incorrect: m.test_result_incorrect,
  avg_interval: m.test_result_avg_interval,
  counter_hits: m.test_result_counter_hits,
  avg_time: m.test_result_avg_time,
  cs: m.test_result_cs,
  perfect: m.test_result_perfect,
  gold: m.test_result_gold,
  items: m.test_result_items,
  final_health: m.test_result_final_health,
};

// Ops reason → user copy; the server's reason vocabulary is ops-facing, so
// the resolution lives client-side with a generic fallback.
const INVALID_REASON_COPY: Record<string, () => string> = {
  BACKGROUND_OVER_BUDGET: m.test_invalid_background,
  ABSURD_REACTION_TIME: m.test_invalid_absurd,
  ABSURD_TIME_SPAN: m.test_invalid_absurd,
  DETAIL_MISMATCH: m.test_invalid_detail,
};

function invalidCopy(reason: string | undefined): string {
  const copy = reason !== undefined ? INVALID_REASON_COPY[reason] : undefined;
  return copy?.() ?? m.test_verdict_invalid();
}

// The grade chip's four-color ladder (prototype .tier) and its grey off
// state — the server's curve-pending verdicts render as 未评级.
const TIERS: Record<string, { readonly cls: string; readonly cn: () => string }> = {
  S: { cls: "text-grade-s bg-[rgba(253,224,71,0.15)]", cn: m.test_grade_s },
  A: { cls: "text-grade-a bg-[rgba(34,197,94,0.15)]", cn: m.test_grade_a },
  B: { cls: "text-grade-b bg-[rgba(59,130,246,0.15)]", cn: m.test_grade_b },
  C: { cls: "text-grade-c bg-[rgba(249,115,22,0.15)]", cn: m.test_grade_c },
};

function TierChip({ grade }: Readonly<{ grade: string | undefined }>) {
  const tier = grade !== undefined ? TIERS[grade] : undefined;
  if (tier === undefined) {
    return (
      <span className="rounded-full bg-[rgba(107,114,128,0.15)] px-2.5 py-0.75 text-[11px] font-extrabold text-[#6b7280]">
        {m.test_grade_off()}
      </span>
    );
  }
  return (
    <span className={`rounded-full px-2.5 py-0.75 text-[11px] font-extrabold ${tier.cls}`}>
      {grade} · {tier.cn()}
    </span>
  );
}

// The retake receipt chip (prototype .st): grey while chances remain,
// amber when spent.
function RetakeChip({ left, max }: Readonly<{ left: number; max: number }>) {
  const spent = left <= 0;
  return (
    <span
      className={`rounded-full px-2.5 py-0.75 text-[10.5px] font-semibold ${
        spent ? "text-status-pending bg-status-pending-bg" : "text-ink-muted bg-status-lock-bg"
      }`}
    >
      {m.test_result_retake_chip({ left, max })}
    </span>
  );
}

// The continue CTA's target: the next undone engine-backed step. With
// the lineup finished it returns undefined and the CTA closes the
// session flow (the report page the prototype routes to is a later
// slice).
function findNextStep(step: StepContext): StepRef | undefined {
  return step.lineup.find(
    (s) => !step.doneCodes.includes(s.code) && getGameModule(s.code) !== null,
  );
}

// The score card: the ability line, then the submit chain's face — big
// score with grade chip, the pending pulse, the invalid card, the error
// card with its retry — and the retake receipt chip.
function VerdictCard({
  verdict,
  step,
  testName,
  onResubmit,
}: Readonly<{
  verdict: SubmitVerdict;
  step: StepContext | null;
  testName: string;
  onResubmit: () => void;
}>) {
  const word = step !== null ? abilityWord(step.ability) : null;
  const maxRetakes = step !== null ? (step.firstSession ? 3 : 1) : 0;
  const retakeLeft = verdict.phase === "done" ? verdict.retestLeft : (step?.retestLeft ?? 0);
  return (
    <div className="border-line bg-surface-card rounded-card border px-4 py-4.5 text-center">
      {word !== null ? (
        <div className="text-ink-tertiary text-[11px]">
          {m.test_result_ability_line({ ability: word.cn, name: testName })}
        </div>
      ) : null}
      {(verdict.phase === "submitting" || verdict.phase === "idle") && (
        <div className="text-ink-tertiary mt-3 animate-pulse text-sm">
          {m.test_verdict_pending()}
        </div>
      )}
      {verdict.phase === "done" && verdict.valid && (
        <>
          <div className="mt-1.5 flex items-baseline justify-center gap-1.5">
            <span className="text-ink text-[58px] leading-none font-black tabular-nums">
              {verdict.authoritativeScore ?? "—"}
            </span>
            <span className="text-ink-tertiary text-sm">{m.test_result_unit()}</span>
          </div>
          <div className="mt-2.5">
            <TierChip grade={verdict.grade} />
            {/* D-AO: the ungraded chip keeps its shape, with the honest
                note on why — the game's norms are still accumulating. */}
            {verdict.grade === undefined && (
              <div className="text-ink-tertiary mt-1.5 text-[10.5px]">
                {m.test_grade_off_note()}
              </div>
            )}
          </div>
        </>
      )}
      {verdict.phase === "done" && !verdict.valid && (
        <div className="text-status-pending rounded-control mt-3 border border-[rgba(245,158,11,0.45)] bg-transparent px-3.25 py-3 text-left text-[12px] leading-[1.85]">
          <span className="font-bold">{m.test_verdict_invalid()}</span>
          <span className="opacity-80"> · {invalidCopy(verdict.invalidReason)}</span>
        </div>
      )}
      {verdict.phase === "error" && (
        <div className="mt-3">
          <div className="text-status-pending text-sm font-bold">{m.test_submit_error()}</div>
          <button
            className="text-ink-tertiary cursor-pointer text-xs underline active:opacity-60"
            onClick={onResubmit}
            type="button"
          >
            {m.common_retry()}
          </button>
        </div>
      )}
      {verdict.phase === "done" && verdict.inSession ? (
        <div className="mt-3 flex flex-wrap justify-center gap-2">
          <RetakeChip left={retakeLeft} max={maxRetakes} />
        </div>
      ) : null}
    </div>
  );
}

/**
 * The result face (prototype test-result): the big score with its grade
 * chip and retake receipt, the fact rows as label/value lines, and the
 * two CTAs — the practice/orange retest and the primary continue, which
 * walks to the next undone engine-backed step or closes the session
 * flow. Markup-only; every number comes from results.ts or the submit
 * verdict.
 */
export function EndFace({
  code,
  record,
  engineParams,
  verdict,
  step,
  testName,
  onResubmit,
  onRetry,
  onExit,
}: Readonly<EndFaceProps>) {
  const navigate = useNavigate();
  const rows = deriveLocalResults(code, record, new EngineParams(engineParams));
  const spent = verdict.phase === "done" && verdict.inSession && verdict.retestLeft <= 0;
  const next = step !== null ? findNextStep(step) : undefined;

  return (
    <div className="bg-canvas absolute inset-0 isolate z-40 overflow-y-auto">
      <div
        aria-hidden
        className="pointer-events-none absolute -top-20 -left-17.5 size-[min(69vw,260px)] rounded-full bg-[#2563eb] opacity-26 blur-[70px]"
      />
      <div
        aria-hidden
        className="bg-ambient-cyan pointer-events-none absolute -right-17.5 -bottom-15 size-[min(59vw,220px)] rounded-full opacity-26 blur-[70px]"
      />
      <div className="relative mx-auto w-full max-w-md px-4 pb-6">
        <div className="flex items-center gap-3 pt-5 pb-2.5">
          <button
            aria-label={m.common_back()}
            className="bg-surface-input text-ink-secondary flex h-10 w-10 shrink-0 items-center justify-center rounded-full active:opacity-60"
            onClick={onExit}
            type="button"
          >
            <ChevronLeft aria-hidden size={20} />
          </button>
          <h1 className="text-ink min-w-0 flex-1 truncate text-[19px] font-bold">
            {m.test_result_title({ name: testName })}
          </h1>
          {step !== null ? (
            <span className="text-ink-tertiary shrink-0 text-[13px] tabular-nums">
              {step.index + 1} / {step.lineup.length}
            </span>
          ) : null}
        </div>
        {step !== null ? (
          <div className="mb-3.5">
            <ProgStrip currentIndex={step.index} doneCodes={step.doneCodes} lineup={step.lineup} />
          </div>
        ) : null}

        <div className="flex flex-col gap-3">
          <VerdictCard onResubmit={onResubmit} step={step} testName={testName} verdict={verdict} />

          <div className="border-line bg-surface-card rounded-card border p-4">
            <div className="text-ink mb-2.5 text-sm font-extrabold">
              {m.test_result_detail_title()}
            </div>
            {rows.map((row) => (
              <div
                className="flex items-center justify-between border-b border-[rgba(55,65,81,0.45)] py-2.25 last:border-b-0"
                key={row.key}
              >
                <span className="text-ink-muted text-[12.5px]">
                  {(ROW_LABELS[row.key] ?? m.test_result_taps)()}
                </span>
                <span className="text-ink text-sm font-extrabold tabular-nums">{row.value}</span>
              </div>
            ))}
          </div>

          <div className="flex flex-col gap-3">
            <Button
              className={`bg-cta-practice shadow-cta-practice rounded-control w-full cursor-pointer px-4 py-3.5 text-[15px] font-bold text-white active:opacity-80 ${
                spent ? "pointer-events-none opacity-45" : ""
              }`}
              onPress={onRetry}
            >
              {spent ? m.test_result_retry_disabled() : m.test_end_retry()}
            </Button>
            {next !== undefined ? (
              <Button
                className="bg-cta-primary shadow-cta-primary rounded-control w-full cursor-pointer px-4 py-3.5 text-[15px] font-bold text-white active:opacity-80"
                onPress={() => void navigate({ to: "/test/$code", params: { code: next.code } })}
              >
                {m.test_result_next({ ability: abilityWord(next.ability).cn })}
              </Button>
            ) : (
              <Button
                className="bg-cta-primary shadow-cta-primary rounded-control w-full cursor-pointer px-4 py-3.5 text-[15px] font-bold text-white active:opacity-80"
                onPress={onExit}
              >
                {m.test_result_finish()}
              </Button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
