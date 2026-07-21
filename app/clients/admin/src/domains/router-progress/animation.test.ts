import { afterEach, describe, expect, it, vi } from "vitest";
import {
  type AnimationRefs,
  cancelProgressAnimation,
  finishProgress,
  startProgress,
} from "./animation";
import type { RouterProgressConfig } from "./config";

const config: RouterProgressConfig = {
  height: 2,
  initialMin: 10,
  initialMax: 10,
  trickleCeiling: 90,
  trickleRate: 0.1,
  finishDelay: 100,
  transitionDuration: 200,
  color: "red",
  cycleDuration: 0,
  position: "top",
  zIndex: 1,
};

function createRefs(): AnimationRefs {
  return {
    width: { current: 0 },
    raf: { current: undefined },
    finishTimer: { current: undefined },
    lastFrameTime: { current: undefined },
  };
}

function createElement() {
  return { style: {} } as HTMLDivElement;
}

afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
});

describe("progress animation", () => {
  it("normalizes trickle progress by elapsed animation time", () => {
    const callbacks: Array<FrameRequestCallback> = [];
    vi.stubGlobal(
      "requestAnimationFrame",
      vi.fn((callback: FrameRequestCallback) => {
        callbacks.push(callback);
        return callbacks.length;
      }),
    );
    vi.stubGlobal("cancelAnimationFrame", vi.fn());
    vi.spyOn(Math, "random").mockReturnValue(0);

    const bar = createElement();
    const container = createElement();
    const refs = createRefs();
    startProgress(bar, container, config, refs);

    callbacks.shift()?.(0);
    callbacks.shift()?.(1000 / 60);

    expect(refs.width.current).toBeCloseTo(18, 5);
    expect(bar.style.width).toBe("18%");
  });

  it("waits for the completion transition before hiding", () => {
    vi.useFakeTimers();
    vi.stubGlobal("cancelAnimationFrame", vi.fn());
    const bar = createElement();
    const container = createElement();
    const refs = createRefs();

    finishProgress(bar, container, config, refs);
    vi.advanceTimersByTime(199);
    expect(container.style.display).toBeUndefined();

    vi.advanceTimersByTime(1);
    expect(container.style.display).toBe("none");
    expect(bar.style.width).toBe("0%");
  });

  it("cancels both scheduled animation work items", () => {
    const cancelAnimationFrame = vi.fn();
    vi.useFakeTimers();
    vi.stubGlobal("cancelAnimationFrame", cancelAnimationFrame);
    const refs = createRefs();
    refs.raf.current = 12;
    refs.finishTimer.current = setTimeout(vi.fn(), 100) as unknown as ReturnType<typeof setTimeout>;

    cancelProgressAnimation(refs);

    expect(cancelAnimationFrame).toHaveBeenCalledWith(12);
    expect(refs.raf.current).toBeUndefined();
    expect(refs.finishTimer.current).toBeUndefined();
  });
});
