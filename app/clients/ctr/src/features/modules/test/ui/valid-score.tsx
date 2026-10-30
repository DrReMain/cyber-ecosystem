import { m } from "#/paraglide/messages";
import type { SubmitVerdict } from "../submit";
import { TierChip } from "./tier-chip";

export type DoneVerdict = Extract<SubmitVerdict, { phase: "done" }>;

/** The submit chain's done-and-valid face: the big score with its unit,
 * the grade chip, and the ungraded note that keeps the chip's shape
 * with the honest why. Norms pending, the authoritative track has no
 * calibrated number — the raw points hold the big slot, labeled as raw
 * so they never masquerade as the calibrated score. */
export function ValidScore({ verdict }: Readonly<{ verdict: DoneVerdict }>) {
  const uncalibrated = verdict.authoritativeScore === undefined;
  const score = verdict.authoritativeScore ?? verdict.rawScore;
  return (
    <>
      <div className="mt-1.5 flex items-baseline justify-center gap-1.5">
        <span className="text-ink text-[58px] leading-none font-black tabular-nums">
          {score ?? "—"}
        </span>
        <span className="text-ink-tertiary text-sm">
          {m.test_result_unit()}
          {uncalibrated && score !== undefined ? ` · ${m.test_result_raw_label()}` : ""}
        </span>
      </div>
      <div className="mt-2.5">
        <TierChip grade={verdict.grade} />
        {/* D-AO: the ungraded chip keeps its shape, with the honest
            note on why — the game's norms are still accumulating. */}
        {verdict.grade === undefined && (
          <div className="text-ink-tertiary mt-1.5 text-[10.5px]">{m.test_grade_off_note()}</div>
        )}
      </div>
    </>
  );
}
