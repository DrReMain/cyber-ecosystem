import { useEffect, useLayoutEffect } from "react";
import { getTextDirection } from "#/paraglide/runtime";
import { useOverflowEdges } from "../use-overflow-edges";

export const useIsomorphicLayoutEffect =
  typeof window === "undefined" ? useEffect : useLayoutEffect;

export function useTabScroll({ activeKey, count }: { activeKey: string; count: number }) {
  const { ref, start: atStart, end: atEnd, overflow, measure } = useOverflowEdges("x");

  // biome-ignore lint/correctness/useExhaustiveDependencies: mount-only by design - the tablist node never remounts
  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const onWheel = (e: WheelEvent) => {
      if (el.scrollWidth <= el.clientWidth) return;
      if (Math.abs(e.deltaX) >= Math.abs(e.deltaY)) return;
      // RTL scrollLeft is negative-going: flip so wheel-down always moves toward the end
      const rtl = getTextDirection() === "rtl";
      e.preventDefault();
      const dy = e.deltaMode === 1 ? e.deltaY * 16 : e.deltaY;
      el.scrollLeft += (rtl ? -1 : 1) * dy * 3;
    };
    el.addEventListener("wheel", onWheel, { passive: false });
    return () => el.removeEventListener("wheel", onWheel);
  }, []);

  useIsomorphicLayoutEffect(() => {
    if (count === 0) return;
    const el = ref.current;
    if (!el) return;
    const active = el.querySelector<HTMLElement>(`[data-tab-key="${CSS.escape(activeKey)}"]`);
    if (!active) return;
    const view = el.getBoundingClientRect();
    const box = active.getBoundingClientRect();
    const pad = 8;
    if (box.left < view.left + pad) {
      el.scrollBy({ left: box.left - view.left - pad });
    } else if (box.right > view.right - pad) {
      el.scrollBy({ left: box.right - view.right + pad });
    }
    measure();
  }, [activeKey, count, measure]);

  const scroll = (direction: 1 | -1) => {
    const el = ref.current;
    if (!el) return;
    // direction is logical (+1 = toward end); the chevron icons are rtl-mirrored, so the physical delta flips too
    const rtl = getTextDirection() === "rtl";
    el.scrollBy({ behavior: "smooth", left: (rtl ? -1 : 1) * direction * (el.clientWidth - 150) });
  };

  return { atEnd, atStart, overflow, ref, scroll };
}
