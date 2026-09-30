import { useTheme } from "@cyber-ecosystem/shared-theme";
import { useMatches, useRouter, useRouterState } from "@tanstack/react-router";
import { theme as antdTheme, ConfigProvider, Menu } from "antd";
import { useAtomValue } from "jotai";
import { useEffect, useMemo, useState } from "react";
import { dashboardPreferencesStore } from "#/stores/dashboard-preferences/store";
import { chainKeysOf } from "../../protocol/chain-keys";
import { navTo, useNavNodes } from "../../protocol/nav";
import { toMenuItems } from "../menu-items";

export function SiderMenu({ collapsed }: Readonly<{ collapsed: boolean }>) {
  const router = useRouter();
  const isDark = useTheme().preference === "dark";
  const { token } = antdTheme.useToken();
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const matches = useMatches();
  const { sidebar } = useAtomValue(dashboardPreferencesStore.atom);
  const chainKeys = useMemo(() => chainKeysOf(matches), [matches]);
  const [openKeys, setOpenKeys] = useState<string[]>(chainKeys);

  useEffect(() => {
    setOpenKeys((prev) =>
      sidebar.accordion ? chainKeys : Array.from(new Set([...prev, ...chainKeys])),
    );
  }, [chainKeys, sidebar.accordion]);

  const navNodes = useNavNodes();
  const items = useMemo(() => toMenuItems(navNodes, collapsed), [navNodes, collapsed]);

  return (
    <ConfigProvider theme={{ components: { Menu: { darkPopupBg: token.colorBgContainer } } }}>
      <Menu
        inlineCollapsed={collapsed}
        items={items}
        mode="inline"
        onClick={({ key }) => navTo(router, key)}
        onOpenChange={(keys) => {
          if (collapsed) return;
          if (!sidebar.accordion) {
            setOpenKeys(keys);
            return;
          }
          const opened = keys.find((key) => !openKeys.includes(key));
          setOpenKeys(opened ? [opened] : []);
        }}
        openKeys={collapsed ? undefined : openKeys}
        selectedKeys={[pathname]}
        style={{ backgroundColor: "transparent", borderInlineEnd: "none", width: "100%" }}
        theme={isDark ? "dark" : "light"}
      />
    </ConfigProvider>
  );
}
