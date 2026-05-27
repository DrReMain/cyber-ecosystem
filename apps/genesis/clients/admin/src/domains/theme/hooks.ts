import type { MouseEvent } from "react"
import { useContext } from "react"
import type { ThemePreference } from "#/domains/theme"
import type { ThemeContextValue } from "./provider"
import { ThemeContext } from "./provider"

export function useTheme(): ThemeContextValue {
  const ctx = useContext(ThemeContext)
  if (!ctx) throw new Error("useTheme must be used within ThemeProvider")
  return ctx
}

export function useToggleTheme() {
  const { preference, setMode } = useTheme()

  return function toggleTheme(e: MouseEvent) {
    const isDark = preference === "dark"
    const nextMode: ThemePreference = isDark ? "light" : "dark"

    const isAppearanceTransition =
      typeof document.startViewTransition === "function" &&
      !window.matchMedia("(prefers-reduced-motion: reduce)").matches

    if (!isAppearanceTransition) {
      setMode(nextMode)
      return
    }

    const x = e.clientX
    const y = e.clientY
    const endRadius = Math.hypot(
      Math.max(x, innerWidth - x),
      Math.max(y, innerHeight - y),
    )

    // Suppress the route-content view transition during theme toggle
    const routeEl = document.querySelector<HTMLElement>(
      "[style*='view-transition-name']",
    )
    const savedVTN = routeEl?.style.viewTransitionName
    if (savedVTN) routeEl?.style.setProperty("view-transition-name", "none")

    const transition = document.startViewTransition(() => {
      setMode(nextMode)
    })

    transition.ready
      .then(() => {
        const clipPath = [
          `circle(0px at ${x}px ${y}px)`,
          `circle(${endRadius}px at ${x}px ${y}px)`,
        ]
        document.documentElement.animate(
          {
            clipPath: isDark ? [...clipPath].reverse() : clipPath,
          },
          {
            duration: 500,
            easing: "cubic-bezier(0.4, 0, 0.2, 1)",
            pseudoElement: isDark
              ? "::view-transition-old(root)"
              : "::view-transition-new(root)",
          },
        )
      })
      .catch(() => {})
    transition.finished.then(() => {
      if (savedVTN && routeEl)
        routeEl.style.setProperty("view-transition-name", savedVTN)
    })
  }
}
