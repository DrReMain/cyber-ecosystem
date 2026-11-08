import { Check } from "lucide-react";
import { m } from "#/paraglide/messages";

/** The history-best bar (prototype cmpBar): the current attempt against
 * the cross-session best on one track — the fill is the current, the
 * mark the best, and an equal-or-higher fill reads as the refreshed
 * best. The best arrives from the digest and includes this attempt, so a
 * tie means this attempt set it; an absent best renders nothing. */
export function CmpBar({ score, best }: Readonly<{ score: number; best: number | undefined }>) {
  if (best === undefined || best <= 0) {
    return null;
  }
  const hi = Math.max(score, best);
  const refreshed = score >= best;
  return (
    <div className="mt-2 rounded-[9px] border border-[rgba(55,65,81,0.5)] bg-[rgba(255,255,255,0.03)] px-3 py-2">
      <div className="flex items-center justify-between text-[11px]">
        <span className="text-ink-tertiary">
          {m.report_cmp_current({ score: Math.round(score * 10) / 10 })}
        </span>
        <span className="font-bold text-[#6EE7B7] tabular-nums">
          {m.report_cmp_best({ score: Math.round(best * 10) / 10 })}
        </span>
      </div>
      <div className="relative mt-1.5 h-1.5 overflow-hidden rounded-full bg-[#111827]">
        <div
          className="from-link to-status-ok h-full rounded-full bg-gradient-to-r"
          style={{ width: `${Math.round((score / hi) * 100)}%` }}
        />
        <span
          aria-hidden
          className="absolute top-0 h-full w-0.5 bg-[#6EE7B7]"
          style={{ left: `${Math.round((best / hi) * 100)}%` }}
        />
      </div>
      <div
        className={`mt-1.5 flex items-center gap-1 text-[10px] ${refreshed ? "text-[#6EE7B7]" : "text-ink-tertiary"}`}
      >
        {refreshed ? <Check aria-hidden size={10} strokeWidth={3} /> : null}
        {refreshed ? m.report_cmp_refreshed() : m.report_cmp_gap()}
      </div>
    </div>
  );
}
