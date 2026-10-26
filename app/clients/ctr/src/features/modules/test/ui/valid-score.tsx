import { m } from "#/paraglide/messages";
import type { SubmitVerdict } from "../submit";
import { TierChip } from "./tier-chip";

export type DoneVerdict = Extract<SubmitVerdict, { phase: "done" }>;

/** The submit chain's done-and-valid face: the big score with its unit,
 * the grade chip, and the ungraded note that keeps the chip's shape
 * with the honest why. */
export function ValidScore({ verdict }: Readonly<{ verdict: DoneVerdict }>) {
  return (
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
          <div className="text-ink-tertiary mt-1.5 text-[10.5px]">{m.test_grade_off_note()}</div>
        )}
      </div>
    </>
  );
}
