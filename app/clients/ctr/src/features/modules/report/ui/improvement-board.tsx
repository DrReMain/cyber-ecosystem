import { m } from "#/paraglide/messages";

export interface ImprovementLine {
  abilityCn: string;
  current: number;
  target: number;
  gap: number;
}

/** The improvement board (unlocked-only, weakest first — the server's
 * order): each ability's run toward the excellence line (the A
 * threshold), the gap in the same percentile currency. */
export function ImprovementBoard({ lines }: Readonly<{ lines: ImprovementLine[] }>) {
  return (
    <div className="border-line bg-surface-card shadow-card rounded-card border px-4 py-3.5 backdrop-blur-md">
      <div className="text-ink mb-2.5 text-sm font-extrabold">{m.report_improve_title()}</div>
      <div className="flex flex-col gap-3">
        {lines.map((line) => (
          <div key={line.abilityCn}>
            <div className="flex items-baseline justify-between">
              <span className="text-ink text-[12.5px] font-bold">{line.abilityCn}</span>
              <span className="text-ink-tertiary text-[10.5px] tabular-nums">
                {m.report_improve_target({ pct: Math.round(line.target) })}
                <span className="text-ink-muted"> · </span>
                {m.report_improve_gap({ pct: Math.round(line.gap) })}
              </span>
            </div>
            <div className="relative mt-1.5 h-[6px] w-full overflow-hidden rounded-full bg-[#111827]">
              <div
                className="h-full rounded-full bg-gradient-to-r from-[#22D3EE] to-[#3B82F6]"
                style={{ width: `${Math.min(100, Math.max(2, line.current))}%` }}
              />
              <div
                aria-hidden
                className="absolute top-0 bottom-0 w-[2px] bg-[#FDE047]"
                style={{ left: `${Math.min(100, Math.max(0, line.target))}%` }}
              />
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
