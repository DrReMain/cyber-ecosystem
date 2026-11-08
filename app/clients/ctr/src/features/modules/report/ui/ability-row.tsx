import { TierChip } from "#/features/modules/test/ui/tier-chip";
import { m } from "#/paraglide/messages";
import { type DeltaWord, type TierWord, zoneCopyOf } from "../vocabulary";
import { AdviceBox } from "./advice-box";
import { CmpBar } from "./cmp-bar";
import { LockLine } from "./lock-line";

// The chip seat's law: a locked row carries no verdict letter (the tier
// is part of the paid trio), the free seat wears the prototype's yellow
// top-score tag until membership opens the letters.
function SeatChip({
  free,
  isSub,
  pct,
  tier,
  unlocked,
}: Readonly<{
  free: boolean;
  isSub: boolean;
  pct: number | null;
  tier: TierWord;
  unlocked: boolean;
}>) {
  const open = free || unlocked;
  if (!open) {
    return null;
  }
  if (isSub) {
    return (
      <span className="text-status-lock rounded-full bg-[rgba(255,255,255,0.06)] px-2 py-0.5 text-[10px] font-semibold">
        {m.report_row_sub()}
      </span>
    );
  }
  if (free && !unlocked) {
    return (
      <span className="text-grade-s rounded-full bg-[rgba(234,179,8,0.18)] px-2 py-0.5 text-[10px] font-extrabold">
        {m.report_free_tag()}
      </span>
    );
  }
  return <TierChip grade={pct === null ? undefined : tier.letter} />;
}

/** The per-ability card (prototype ③): ability identity, the percentile
 * bar and score, then the detail seat — open for the free seat (the
 * best measured main ability's trio: tier, analysis, advice) or any
 * unlocked row, locked teaser otherwise. The analysis copy is
 * client-held five-zone wording — the contract carries numbers, and the
 * configuration face lands with E9. */
export function AbilityRow({
  abilityCn,
  free,
  gradient,
  historyBest,
  isSub,
  pct,
  score,
  tier,
  delta,
  unlocked,
}: Readonly<{
  abilityCn: string;
  free: boolean;
  gradient: string;
  historyBest: number | undefined;
  isSub: boolean;
  pct: number | null;
  score: number;
  tier: TierWord;
  delta: DeltaWord | null;
  unlocked: boolean;
}>) {
  const open = free || unlocked;
  const zone = zoneCopyOf(abilityCn, pct, score);
  return (
    <div
      className={`border-line bg-surface-card shadow-card rounded-card border p-3.5 backdrop-blur-md ${free && !unlocked ? "border-[rgba(234,179,8,0.45)]" : ""}`}
    >
      <div className="flex items-center justify-between gap-2">
        <span className="text-ink text-[13.5px] font-bold">{abilityCn}</span>
        <SeatChip free={free} isSub={isSub} pct={pct} tier={tier} unlocked={unlocked} />
      </div>
      <div className="mt-2 flex items-center gap-2.5">
        {/* The identity block is wordless in the prototype — a gradient
            plate, never a two-character abbreviation. */}
        <div
          aria-hidden
          className={`shadow-inset-ring h-10 w-10 shrink-0 rounded-xl ${gradient}`}
        />
        <div className="min-w-0 flex-1">
          {pct !== null ? (
            <div className="h-1.5 w-full overflow-hidden rounded-full bg-[#111827]">
              <div
                className="from-link to-status-ok h-full rounded-full bg-linear-to-r"
                style={{ width: `${Math.min(100, Math.max(2, pct))}%` }}
              />
            </div>
          ) : (
            <div className="text-ink-tertiary text-[10.5px]">{m.report_axis_score_only()}</div>
          )}
          {/* An open seat with no cohort answer keeps the bar's score-only
              note — the detail renders nothing rather than repeat the
              line; the history-best bar still reads the digest. */}
          {open ? (
            zone !== null ? (
              <>
                <div className="text-ink-tertiary mt-2 text-[11px] leading-[1.6]">
                  {zone.analysis}
                </div>
                <AdviceBox advice={zone.advice} delta={delta} />
              </>
            ) : null
          ) : (
            <LockLine abilityCn={abilityCn} />
          )}
          {open ? <CmpBar best={historyBest} score={score} /> : null}
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
