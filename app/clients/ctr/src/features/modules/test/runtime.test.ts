import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { GameClock } from "./clock";
import { EngineParams, type GameInstance, type GameModule, type GameModuleCtx } from "./contract";
import { TestRunHost } from "./runtime";

// A stub module that parks its ctx so tests can drive collect/end, and
// counts the pause/resume/destroy forwards.
function stubModule() {
  const instance: GameInstance = {
    start: vi.fn(),
    pause: vi.fn(),
    resume: vi.fn(),
    destroy: vi.fn(),
  };
  let ctx: GameModuleCtx | null = null;
  const module: GameModule = {
    code: "stub",
    create: (received) => {
      ctx = received;
      return instance;
    },
  };
  return {
    module,
    instance,
    ctx: () => {
      if (ctx === null) {
        throw new Error("module not created");
      }
      return ctx;
    },
  };
}

function makeHost(overrides: Partial<ConstructorParameters<typeof TestRunHost>[0]> = {}) {
  const stub = stubModule();
  const clock = new GameClock({
    durationMs: 10_000,
    schedule: (delayMs, fn) => {
      const id = setInterval(fn, delayMs);
      return { clear: () => clearInterval(id) };
    },
  });
  let visibilityHandler: ((hidden: boolean) => void) | null = null;
  const onEnd = vi.fn();
  const host = new TestRunHost({
    module: stub.module,
    clock,
    mount: document.createElement("div"),
    params: new EngineParams({}),
    durationMs: 10_000,
    locale: "zh-CN",
    now: () => Date.now(),
    onEnd,
    bindVisibility: (handler) => {
      visibilityHandler = handler;
      return () => {
        visibilityHandler = null;
      };
    },
    ...overrides,
  });
  return { host, clock, stub, onEnd, flip: (hidden: boolean) => visibilityHandler?.(hidden) };
}

describe("TestRunHost", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("marks a background span with elapsed at hidden and wall duration", () => {
    const { host, clock, flip, onEnd } = makeHost();
    host.start();
    vi.advanceTimersByTime(1000);
    flip(true);
    vi.advanceTimersByTime(500);
    // Hidden: the clock froze at the hidden moment.
    expect(clock.elapsedMs()).toBe(1000);
    flip(false);
    expect(onEnd).not.toHaveBeenCalled();
    expect(host.record().backgroundSpans).toEqual([{ elapsedMs: 1000, durationMs: 500 }]);
    vi.advanceTimersByTime(1000);
    // Resumed: active time flows again from 1000.
    expect(clock.elapsedMs()).toBe(2000);
    host.destroy();
  });

  it("ignores background flips before the run starts (pre-start noise)", () => {
    const { host, clock, stub, flip } = makeHost();
    flip(true);
    vi.advanceTimersByTime(3000);
    flip(false);
    // No span, no pause forwarded, and the clock keeps a clean origin —
    // a pre-start pause would have produced negative elapsed.
    expect(host.record().backgroundSpans).toEqual([]);
    expect(stub.instance.pause).not.toHaveBeenCalled();
    host.start();
    vi.advanceTimersByTime(1000);
    expect(clock.elapsedMs()).toBe(1000);
    host.destroy();
  });

  it("keeps a user pause across a background round-trip", () => {
    const { host, clock, stub, flip } = makeHost();
    host.start();
    vi.advanceTimersByTime(1000);
    host.pause();
    flip(true);
    vi.advanceTimersByTime(300);
    flip(false);
    // Still user-paused: clock frozen, instance not resumed.
    expect(clock.elapsedMs()).toBe(1000);
    expect(stub.instance.resume).not.toHaveBeenCalled();
    host.resume();
    expect(stub.instance.resume).toHaveBeenCalledTimes(1);
    vi.advanceTimersByTime(1000);
    expect(clock.elapsedMs()).toBe(2000);
    host.destroy();
  });

  it("caps events at 4096 and ends exactly once", () => {
    const { host, stub, onEnd } = makeHost();
    host.start();
    const { collect } = stub.ctx();
    for (let i = 0; i < 4200; i += 1) {
      collect({ i });
    }
    expect(host.record().events.length).toBe(4096);
    stub.ctx().end("failed");
    stub.ctx().end("aborted");
    expect(onEnd).toHaveBeenCalledTimes(1);
    expect(onEnd.mock.calls[0]?.[0]?.exit).toBe("failed");
    // Post-end collects are dropped.
    collect({ late: true });
    expect(host.record().events.length).toBe(4096);
    host.destroy();
  });

  it("forwards the clock's natural finish exactly once as completed", () => {
    const { host, onEnd, stub } = makeHost();
    host.start();
    vi.advanceTimersByTime(10_000);
    expect(onEnd).toHaveBeenCalledTimes(1);
    expect(onEnd.mock.calls[0]?.[0]?.exit).toBe("completed");
    // The game face freezes at end — no timer churn behind the overlay.
    expect(stub.instance.pause).toHaveBeenCalled();
    host.destroy();
  });
});
