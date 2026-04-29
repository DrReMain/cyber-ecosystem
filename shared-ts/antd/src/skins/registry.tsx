import type { ReactNode } from "react"
import type { SkinPlugin } from "./types"
import defaultSkin from "../skin-presets/default"
import doodleSkin from "../skin-presets/doodle"
import heroSkin from "../skin-presets/hero"
import illustrationSkin from "../skin-presets/illustration"
import lumaSkin from "../skin-presets/luma"
import lyraSkin from "../skin-presets/lyra"

const SKINS: Record<string, SkinPlugin> = {}
for (const skin of [defaultSkin, doodleSkin, heroSkin, illustrationSkin, lumaSkin, lyraSkin]) {
  SKINS[skin.id] = skin
}

export function getSkin(id: string): SkinPlugin | undefined {
  return SKINS[id]
}

export function getAllSkins(): SkinPlugin[] {
  return Object.values(SKINS)
}

export function getSkinIds(): string[] {
  return Object.keys(SKINS)
}

export function SkinSwitcher({
  skinId,
  isDark,
  compact,
  children,
}: {
  skinId: string
  isDark: boolean
  compact: boolean
  children: ReactNode
}) {
  const skin = getSkin(skinId)
  const Provider = skin?.Provider
  if (!Provider) return <>{children}</>
  return (
    <Provider isDark={isDark} compact={compact}>
      {children}
    </Provider>
  )
}
