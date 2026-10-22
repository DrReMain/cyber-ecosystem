import { useQuery } from "@connectrpc/connect-query";
import { getProfile } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile-ProfileService_connectquery";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { toast } from "sonner";
import { SoonCard } from "#/components/soon-card";
import type { AmbientOrb } from "#/features/app/ambient/ambient-backdrop";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { StateScreen } from "#/features/app/states/state-screen";
import { TabBar } from "#/features/app/tabbar";
import { PROFILE_STALE_MS } from "#/features/auth/profile-cache";
import { broadcastSessionKick } from "#/features/auth/session-kick";
import { m } from "#/paraglide/messages";
import { logoutFn } from "./logout.fn";
import { LogoutSheet } from "./logout-sheet";
import { DataCard } from "./ui/data-card";
import { IdentityCard } from "./ui/identity-card";
import { SecurityCard } from "./ui/security-card";

/* The overview page's orb pair (prototype profile-overview). */
const overviewOrbs: readonly AmbientOrb[] = [
  { className: "top-[-70px] left-[-60px] size-[min(62vw,240px)]", tone: "purple" },
  { className: "bottom-[-50px] right-[-60px] size-[min(51vw,200px)]", tone: "cyan" },
];

export function ProfilePage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  // Unauthenticated never mounts this page: the /profile layout's session
  // guard (features/auth/guard.ts) redirects in beforeLoad, so this query
  // only ever fires authed. Errors escalate to the route boundary per the
  // global throwOnError — this page owns only the pending face.
  const profileQuery = useQuery(getProfile, {}, { staleTime: PROFILE_STALE_MS });
  const [loggingOut, setLoggingOut] = useState(false);
  const [sheetOpen, setSheetOpen] = useState(false);

  if (profileQuery.isPending) {
    return (
      <AmbientLayout orbs={overviewOrbs}>
        <StateScreen variant="loading" />
      </AmbientLayout>
    );
  }

  const profile = profileQuery.data?.profile ?? undefined;

  async function onLogout() {
    setLoggingOut(true);
    try {
      await logoutFn();
      queryClient.clear();
      // Boots the other tabs off the dead session (no expired notice — a
      // deliberate sign-out is expected, not a surprise).
      broadcastSessionKick();
      await navigate({ to: "/login" });
    } finally {
      setLoggingOut(false);
      setSheetOpen(false);
    }
  }

  return (
    <AmbientLayout orbs={overviewOrbs}>
      <div className="flex w-full flex-col gap-3 pb-20">
        <div className="flex items-center justify-between">
          <h1 className="text-ink text-[19px] font-bold">{m.profile_title()}</h1>
          <button
            className="tap-zone text-danger-text rounded-full border border-[rgba(239,68,68,0.6)] px-[11px] py-1 text-[11.5px] font-extrabold active:opacity-60 disabled:opacity-50"
            disabled={loggingOut}
            onClick={() => setSheetOpen(true)}
            type="button"
          >
            {m.profile_logout()}
          </button>
        </div>

        <IdentityCard profile={profile} />
        <DataCard profile={profile} />
        <SecurityCard />
        <SoonCard description={m.profile_soon_desc()} title={m.profile_soon_title()} />

        <div className="border-line-accent-blue bg-identity rounded-card border px-4 py-[15px]">
          <div className="text-ink mb-2 text-[14.5px] font-extrabold">{m.profile_about()}</div>
          <div className="text-ink-muted flex gap-[14px] text-xs">
            {[m.login_agree_user_name(), m.login_agree_privacy_name(), m.profile_about_icp()].map(
              (label) => (
                <button
                  className="tap-zone py-1 underline active:opacity-60"
                  key={label}
                  onClick={() => toast.info(m.common_soon())}
                  type="button"
                >
                  {label}
                </button>
              ),
            )}
          </div>
        </div>

        <div className="text-ink-tertiary mt-2 text-center text-[10.5px] leading-[1.95]">
          {m.login_footer_company()}
          <br />
          {m.login_footer_icp()}
          <br />
          {m.login_footer_email()}
        </div>
      </div>
      <TabBar active="me" />
      <LogoutSheet
        onCancel={() => setSheetOpen(false)}
        onConfirm={() => void onLogout()}
        open={sheetOpen}
      />
    </AmbientLayout>
  );
}
