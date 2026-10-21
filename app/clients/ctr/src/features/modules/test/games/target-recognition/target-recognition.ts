// target_recognition — friend/foe discrimination. Legacy semantic anchors:
// one target visible at a time drawn by type lot (enemy 60% / friendly 25%
// / neutral 15%), replaced on a 700 ms cadence, the first target appearing
// immediately at start; taps are hit-tested geometrically at the stage
// level (the slop zone reaches past the small visual), a hit records the
// click with its type and reaction — correctness is the server's judgment,
// never asserted here — and a blank tap counts locally without a fact
// (the contract's target_id cannot name air). Positions are
// container-relative. The run itself is owned by the clock.
import type { GameInstance, GameModule, GameModuleCtx } from "../../contract";
import {
  distance,
  drawType,
  isHit,
  type Point,
  randomCenter,
  type SpawnArea,
  TARGET_RECOGNITION_DEFAULTS,
  type TargetRecognitionParams,
  TYPE_SPECS,
  type TypeSpec,
} from "./logic";

// Timer accounting follows the engine's pause law: deadlines are parked as
// remaining-ms at pause and re-armed at resume.
interface LiveTarget {
  readonly id: string;
  readonly el: HTMLElement;
  readonly center: Point;
  readonly spec: TypeSpec;
  readonly spawnAt: number;
  /** Absolute expiry (ctx.now base); kept fresh by every re-arm. */
  expiresAt: number;
  /** Lifetime left when parked by a pause; null while the timer runs. */
  parkedMs: number | null;
  expiryTimer: ReturnType<typeof setTimeout> | null;
}

const TYPE_BG: Record<string, string> = {
  enemy: "bg-rec-enemy",
  friendly: "bg-rec-friendly",
  neutral: "bg-rec-neutral",
};

function spawnAreaOf(el: HTMLElement): SpawnArea {
  const rect = el.getBoundingClientRect();
  return { width: rect.width, height: rect.height };
}

