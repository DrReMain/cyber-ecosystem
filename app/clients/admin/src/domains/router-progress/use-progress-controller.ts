import { useEffect, useMemo, useRef } from "react";
import {
  type AnimationRefs,
  cancelProgressAnimation,
  finishProgress,
  startProgress,
} from "./animation";
import type { RouterProgressConfig } from "./config";

function resetProgress(
  bar: HTMLDivElement | null,
  container: HTMLDivElement | null,
  refs: AnimationRefs,
  isActive: { current: boolean },
) {
  cancelProgressAnimation(refs);
  isActive.current = false;
  refs.width.current = 0;
  if (bar) {
    bar.style.transition = "none";
    bar.style.width = "0%";
  }
  if (container) container.style.display = "none";
}

/**
 * Owns the imperative DOM work required to animate a progress indicator.
 * It is router-agnostic so other loading sources can reuse the controller.
 */
export function useProgressController(config: RouterProgressConfig, isLoading: boolean) {
  const containerRef = useRef<HTMLDivElement>(null);
  const barRef = useRef<HTMLDivElement>(null);
  const configRef = useRef(config);
  const isActiveRef = useRef(false);
  const widthRef = useRef(0);
  const rafRef = useRef<number | undefined>(undefined);
  const finishTimerRef = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  const lastFrameTimeRef = useRef<number | undefined>(undefined);
  const refs = useMemo<AnimationRefs>(
    () => ({
      width: widthRef,
      raf: rafRef,
      finishTimer: finishTimerRef,
      lastFrameTime: lastFrameTimeRef,
    }),
    [],
  );

  configRef.current = config;

  useEffect(() => {
    const bar = barRef.current;
    const container = containerRef.current;
    if (!(bar && container)) return;
    if (isLoading && !isActiveRef.current) {
      isActiveRef.current = true;
      startProgress(bar, container, configRef.current, refs);
    } else if (!isLoading && isActiveRef.current) {
      isActiveRef.current = false;
      finishProgress(bar, container, configRef.current, refs);
    }
  }, [isLoading, refs]);

  useEffect(
    () => () => resetProgress(barRef.current, containerRef.current, refs, isActiveRef),
    [refs],
  );

  return { containerRef, barRef };
}
