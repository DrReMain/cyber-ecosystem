import { m } from "#/paraglide/messages";

export interface ComparisonLine {
  abilityCn: string;
  previous: number | null;
  delta: { text: string; tone: "up" | "down" | "first" };
}

/** The last-attempt comparison card (unlocked-only): one line per ability
 * — the previous attempt's percentile and the delta, a first measurement
 * rendered as its own word, never as zero. */
export function ComparisonCard({ lines }: Readonly<{ lines: ComparisonLine[] }>) {
  return (
    <div className="border-line bg-surface-card shadow-card rounded-card border px-4 py-3.5 backdrop-blur-md">
      <div className="text-ink mb-2.5 text-sm font-extrabold">{m.report_cmp_title()}</div>
      <div className="flex flex-col">
        {lines.map((line) => (
          <div
            className="flex items-center justify-between border-b border-[rgba(55,65,81,0.45)] py-2.25 last:border-b-0"
            key={line.abilityCn}
          >
            <span className="text-ink-muted text-[12.5px]">{line.abilityCn}</span>
            <span className="flex items-baseline gap-2.5">
              <span className="text-ink-tertiary text-[11px] tabular-nums">
                {line.previous === null
                  ? m.report_cmp_no_prev()
                  : m.report_cmp_prev({ pct: Math.round(line.previous) })}
              </span>
              <span
                className={`text-[13px] font-extrabold tabular-nums ${
                  line.delta.tone === "up"
                    ? "text-status-ok"
                    : line.delta.tone === "down"
                      ? "text-[#F87171]"
                      : "text-ink-tertiary"
                }`}
              >
                {line.delta.text}
              </span>
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
