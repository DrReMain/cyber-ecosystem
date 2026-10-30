import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import { getSessionReport } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/report-ReportService_connectquery";
import { useNavigate } from "@tanstack/react-router";
import { ChevronLeft } from "lucide-react";
import { Button } from "react-aria-components";
import { abilityWord } from "#/features/modules/match/vocabulary";
import { m } from "#/paraglide/messages";
import { type AttemptRecord, EngineParams, type StepContext, type StepRef } from "../contract";
import { getGameModule } from "../registry";
import { deriveLocalResults } from "../results";
import type { SubmitVerdict } from "../submit";
import { InvalidNotice } from "./invalid-notice";
import { ProgStrip } from "./prog-strip";
import { RetakeChip } from "./retake-chip";
import { SubmitError } from "./submit-error";
import { ValidScore } from "./valid-score";

interface EndFaceProps {
  readonly code: string;
  /** The just-run attempt's local record — null in the review mode (a
   * read-back visit has no local facts to derive, the fact rows stay
   * absent rather than invented). */
  readonly record: AttemptRecord | null;
  /** The attempt's boot params — the derivations re-read the same thresholds the game ran with. */
  readonly engineParams: Readonly<Record<string, string>>;
  readonly verdict: SubmitVerdict;
  /** The live session binding — null after a practice run. */
  readonly step: StepContext | null;
  readonly testName: string;
  /** The session's read-back face key — the settlement tags and the
   * attempts note ride it; absent on practice runs. */
  readonly sessionId: string | undefined;
  /** Read-back mode: the face renders the session's stored verdict, no
   * local run, no retake entry. */
  readonly review: boolean;
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

// The session's retake budget: the first run of the lineup carries
// three chances, later sessions one.
function retakeBudget(step: StepContext | null): number {
  return step !== null ? (step.firstSession ? 3 : 1) : 0;
}

// The continue CTA's target: the next undone engine-backed step. With
// the lineup finished it returns undefined and the CTA closes the flow
// into the session-done page — the session's own read-back face carries
// the report entry.
function findNextStep(step: StepContext): StepRef | undefined {
  return step.lineup.find(
    (s) => !step.doneCodes.includes(s.code) && getGameModule(s.code) !== null,
  );
}

// The transport flattens the WKT wrappers: the line's optional numbers
// arrive as plain scalars-or-undefined, narrowed here once.
interface WireLine {
  bestScore: number;
  bestGrade: string | undefined;
  percentile: number | undefined;
  prevSessionScore: number | undefined;
  historyBestScore: number | undefined;
  attemptScores: number[];
}

/** The "vs last session" tag: the previous session's same-ability score
 * with its signed delta, or the honest first-attempt word — never zero
 * masquerading as a delta. */
function PrevTag({ line }: Readonly<{ line: WireLine | undefined }>) {
  if (line === undefined || line.prevSessionScore === undefined) {
    return (
      <span className="text-ink-muted bg-status-lock-bg rounded-full px-2.5 py-0.75 text-[10.5px] font-semibold">
        {m.test_result_first()}
      </span>
    );
  }
  const delta = line.bestScore - line.prevSessionScore;
  const up = delta >= 0;
  return (
    <span
      className={`rounded-full px-2.5 py-0.75 text-[10.5px] font-semibold ${
        up
          ? "text-status-ok bg-[rgba(34,197,94,0.12)]"
          : "bg-[rgba(248,113,113,0.12)] text-[#F87171]"
      }`}
    >
      {m.test_result_prev({ delta: `${up ? "+" : ""}${Math.round(delta * 10) / 10}` })}
    </span>
  );
}

/** The cross-session best tag (prototype .st bestTag): the bar the
 * comparison track marks, absent while the ledger has nothing. */
function BestTag({ line }: Readonly<{ line: WireLine | undefined }>) {
  if (line === undefined || line.historyBestScore === undefined) {
    return null;
  }
  return (
    <span className="text-ink-muted bg-status-lock-bg rounded-full px-2.5 py-0.75 text-[10.5px] font-semibold tabular-nums">
      {m.test_result_best({ score: Math.round(line.historyBestScore * 10) / 10 })}
    </span>
  );
}

/** The attempts receipt (prototype's note card): every valid attempt's
 * score in order with the taken one marked, then the rubric in one
 * honest line — a single attempt reads too, as the prototype's note
 * does. */
function AttemptsNote({ line }: Readonly<{ line: WireLine | undefined }>) {
  // toJson elides empty repeated fields, so a line without attempts may
  // arrive without the key at all — normalize before reading.
  const scores = line?.attemptScores ?? [];
  if (scores.length === 0) {
    return null; // an unscored line has no receipt to tell
  }
  const best = Math.max(...scores);
  // The receipt's identity is its position — repeated scores are legal
  // and carry no distinguishing content of their own — derived here so
  // the render's key never touches the index the array-key lint guards.
  const entries = scores.map((score, i) => ({
    key: `${i}-${score}`,
    leading: i > 0,
    score,
  }));
  return (
    <div className="border-line bg-surface-content text-ink-muted rounded-card border px-3.5 py-3 text-[11px] leading-[1.8]">
      {m.test_attempts_pre({ n: scores.length })}
      {entries.map(({ key, leading, score }) => (
        <span key={key}>
          {leading ? (
            <span aria-hidden className="text-ink-tertiary">
              {" / "}
            </span>
          ) : null}
          {score === best ? (
            <span className="text-grade-s font-bold tabular-nums">
              {Math.round(score * 10) / 10}
              <span className="ml-0.5 rounded-[4px] bg-[rgba(253,224,71,0.14)] px-1 py-px text-[9px] font-semibold">
                {m.test_attempts_taken()}
              </span>
            </span>
          ) : (
            <span className="text-ink-tertiary tabular-nums">{Math.round(score * 10) / 10}</span>
          )}
        </span>
      ))}
      {m.test_attempts_note()}
    </div>
  );
}

// The score card: the ability line, then the verdict chain's face — big
// score with grade chip, the pending pulse, the invalid card, the error
// card with its retry — and the settlement's three receipt tags.
function VerdictCard({
  verdict,
  step,
  testName,
  line,
  onResubmit,
}: Readonly<{
  verdict: SubmitVerdict;
  step: StepContext | null;
  testName: string;
  line: WireLine | undefined;
  onResubmit: () => void;
}>) {
  const word = step !== null ? abilityWord(step.ability) : null;
  const maxRetakes = retakeBudget(step);
  const retakeLeft = verdict.phase === "done" ? verdict.retestLeft : (step?.retestLeft ?? 0);
  return (
    <div className="border-line bg-surface-card shadow-card rounded-card border px-4 py-4.5 text-center backdrop-blur-md">
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
      {verdict.phase === "done" && verdict.valid ? <ValidScore verdict={verdict} /> : null}
      {verdict.phase === "done" && !verdict.valid ? (
        verdict.invalidReason !== undefined ? (
          <InvalidNotice reason={verdict.invalidReason} />
        ) : (
          // The read-back with no stored line: not a void — the honest
          // no-score word the prototype's own note carries.
          <div className="text-ink-tertiary mt-3 text-sm">{m.test_review_no_score()}</div>
        )
      ) : null}
      {verdict.phase === "error" ? <SubmitError onResubmit={onResubmit} /> : null}
      {verdict.phase === "done" && verdict.inSession ? (
        <div className="mt-2.5 flex flex-wrap items-center justify-center gap-2">
          <PrevTag line={line} />
          <RetakeChip left={retakeLeft} max={maxRetakes} />
          <BestTag line={line} />
        </div>
      ) : null}
    </div>
  );
}

// The read-back mode's verdict — the stored line is the whole truth: no
// local run, no submit chain. A pending read keeps the pulse; a done
// read with no line renders the honest no-score face (the reason stays
// undefined so it never wears the void card — an unplayed, a voided,
// and a norms-pending game all read the same here until the report
// carries more truth).
function reviewVerdictOf(line: WireLine | undefined, step: StepContext | null): SubmitVerdict {
  if (line === undefined) {
    return {
      phase: "done",
      valid: false,
      invalidReason: undefined,
      authoritativeScore: undefined,
      rawScore: undefined,
      grade: undefined,
      retestLeft: step?.retestLeft ?? 0,
      inSession: true,
    };
  }
  return {
    phase: "done",
    valid: true,
    invalidReason: undefined,
    authoritativeScore: line.bestScore,
    rawScore: undefined,
    grade: line.bestGrade,
    retestLeft: step?.retestLeft ?? 0,
    inSession: true,
  };
}

// The face's one verdict: the submit chain's on a live run, the stored
// line's on a read-back — the pending read keeps the pulse face.
function shownVerdictOf(
  review: boolean,
  pending: boolean,
  line: WireLine | undefined,
  step: StepContext | null,
  verdict: SubmitVerdict,
): SubmitVerdict {
  if (!review) {
    return verdict;
  }
  if (pending) {
    return { phase: "idle" };
  }
  return reviewVerdictOf(line, step);
}

/**
 * The result face (prototype test-result): the big score with its grade
 * chip, the settlement's three receipt tags (vs-last-session, retake
 * budget, cross-session best), the attempts note, the fact rows as
 * label/value lines, and the two CTAs — the retake, greyed dead whenever
 * the read-back mode or a spent budget holds it (the prototype's ro face
 * keeps the button visible-but-dead rather than hiding it), with the
 * ruled exit line under it — and the primary continue, which walks to the next
 * undone engine-backed step or closes the flow into the session-done
 * page. Markup-only; every number comes from results.ts, the submit
 * verdict, or the session's read-back face.
 */
export function EndFace({
  code,
  record,
  engineParams,
  verdict,
  step,
  testName,
  sessionId,
  review,
  onResubmit,
  onRetry,
  onExit,
}: Readonly<EndFaceProps>) {
  const navigate = useNavigate();
  // The session's read-back face: the settlement tags and the attempts
  // note ride it; a practice run has no session to read.
  const reportQuery = useConnectQuery(
    getSessionReport,
    { sessionId: sessionId ?? "" },
    { enabled: sessionId !== undefined },
  );
  const line = (reportQuery.data?.abilities ?? []).find((a) => a.gameCode === code) as
    | WireLine
    | undefined;
  const shown = shownVerdictOf(review, reportQuery.isPending, line, step, verdict);

  const rows =
    record !== null ? deriveLocalResults(code, record, new EngineParams(engineParams)) : null;
  const spent = shown.phase === "done" && shown.inSession && shown.retestLeft <= 0;
  const next = step !== null ? findNextStep(step) : undefined;
  const closeFlow = () =>
    sessionId !== undefined
      ? void navigate({ to: "/report/session/$sessionId", params: { sessionId } })
      : onExit();

  // overflow-x-hidden: both orbs bleed past the side edges by design, and
  // a lone overflow-y-auto computes overflow-x to auto (the CSS
  // visible-with-non-visible rule) — the bleed then buys the face a
  // horizontal scrollbar. x stays clipped, y scrolls.
  return (
    <div className="bg-canvas absolute inset-0 isolate z-40 overflow-x-hidden overflow-y-auto">
      <div
        aria-hidden
        className="pointer-events-none absolute -top-20 -left-17.5 size-[260px] rounded-full bg-[#2563eb] opacity-26 blur-[70px]"
      />
      <div
        aria-hidden
        className="bg-ambient-cyan pointer-events-none absolute -right-17.5 -bottom-15 size-[220px] rounded-full opacity-26 blur-[70px]"
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
          <VerdictCard
            line={line}
            onResubmit={onResubmit}
            step={step}
            testName={testName}
            verdict={shown}
          />

          <AttemptsNote line={line} />

          {rows !== null ? (
            <div className="border-line bg-surface-card shadow-card rounded-card border p-4 backdrop-blur-md">
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
          ) : null}

          <div className="flex flex-col gap-3">
            {/* The prototype's read-back keeps the retake visible but dead
                (the ro face and a spent budget both grey it); the ruled
                spent-exit line rides under the button. */}
            <Button
              className={`bg-cta-practice shadow-cta-practice rounded-control w-full cursor-pointer px-4 py-3.5 text-[15px] font-bold text-white active:opacity-80 ${
                spent || review ? "pointer-events-none opacity-45" : ""
              }`}
              onPress={onRetry}
            >
              {spent ? m.test_result_retry_disabled() : m.test_end_retry()}
            </Button>
            {spent ? (
              <p className="text-[#FCD34D] px-1 text-[11.5px] leading-[1.7]">
                {m.test_intro_blocked()}
              </p>
            ) : null}
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
                onPress={closeFlow}
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
