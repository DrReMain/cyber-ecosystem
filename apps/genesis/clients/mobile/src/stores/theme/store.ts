import { atom } from "jotai"
import { Appearance } from "react-native"
import { z } from "zod"
import { storage } from "@/lib/mmkv"
import { defineStore } from "@/stores/_core/define-store"

const THEME_MODES = ["light", "dark", "system"] as const
export type ThemeMode = (typeof THEME_MODES)[number]
const ThemeModeSchema = z.enum(THEME_MODES)

function applyTheme(mode: ThemeMode) {
  Appearance.setColorScheme(mode === "system" ? "unspecified" : mode)
}

export function initTheme() {
  const raw = storage.getString("theme")
  if (!raw) return
  try {
    const parsed = JSON.parse(raw) as ThemeMode
    if (
      typeof parsed === "string" &&
      (THEME_MODES as readonly string[]).includes(parsed)
    ) {
      applyTheme(parsed)
    }
  } catch {}
}

export const themeStore = defineStore<ThemeMode>("theme", "system", {
  persist: true,
  schema: ThemeModeSchema,
})

export const setThemeAtom = atom(null, (_get, set, mode: ThemeMode) => {
  set(themeStore.atom, mode)
  applyTheme(mode)
})
