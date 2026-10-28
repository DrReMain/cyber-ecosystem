import { m } from "#/paraglide/messages";

/** The genre stats (prototype profile-overview ⑥): sessions per genre on
 * one shared track — the fuller genre sets the scale, a silent genre
 * still shows its honest zero. */
export function GenreStatsCard({ fps, moba }: Readonly<{ fps: number; moba: number }>) {
  const rows = [
    { key: "FPS", n: fps },
    { key: "MOBA", n: moba },
  ];
  const max = Math.max(fps, moba, 1);
  return (
    <div className="border-line-accent-blue bg-identity rounded-card border px-4 py-[15px]">
      <div className="text-ink mb-[11px] text-[14.5px] font-extrabold">
        {m.profile_stats_title()}
      </div>
      {rows.map((row) => (
        <div className="flex items-center gap-2.5 py-1.5" key={row.key}>
          <span className="text-ink-muted w-[38px] shrink-0 text-[11.5px]">{row.key}</span>
          <div className="h-[6px] min-w-0 flex-1 overflow-hidden rounded-full bg-[rgba(55,65,81,0.6)]">
            <div
              className="h-full rounded-full bg-gradient-to-r from-[#2563EB] to-[#06B6D4]"
              style={{ width: `${Math.round((row.n / max) * 100)}%` }}
            />
          </div>
          <span className="text-ink w-[34px] shrink-0 text-right text-xs font-extrabold tabular-nums">
            {m.profile_stats_count({ n: row.n })}
          </span>
        </div>
      ))}
    </div>
  );
}
