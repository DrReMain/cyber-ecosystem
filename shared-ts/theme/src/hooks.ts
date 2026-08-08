import type { MouseEvent } from "react";
import { useContext } from "react";
import type { ThemePreference } from "./cookie";
import type { ThemeContextValue } from "./provider";
import { ThemeContext } from "./provider";

export function useTheme(): ThemeContextValue {
  const ctx = useContext(ThemeContext);
  if (!ctx) throw new Error("useTheme must be used within ThemeProvider");
  return ctx;
}

export function useToggleTheme() {
  const { preference, setMode } = useTheme();

  return function toggleTheme(e: MouseEvent) {
    const isDark = preference === "dark";
    const nextMode: ThemePreference = isDark ? "light" : "dark";

    const isAppearanceTransition =
      typeof document.startViewTransition === "function" &&
      !window.matchMedia("(prefers-reduced-motion: reduce)").matches;

    if (!isAppearanceTransition) {
      setMode(nextMode);
      return;
    }

    const x = e.clientX;
    const y = e.clientY;
    const endRadius = Math.hypot(
      Math.max(x, window.innerWidth - x),
      Math.max(y, window.innerHeight - y),
    );

    // [data-theme-transition] makes route-transition css stand down during
    // this morph (see ./route-transition/route-transition.css) — no JS import.
    const root = document.documentElement;
    root.dataset.themeTransition = "true";

    const transition = document.startViewTransition(() => {
      setMode(nextMode);
    });

    transition.ready
      .then(() => {
        const clipPath = [
          `circle(0px at ${x}px ${y}px)`,
          `circle(${endRadius}px at ${x}px ${y}px)`,
        ];
        document.documentElement.animate(
          {
            clipPath: isDark ? [...clipPath].reverse() : clipPath,
          },
          {
            duration: 500,
            easing: "cubic-bezier(0.4, 0, 0.2, 1)",
            pseudoElement: isDark ? "::view-transition-old(root)" : "::view-transition-new(root)",
          },
        );
      })
      .catch(() => {});
    transition.finished.then(
      () => {
        delete root.dataset.themeTransition;
      },
      () => {
        delete root.dataset.themeTransition;
      },
    );
  };
}
