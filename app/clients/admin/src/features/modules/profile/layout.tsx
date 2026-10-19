import { useQuery } from "@tanstack/react-query";
import { Outlet, useRouterState } from "@tanstack/react-router";
import clsx from "clsx";
import { useAtomValue } from "jotai";
import { Bot, KeyRound } from "lucide-react";
import { useFileUrl } from "#/features/app/file/use-file-url";
import { m } from "#/paraglide/messages";
import { dashboardPreferencesUiStore } from "#/stores/dashboard-preferences/ui-store";
import { sessionQuery } from "../../layout-dashboard/auth/session.fn";
import { ProfileIdentity } from "./ui/profile-identity";
import { ProfileMenu, type ProfileMenuItem } from "./ui/profile-menu";

export function ProfileLayout() {
  const { data } = useQuery(sessionQuery);
  const user = data?.status === "authed" ? data.user : undefined;
  const email = user?.email ?? "";
  const name = email.split("@")[0] ?? "";
  const avatarUrl = useFileUrl(user?.avatar);
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const { maximized } = useAtomValue(dashboardPreferencesUiStore.atom);

  const items: ProfileMenuItem[] = [
    { icon: KeyRound, label: m.profile_password_title(), path: "/dashboard/profile/password" },
    { icon: Bot, label: m.profile_agents_title(), path: "/dashboard/profile/agents" },
  ];

  return (
    <div className="grid items-start gap-4 p-4 lg:grid-cols-[264px_minmax(0,1fr)]">
      <aside
        className={clsx(
          "flex h-fit flex-col gap-4 lg:sticky",
          maximized ? "lg:top-13.5" : "lg:top-27.5",
        )}
      >
        <ProfileIdentity avatarUrl={avatarUrl} email={email} name={name} />
        <ProfileMenu activePath={pathname} items={items} />
      </aside>
      <div className="min-w-0">
        <Outlet />
      </div>
    </div>
  );
}