export const targetRecognitionModule: GameModule = {
  code: "target_recognition",
  create(ctx: GameModuleCtx): GameInstance {
    const params: TargetRecognitionParams = {
      spawnIntervalMs: ctx.params.num(
        "spawn_interval_ms",
        TARGET_RECOGNITION_DEFAULTS.spawnIntervalMs,
      ),
      spawnPaddingPx: ctx.params.num(
        "spawn_padding_px",
        TARGET_RECOGNITION_DEFAULTS.spawnPaddingPx,
      ),
      clickDetectionRangePx: ctx.params.num(
        "click_detection_range_px",
        TARGET_RECOGNITION_DEFAULTS.clickDetectionRangePx,
      ),
      invalidClickMarginPx: ctx.params.num(
        "invalid_click_margin_px",
        TARGET_RECOGNITION_DEFAULTS.invalidClickMarginPx,
      ),
    };

    const stage = document.createElement("div");
    stage.className = "h-full w-full touch-none overflow-hidden relative select-none";
    let started = false;
    let paused = false;
    let destroyed = false;
    let targetSeq = 0;
    const spawned = { enemy: 0, friendly: 0, neutral: 0 };
    let blankClicks = 0;
    let target: LiveTarget | null = null;
    let spawnTimer: ReturnType<typeof setTimeout> | null = null;
    let spawnDeadline = 0;
    let spawnParkedMs: number | null = null;

    function clearTimer(timer: ReturnType<typeof setTimeout> | null): null {
      if (timer !== null) {
        clearTimeout(timer);
      }
      return null;
    }

    function removeTarget(): void {
      if (target === null) {
        return;
      }
      target.expiryTimer = clearTimer(target.expiryTimer);
      target.el.remove();
      target = null;
    }

    function onExpiry(state: LiveTarget): void {
      state.expiryTimer = null;
      state.parkedMs = null;
      // The cadence normally replaces a target first; expiry is the safety
      // net at the declared per-type lifespan.
      if (target === state) {
        removeTarget();
      }
    }

    function armExpiry(state: LiveTarget): void {
      if (paused || destroyed) {
        return;
      }
      const remaining = Math.max(0, state.parkedMs ?? state.expiresAt - ctx.now());
      state.parkedMs = null;
      state.expiresAt = ctx.now() + remaining;
      state.expiryTimer = setTimeout(() => onExpiry(state), remaining);
    }

    function armSpawn(remainingMs: number): void {
      if (paused || destroyed || spawnTimer !== null) {
        return;
      }
      const remaining = Math.max(0, remainingMs);
      spawnDeadline = ctx.now() + remaining;
      spawnTimer = setTimeout(onSpawnTick, remaining);
    }

    function onSpawnTick(): void {
      spawnTimer = null;
      if (paused || destroyed) {
        return;
      }
      spawnOnce();
      armSpawn(params.spawnIntervalMs);
    }

    function spawnOnce(): void {
      const type = drawType();
      spawned[type] += 1;
      ctx.stat(`spawned_${type}`, spawned[type]);
      const spec = TYPE_SPECS.find((entry) => entry.type === type) as TypeSpec;
      const pos = randomCenter(spawnAreaOf(stage), params.spawnPaddingPx);
      targetSeq += 1;
      const el = document.createElement("div");
      const size = `${spec.diameterPx}px`;
      el.className = `absolute flex items-center justify-center rounded-full border-2 border-white/20 font-bold text-ink text-[10px] select-none ${TYPE_BG[type]}`;
      el.style.width = size;
      el.style.height = size;
      el.style.left = `${pos.x - spec.diameterPx / 2}px`;
      el.style.top = `${pos.y - spec.diameterPx / 2}px`;
      el.textContent = spec.value > 0 ? `+${spec.value}` : `${spec.value}`;
      removeTarget();
      const state: LiveTarget = {
        id: String(targetSeq),
        el,
        center: pos,
        spec,
        spawnAt: ctx.now(),
        expiresAt: ctx.now() + spec.lifespanMs,
        parkedMs: null,
        expiryTimer: null,
      };
      target = state;
      stage.appendChild(el);
      armExpiry(state);
    }

    function onStageTap(event: PointerEvent): void {
      if (destroyed || !started || paused) {
        return;
      }
      if (event.pointerType === "mouse" && event.button !== 0) {
        return;
      }
      const rect = stage.getBoundingClientRect();
      const click = { x: event.clientX - rect.left, y: event.clientY - rect.top };
      if (
        target !== null &&
        isHit(distance(target.center, click), target.spec.diameterPx, params)
      ) {
        const current = target;
        removeTarget();
        ctx.collect({
          targetId: current.id,
          targetType: current.spec.type,
          reactionTimeMs: Math.max(0, Math.round(ctx.now() - current.spawnAt)),
          elapsedMs: Math.max(0, Math.round(ctx.clock.elapsedMs())),
        });
        return;
      }
      // Blank air (or an edge-grazing tap): local fact only — the frozen
      // contract's target_id cannot name air (legacy law carried over).
      blankClicks += 1;
      ctx.stat("blankClicks", blankClicks);
    }

    return {
      start(): void {
        if (destroyed || started) {
          return;
        }
        started = true;
        ctx.mount.appendChild(stage);
        stage.addEventListener("pointerdown", onStageTap);
        // The first target shows immediately (legacy law), the cadence after.
        spawnOnce();
        armSpawn(params.spawnIntervalMs);
      },
      pause(): void {
        if (paused || destroyed) {
          return;
        }
        paused = true;
        if (spawnTimer !== null) {
          spawnParkedMs = Math.max(0, spawnDeadline - ctx.now());
          spawnTimer = clearTimer(spawnTimer);
        }
        if (target !== null && target.expiryTimer !== null) {
          target.parkedMs = Math.max(0, target.expiresAt - ctx.now());
          target.expiryTimer = clearTimer(target.expiryTimer);
        }
      },
      resume(): void {
        if (!paused || destroyed) {
          return;
        }
        paused = false;
        armSpawn(spawnParkedMs ?? params.spawnIntervalMs);
        spawnParkedMs = null;
        if (target !== null) {
          armExpiry(target);
        }
      },
      destroy(): void {
        if (destroyed) {
          return;
        }
        destroyed = true;
        spawnTimer = clearTimer(spawnTimer);
        removeTarget();
        stage.removeEventListener("pointerdown", onStageTap);
        stage.remove();
      },
    };
  },
};
