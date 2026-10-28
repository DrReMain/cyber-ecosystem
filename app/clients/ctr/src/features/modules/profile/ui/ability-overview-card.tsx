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
    <div className="border-line-accent-blue bg-identity rounded-card border px-4 py-[15px]">
      <div className="mb-[11px] flex items-center justify-between">
        <span className="text-ink text-[14.5px] font-extrabold">{m.profile_ability_title()}</span>
        <span className="text-ink-tertiary text-[10.5px]">
          {m.profile_ability_sub({ n: tested })}
        </span>
      </div>
      {rows.map((row) =>
        row.best !== undefined ? (
          <div className="py-[7px]" key={row.ability}>
            <div className="mb-1.5 flex items-center gap-2">
              <div
                aria-hidden
                className={`${row.gradient} shadow-inset-ring flex h-[26px] w-[26px] shrink-0 items-center justify-center rounded-[9px] text-[10px] font-bold text-white`}
              >
                {row.abilityCn.slice(0, 2)}
              </div>
              <span className="text-ink text-[12.5px] font-extrabold">{row.abilityCn}</span>
              <TierChip grade={row.grade} />
              <span className="text-ink-muted ml-auto text-[10.5px] tabular-nums">
                {m.profile_ability_best({ score: Math.round(row.best * 10) / 10 })}
              </span>
            </div>
            <div className="h-[6px] w-full overflow-hidden rounded-full bg-[rgba(55,65,81,0.6)]">
              <div
                className="h-full rounded-full bg-gradient-to-r from-[#2563EB] to-[#06B6D4]"
                style={{ width: `${row.pct ?? 0}%` }}
              />
            </div>
          </div>
        ) : (
          <div className="py-[7px] opacity-70" key={row.ability}>
            <div className="mb-1 flex items-center gap-2">
              <div
                aria-hidden
                className={`${row.gradient} shadow-inset-ring flex h-[26px] w-[26px] shrink-0 items-center justify-center rounded-[9px] text-[10px] font-bold text-white opacity-40`}
              >
                {row.abilityCn.slice(0, 2)}
              </div>
              <span className="text-[12.5px] font-extrabold text-[#6b7280]">{row.abilityCn}</span>
              <span className="rounded-full bg-[rgba(255,255,255,0.06)] px-2 py-0.5 text-[9.5px] font-semibold text-[#6b7280]">
                {m.profile_ability_untested()}
              </span>
            </div>
            <div className="text-ink-tertiary pl-[34px] text-[10px]">{row.hint}</div>
          </div>
        ),
      )}
    </div>
  );
}
