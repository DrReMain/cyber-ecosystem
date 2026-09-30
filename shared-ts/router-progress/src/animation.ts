import type { RouterProgressConfig } from "./config";

type Timer = ReturnType<typeof setTimeout>;
interface Ref<T> {
  current: T;
}

export interface AnimationRefs {
  width: Ref<number>;
  raf: Ref<number | undefined>;
  finishTimer: Ref<Timer | undefined>;
  lastFrameTime: Ref<number | undefined>;
}

const FRAME_DURATION = 1000 / 60;

export function cancelProgressAnimation(refs: AnimationRefs) {
  if (refs.raf.current !== undefined) {
    cancelAnimationFrame(refs.raf.current);
    refs.raf.current = undefined;
  }
  if (refs.finishTimer.current !== undefined) {
    clearTimeout(refs.finishTimer.current);
    refs.finishTimer.current = undefined;
  }
  refs.lastFrameTime.current = undefined;
}

export function startProgress(
  bar: HTMLDivElement,
  container: HTMLDivElement,
  config: RouterProgressConfig,
  refs: AnimationRefs,
) {
  cancelProgressAnimation(refs);
  const initial = config.initialMin + Math.random() * (config.initialMax - config.initialMin);
  refs.width.current = initial;
  container.style.display = "block";
  bar.style.transition = "none";
  bar.style.width = `${initial}%`;

  const tick = (timestamp: number) => {
    if (refs.width.current >= config.trickleCeiling) return;

    const previousFrameTime = refs.lastFrameTime.current;
    refs.lastFrameTime.current = timestamp;
    if (previousFrameTime !== undefined) {
      const elapsedFrames = Math.min((timestamp - previousFrameTime) / FRAME_DURATION, 60);
      const remaining = config.trickleCeiling - refs.width.current;
      const progressFactor = 1 - (1 - config.trickleRate) ** elapsedFrames;
      refs.width.current += remaining * progressFactor;
      bar.style.width = `${refs.width.current}%`;
    }

    if (refs.width.current < config.trickleCeiling) {
      refs.raf.current = requestAnimationFrame(tick);
    } else {
      refs.raf.current = undefined;
    }
  };
  refs.raf.current = requestAnimationFrame(tick);
}

export function finishProgress(
  bar: HTMLDivElement,
  container: HTMLDivElement,
  config: RouterProgressConfig,
  refs: AnimationRefs,
) {
  if (refs.raf.current !== undefined) {
    cancelAnimationFrame(refs.raf.current);
    refs.raf.current = undefined;
  }
  if (refs.finishTimer.current !== undefined) {
    clearTimeout(refs.finishTimer.current);
  }
  refs.lastFrameTime.current = undefined;
  bar.style.transition = `width ${config.transitionDuration}ms ease-out`;
  bar.style.width = "100%";
  refs.finishTimer.current = setTimeout(
    () => {
      container.style.display = "none";
      bar.style.transition = "none";
      bar.style.width = "0%";
      refs.width.current = 0;
      refs.finishTimer.current = undefined;
    },
    Math.max(config.finishDelay, config.transitionDuration),
  );
}
