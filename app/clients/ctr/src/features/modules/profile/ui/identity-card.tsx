import type { Profile as ProfilePB } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile_pb";
import { User } from "lucide-react";
import { toast } from "sonner";
import { m } from "#/paraglide/messages";
import { bandLabel } from "../bands";

/**
 * The profile overview's header card: avatar chip, nickname, quick-action
 * pills, and the band / preference meta line.
 */
export function IdentityCard({ profile }: Readonly<{ profile: ProfilePB | undefined }>) {
  const games = profile?.games ?? [];
  const meta = [
    bandLabel(profile?.ageBand),
    games.length > 0 ? m.profile_meta_games({ n: games.length }) : null,
  ]
    .filter((part): part is string => part !== null && part !== undefined)
    .join(" · ");
  return (
    <div className="border-line-accent-blue bg-identity rounded-card border px-4 py-[15px]">
      <div className="flex items-center gap-3">
        <div className="bg-logo shadow-inset-ring flex h-[46px] w-[46px] shrink-0 items-center justify-center rounded-xl">
          <User aria-hidden className="text-white" size={20} />
        </div>
        <div className="min-w-0 flex-1">
          <div className="text-ink truncate text-base font-black">
            {profile?.nickname ?? m.profile_unset_nick()}
          </div>
          <div className="mt-[7px] flex gap-[7px]">
            {[m.profile_msg(), m.profile_feedback()].map((label) => (
              <button
                className="tap-zone bg-ink/8 text-ink-secondary rounded-full px-2.5 py-[3px] text-[10.5px] font-semibold active:opacity-60"
                key={label}
                onClick={() => toast.info(m.common_soon())}
                type="button"
              >
                {label}
              </button>
            ))}
          </div>
        </div>
      </div>
      {meta !== "" && <p className="text-ink-tertiary mt-[11px] text-[11px]">{meta}</p>}
      {/* The prototype's four-stat strip. Values stay "—" until the records
       * read API lands (P18) — unknown, honestly shown, never zero-asserted. */}
      <div className="mt-3 flex border-t border-[rgba(55,65,81,0.55)] pt-[11px]">
        {[
          { color: "text-ink", label: m.profile_stat_total() },
          { color: "text-status-ok", label: m.profile_stat_comp() },
          { color: "text-[#fde047]", label: m.profile_stat_best() },
          { color: "text-ink", label: m.profile_stat_prog() },
        ].map((cell) => (
          <div className="flex-1 text-center" key={cell.label}>
            <div className={`text-[19px] leading-[1.15] font-black ${cell.color}`}>—</div>
            <div className="text-ink-tertiary mt-[3px] text-[10.5px]">{cell.label}</div>
          </div>
        ))}
      </div>
    </div>
  );
}
