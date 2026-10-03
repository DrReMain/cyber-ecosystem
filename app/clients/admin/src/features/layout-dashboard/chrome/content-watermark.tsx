import { useRouteContext } from "@tanstack/react-router";
import { theme as antdTheme, Watermark } from "antd";
import { useAtomValue } from "jotai";
import type { PropsWithChildren } from "react";
import { APP_NAME } from "#/config";
import { watermarkEnabledAtom } from "#/stores/dashboard-preferences/store";
import { AREA_PATH } from "../area";

export function ContentWatermark({ children }: Readonly<PropsWithChildren>) {
  const enabled = useAtomValue(watermarkEnabledAtom);
  const { user } = useRouteContext({ from: AREA_PATH });
  const { token } = antdTheme.useToken();

  return (
    <Watermark
      className="overflow-visible! flex min-h-full w-full flex-col"
      content={enabled ? [APP_NAME, user.email ?? ""] : undefined}
      font={{ color: token.colorTextQuaternary }}
      zIndex={20}
    >
      {children}
    </Watermark>
  );
}
