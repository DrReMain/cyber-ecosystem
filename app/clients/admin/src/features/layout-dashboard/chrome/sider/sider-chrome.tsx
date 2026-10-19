import { useTheme } from "@cyber-ecosystem/shared-theme";
import { Image } from "@unpic/react";
import clsx from "clsx";
import { useAtomValue, useSetAtom } from "jotai";
import { ChevronsLeft, ChevronsRight, Pin, PinOff } from "lucide-react";
import { APP_NAME } from "#/config";
import { m } from "#/paraglide/messages";
import { dashboardPreferencesStore } from "#/stores/dashboard-preferences/store";
import { ScrollFade } from "../scroll-fade";
import { useOverflowEdges } from "../use-overflow-edges";
import { SiderMenu } from "./sider-menu";
import { SiderUserCard } from "./sider-user-card";

const TRIGGER_BTN =
  "inline-flex size-7 cursor-pointer items-center justify-center rounded-md text-ink-tertiary transition-colors hover:bg-fill-secondary hover:text-ink focus-visible:outline-2 focus-visible:outline-primary";

interface SiderChromeProps {
  collapsed: boolean;
  onToggleCollapse: () => void;
}

export function SiderChrome({ collapsed, onToggleCollapse }: Readonly<SiderChromeProps>) {
  const isDark = useTheme().preference === "dark";
  const { sidebar } = useAtomValue(dashboardPreferencesStore.atom);
  const setPreferences = useSetAtom(dashboardPreferencesStore.atom);
  const {
    ref: menuRef,
    overflow: menuOverflow,
    start: menuAtTop,
    end: menuAtBottom,
  } = useOverflowEdges("y");

  return (
    <>
      <div
        className={clsx(
          "flex h-14 flex-none items-center border-line-soft border-b font-mono font-semibold text-[12.5px] tracking-[0.2em]",
          collapsed ? "justify-center" : "gap-2.75 px-4",
        )}
      >
        <Image
          alt=""
          aria-hidden
          className="size-5.5 flex-none"
          height={22}
          src={isDark ? "/mark-white.svg" : "/mark-black.svg"}
          width={22}
        />
        {collapsed ? null : <span>{APP_NAME}</span>}
      </div>

      {/* Menu region - many entries must scroll within the viewport. */}
      <div className="relative min-h-0 flex-1">
        <div className="h-full overflow-y-auto py-2" ref={menuRef}>
          <SiderMenu collapsed={collapsed} />
        </div>
        <ScrollFade axis="y" show={menuOverflow && !menuAtTop} side="start" />
        <ScrollFade axis="y" show={menuOverflow && !menuAtBottom} side="end" />
      </div>

      {/* Reserved slot: signed-in identity card. */}
      <SiderUserCard collapsed={collapsed} />

      <div
        className={clsx(
          "hidden h-12 flex-none items-center border-line-soft border-t px-3 lg:flex",
          collapsed ? "justify-center" : "justify-between",
        )}
      >
        <button
          aria-label={
            collapsed ? m.layout_dashboard_sidebar_expand() : m.layout_dashboard_sidebar_collapse()
          }
          className={TRIGGER_BTN}
          onClick={onToggleCollapse}
          type="button"
        >
          {collapsed ? (
            <ChevronsRight aria-hidden className="size-4 rtl:-scale-x-100" />
          ) : (
            <ChevronsLeft aria-hidden className="size-4 rtl:-scale-x-100" />
          )}
        </button>

        {collapsed ? null : (
          <button
            aria-label={
              sidebar.fixed ? m.layout_dashboard_sidebar_pin() : m.layout_dashboard_sidebar_unpin()
            }
            className={TRIGGER_BTN}
            onClick={() =>
              setPreferences((d) => {
                d.sidebar.fixed = !d.sidebar.fixed;
              })
            }
            type="button"
          >
            {sidebar.fixed ? (
              <Pin aria-hidden className="size-4" />
            ) : (
              <PinOff aria-hidden className="size-4" />
            )}
          </button>
        )}
      </div>
    </>
  );
}
