import type { ReactNode } from "react"

export interface SkinPlugin {
  id: string
  name: string
  bodyBg: {
    light: string
    dark: string
  }
  Provider: (props: {
    isDark: boolean
    compact: boolean
    children: ReactNode
  }) => ReactNode
}

export type ThemePreference = "light" | "dark"
