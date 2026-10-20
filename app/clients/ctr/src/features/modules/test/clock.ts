// GameClock — the single timing implementation that retires the legacy
// thirteen hand-rolled counters. A wall-clock base (zero drift: elapsed is
// derived at read time, never accumulated from ticks) plus frozen-segment
// accounting that covers every legacy need: user pause and background spans
// (pause), phase holds (hold), bonus time (addTime), and a game-chosen
// start moment (first-deduction starts). Pure logic: the time source and
// scheduler are injectable, so fake timers drive it exactly.

export type ClockNow = () => number;

export interface ClockTimer {
  clear(): void;
}

export type ClockScheduler = (delayMs: number, fn: () => void) => ClockTimer;

export interface ClockSnapshot {
  readonly elapsedMs: number;
  readonly remainingMs: number;
  readonly started: boolean;
  readonly frozen: boolean;
  readonly finished: boolean;
}

export interface GameClockOptions {
  readonly durationMs: number;
  /** Display tick cadence (legacy: 10ms). */
  readonly tickMs?: number;
  readonly now?: ClockNow;
  readonly schedule?: ClockScheduler;
}

function defaultSchedule(delayMs: number, fn: () => void): ClockTimer {
  const id = setInterval(fn, delayMs);
  return { clear: () => clearInterval(id) };
}

export class GameClock {
  private readonly durationMs: number;
  private readonly tickMs: number;
  private readonly now: ClockNow;
  private readonly schedule: ClockScheduler;

  private originMs: number | null = null;
  private freezeStartMs: number | null = null;
  private frozenTotalMs = 0;
  private pauseOpen = false;
  private holdOpen = false;
  private addedMs = 0;
  private finished = false;
  private ticker: ClockTimer | null = null;
  private readonly tickSubs = new Set<(snapshot: ClockSnapshot) => void>();
  private readonly finishSubs = new Set<() => void>();

  constructor(options: GameClockOptions) {
    this.durationMs = options.durationMs;
    this.tickMs = options.tickMs ?? 10;
    this.now = options.now ?? Date.now;
    this.schedule = options.schedule ?? defaultSchedule;
  }

  /** Idempotent; the game decides when (first meaningful deduction starts). */
  start(): void {
    if (this.originMs !== null) {
      return;
    }
    this.originMs = this.now();
    this.ticker = this.schedule(this.tickMs, () => this.handleTick());
    this.handleTick();
  }

  /** Idempotent; opens the shared freeze segment (user pause / background). */
  pause(): void {
    this.setFreezeFlag("pause", true);
  }

  /** Closes the pause side; a still-open hold keeps the clock frozen. */
  resume(): void {
    this.setFreezeFlag("pause", false);
  }

  /** Idempotent; opens the shared freeze segment (phase hold). */
  hold(): void {
    this.setFreezeFlag("hold", true);
  }

  releaseHold(): void {
    this.setFreezeFlag("hold", false);
  }

  /** Extends the run (legacy: perfect-parry bonus); ignored after finish. */
  addTime(ms: number): void {
    if (this.finished) {
      return;
    }
    this.addedMs += ms;
  }

  /** Active time: wall time minus every frozen segment (union semantics). */
  elapsedMs(): number {
    if (this.originMs === null) {
      return 0;
    }
    const now = this.now();
    const openFreeze = this.freezeStartMs !== null ? now - this.freezeStartMs : 0;
    return now - this.originMs - this.frozenTotalMs - openFreeze;
  }

  remainingMs(): number {
    return Math.max(0, this.durationMs + this.addedMs - this.elapsedMs());
  }

  snapshot(): ClockSnapshot {
    return {
      elapsedMs: this.elapsedMs(),
      remainingMs: this.remainingMs(),
      started: this.originMs !== null,
      frozen: this.pauseOpen || this.holdOpen,
      finished: this.finished,
    };
  }

  /** Cadence-throttled subscription; the current snapshot fires immediately. */
  onTick(fn: (snapshot: ClockSnapshot) => void): () => void {
    this.tickSubs.add(fn);
    fn(this.snapshot());
    return () => {
      this.tickSubs.delete(fn);
    };
  }

  /** Fires exactly once, when a tick first observes remaining at zero. */
  onFinish(fn: () => void): () => void {
    this.finishSubs.add(fn);
    return () => {
      this.finishSubs.delete(fn);
    };
  }

  dispose(): void {
    this.ticker?.clear();
    this.ticker = null;
    this.tickSubs.clear();
    this.finishSubs.clear();
  }

  private setFreezeFlag(side: "pause" | "hold", open: boolean): void {
    const wasFrozen = this.pauseOpen || this.holdOpen;
    if (side === "pause") {
      this.pauseOpen = open;
    } else {
      this.holdOpen = open;
    }
    const isFrozen = this.pauseOpen || this.holdOpen;
    if (!wasFrozen && isFrozen) {
      this.freezeStartMs = this.now();
    } else if (wasFrozen && !isFrozen) {
      if (this.freezeStartMs !== null) {
        this.frozenTotalMs += this.now() - this.freezeStartMs;
        this.freezeStartMs = null;
      }
    }
  }

  private handleTick(): void {
    const snapshot = this.snapshot();
    for (const sub of this.tickSubs) {
      sub(snapshot);
    }
    if (!this.finished && snapshot.remainingMs <= 0 && this.originMs !== null) {
      this.finished = true;
      for (const sub of this.finishSubs) {
        sub();
      }
      // The run is over and addTime is ignored past this point: stop ticking.
      this.ticker?.clear();
      this.ticker = null;
    }
  }
}
