import { theme as antdTheme, ConfigProvider } from "antd"
import type { ReactNode } from "react"
import type { SkinPlugin } from "../skins/types"

function DefaultProvider({
  isDark,
  compact,
  children,
}: {
  isDark: boolean
  compact: boolean
  children: ReactNode
}) {
  const base = isDark ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm
  return (
    <ConfigProvider
      theme={{
        algorithm: compact ? [base, antdTheme.compactAlgorithm] : base,
      }}
    >
      {children}
    </ConfigProvider>
  )
}

const defaultSkin: SkinPlugin = {
  id: "default",
  name: "Default",
  bodyBg: {
    light: "#ffffff",
    dark: "#000000",
  },
  Provider: DefaultProvider,
}

export default defaultSkin
