import { useRouteContext } from "@tanstack/react-router";
import { theme as antdTheme, Watermark } from "antd";
import { useAtomValue } from "jotai";
import type { PropsWithChildren } from "react";
import { APP_NAME } from "#/config";
import { dashboardPreferencesStore } from "#/stores/dashboard-preferences/store";
import { AREA_PATH } from "../area";

export function ContentWatermark({ children }: Readonly<PropsWithChildren>) {
  const { watermark } = useAtomValue(dashboardPreferencesStore.atom);
  const { user } = useRouteContext({ from: AREA_PATH });
  const { token } = antdTheme.useToken();

  if (!watermark.enabled) {
    return children;
  }

  return (
    <Watermark
      className="flex min-h-full w-full flex-col"
      content={[APP_NAME, user.email ?? ""]}
      font={{ color: token.colorTextQuaternary }}
      zIndex={20}
    >
      {children}
    </Watermark>
  );
}
