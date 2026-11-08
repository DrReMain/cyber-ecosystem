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
    <div className="border-line-accent-blue bg-identity rounded-card border px-4 py-3.75">
      <div className="text-ink mb-2.75 text-[14.5px] font-extrabold">{m.profile_stats_title()}</div>
      {rows.map((row) => (
        <div className="flex items-center gap-2.5 py-1.5" key={row.key}>
          <span className="text-ink-muted w-9.5 shrink-0 text-[11.5px]">{row.key}</span>
          <div className="bg-line-soft h-1.5 min-w-0 flex-1 overflow-hidden rounded-full">
            <div
              className="from-ambient-blue to-ambient-cyan h-full rounded-full bg-gradient-to-r"
              style={{ width: `${Math.round((row.n / max) * 100)}%` }}
            />
          </div>
          <span className="text-ink w-8.5 shrink-0 text-right text-xs font-extrabold tabular-nums">
            {m.profile_stats_count({ n: row.n })}
          </span>
        </div>
      ))}
    </div>
  );
}
