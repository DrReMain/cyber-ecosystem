import { type RefObject, useCallback, useEffect, useRef, useState } from "react";

interface OverflowEdgeState {
  end: boolean;
  overflow: boolean;
  start: boolean;
}

export function useOverflowEdges<T extends HTMLElement = HTMLDivElement>(
  axis: "x" | "y",
  target?: RefObject<T | null>,
) {
  const innerRef = useRef<T | null>(null);
  const ref = target ?? innerRef;
  const [edge, setEdge] = useState<OverflowEdgeState>({ end: false, overflow: false, start: true });

  const measure = useCallback(() => {
    const el = ref.current;
    if (!el) return;
    const max = axis === "x" ? el.scrollWidth - el.clientWidth : el.scrollHeight - el.clientHeight;
    const pos = axis === "x" ? Math.abs(el.scrollLeft) : el.scrollTop;
    setEdge((prev) => {
      const next =
        max <= 1
          ? { end: true, overflow: false, start: true }
          : { end: pos >= max - 1, overflow: true, start: pos <= 1 };
      return prev.end === next.end && prev.overflow === next.overflow && prev.start === next.start
        ? prev
        : next;
    });
  }, [axis, ref]);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const observer = new ResizeObserver(measure);
    observer.observe(el);
    if (el.firstElementChild instanceof HTMLElement) observer.observe(el.firstElementChild);
    if (el.lastElementChild instanceof HTMLElement) observer.observe(el.lastElementChild);
    el.addEventListener("scroll", measure, { passive: true });
    measure();
    return () => {
      observer.disconnect();
      el.removeEventListener("scroll", measure);
    };
  }, [ref, measure]);

  return { ...edge, measure, ref };
}
