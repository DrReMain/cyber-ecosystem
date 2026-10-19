import type { RefObject } from "react";
import { useEffect, useLayoutEffect, useRef } from "react";

const useIsomorphicLayoutEffect = typeof window === "undefined" ? useEffect : useLayoutEffect;

export function usePerTabScroll(ref: RefObject<HTMLElement | null>, pathname: string) {
  const positions = useRef(new Map<string, number>());
  const prev = useRef(pathname);

  useIsomorphicLayoutEffect(() => {
    const el = ref.current;
    if (el) positions.current.set(prev.current, el.scrollTop);
    prev.current = pathname;
    el?.scrollTo({ top: positions.current.get(pathname) ?? 0 });
  }, [pathname, ref]);
}
