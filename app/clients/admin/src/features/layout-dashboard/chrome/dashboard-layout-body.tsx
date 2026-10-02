import { useRouteContext } from "@tanstack/react-router";
import clsx from "clsx";
import { useAtomValue, useSetAtom } from "jotai";
import { useEffect } from "react";
import { dashboardPreferencesStore } from "#/stores/dashboard-preferences/store";
import { dashboardPreferencesUiStore } from "#/stores/dashboard-preferences/ui-store";
import { AREA_PATH } from "../area";
import { FixedLayout } from "./fixed-layout";
import { MainArea } from "./main-area";
import { FloatingSider } from "./sider/floating-sider";
import { reconcileTabSession } from "./tab/tab-session";
import { useIsNarrowViewport } from "./use-viewport";

export function DashboardLayoutBody() {
  const { user } = useRouteContext({ from: AREA_PATH });
  const { sidebar } = useAtomValue(dashboardPreferencesStore.atom);
  const { maximized } = useAtomValue(dashboardPreferencesUiStore.atom);
  const setPreferences = useSetAtom(dashboardPreferencesStore.atom);
  const narrow = useIsNarrowViewport();

  useEffect(() => {
    reconcileTabSession(user.id);
  }, [user.id]);

  if (sidebar.fixed) {
    return (
      <FixedLayout
        collapsed={sidebar.collapsed || narrow}
        maximized={maximized}
        onToggleCollapse={() =>
          setPreferences((d) => {
            d.sidebar.collapsed = !d.sidebar.collapsed;
          })
        }
        visible={sidebar.visible}
      />
    );
  }

  return (
    <>
      <div className={maximized ? "hidden" : "contents"}>
        <FloatingSider />
      </div>
      {/* ms-16 (4rem) pairs with SIDEBAR_COLLAPSED_WIDTH in stores/preferences */}
      <MainArea className={clsx(sidebar.visible && !maximized && "ms-16")} />
    </>
  );
}
