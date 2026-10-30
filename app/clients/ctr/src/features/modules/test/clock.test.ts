import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { GameClock } from "./clock";

// Fake timers own both Date.now and setInterval, so the default (injected)
// source and scheduler are driven exactly by advanceTimersByTime.
function makeClock(durationMs: number): GameClock {
  return new GameClock({
    durationMs,
    schedule: (delayMs, fn) => {
      const id = setInterval(fn, delayMs);
      return { clear: () => clearInterval(id) };
    },
  });
}

describe("GameClock", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("derives elapsed from the wall clock with zero drift", () => {
    const clock = makeClock(10_000);
    clock.start();
    // Irregular advances: tick accumulation would drift; read-time
    // derivation must stay exact.
    vi.advanceTimersByTime(333);
    vi.advanceTimersByTime(1001);
    vi.advanceTimersByTime(2666);
    expect(clock.elapsedMs()).toBe(4000);
    vi.advanceTimersByTime(6000);
    expect(clock.elapsedMs()).toBe(10_000);
    clock.dispose();
  });

  it("defers the deadline across a pause and never counts paused time", () => {
    const clock = makeClock(10_000);
    const finish = vi.fn();
    clock.onFinish(finish);
    clock.start();
    vi.advanceTimersByTime(5000);
    clock.pause();
    vi.advanceTimersByTime(3000);
    // Frozen for the whole pause segment.
    expect(clock.elapsedMs()).toBe(5000);
    clock.resume();
    vi.advanceTimersByTime(5000);
    expect(finish).toHaveBeenCalledTimes(1);
    expect(clock.elapsedMs()).toBe(10_000);
    clock.dispose();
  });

  it("keeps the clock frozen while a hold stays open under a resumed pause", () => {
    const clock = makeClock(10_000);
    clock.start();
    vi.advanceTimersByTime(1000);
    clock.pause();
    clock.hold();
    clock.resume();
    vi.advanceTimersByTime(2000);
    // pause closed, hold still open: still frozen (union semantics, no
    // double-count of the overlap).
    expect(clock.elapsedMs()).toBe(1000);
    clock.releaseHold();
    vi.advanceTimersByTime(1000);
    expect(clock.elapsedMs()).toBe(2000);
    clock.dispose();
  });

  it("extends the run with addTime and ignores it after finish", () => {
    const clock = makeClock(10_000);
    const finish = vi.fn();
    clock.onFinish(finish);
    clock.start();
    vi.advanceTimersByTime(9000);
    clock.addTime(1000);
    expect(clock.remainingMs()).toBe(2000);
    vi.advanceTimersByTime(2000);
    expect(finish).toHaveBeenCalledTimes(1);
    clock.addTime(5000);
    expect(clock.remainingMs()).toBe(0);
    clock.dispose();
  });

  it("holds a phase without touching the pause side", () => {
    const clock = makeClock(10_000);
    clock.start();
    vi.advanceTimersByTime(1000);
    clock.hold();
    vi.advanceTimersByTime(2000);
    expect(clock.elapsedMs()).toBe(1000);
    clock.releaseHold();
    vi.advanceTimersByTime(1000);
    expect(clock.elapsedMs()).toBe(2000);
    clock.dispose();
  });

  it("fires onFinish exactly once across the zero crossing", () => {
    const clock = makeClock(1000);
    const finish = vi.fn();
    clock.onFinish(finish);
    clock.start();
    vi.advanceTimersByTime(999_999);
    expect(finish).toHaveBeenCalledTimes(1);
    expect(clock.snapshot().finished).toBe(true);
    clock.dispose();
  });

  it("stays idle before start: no ticks run, no finish, zero elapsed", () => {
    const clock = makeClock(1000);
    const tick = vi.fn();
    const finish = vi.fn();
    clock.onTick(tick);
    clock.onFinish(finish);
    const immediate = tick.mock.calls.length; // the first frame on subscribe
    vi.advanceTimersByTime(5000);
    expect(tick.mock.calls.length).toBe(immediate);
    expect(finish).not.toHaveBeenCalled();
    expect(clock.elapsedMs()).toBe(0);
    clock.dispose();
  });

  it("ticks at the configured cadence", () => {
    const clock = makeClock(10_000);
    const tick = vi.fn();
    clock.onTick(tick);
    expect(tick).toHaveBeenCalledTimes(1); // first frame on subscribe
    clock.start();
    vi.advanceTimersByTime(30);
    // start fires one tick, then 30ms of cadence = 3 more.
    expect(tick.mock.calls.length).toBe(5);
    clock.dispose();
  });

  it("keeps elapsed + remaining equal to duration + added while running", () => {
    const clock = makeClock(10_000);
    clock.start();
    clock.addTime(1000);
    vi.advanceTimersByTime(2500);
    clock.pause();
    vi.advanceTimersByTime(800);
    clock.hold();
    clock.resume();
    vi.advanceTimersByTime(1200);
    clock.releaseHold();
    const { elapsedMs, remainingMs } = clock.snapshot();
    expect(elapsedMs + remainingMs).toBe(11_000);
    clock.dispose();
  });
});
