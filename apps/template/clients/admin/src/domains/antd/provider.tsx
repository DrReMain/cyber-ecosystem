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
            // Align antd breakpoints with Tailwind by name and value
            // xs=0 (base/mobile) sm=640 md=768 lg=1024 xl=1280 xxl=1536
            screenXS: 0,
            screenXSMin: 0,
            screenXSMax: 639,
            screenSM: 640,
            screenSMMin: 640,
            screenSMMax: 767,
            screenMD: 768,
            screenMDMin: 768,
            screenMDMax: 1023,
            screenLG: 1024,
            screenLGMin: 1024,
            screenLGMax: 1279,
            screenXL: 1280,
            screenXLMin: 1280,
            screenXLMax: 1535,
            screenXXL: 1536,
            screenXXLMin: 1536,
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
