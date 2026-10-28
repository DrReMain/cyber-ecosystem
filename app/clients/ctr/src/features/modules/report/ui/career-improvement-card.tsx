import { m } from "#/paraglide/messages";
import type { WireCareerImprovement } from "../career";
import { dimensionWord } from "../vocabulary";

/** One dimension's gap row: the current value against the server-owned
 * excellence line with the gap in percent of it, and the prescription
 * copy the seeded library carries — a dimension already at the line
 * states a zero gap, never a negative one. */
export function CareerImprovementCard({ imp }: Readonly<{ imp: WireCareerImprovement }>) {
  const pct = Math.max(0, Math.min(100, imp.current));
  return (
    <div className="border-line bg-surface-card shadow-card rounded-card border px-4 py-3.5 backdrop-blur-md">
      <div className="flex items-baseline justify-between gap-3">
        <span className="text-ink text-[13px] font-bold">{dimensionWord(imp.dimension)}</span>
        <span className="text-ink-tertiary text-[11px] tabular-nums">
          {m.report_career_gap({
            current: Math.round(imp.current),
            target: Math.round(imp.target),
            gap: imp.gapPercent,
          })}
        </span>
      </div>
      <div className="bg-line mt-2 h-1.5 w-full overflow-hidden rounded-full">
        <div className="bg-identity h-full rounded-full" style={{ width: `${pct}%` }} />
      </div>
      <ul className="mt-2 space-y-1">
        {imp.suggestions.map((s) => (
          <li className="text-ink-secondary text-[11.5px] leading-relaxed" key={s}>
            · {s}
          </li>
        ))}
      </ul>
    </div>
  );
}
