import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import { getProfile } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile-ProfileService_connectquery";
import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { Bell } from "lucide-react";
import { toast } from "sonner";
import { SoonCard } from "#/components/soon-card";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { TabBar } from "#/features/app/tabbar";
import { sessionQuery } from "#/features/auth/session.fn";
import { m } from "#/paraglide/messages";

export function HomePage() {
  const navigate = useNavigate();
  const session = useQuery({ ...sessionQuery, throwOnError: false });
  const authed = session.data?.status === "authed";
  // The profile fetch rides the session verdict — anonymous never fires it,
  // so the member probe cannot 401 its way into the route boundary. While
  // authed-but-loading the greet still shows the guest name for a beat.
  const profileQuery = useConnectQuery(getProfile, {}, { enabled: authed });
  const profile = profileQuery.data?.profile;
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
    <AmbientLayout>
      <div className="flex w-full flex-col gap-3 pb-20">
        <div className="flex items-start justify-between">
          <div>
            <div className="text-[11px] text-ink-tertiary">
              {authed ? m.home_greet_logged() : m.home_greet_guest()}
            </div>
            <h1 className="mt-0.5 font-black text-ink text-xl">
              {authed && profile?.nickname != null ? profile.nickname : m.home_guest_name()}
            </h1>
          </div>
          <button
            aria-label={m.profile_msg()}
            className="flex h-11 w-11 items-center justify-center rounded-full border border-line text-ink-secondary active:opacity-60"
            onClick={() => toast.info(m.common_soon())}
            type="button"
          >
            <Bell aria-hidden size={17} />
          </button>
        </div>

        <div className="rounded-lg border border-line bg-surface-card p-4">
          <div className="font-extrabold text-base text-ink">{m.home_cta_title()}</div>
          <p className="mt-1 text-[11.5px] text-ink-tertiary leading-relaxed">
            {authed && !hasGames ? m.home_cta_desc_need_games() : m.home_cta_desc_ready()}
          </p>
          <button
            className="mt-3.5 w-full rounded-sm bg-cta-primary px-4 py-3 font-semibold text-white active:opacity-80"
            onClick={onStart}
            type="button"
          >
            {!authed ? m.home_cta_login() : !hasGames ? m.home_cta_pick() : m.home_cta_start()}
          </button>
          {!(authed && hasGames) && (
            <p className="mt-2 text-center text-[11px] text-ink-tertiary">
              {!authed ? m.home_hint_login() : m.home_hint_games()}
            </p>
          )}
        </div>

        <SoonCard description={m.home_soon_desc()} title={m.home_soon_title()} />
      </div>
      <TabBar active="home" />
    </AmbientLayout>
  );
}
