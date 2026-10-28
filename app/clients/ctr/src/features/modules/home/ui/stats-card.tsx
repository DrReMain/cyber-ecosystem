import { m } from "#/paraglide/messages";

/** The stat card (prototype home's 数据统计): the cross-session counts —
 * sessions done, core-ability coverage over the six-word vocabulary, the
 * reaction best — plus the prototype's own nudge while the coverage
 * stays short of the six the cognitive report needs. Absent values stay
 * dashes, never zeros. */
export function StatsCard({
  covered,
  reactionBest,
  total,
}: Readonly<{ covered: number; reactionBest: number | undefined; total: number }>) {
  const cells = [
    { key: "sessions", label: m.home_stat_sessions(), value: String(total) },
    { key: "covered", label: m.home_stat_covered(), value: `${covered} / 6` },
    {
      key: "reaction",
      label: m.home_stat_reaction(),
      value: reactionBest === undefined ? "—" : String(Math.round(reactionBest * 10) / 10),
    },
  ];
  return (
    <div className="border-line bg-surface-card shadow-card rounded-card border px-4 py-3.5 backdrop-blur-md">
      <div className="text-ink mb-2.5 text-sm font-extrabold">{m.home_stat_title()}</div>
      <div className="flex">
        {cells.map((cell) => (
          <div className="flex-1 text-center" key={cell.key}>
            <div className="text-ink text-[15px] font-black tabular-nums">{cell.value}</div>
            <div className="text-ink-tertiary mt-0.5 text-[10px]">{cell.label}</div>
          </div>
        ))}
      </div>
      {covered < 6 ? (
        <div className="text-ink-tertiary mt-2.5 text-[11px] leading-[1.6]">
          {m.home_stat_covered_note({ covered })}
        </div>
      ) : null}
    </div>
  );
}
