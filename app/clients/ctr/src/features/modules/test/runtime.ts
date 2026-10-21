// TestRunHost — the runtime spine between the React shell and a game
// module: it owns the clock and instance lifecycle, forwards the user
// pause, marks background spans on visibility flips (the engine only
// marks — the server judges validity), caps the raw-detail buffers at the
// frozen contract limits (4096 events / 64 spans), and guarantees the run
// ends exactly once. Listener seams are injectable so the accounting is
// unit-testable without a real document.
import type { GameClock } from "./clock";
import type {
  AttemptRecord,
  EngineParams,
  GameExit,
  GameInstance,
  GameModule,
  GameModuleCtx,
} from "./contract";

export interface TestRunHostOptions {
  readonly module: GameModule;
  readonly clock: GameClock;
  readonly mount: HTMLElement;
  readonly params: EngineParams;
  readonly durationMs: number;
  readonly locale: string;
  now(): number;
  readonly onEnd: (record: AttemptRecord) => void;
  /** Test seam; default binds visibilitychange on document (client only). */
  readonly bindVisibility?: (handler: (hidden: boolean) => void) => () => void;
}

const MAX_EVENTS = 4096;
const MAX_SPANS = 64;

function defaultBindVisibility(handler: (hidden: boolean) => void): () => void {
  if (typeof document === "undefined") {
    return () => {};
  }
  const listener = () => handler(document.hidden);
  document.addEventListener("visibilitychange", listener);
  return () => document.removeEventListener("visibilitychange", listener);
}

export class TestRunHost {
  private readonly events: Record<string, unknown>[] = [];
  private readonly spans: { elapsedMs: number; durationMs: number }[] = [];
  private readonly stats: Record<string, number> = {};
  private readonly now: () => number;
  private readonly clock: GameClock;
  private readonly instance: GameInstance;
  private readonly offFinish: () => void;
  private readonly unbindVisibility: () => void;
  private readonly onEnd: (record: AttemptRecord) => void;
  private exit: GameExit | null = null;
  private started = false;
  private userPaused = false;
  private backgroundOpen = false;
  private spanStartElapsedMs = 0;
  private spanStartWallMs = 0;
  private ended = false;
  private destroyed = false;

  constructor(options: TestRunHostOptions) {
    this.now = options.now;
    this.clock = options.clock;
    this.onEnd = options.onEnd;
    const ctx: GameModuleCtx = {
      mount: options.mount,
      clock: options.clock,
      params: options.params,
      durationMs: options.durationMs,
      locale: options.locale,
      now: options.now,
      collect: (event) => this.collect(event),
      stat: (key, value) => {
        this.stats[key] = value;
      },
      end: (exit) => this.endByGame(exit),
    };
    this.instance = options.module.create(ctx);
    this.offFinish = options.clock.onFinish(() => this.endByClock());
    this.unbindVisibility = (options.bindVisibility ?? defaultBindVisibility)((hidden) =>
      this.onVisibility(hidden),
    );
  }

  start(): void {
    this.started = true;
    this.clock.start();
    this.instance.start();
  }

  pause(): void {
    if (this.userPaused || this.ended) {
      return;
    }
    this.userPaused = true;
    this.instance.pause();
    this.clock.pause();
  }

  resume(): void {
    if (!this.userPaused || this.ended) {
      return;
    }
    this.userPaused = false;
    if (!this.backgroundOpen) {
      this.instance.resume();
      this.clock.resume();
    }
  }

  record(): AttemptRecord {
    return {
      events: [...this.events],
      backgroundSpans: [...this.spans],
      stats: { ...this.stats },
      exit: this.exit ?? "aborted",
    };
  }

  destroy(): void {
    if (this.destroyed) {
      return;
    }
    this.destroyed = true;
    this.unbindVisibility();
    this.offFinish();
    this.instance.destroy();
    this.clock.dispose();
  }

  private collect(event: Record<string, unknown>): void {
    if (this.ended) {
      return;
    }
    if (this.events.length < MAX_EVENTS) {
      this.events.push(event);
    }
  }

  private onVisibility(hidden: boolean): void {
    // Before the run starts (intro ritual / countdown) a background flip
    // is pre-run noise: pausing then would freeze a clock with no origin
    // (negative elapsed) and park the game in a pause it never left.
    if (this.ended || !this.started) {
      return;
    }
    if (hidden) {
      this.backgroundOpen = true;
      this.spanStartElapsedMs = this.clock.elapsedMs();
      this.spanStartWallMs = this.now();
      if (!this.userPaused) {
        this.instance.pause();
        this.clock.pause();
      }
      return;
    }
    if (!this.backgroundOpen) {
      return;
    }
    this.backgroundOpen = false;
    if (this.spans.length < MAX_SPANS) {
      this.spans.push({
        elapsedMs: this.spanStartElapsedMs,
        durationMs: Math.max(0, this.now() - this.spanStartWallMs),
      });
    }
    if (!this.userPaused) {
      this.instance.resume();
      this.clock.resume();
    }
  }

  private endByClock(): void {
    this.finalize("completed");
  }

  private endByGame(exit: GameExit): void {
    if (this.ended) {
      return;
    }
    // A game-initiated end stops the clock too — no finish fires after.
    this.clock.dispose();
    this.finalize(exit);
  }

  private finalize(exit: GameExit): void {
    if (this.ended) {
      return;
    }
    this.ended = true;
    this.exit = exit;
    this.unbindVisibility();
    // Freeze the game face at end — without this, a game's own timers
    // (red_dot's target-swap chain) keep firing behind the end overlay.
    this.instance.pause();
    this.onEnd(this.record());
  }
}
