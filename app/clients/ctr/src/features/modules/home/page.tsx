import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { Bell } from "lucide-react";
import { toast } from "sonner";
import { SoonCard } from "#/components/soon-card";
import type { AmbientOrb } from "#/features/app/ambient/ambient-backdrop";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { TabBar } from "#/features/app/tabbar";
import { sessionQuery } from "#/features/auth/session.fn";
import { m } from "#/paraglide/messages";

/* The home page's orb pair (prototype home). */
const homeOrbs: readonly AmbientOrb[] = [
  { className: "top-[-60px] left-[-70px] size-[min(62vw,240px)]", tone: "blue" },
  { className: "top-[280px] right-[-70px] size-[min(51vw,200px)]", tone: "purple" },
];

export function HomePage() {
  const navigate = useNavigate();
  const session = useQuery({ ...sessionQuery, throwOnError: false });
  // One session truth: the probe carries the profile (it proves the
  // credential with getProfile), so the greet, the nickname, and the games
  // gate all read it — profile writes invalidate it, and the dehydrated
  // SSR verdict delivers the final state on first paint.
  const authedSession = session.data?.status === "authed" ? session.data : undefined;
  const authed = authedSession !== undefined;
  const profile = authedSession?.profile;
  const hasGames = (profile?.games ?? []).length > 0;

  function onStart() {
    if (!authed) {
      void navigate({ to: "/login", search: { next: "/" } });
      return;
    }
    if (!hasGames) {
      void navigate({ to: "/profile/setup/games" });
      return;
    }
    void navigate({ to: "/match" });
  }

  return (
    <AmbientLayout chrome={<TabBar active="home" />} orbs={homeOrbs}>
      <div className="flex w-full flex-col gap-3 pb-20 md:pb-0">
        <div className="flex items-start justify-between">
          <div>
            <div className="text-ink-tertiary text-[11px]">
              {authed ? m.home_greet_logged() : m.home_greet_guest()}
            </div>
            <h1 className="text-ink mt-0.5 text-xl font-black">
              {authed && profile?.nickname != null ? profile.nickname : m.home_guest_name()}
            </h1>
          </div>
          <button
            aria-label={m.profile_msg()}
            className="bg-surface-input text-ink-secondary flex h-10 w-10 items-center justify-center rounded-full active:opacity-60"
            onClick={() => toast.info(m.common_soon())}
            type="button"
          >
            <Bell aria-hidden size={20} />
          </button>
        </div>

        <div className="border-line bg-surface-card shadow-card rounded-card border px-4 py-[18px]">
          <div className="text-ink text-base font-extrabold">{m.home_cta_title()}</div>
          <p className="text-ink-tertiary mt-[5px] text-[11.5px] leading-[1.7]">
            {authed && !hasGames ? m.home_cta_desc_need_games() : m.home_cta_desc_ready()}
          </p>
          <button
            className="bg-cta-primary shadow-cta-primary rounded-control mt-3.5 w-full px-4 py-[14px] text-[15px] font-bold text-white active:opacity-80"
            onClick={onStart}
            type="button"
          >
            {!authed ? m.home_cta_login() : !hasGames ? m.home_cta_pick() : m.home_cta_start()}
          </button>
          {!(authed && hasGames) && (
            <p className="text-ink-tertiary mt-[9px] text-center text-[11px]">
              {!authed ? m.home_hint_login() : m.home_hint_games()}
            </p>
          )}
        </div>

        <SoonCard description={m.home_soon_desc()} title={m.home_soon_title()} />
      </div>
    </AmbientLayout>
  );
}
