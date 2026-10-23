import { Button } from "react-aria-components";
import { m } from "#/paraglide/messages";
import { type AttemptRecord, EngineParams } from "../contract";
import { deriveLocalResults } from "../results";
import type { SubmitVerdict } from "../submit";

interface EndFaceProps {
  readonly code: string;
  readonly record: AttemptRecord;
  /** The attempt's boot params — the derivations re-read the same thresholds the game ran with. */
  readonly engineParams: Readonly<Record<string, string>>;
  readonly verdict: SubmitVerdict;
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

// The per-game end face: the server's verdict card (dual-track score and
// grade, the invalid face, the retake receipt) above the local fact rows.
// Markup-only; every number comes from results.ts or the submit verdict.
export function EndFace({
  code,
  record,
  engineParams,
  verdict,
  onResubmit,
  onRetry,
  onExit,
}: Readonly<EndFaceProps>) {
  const rows = deriveLocalResults(code, record, new EngineParams(engineParams));
  return (
    <div className="absolute inset-0 z-40 flex flex-col overflow-y-auto bg-black/80 p-6 backdrop-blur-sm">
      <div className="mx-auto flex w-full max-w-xs flex-1 flex-col justify-center py-6">
        <h2 className="text-ink text-center text-xl font-extrabold">{m.test_end_title()}</h2>
        <div className="mt-4">
          {verdict.phase === "submitting" && (
            <div className="border-line bg-surface-card text-ink-tertiary rounded-card animate-pulse border px-4 py-3 text-center text-sm">
              {m.test_verdict_pending()}
            </div>
          )}
          {verdict.phase === "done" && verdict.valid && (
            <div className="border-line bg-surface-card rounded-card border px-4 py-3 text-center">
              <div className="text-ink-tertiary text-xs">{m.test_result_score()}</div>
              <div className="text-ink mt-0.5 text-2xl font-black">
                {verdict.authoritativeScore ?? "—"}
                {verdict.grade !== undefined && (
                  <span className="text-cta-primary ml-2 align-middle text-base font-extrabold">
                    {verdict.grade}
                  </span>
                )}
              </div>
              {verdict.inSession ? (
                <div className="text-status-ok mt-1 text-[11px] font-bold">
                  {verdict.retestLeft > 0
                    ? m.test_verdict_left({ n: verdict.retestLeft })
                    : m.test_verdict_left_zero()}
                </div>
              ) : null}
            </div>
          )}
          {verdict.phase === "done" && !verdict.valid && (
            <div className="border-status-pending/45 bg-surface-card text-status-pending rounded-card border px-4 py-3 text-center text-sm">
              <span className="font-bold">{m.test_verdict_invalid()}</span>
              <span className="opacity-80"> · {invalidCopy(verdict.invalidReason)}</span>
            </div>
          )}
          {verdict.phase === "error" && (
            <div className="border-line bg-surface-card rounded-card border px-4 py-3 text-center">
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
        </div>
        <div className="mt-3 grid grid-cols-2 gap-2.5">
          {rows.map((row) => (
            <div
              className="rounded-card border-line bg-surface-card border px-3 py-2.5"
              key={row.key}
            >
              <div className="text-ink-tertiary text-xs">
                {(ROW_LABELS[row.key] ?? m.test_result_taps)()}
              </div>
              <div className="text-ink mt-0.5 text-lg font-bold">{row.value}</div>
            </div>
          ))}
        </div>
        <div className="mt-6 flex flex-col gap-3">
          {/* Retry is the practice/orange CTA (prototype btn-org), not the primary blue. */}
          <Button
            className="rounded-control bg-cta-practice w-full cursor-pointer px-4 py-3 font-semibold text-white active:opacity-80"
            onPress={onRetry}
          >
            {m.test_end_retry()}
          </Button>
          <Button
            className="rounded-control border-line text-ink-secondary w-full cursor-pointer border px-4 py-3 font-semibold active:opacity-80"
            onPress={onExit}
          >
            {m.test_end_exit()}
          </Button>
        </div>
      </div>
    </div>
  );
}
