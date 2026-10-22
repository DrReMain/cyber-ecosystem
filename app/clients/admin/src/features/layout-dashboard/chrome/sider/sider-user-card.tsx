import { useQuery } from "@tanstack/react-query";
import { useRouter } from "@tanstack/react-router";
import { Avatar, Tooltip } from "antd";
import clsx from "clsx";
import { useFileUrl } from "#/features/app/file/use-file-url";
import { m } from "#/paraglide/messages";
import { getTextDirection } from "#/paraglide/runtime";
import { sessionQuery } from "../../auth/session.fn";

export function SiderUserCard({ collapsed }: Readonly<{ collapsed: boolean }>) {
  const router = useRouter();
  const { data } = useQuery(sessionQuery);
  const account = data?.status === "authed" ? data.user.account : "";
  const name = account.split("@")[0] ?? "";
  const avatarUrl = useFileUrl(data?.status === "authed" ? data.user.avatar : undefined);

  return (
    <button
      aria-label={m.layout_dashboard_sidebar_profile()}
      className={clsx(
        "border-line-soft hover:bg-fill-secondary focus-visible:outline-primary flex h-16 w-full flex-none cursor-pointer items-center border-t px-4 text-start transition-colors focus-visible:outline-2",
        collapsed ? "justify-center px-0" : "gap-2.5",
      )}
      onClick={() => router.navigate({ to: "/dashboard/profile" })}
      type="button"
    >
      <Tooltip
        placement={getTextDirection() === "rtl" ? "left" : "right"}
        title={account || undefined}
      >
        <Avatar className="flex-none" size={collapsed ? 30 : 34} src={avatarUrl}>
          {name.charAt(0).toUpperCase()}
        </Avatar>
      </Tooltip>
      {collapsed ? null : (
        <div className="flex min-w-0 flex-col">
          <span className="truncate text-[13px] font-medium">{name}</span>
          <span className="text-ink-tertiary truncate font-mono text-[11px]">{account}</span>
        </div>
      )}
    </button>
  );
}
