import type { GameClock } from "./clock";

// Engine-params keys are owned by the engine registry (single source: the
// metric spec): the client consumes them, never defines them. Values are
// strings on the wire; parsing goes through EngineParams, and a missing or
// non-numeric entry falls back to the game's compiled-in default — the
// legacy-calibrated constants, which are the semantic anchor.
export class EngineParams {
  private readonly kv: Readonly<Record<string, string>>;

  constructor(kv: Readonly<Record<string, string>>) {
    this.kv = kv;
  }

  num(key: string, fallback: number): number {
    const raw = this.kv[key];
    if (raw === undefined) {
      return fallback;
    }
    const value = Number(raw);
    return Number.isFinite(value) ? value : fallback;
  }

  str(key: string, fallback: string): string {
    return this.kv[key] ?? fallback;
  }

  bool(key: string, fallback: boolean): boolean {
    const raw = this.kv[key];
    return raw === undefined ? fallback : raw === "true";
  }
}

export type GameExit = "completed" | "aborted" | "failed";

// The one attempt-shaped record a run accumulates. Events are per-game
// shapes mirroring the frozen proto Detail items (gen types at the game
// side); the submit seam itself stays deferred until the submit contract
// freezes.
export interface AttemptRecord {
  readonly events: readonly Record<string, unknown>[];
  readonly backgroundSpans: readonly { elapsedMs: number; durationMs: number }[];
  readonly exit: GameExit;
}

// What a game receives from the host. The mount element is host-owned —
// games write only their own children into it.
export interface GameModuleCtx {
  readonly mount: HTMLElement;
  readonly clock: GameClock;
  readonly params: EngineParams;
  /** config.duration_seconds × 1000 — the single true source for the run length. */
  readonly durationMs: number;
  readonly locale: string;
  /** The clock's time source, handed through so reaction windows share it. */
  now(): number;
  /** Append one raw-detail event (per-game proto Detail shape). Capped at 4096. */
  collect(event: Record<string, unknown>): void;
  /** A game's own natural end (unused when the clock alone ends the run). */
  end(exit: GameExit): void;
}

export interface GameInstance {
  start(): void;
  pause(): void;
  resume(): void;
  destroy(): void;
}

// The plugin contract: a game is a pure TS class rendering into the mount
// element itself; the React host owns only the shell (HUD, overlays, route).
// Games must not import auth, services, or stores — the boundary law the
// biome override on games/** enforces.
export interface GameModule {
  readonly code: string;
  create(ctx: GameModuleCtx): GameInstance;
}
