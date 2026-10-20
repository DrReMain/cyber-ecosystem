import { describe, expect, it } from "vitest";
import { createFrameLoop } from "./frame-loop";

function manualFrames() {
  const frames: ((timeMs: number) => void)[] = [];
  return {
    requestFrame: (cb: (timeMs: number) => void) => {
      frames.push(cb);
      return frames.length;
    },
    cancelFrame: () => {},
    emit: (timeMs: number) => {
      const cb = frames[frames.length - 1];
      if (cb !== undefined) {
        cb(timeMs);
      }
    },
  };
}

describe("createFrameLoop", () => {
  it("steps with real dt and clamps long gaps to the max", () => {
    const source = manualFrames();
    const steps: number[] = [];
    const loop = createFrameLoop((dtSec) => steps.push(dtSec), {
      requestFrame: source.requestFrame,
      cancelFrame: source.cancelFrame,
      maxDtMs: 50,
    });
    loop.start();
    source.emit(0);
    source.emit(16);
    source.emit(32);
    source.emit(3032); // 3s gap → clamped to 50ms
    source.emit(3048);
    expect(steps.map((dt) => Math.round(dt * 1000))).toEqual([16, 16, 50, 16]);
    loop.stop();
  });

  it("stop is idempotent and freezes the loop", () => {
    const source = manualFrames();
    const steps: number[] = [];
    const loop = createFrameLoop((dtSec) => steps.push(dtSec), {
      requestFrame: source.requestFrame,
      cancelFrame: source.cancelFrame,
    });
    loop.start();
    source.emit(0);
    loop.stop();
    loop.stop();
    source.emit(100);
    source.emit(200);
    expect(steps).toEqual([]);
  });
});
