import { theme as antdTheme, ConfigProvider } from "antd";
import type { ReactNode } from "react";
import type { SkinPlugin } from "../skins";

const LIGHT_TOKENS = antdTheme.getDesignToken();
const DARK_TOKENS = antdTheme.getDesignToken({ algorithm: antdTheme.darkAlgorithm });

function DefaultProvider({
  isDark,
  compact,
  children,
}: {
  isDark: boolean;
  compact: boolean;
  children: ReactNode;
}) {
  const base = isDark ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm;
  const tokens = isDark ? DARK_TOKENS : LIGHT_TOKENS;
  return (
    <ConfigProvider
      theme={{
        algorithm: compact ? [base, antdTheme.compactAlgorithm] : base,
        token: tokens,
        components: {
          Layout: {
            bodyBg: tokens.colorBgLayout,
            headerBg: tokens.colorBgContainer,
            headerHeight: 56,
            siderBg: DARK_TOKENS.colorBgContainer,
          },
          Menu: {
            itemBorderRadius: 8,
            itemHeight: 40,
            darkItemBg: "transparent",
            darkSubMenuItemBg: "transparent",
          },
        },
      }}
    >
      {children}
    </ConfigProvider>
  );
}

const defaultSkin: SkinPlugin = {
  id: "default",
  name: "Default",
  bodyBg: {
    light: LIGHT_TOKENS.colorBgLayout,
    dark: DARK_TOKENS.colorBgLayout,
  },
  Provider: DefaultProvider,
};

export default defaultSkin;
