import { useQuery } from "@tanstack/react-query";
import { Avatar, Tooltip } from "antd";
import clsx from "clsx";
import { useFileUrl } from "#/features/app/file/use-file-url";
import { getTextDirection } from "#/paraglide/runtime";
import { sessionQuery } from "../../auth/session.fn";

export function SiderUserCard({ collapsed }: Readonly<{ collapsed: boolean }>) {
  const { data } = useQuery(sessionQuery);
  const email = data?.status === "authed" ? data.user.email : "";
  const name = email.split("@")[0] ?? "";
  const avatarUrl = useFileUrl(data?.status === "authed" ? data.user.avatar : undefined);

  return (
    <div
      className={clsx(
        "flex h-16 flex-none items-center border-line-soft border-t px-4",
        collapsed ? "justify-center px-0" : "gap-2.5",
      )}
    >
      <Tooltip
        placement={getTextDirection() === "rtl" ? "left" : "right"}
        title={email || undefined}
      >
        <Avatar className="flex-none" size={collapsed ? 30 : 34} src={avatarUrl}>
          {name.charAt(0).toUpperCase()}
        </Avatar>
      </Tooltip>
      {collapsed ? null : (
        <div className="flex min-w-0 flex-col">
          <span className="truncate font-medium text-[13px]">{name}</span>
          <span className="truncate font-mono text-[11px] text-ink-tertiary">{email}</span>
        </div>
      )}
    </div>
  );
}
