import { ThemeToggle } from "@cyber-ecosystem/shared-theme";
import { useFullscreen } from "ahooks";
import { App, Layout } from "antd";
import clsx from "clsx";
import { useAtomValue, useSetAtom } from "jotai";
import {
  LogOut,
  Maximize,
  Minimize,
  Moon,
  PanelLeftClose,
  PanelLeftOpen,
  Settings,
  Sun,
} from "lucide-react";
import { useLayoutEffect, useState } from "react";
import { LocaleSelect } from "#/domains/i18n";
import { m } from "#/paraglide/messages";
import { dashboardPreferencesStore } from "#/stores/dashboard-preferences/store";
import { dashboardPreferencesUiStore } from "#/stores/dashboard-preferences/ui-store";
import { useLogout } from "../auth/use-logout";
import { Breadcrumb } from "./breadcrumb";
import { ChipIconButton } from "./chip-icon-button";
import { SettingsDrawer } from "./settings/settings-drawer";

interface TopbarProps {
  className?: string;
}

export function Topbar({ className }: Readonly<TopbarProps>) {
  const { drawerOpen: settingsOpen } = useAtomValue(dashboardPreferencesUiStore.atom);
  const setPreferencesUi = useSetAtom(dashboardPreferencesUiStore.atom);
  const { sidebar } = useAtomValue(dashboardPreferencesStore.atom);
  const setPreferences = useSetAtom(dashboardPreferencesStore.atom);
  const { modal } = App.useApp();
  const logout = useLogout();
  const [isFullscreen, { isEnabled, toggleFullscreen }] = useFullscreen(
    () => document.documentElement,
  );
  const [isClient, setIsClient] = useState(false);
  useLayoutEffect(() => setIsClient(true), []);

  return (
    <Layout.Header
      className={clsx(
        "sticky top-0 z-40 flex items-center justify-between border-line-soft border-b px-4",
        className,
      )}
    >
      <div className="flex items-center gap-3">
        <ChipIconButton
          label={m.layout_dashboard_topbar_sidebar()}
          onClick={() =>
            setPreferences((d) => {
              d.sidebar.visible = !d.sidebar.visible;
            })
          }
          tooltip={false}
        >
          {sidebar.visible ? (
            <PanelLeftClose aria-hidden className="size-3 rtl:-scale-x-100" />
          ) : (
            <PanelLeftOpen aria-hidden className="size-3 rtl:-scale-x-100" />
          )}
        </ChipIconButton>

        <Breadcrumb />
      </div>

      <div className="relative flex items-center gap-2 text-base">
        <LocaleSelect />

        <ThemeToggle>
          {({ isDark, toggle }) => (
            <ChipIconButton label={m.layout_dashboard_topbar_theme()} onClick={toggle}>
              {isDark ? (
                <Sun aria-hidden className="size-3" />
              ) : (
                <Moon aria-hidden className="size-3" />
              )}
            </ChipIconButton>
          )}
        </ThemeToggle>

        <ChipIconButton
          className="max-lg:hidden"
          disabled={isClient ? !isEnabled : false}
          label={m.layout_dashboard_topbar_fullscreen()}
          onClick={toggleFullscreen}
        >
          {isFullscreen ? (
            <Minimize aria-hidden className="size-3" />
          ) : (
            <Maximize aria-hidden className="size-3" />
          )}
        </ChipIconButton>

        <ChipIconButton
          label={m.layout_dashboard_topbar_settings()}
          onClick={() =>
            setPreferencesUi((d) => {
              d.drawerOpen = true;
            })
          }
        >
          <Settings aria-hidden className="size-3" />
        </ChipIconButton>

        <ChipIconButton
          label={m.layout_dashboard_logout()}
          onClick={() => {
            if (logout.isPending) return;
            modal.confirm({
              icon: null,
              title: m.layout_dashboard_logout(),
              content: m.layout_dashboard_logout_confirm_content(),
              okText: m.layout_dashboard_logout_confirm_ok(),
              cancelButtonProps: { variant: "filled", color: "default" },
              okButtonProps: { variant: "filled", color: "danger" },
              onOk: () => logout.mutateAsync(),
            });
          }}
        >
          <LogOut aria-hidden className="size-3" />
        </ChipIconButton>
      </div>

      <SettingsDrawer
        onClose={() =>
          setPreferencesUi((d) => {
            d.drawerOpen = false;
          })
        }
        open={settingsOpen}
      />
    </Layout.Header>
  );
}
