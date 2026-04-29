import { App, ConfigProvider } from "antd"
import type { PropsWithChildren } from "react"
import { useTheme } from "#/domains/theme"
import { useAntdLocale } from "./locale"
import { AntdRegistry } from "./registry"
import showInsetEffect from "./show-inset-effect"
import { SkinSwitcher } from "./skins"

export function AntdProvider({ children }: Readonly<PropsWithChildren>) {
  const { skinId, preference, compact } = useTheme()
  const isDark = preference === "dark"
  const { locale, direction } = useAntdLocale()

  return (
    <AntdRegistry>
      <ConfigProvider
        direction={direction}
        form={{ requiredMark: "optional" }}
        input={{ autoComplete: "off", allowClear: true }}
        locale={locale}
        theme={{
          token: {
            // Compatible with Tailwind: sm=640 md=768 lg=1024 xl=1280 2xl=1536
            // antd has 7 tiers (xs..xxxl), Tailwind has 5 (sm..2xl)
            // Mapping: xs=sm, sm=md, md=lg, lg=xl, xl=2xl, xxl/xxxl=antd defaults
            screenXS: 640,
            screenXSMin: 640,
            screenXSMax: 767,
            screenSM: 768,
            screenSMMin: 768,
            screenSMMax: 1023,
            screenMD: 1024,
            screenMDMin: 1024,
            screenMDMax: 1279,
            screenLG: 1280,
            screenLGMin: 1280,
            screenLGMax: 1535,
            screenXL: 1536,
            screenXLMin: 1536,
            screenXLMax: 1599,
            screenXXL: 1600,
            screenXXLMin: 1600,
            screenXXLMax: 1919,
          },
        }}
        wave={{ showEffect: showInsetEffect }}
      >
        <SkinSwitcher compact={compact} isDark={isDark} skinId={skinId}>
          <App>{children}</App>
        </SkinSwitcher>
      </ConfigProvider>
    </AntdRegistry>
  )
}
