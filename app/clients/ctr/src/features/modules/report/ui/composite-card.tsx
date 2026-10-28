import { TierChip } from "#/features/modules/test/ui/tier-chip";
import { m } from "#/paraglide/messages";
import type { TierWord } from "../vocabulary";

/** The headline card (prototype ①): the composite percentile in the
 * "top X%" phrasing with the tier chip, the beaten share on the right —
 * and the honest no-verdict face when no measured ability carries a
 * percentile. Markup-only; every number arrives as a prop. */
export function CompositeCard({
  composite,
  tier,
}: Readonly<{
  composite: number | undefined;
  tier: TierWord;
}>) {
  const hasComposite = composite !== undefined && composite !== null;
  return (
    <div className="border-line bg-surface-card shadow-card rounded-card relative overflow-hidden border px-4 py-4.5 backdrop-blur-md">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_85%_0%,rgba(6,182,212,0.14),transparent_55%)]"
      />
      <div className="relative flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="text-ink-tertiary text-[11px]">{m.report_composite_label()}</div>
          {hasComposite ? (
            <div className="mt-1 flex flex-wrap items-baseline gap-2.5">
              <span className="text-ink text-[38px] leading-none font-black tracking-tight tabular-nums">
                {m.report_composite_top({ pct: Math.max(1, Math.round(100 - composite)) })}
              </span>
              <TierChip grade={tier.letter === "·" ? undefined : tier.letter} />
            </div>
          ) : (
            <div className="mt-2 flex flex-wrap items-baseline gap-2.5">
              <span className="text-ink text-[30px] leading-none font-black">—</span>
              <span className="text-ink-tertiary text-[11px] leading-snug">
                {m.report_composite_none()}
              </span>
            </div>
          )}
        </div>
        {hasComposite ? (
          <div className="shrink-0 text-right">
            <div className="text-ink-tertiary text-[10px]">{m.report_composite_beat_label()}</div>
            <div className="text-status-ok mt-0.5 text-[22px] font-extrabold tabular-nums">
              {Math.round(composite)}%
            </div>
          </div>
        ) : null}
      </div>
    </div>
  );
}
