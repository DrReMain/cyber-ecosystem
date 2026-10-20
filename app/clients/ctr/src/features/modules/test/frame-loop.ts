// The rAF side-kick to GameClock: continuous-motion games (legacy
// skill_prediction, tracking_moving) step on frames with a clamped dt so a
// stalled tab never teleports entities (legacy clamped at 50ms). The frame
// source is injectable for tests.

export interface FrameLoop {
  start(): void;
  stop(): void;
}

export interface FrameLoopOptions {
  readonly requestFrame?: (cb: (timeMs: number) => void) => number;
  readonly cancelFrame?: (handle: number) => void;
  /** Legacy dt clamp: 50ms. */
  readonly maxDtMs?: number;
}

function defaultRequestFrame(cb: (timeMs: number) => void): number {
  return requestAnimationFrame(cb);
}

export function createFrameLoop(
  step: (dtSec: number, nowMs: number) => void,
  options: FrameLoopOptions = {},
): FrameLoop {
  const requestFrame = options.requestFrame ?? defaultRequestFrame;
  const cancelFrame = options.cancelFrame ?? cancelAnimationFrame;
  const maxDtMs = options.maxDtMs ?? 50;

  let running = false;
  let handle: number | null = null;
  let lastMs: number | null = null;

  function frame(timeMs: number): void {
    if (!running) {
      return;
    }
    if (lastMs !== null) {
      const dtMs = Math.min(timeMs - lastMs, maxDtMs);
      step(dtMs / 1000, timeMs);
    }
    lastMs = timeMs;
    handle = requestFrame(frame);
  }

  return {
    start(): void {
      if (running) {
        return;
      }
      running = true;
      lastMs = null;
      handle = requestFrame(frame);
    },
    stop(): void {
      if (!running) {
        return;
      }
      running = false;
      if (handle !== null) {
        cancelFrame(handle);
        handle = null;
      }
    },
  };
}
