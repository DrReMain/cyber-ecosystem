import type { SubmitTestAttemptResponse } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test_pb";
import { ErrorFallback } from "#/features/app/error/error-fallback";
import { LoadingStrip } from "#/features/app/states/loading-strip";
import { m } from "#/paraglide/messages";

export type Verdict =
  | { readonly phase: "pending" }
  | {
      readonly phase: "error";
      readonly error: unknown;
      readonly onResubmit: () => void;
    }
  | { readonly phase: "done"; readonly receipt: SubmitTestAttemptResponse }
  /** The attempt never crossed the seam (unmapped code, non-completed
   * exit) — no verdict exists, the end face stays local-only. */
  | { readonly phase: "local" };

interface VerdictCardProps {
  readonly verdict: Verdict;
}

// The transport re-serializes responses to JSON — WKT wrappers arrive
// flattened to scalars while the gen types claim message shapes. One cast
// at this boundary, then plain runtime guards everywhere downstream.
interface FlatReceipt {
  readonly valid: boolean;
  readonly rawScore?: number;
  readonly authoritativeScore?: number;
  readonly grade?: string;
  readonly invalidReason?: string;
  readonly retestLeft: number;
}

function flattenReceipt(receipt: SubmitTestAttemptResponse): FlatReceipt {
  return receipt as unknown as FlatReceipt;
}

// Grade chip classes per four-color law {S | A | B | C} — open vocabulary,
// an unknown word falls back to the neutral chip, never to no chip.
const GRADE_CLASSES: Record<string, string> = {
  S: "border-grade-s/40 bg-grade-s/10 text-grade-s",
  A: "border-grade-a/40 bg-grade-a/10 text-grade-a",
  B: "border-grade-b/40 bg-grade-b/10 text-grade-b",
  C: "border-grade-c/40 bg-grade-c/10 text-grade-c",
};

// Server invalid-reason words (ops vocabulary) → user copy; an unmapped
// word renders the generic fallback, never the raw word.
const INVALID_COPY: Record<string, () => string> = {
  BACKGROUND_OVER_BUDGET: m.test_verdict_invalid_background,
  ABSURD_REACTION_TIME: m.test_verdict_invalid_absurd_reaction,
  ABSURD_TIME_SPAN: m.test_verdict_invalid_span,
  DETAIL_MISMATCH: m.test_verdict_invalid_mismatch,
};

function invalidCopy(reason: string | undefined): string {
  if (reason !== undefined && INVALID_COPY[reason] !== undefined) {
    return INVALID_COPY[reason]();
  }
  return m.test_verdict_invalid_fallback();
}

function fmtScore(value: number): string {
  return Number.isInteger(value) ? String(value) : value.toFixed(1);
}

/**
 * The server verdict receipt — the end face's primary block. Valid attempts
 * show the dual-track scores (authoritative absent = the kind's percentile
 * curve is still calibrating, degraded to the raw score with a note);
 * invalid attempts show the refusal with its reason copy; a transport
 * refusal renders through the shared error row with resubmit. Props-only:
 * numbers arrive pre-judged, submission lives in the wiring layer.
 */
export function VerdictCard({ verdict }: Readonly<VerdictCardProps>) {
  if (verdict.phase === "local") {
    return null;
  }

  if (verdict.phase === "pending") {
    return (
      <div className="border-line bg-surface-deep rounded-card border px-4 py-5" role="status">
        <span className="sr-only">{m.test_verdict_pending()}</span>
        <LoadingStrip tone="ink" />
      </div>
    );
  }

  if (verdict.phase === "error") {
    return (
      <div className="border-line bg-surface-deep rounded-card border">
        <ErrorFallback error={verdict.error} onRetry={verdict.onResubmit} />
      </div>
    );
  }

  const receipt = flattenReceipt(verdict.receipt);

  if (!receipt.valid) {
    return (
      <div className="border-danger/40 bg-surface-deep rounded-card border px-4 py-4">
        <span className="border-danger/40 bg-danger/10 text-danger-text inline-flex rounded-full border px-2.5 py-0.75 font-mono text-[10px] tracking-[0.18em] uppercase">
          {m.test_verdict_invalid()}
        </span>
        <p className="text-ink-secondary mt-2.5 text-[12.5px] leading-relaxed">
          {invalidCopy(receipt.invalidReason)}
        </p>
      </div>
    );
  }

  const grade = receipt.grade;
  const gradeClass = grade !== undefined ? (GRADE_CLASSES[grade] ?? null) : null;

  return (
    <div className="border-line bg-surface-deep rounded-card border px-4 py-4">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="text-ink-tertiary text-xs">{m.test_verdict_raw()}</div>
          <div className="text-ink mt-0.5 text-3xl leading-none font-black tracking-tight">
            {receipt.rawScore !== undefined ? fmtScore(receipt.rawScore) : "—"}
          </div>
        </div>
        {grade !== undefined && (
          <span
            className={`inline-flex h-11 w-11 shrink-0 items-center justify-center rounded-full border text-xl font-black ${gradeClass ?? "border-line bg-surface-input text-ink-secondary"}`}
          >
            <span aria-hidden>{grade}</span>
            <span className="sr-only">{m.test_verdict_grade({ grade })}</span>
          </span>
        )}
      </div>
      <div className="border-line-soft mt-3.5 border-t pt-3">
        {receipt.authoritativeScore !== undefined ? (
          <div className="flex items-baseline justify-between">
            <span className="text-ink-tertiary text-xs">{m.test_verdict_authoritative()}</span>
            <span className="text-ink text-base font-bold">
              {fmtScore(receipt.authoritativeScore)}
            </span>
          </div>
        ) : (
          <p className="text-ink-tertiary text-[11px] leading-relaxed">
            {m.test_verdict_calibration()}
          </p>
        )}
        {receipt.retestLeft > 0 && (
          <div className="text-status-ok mt-2 text-right text-[11.5px] font-bold">
            {m.test_verdict_retest({ n: receipt.retestLeft })}
          </div>
        )}
      </div>
    </div>
  );
}
