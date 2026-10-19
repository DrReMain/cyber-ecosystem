import { Splitter } from "antd";
import clsx from "clsx";
import { useAtomValue, useSetAtom } from "jotai";
import { useEffect, useState } from "react";
import {
  dashboardPreferencesStore,
  SIDEBAR_COLLAPSED_WIDTH,
  SIDEBAR_WIDTH_MAX,
  SIDEBAR_WIDTH_MIN,
  sidebarWidthAtom,
} from "#/stores/dashboard-preferences/store";
import { MainArea } from "./main-area";
import { InlineSider } from "./sider/inline-sider";

interface FixedLayoutProps {
  collapsed: boolean;
  maximized: boolean;
  onToggleCollapse: () => void;
  visible: boolean;
}

export function FixedLayout({
  collapsed,
  maximized,
  onToggleCollapse,
  visible,
}: Readonly<FixedLayoutProps>) {
  const { sidebar } = useAtomValue(dashboardPreferencesStore.atom);
  const setSidebarWidth = useSetAtom(sidebarWidthAtom);
  const [liveWidth, setLiveWidth] = useState(sidebar.width);
  useEffect(() => {
    setLiveWidth(sidebar.width);
  }, [sidebar.width]);
  const siderHidden = !visible || maximized;

  return (
    <Splitter
      className={clsx(
        "h-svh",
        "[&_.ant-splitter-bar]:z-99",
        siderHidden && "[&_.ant-splitter-bar]:hidden",
      )}
      onResize={([w]) => {
        if (typeof w === "number") setLiveWidth(w);
      }}
      onResizeEnd={([w]) => {
        if (typeof w === "number") setSidebarWidth(w);
      }}
    >
      <Splitter.Panel
        className={siderHidden ? "hidden" : undefined}
        max={collapsed ? SIDEBAR_COLLAPSED_WIDTH : SIDEBAR_WIDTH_MAX}
        min={collapsed ? SIDEBAR_COLLAPSED_WIDTH : SIDEBAR_WIDTH_MIN}
        resizable={!collapsed && visible}
        size={collapsed ? SIDEBAR_COLLAPSED_WIDTH : liveWidth}
      >
        <InlineSider collapsed={collapsed} onToggleCollapse={onToggleCollapse} />
      </Splitter.Panel>
      {/* grow! must beat antd's inline flex sizing: after a drag, both panels carry
          inline sizes (flexGrow 0), leaving a gap without the !important */}
      <Splitter.Panel className={siderHidden ? "grow!" : undefined}>
        <MainArea />
      </Splitter.Panel>
    </Splitter>
  );
}
