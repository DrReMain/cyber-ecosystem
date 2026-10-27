import { TierChip } from "#/features/modules/test/ui/tier-chip";
import { m } from "#/paraglide/messages";
import { type DeltaWord, type TierWord, zoneCopyOf } from "../vocabulary";
import { AdviceBox } from "./advice-box";
import { LockLine } from "./lock-line";

/** The per-ability card (prototype ③): ability identity, the percentile
 * bar and score, then either the unlocked detail (three-zone analysis
 * copy + the last-attempt delta) or the locked teaser line. The analysis
 * copy is client-held five-zone wording — the contract carries numbers,
 * and the configuration face lands with E9. */
export function AbilityRow({
  abilityCn,
  gradient,
  isSub,
  pct,
  score,
  tier,
  delta,
  unlocked,
}: Readonly<{
  abilityCn: string;
  gradient: string;
  isSub: boolean;
  pct: number | null;
  score: number;
  tier: TierWord;
  delta: DeltaWord | null;
  unlocked: boolean;
}>) {
  const zone = zoneCopyOf(abilityCn, pct, score);
  return (
    <div className="border-line bg-surface-card shadow-card rounded-card border p-3.5 backdrop-blur-md">
      <div className="flex items-center justify-between gap-2">
        <span className="text-ink text-[13.5px] font-bold">{abilityCn}</span>
        {isSub ? (
          <span className="rounded-full bg-[rgba(255,255,255,0.06)] px-2 py-0.5 text-[10px] font-semibold text-[#9CA3AF]">
            {m.report_row_sub()}
          </span>
        ) : (
          <TierChip grade={pct === null ? undefined : tier.letter} />
        )}
      </div>
      <div className="mt-2 flex items-center gap-2.5">
        <div
          aria-hidden
          className={`shadow-inset-ring flex h-10 w-10 shrink-0 items-center justify-center rounded-xl text-[11px] font-bold text-white ${gradient}`}
        >
          {abilityCn.slice(0, 2)}
        </div>
        <div className="min-w-0 flex-1">
          {pct !== null ? (
            <div className="h-[6px] w-full overflow-hidden rounded-full bg-[#111827]">
              <div
                className="h-full rounded-full bg-gradient-to-r from-[#22D3EE] to-[#22C55E]"
                style={{ width: `${Math.min(100, Math.max(2, pct))}%` }}
              />
            </div>
          ) : (
            <div className="text-ink-tertiary text-[10.5px]">{m.report_axis_score_only()}</div>
          )}
          {/* Unlocked with no cohort answer: the bar's score-only note
              already says it — the detail seat renders nothing rather
              than repeat the line. */}
          {unlocked ? (
            zone !== null ? (
              <>
                <div className="text-ink-tertiary mt-2 text-[11px] leading-[1.6]">
                  {zone.analysis}
                </div>
                <AdviceBox advice={zone.advice} delta={delta} />
              </>
            ) : null
          ) : (
            <LockLine />
          )}
        </div>
        <div className="shrink-0 text-right">
          <div className="text-ink text-base font-extrabold tabular-nums">
            {m.report_row_score({ score: Math.round(score * 10) / 10 })}
          </div>
          {pct !== null ? (
            <div className="text-ink-tertiary mt-0.5 text-[9.5px] tabular-nums">
              {m.report_row_beat({ pct: Math.round(pct) })}
            </div>
          ) : null}
        </div>
      </div>
    </div>
  );
}
