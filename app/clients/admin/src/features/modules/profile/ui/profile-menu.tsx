import { Link } from "@tanstack/react-router";
import clsx from "clsx";
import type { LucideIcon } from "lucide-react";

export interface ProfileMenuItem {
  icon: LucideIcon;
  label: string;
  path: string;
}

interface ProfileMenuProps {
  activePath: string;
  items: ProfileMenuItem[];
}

export function ProfileMenu({ activePath, items }: Readonly<ProfileMenuProps>) {
  return (
    <nav className="flex flex-col gap-1">
      {items.map((item) => {
        const active = activePath === item.path;
        const Icon = item.icon;
        return (
          <Link
            aria-current={active ? "page" : undefined}
            className={clsx(
              "flex items-center gap-2.5 rounded-md px-3 py-2 text-[13px] transition-colors",
              active
                ? "bg-fill-secondary font-medium text-primary"
                : "text-ink-secondary hover:bg-fill-secondary hover:text-ink",
            )}
            key={item.path}
            to={item.path as never}
          >
            <Icon aria-hidden className="size-4 flex-none" />
            <span className="truncate">{item.label}</span>
          </Link>
        );
      })}
    </nav>
  );
}
