import type { MouseEvent } from "react";
import { useContext } from "react";
import { flushSync } from "react-dom";
import type { ThemePreference } from "./cookie";
import type { ThemeContextValue } from "./provider";
import { ThemeContext } from "./provider";

// Shared with the toggle button's delayed icon swap (toggle.tsx) so the icon
// lands when the morph does.
export const THEME_MORPH_MS = 800;

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

    // Percentages, not pixels: px coordinates on root view-transition pseudos
    // resolve in Snapshot Containing Block space (Chromium bug — offset under
    // page zoom, e.g. macOS Retina Chrome). Relative values cancel that
    // mismatch: the circle anchors at the click point whatever the snapshot
    // box is scaled to, and stay pixel-identical at 100% zoom.
    const x = e.clientX;
    const y = e.clientY;
    const width = window.innerWidth;
    const height = window.innerHeight;
    const xPct = (x / width) * 100;
    const yPct = (y / height) * 100;
    const endRadius = Math.hypot(Math.max(x, width - x), Math.max(y, height - y));
    // circle() percentage radii resolve against sqrt(w² + h²) / √2.
    const radiusDenominator = Math.hypot(width, height) / Math.SQRT2;
    const endRadiusPct = (endRadius / radiusDenominator) * 100;

    // [data-theme-transition] makes route-transition css stand down during
    // this morph (see ./route-transition/route-transition.css) — no JS import.
    const root = document.documentElement;
    root.dataset.themeTransition = "true";

    const transition = document.startViewTransition(() => {
      // The new snapshot is captured once this callback settles, so the flip
      // must land inside it, not one React tick later: flushSync commits
      // React now, and the class is set imperatively because its effect runs
      // post-commit and the z-index branches in view-transition.css key off
      // it during the morph.
      root.classList.toggle("dark", nextMode === "dark");
      flushSync(() => setMode(nextMode));
    });

    transition.ready
      .then(() => {
        const clipPath = [
          `circle(0% at ${xPct}% ${yPct}%)`,
          `circle(${endRadiusPct}% at ${xPct}% ${yPct}%)`,
        ];
        const morph = document.documentElement.animate(
          {
            clipPath: isDark ? [...clipPath].reverse() : clipPath,
          },
          {
            // Symmetric ease-in-out: revealed area grows with r², so an
            // ease-out front-loads most of the coverage into the opening
            // and reads as fast at any duration — a symmetric curve spreads
            // the sweep. fill: "forwards" holds the end keyframe until the
            // pseudo tree is torn down: the clip runs on the compositor
            // while teardown is a main-thread mutation, and a released
            // effect can paint one frame of the un-clipped layer first.
            duration: THEME_MORPH_MS,
            easing: "cubic-bezier(0.65, 0, 0.35, 1)",
            fill: "forwards",
            pseudoElement: isDark ? "::view-transition-old(root)" : "::view-transition-new(root)",
          },
        );
        // Cancel at teardown: pseudo resolution is by name, so a filled
        // animation that outlives this transition re-attaches to the next
        // transition's same-named pseudo — a stale clip blanks one of its
        // layers, and leftover animations fight for the property. fill and
        // cancel are a pair.
        transition.finished.finally(() => morph.cancel());
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
