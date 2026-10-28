import { m } from "#/paraglide/messages";
import type { WireCareerRecommendation } from "../career";
import { dimensionWord } from "../vocabulary";

/** One seeded area's recommendation row: the match as a progress bar over
 * the 0-100 floor, the basis dimensions in the frozen matrix's own words,
 * and the library's own facts (washout, training load, advice) rendered
 * as data the row carries — the page never words them itself. */
export function CareerAreaCard({ rec }: Readonly<{ rec: WireCareerRecommendation }>) {
  const pct = Math.max(0, Math.min(100, rec.matchedScore));
  const basis = rec.basisDimensions.map(dimensionWord).join(" · ");
  return (
    <div className="border-line bg-surface-card shadow-card rounded-card border px-4 py-3.5 backdrop-blur-md">
      <div className="flex items-baseline justify-between gap-3">
        <span className="text-ink text-[13.5px] font-bold">{rec.name}</span>
        <span className="text-identity text-[12px] font-bold tabular-nums">
          {m.report_career_matched({ score: Math.round(pct) })}
        </span>
      </div>
      <div className="bg-line mt-2 h-1.5 w-full overflow-hidden rounded-full">
        <div className="bg-identity h-full rounded-full" style={{ width: `${pct}%` }} />
      </div>
      <div className="text-ink-tertiary mt-2 text-[11px]">
        {m.report_career_basis({ dims: basis })}
      </div>
      <div className="mt-2 flex flex-wrap gap-x-4 gap-y-1">
        <span className="text-ink-secondary text-[11px] tabular-nums">
          {m.report_career_elimination({ rate: Math.round(rec.eliminationRate) })}
        </span>
        <span className="text-ink-secondary text-[11px]">
          {m.report_career_intensity({ value: rec.trainingIntensity })}
        </span>
      </div>
      <p className="text-ink-secondary mt-2 text-[12px] leading-relaxed">{rec.suggestion}</p>
    </div>
  );
}
