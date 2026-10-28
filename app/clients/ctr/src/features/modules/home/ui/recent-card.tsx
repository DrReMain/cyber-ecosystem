import { useNavigate } from "@tanstack/react-router";
import { ChevronRight } from "lucide-react";
import { TierChip } from "#/features/modules/test/ui/tier-chip";
import { m } from "#/paraglide/messages";

/** One percentile cell of the recent card's strip. */
export interface RecentItem {
  readonly abilityCn: string;
  readonly pct: number;
}

/** The recent card (prototype home's 最近一次测评): the latest scored
 * session's headline — the composite percentile with its tier, the date
 * and lineup size, and the first measured abilities as a percentile
 * strip — tapping through to the anchor's report. A thin cohort keeps
 * the number honest: no composite renders the dash, no verdict keeps the
 * chip away, never a zero masquerading as a score. */
export function RecentCard({
  anchorId,
  composite,
  count,
  date,
  genreWord,
  items,
  tierLetter,
}: Readonly<{
  anchorId: string;
  composite: number | undefined;
  count: number;
  date: string;
  genreWord: string;
  items: readonly RecentItem[];
  tierLetter: string | undefined;
}>) {
  const navigate = useNavigate();
  return (
    <div className="border-line bg-surface-card shadow-card rounded-card border px-4 py-4 backdrop-blur-md">
      <div className="mb-2.5 flex items-center justify-between">
        <span className="text-ink text-sm font-extrabold">{m.home_recent_title()}</span>
        <button
          className="text-link tap-zone flex items-center gap-0.5 text-[11.5px] font-bold active:opacity-60"
          onClick={() => void navigate({ to: "/report/$recordId", params: { recordId: anchorId } })}
          type="button"
        >
          {m.home_recent_go()}
          <ChevronRight aria-hidden size={12} />
        </button>
      </div>
      <div className="flex items-end gap-3">
        <div>
          <div className="text-ink text-[38px] leading-none font-black tabular-nums">
            {composite === undefined ? "—" : Math.round(composite)}
          </div>
          <div className="text-ink-tertiary mt-1 text-[10.5px]">{m.home_recent_comp()}</div>
        </div>
        {composite !== undefined ? (
          <div className="pb-3.25">
            <TierChip grade={tierLetter} />
          </div>
        ) : null}
        <div className="text-ink-tertiary ml-auto pb-3.25 text-right text-[10.5px] leading-[1.5]">
          {date}
          <br />
          {genreWord} · {m.home_recent_items({ n: count })}
        </div>
      </div>
      {items.length > 0 ? (
        <div className="border-line mt-3 flex border-t pt-3">
          {items.map((it) => (
            <div className="flex-1 text-center" key={it.abilityCn}>
              <div className="text-ink text-[15px] font-black tabular-nums">
                {Math.round(it.pct)}
              </div>
              <div className="text-ink-tertiary mt-0.5 text-[10px]">{it.abilityCn}</div>
            </div>
          ))}
        </div>
      ) : null}
    </div>
  );
}
