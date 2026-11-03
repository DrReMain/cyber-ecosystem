import { TierChip } from "#/features/modules/test/ui/tier-chip";
import { m } from "#/paraglide/messages";
import type { TierWord } from "../vocabulary";

/** The cognitive report's head (prototype report head): the composite as
 * a raw 0-100 ring — the dimension mean, not a percentile, so no
 * "top X%" inversion — with the tier chip and the two counters beside
 * it; a null composite keeps the ring, honestly empty. */
export function CognitiveHead({
  composite,
  covered,
  tested,
  tier,
  total,
}: Readonly<{
  composite: number | undefined;
  covered: number;
  tested: number;
  tier: TierWord;
  total: number;
}>) {
  const has = composite !== undefined && composite !== null;
  const circumference = 2 * Math.PI * 46;
  const on = circumference * (Math.max(0, Math.min(100, composite ?? 0)) / 100);
  return (
    <div className="border-line bg-surface-card shadow-card rounded-card border px-4 py-4 backdrop-blur-md">
      <div className="flex items-center gap-4">
        <div className="relative size-28 shrink-0">
          <svg aria-hidden className="-rotate-90" viewBox="0 0 112 112">
            <circle cx={56} cy={56} fill="none" r={46} stroke="rgba(55,65,81,.75)" strokeWidth={8} />
            {has ? (
              <circle
                cx={56}
                cy={56}
                fill="none"
                r={46}
                stroke="#3B82F6"
                strokeDasharray={`${on.toFixed(1)} ${(circumference - on).toFixed(1)}`}
                strokeLinecap="round"
                strokeWidth={8}
              />
            ) : null}
          </svg>
          <div className="absolute inset-0 flex flex-col items-center justify-center">
            <span className="text-ink text-[27px] leading-none font-black tabular-nums">
              {has ? Math.round(composite) : "—"}
            </span>
            <span className="text-ink-tertiary mt-0.5 text-[9.5px]">
              {has ? m.report_cognitive_composite_label() : m.report_cognitive_none()}
            </span>
          </div>
        </div>
        <div className="min-w-0 flex-1">
          {has ? <TierChip grade={tier.letter === "·" ? undefined : tier.letter} /> : null}
          <div className="text-ink-tertiary mt-2 text-[12px] leading-[1.7]">
            {m.report_cognitive_dims_tested({ tested, total })}
            <br />
            {m.report_cognitive_abilities_covered({ covered })}
          </div>
          {tested < total ? (
            <div className="text-ink-tertiary mt-1.5 text-[11px]">
              {m.report_cognitive_full_note()}
            </div>
          ) : null}
        </div>
      </div>
    </div>
  );
}
