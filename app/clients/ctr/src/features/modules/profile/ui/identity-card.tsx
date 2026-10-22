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
    <div className="border-line-accent-blue bg-identity rounded-2xl border p-4">
      <div className="flex items-center gap-3">
        <div className="bg-cta-primary flex h-11 w-11 shrink-0 items-center justify-center rounded-xl">
          <User aria-hidden className="text-white" size={20} />
        </div>
        <div className="min-w-0 flex-1">
          <div className="text-ink truncate text-base font-black">
            {profile?.nickname ?? m.profile_unset()}
          </div>
          <div className="mt-1.5 flex gap-1.5">
            {[m.profile_msg(), m.profile_feedback()].map((label) => (
              <button
                className="tap-zone bg-ink/8 text-ink-secondary rounded-full px-3 py-1 text-xs active:opacity-60"
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
      {meta !== "" && <p className="text-ink-tertiary mt-3 text-[11px]">{meta}</p>}
    </div>
  );
}
