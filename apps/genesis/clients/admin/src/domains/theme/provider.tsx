import type { PropsWithChildren } from "react"
import { createContext, useCallback, useEffect, useRef, useState } from "react"
import type { ThemePreference } from "#/domains/theme"
import {
  setThemeCookie,
  type ThemeCookieData,
  type ThemeMode,
} from "#/domains/theme/cookie"

export interface ThemeContextValue {
  skinId: string
  mode: ThemeMode
  preference: ThemePreference
  compact: boolean
  setSkinId: (id: string) => void
  setMode: (m: ThemeMode) => void
  setCompact: (c: boolean) => void
}

export const ThemeContext = createContext<ThemeContextValue | null>(null)

function getSystemPreference(): ThemePreference {
  return window.matchMedia("(prefers-color-scheme: dark)").matches
    ? "dark"
    : "light"
}

function useThemeSync(
  skinId: string,
  mode: ThemeMode,
  preference: ThemePreference,
  compact: boolean,
  setPreference: (p: ThemePreference) => void,
) {
  const resolved = useRef(false)

  // biome-ignore lint/correctness/useExhaustiveDependencies: mount-only
  useEffect(() => {
    if (resolved.current) return
    resolved.current = true
    if (mode === "system") {
      const systemPref = getSystemPreference()
      if (systemPref !== preference) {
        setPreference(systemPref)
        setThemeCookie(skinId, "system", systemPref, compact)
      }
    }
  }, [])

  useEffect(() => {
    if (mode !== "system") return
    const mq = window.matchMedia("(prefers-color-scheme: dark)")
    const handler = (e: MediaQueryListEvent) => {
      const next: ThemePreference = e.matches ? "dark" : "light"
      setPreference(next)
      setThemeCookie(skinId, "system", next, compact)
    }
    mq.addEventListener("change", handler)
    return () => mq.removeEventListener("change", handler)
  }, [skinId, mode, compact, setPreference])

  useEffect(() => {
    document.documentElement.classList.toggle("dark", preference === "dark")
  }, [preference])
}

interface IProps {
  initialTheme: ThemeCookieData
}

export function ThemeProvider({
  children,
  initialTheme,
}: Readonly<PropsWithChildren<IProps>>) {
  const [skinId, setSkinIdState] = useState(initialTheme.skinId)
  const [mode, setModeState] = useState<ThemeMode>(initialTheme.mode)
  const [preference, setPreference] = useState<ThemePreference>(
    initialTheme.preference,
  )
  const [compact, setCompactState] = useState(initialTheme.compact)

  useThemeSync(skinId, mode, preference, compact, setPreference)

  const setSkinId = useCallback(
    (id: string) => {
      setSkinIdState(id)
      setThemeCookie(id, mode, preference, compact)
    },
    [mode, preference, compact],
  )

  const setMode = useCallback(
    (m: ThemeMode) => {
      setModeState(m)
      const next = m === "system" ? getSystemPreference() : m
      setPreference(next)
      setThemeCookie(skinId, m, next, compact)
    },
    [skinId, compact],
  )

  const setCompact = useCallback(
    (c: boolean) => {
      setCompactState(c)
      setThemeCookie(skinId, mode, preference, c)
    },
    [skinId, mode, preference],
  )

  return (
    <ThemeContext
      value={{
        skinId,
        mode,
        preference,
        compact,
        setSkinId,
        setMode,
        setCompact,
      }}
    >
      {children}
    </ThemeContext>
  )
}
