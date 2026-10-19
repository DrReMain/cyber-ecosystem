import { useTheme } from "@cyber-ecosystem/shared-theme";
import { Layout } from "antd";
import clsx from "clsx";
import { useAtomValue } from "jotai";
import { useEffect, useState } from "react";
import {
  dashboardPreferencesStore,
  SIDEBAR_COLLAPSED_WIDTH,
} from "#/stores/dashboard-preferences/store";
import { useIsNarrowViewport } from "../use-viewport";
import { SiderChrome } from "./sider-chrome";

export function FloatingSider() {
  const isDark = useTheme().preference === "dark";
  const { sidebar } = useAtomValue(dashboardPreferencesStore.atom);
  const narrow = useIsNarrowViewport();
  const [hoverExpanded, setHoverExpanded] = useState(false);

  useEffect(() => {
    if (narrow) setHoverExpanded(false);
  }, [narrow]);

  if (!sidebar.visible) return null;

  return (
    <Layout.Sider
      className={clsx(
        "fixed inset-s-0 top-0 bottom-0 z-50 border-line-soft border-e",
        hoverExpanded && "shadow-xl",
      )}
      collapsed={!hoverExpanded}
      collapsedWidth={SIDEBAR_COLLAPSED_WIDTH}
      onMouseEnter={() => {
        if (!narrow) setHoverExpanded(true);
      }}
      onMouseLeave={() => setHoverExpanded(false)}
      theme={isDark ? "dark" : "light"}
      width={sidebar.width}
    >
      <div className="flex h-full flex-col">
        <SiderChrome
          collapsed={!hoverExpanded}
          onToggleCollapse={() => setHoverExpanded((v) => !v)}
        />
      </div>
    </Layout.Sider>
  );
}
