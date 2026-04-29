import { createServerFn } from "@tanstack/react-start"
import { getCookie, getRequestHeader } from "@tanstack/react-start/server"
import { z } from "zod"
import { readCookie, writeCookie } from "#/lib/cookie"

export type ThemePreference = "light" | "dark"
export type ThemeMode = "light" | "dark" | "system"

const ThemeCookieSchema = z.object({
  skinId: z.string().default("default"),
  mode: z.enum(["light", "dark", "system"]).default("light"),
  preference: z.enum(["light", "dark"]).default("light"),
  compact: z.boolean().default(false),
})

export type ThemeCookieData = z.infer<typeof ThemeCookieSchema>

const THEME_COOKIE_KEY = "theme"

const THEME_DEFAULT: ThemeCookieData = {
  skinId: "default",
  mode: "light",
  preference: "light",
  compact: false,
}

export const getThemeFromServer = createServerFn({ method: "GET" }).handler(
  async (): Promise<ThemeCookieData> => {
    const raw = getCookie(THEME_COOKIE_KEY)
    const data = readCookie(raw, ThemeCookieSchema, THEME_DEFAULT)

    if (data.mode === "system") {
      const headerVal = getRequestHeader("sec-ch-prefers-color-scheme")
      const headerPref =
        headerVal === "dark" || headerVal === "light" ? headerVal : undefined
      data.preference = headerPref ?? data.preference
    } else {
      data.preference = data.mode
    }

    return data
  },
)

export function setThemeCookie(
  skinId: string,
  mode: ThemeMode,
  preference: ThemePreference,
  compact: boolean,
): void {
  writeCookie(THEME_COOKIE_KEY, { skinId, mode, preference, compact })
}
