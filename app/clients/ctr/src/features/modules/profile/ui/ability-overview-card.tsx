import { TierChip } from "#/features/modules/test/ui/tier-chip";
import { m } from "#/paraglide/messages";
import type { AbilityOverviewRow } from "../overview";

/** The ability overview (prototype profile-overview ④): one row per
 * core ability — the identity chip, the tier, the cross-session best and
 * its percentile track; an untested row sinks with the way in named.
 * The track stays one gradient for every row: the chip already carries
 * the ability's identity, a per-tier color on the bar would press two
 * color meanings into one row. */
export function AbilityOverviewCard({ rows }: Readonly<{ rows: readonly AbilityOverviewRow[] }>) {
  const tested = rows.filter((r) => r.best !== undefined).length;
  return (
    <div className="border-line-accent-blue bg-identity rounded-card border px-4 py-3.75">
      <div className="mb-2.75 flex items-center justify-between">
        <span className="text-ink text-[14.5px] font-extrabold">{m.profile_ability_title()}</span>
        <span className="text-ink-tertiary text-[10.5px]">
          {m.profile_ability_sub({ n: tested })}
        </span>
      </div>
      {rows.map((row) =>
        row.best !== undefined ? (
          <div className="py-1.75" key={row.ability}>
            <div className="mb-1.5 flex items-center gap-2">
              {/* The prototype's identity block is a bare gradient chip —
               * no text rides inside it (the row's name carries the word). */}
              <div
                aria-hidden
                className={`${row.gradient} shadow-inset-ring h-6.5 w-6.5 shrink-0 rounded-[9px]`}
              />
              <span className="text-ink text-[12.5px] font-extrabold">{row.abilityCn}</span>
              <TierChip grade={row.grade} />
              <span className="text-ink-muted ml-auto text-[10.5px] tabular-nums">
                {m.profile_ability_best({ score: Math.round(row.best * 10) / 10 })}
              </span>
            </div>
            <div className="bg-line-soft h-1.5 w-full overflow-hidden rounded-full">
              <div
                className="from-ambient-blue to-ambient-cyan h-full rounded-full bg-gradient-to-r"
                style={{ width: `${row.pct ?? 0}%` }}
              />
            </div>
          </div>
        ) : (
          <div className="py-1.75 opacity-70" key={row.ability}>
            <div className="mb-1 flex items-center gap-2">
              <div
                aria-hidden
                className={`${row.gradient} shadow-inset-ring h-6.5 w-6.5 shrink-0 rounded-[9px] opacity-40`}
              />
              <span className="text-status-off text-[12.5px] font-extrabold">{row.abilityCn}</span>
              <span className="text-status-off rounded-full bg-[rgba(255,255,255,0.06)] px-2 py-0.5 text-[9.5px] font-semibold">
                {m.profile_ability_untested()}
              </span>
            </div>
            <div className="text-ink-tertiary pl-8.5 text-[10px]">{row.hint}</div>
          </div>
        ),
      )}
    </div>
  );
}
