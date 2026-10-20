// red_dot — the first GameModule: a red target spawns at a random padded
// position, lives for its configured lifetime, and any tap (hit or miss)
// records one event and swaps the target. Legacy semantic anchors: hit at
// 2× the visual radius, deviation at one decimal, reaction measured from
// the target's spawn, blank taps recorded then swapped. The run itself is
// owned by the clock — this class never ends it.
import type { GameInstance, GameModule, GameModuleCtx } from "../../contract";
import {
  deviationOf,
  isHit,
  outsidePaddedArea,
  type Point,
  RED_DOT_DEFAULTS,
  randomPosition,
  round1,
  type SpawnArea,
} from "./logic";

interface TargetState {
  readonly el: HTMLElement;
  readonly pos: Point;
  readonly spawnAt: number;
  /** Lifetime left when frozen by a pause; null while the timer runs. */
  frozenRemainingMs: number | null;
  swapTimer: ReturnType<typeof setTimeout> | null;
}

function spawnAreaOf(el: HTMLElement): SpawnArea {
  const rect = el.getBoundingClientRect();
  return { width: rect.width, height: rect.height };
}

export const redDotModule: GameModule = {
  code: "red_dot",
  create(ctx: GameModuleCtx): GameInstance {
    const params = {
      targetRadiusPx: ctx.params.num("target_size_px", RED_DOT_DEFAULTS.targetRadiusPx),
      targetLifetimeMs: ctx.params.num("target_lifetime_ms", RED_DOT_DEFAULTS.targetLifetimeMs),
      spawnPaddingPx: ctx.params.num("spawn_padding_px", RED_DOT_DEFAULTS.spawnPaddingPx),
    };

    const stage = document.createElement("div");
    stage.className = "h-full w-full touch-none overflow-hidden relative select-none";
    let started = false;
    let paused = false;
    let destroyed = false;
    let target: TargetState | null = null;
    let observer: ResizeObserver | null = null;

    function clearSwapTimer(state: TargetState): void {
      if (state.swapTimer !== null) {
        clearTimeout(state.swapTimer);
        state.swapTimer = null;
      }
    }

    function removeTarget(): void {
      if (target === null) {
        return;
      }
      clearSwapTimer(target);
      target.el.remove();
      target = null;
    }

    function spawnTarget(): void {
      removeTarget();
      const area = spawnAreaOf(stage);
      const pos = randomPosition(area, params);
      const el = document.createElement("div");
      const size = `${params.targetRadiusPx * 2}px`;
      el.className = "absolute rounded-full bg-danger";
      el.style.width = size;
      el.style.height = size;
      el.style.left = `${round1(pos.x - params.targetRadiusPx)}px`;
      el.style.top = `${round1(pos.y - params.targetRadiusPx)}px`;
      stage.appendChild(el);
      const state: TargetState = {
        el,
        pos,
        spawnAt: ctx.now(),
        frozenRemainingMs: null,
        swapTimer: null,
      };
      target = state;
      armLifetime(state);
    }

    function armLifetime(state: TargetState): void {
      if (paused || destroyed) {
        return;
      }
      // A lifetime that fully burned out during a pause swaps on the next
      // tick (legacy law: an expired target never lingers).
      const remaining = Math.max(0, state.frozenRemainingMs ?? params.targetLifetimeMs);
      state.frozenRemainingMs = null;
      state.swapTimer = setTimeout(() => {
        state.swapTimer = null;
        if (!(paused || destroyed)) {
          spawnTarget();
        }
      }, remaining);
    }

    function onPointerDown(event: PointerEvent): void {
      if (destroyed || !started || paused || target === null) {
        return;
      }
      const rect = stage.getBoundingClientRect();
      const click = { x: event.clientX - rect.left, y: event.clientY - rect.top };
      const deviation = deviationOf(target.pos, click);
      ctx.collect({
        target: { x: round1(target.pos.x), y: round1(target.pos.y) },
        click: { x: round1(click.x), y: round1(click.y) },
        reactionTimeMs: Math.max(0, Math.round(ctx.now() - target.spawnAt)),
        deviation,
        isHit: isHit(deviation, params),
      });
      // Any tap — hit or blank — swaps the target (legacy law).
      spawnTarget();
    }

    function onResize(): void {
      if (target === null || destroyed) {
        return;
      }
      // A shrunk stage invalidates off-screen targets: respawn in-bounds.
      if (outsidePaddedArea(target.pos, spawnAreaOf(stage), params)) {
        spawnTarget();
      }
    }

    return {
      start(): void {
        if (destroyed || started) {
          return;
        }
        started = true;
        ctx.mount.appendChild(stage);
        stage.addEventListener("pointerdown", onPointerDown);
        observer = new ResizeObserver(onResize);
        observer.observe(stage);
        spawnTarget();
      },
      pause(): void {
        if (paused || destroyed) {
          return;
        }
        paused = true;
        if (target !== null) {
          // Freeze the lifetime mid-flight: park what is left, re-arm on
          // resume so a pause never burns a target's remaining time.
          if (target.swapTimer !== null) {
            clearTimeout(target.swapTimer);
            target.swapTimer = null;
            target.frozenRemainingMs = Math.max(
              0,
              target.spawnAt + params.targetLifetimeMs - ctx.now(),
            );
          }
        }
      },
      resume(): void {
        if (!paused || destroyed) {
          return;
        }
        paused = false;
        if (target !== null) {
          armLifetime(target);
        }
      },
      destroy(): void {
        if (destroyed) {
          return;
        }
        destroyed = true;
        removeTarget();
        observer?.disconnect();
        observer = null;
        stage.removeEventListener("pointerdown", onPointerDown);
        stage.remove();
      },
    };
  },
};
