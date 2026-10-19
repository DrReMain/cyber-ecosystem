import { Outlet, useRouterState } from "@tanstack/react-router";
import { FloatButton, Layout } from "antd";
import clsx from "clsx";
import { useAtomValue } from "jotai";
import { useRef, useState } from "react";
import { dashboardPreferencesUiStore } from "#/stores/dashboard-preferences/ui-store";
import { ContentWatermark } from "./content-watermark";
import { Tabbar } from "./tab/tabbar";
import { usePerTabScroll } from "./tab/use-per-tab-scroll";
import { Topbar } from "./topbar";
import { useOverflowEdges } from "./use-overflow-edges";

interface MainAreaProps {
  className?: string;
}

export function MainArea({ className }: Readonly<MainAreaProps>) {
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const { maximized } = useAtomValue(dashboardPreferencesUiStore.atom);
  const layoutRef = useRef<HTMLDivElement>(null);
  const [refreshTick, setRefreshTick] = useState(0);
  usePerTabScroll(layoutRef, pathname);
  const { overflow, start: atTop } = useOverflowEdges("y", layoutRef);

  return (
    <Layout className={clsx("h-svh overflow-auto", className)} ref={layoutRef}>
      <Topbar className={maximized ? "hidden" : undefined} />
      <Tabbar onRefresh={() => setRefreshTick((t) => t + 1)} scrolled={overflow && !atTop} />
      <Layout.Content key={refreshTick}>
        <ContentWatermark>
          <Outlet />
        </ContentWatermark>
      </Layout.Content>
      <FloatButton.BackTop
        style={{ position: "fixed", insetBlockEnd: 40, insetInlineEnd: 24 }}
        target={() => layoutRef.current ?? document.body}
      />
    </Layout>
  );
}
