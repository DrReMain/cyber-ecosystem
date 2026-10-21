// flash_aiming — snap shots at flashing targets. Legacy semantic anchors:
// a spawn cadence of one target every 800 ms with only the newest visible,
// taps recorded only on the target itself (blank air is ignored, unlike
// red_dot), aim time measured from the target's spawn, and every spawned
// target either clicked or expiring as a miss — a replaced target is gone
// from view but its miss still lands at its lifetime's end. Positions are
// container-relative; no rotate hack. The run itself is owned by the clock —
// this class never ends it.
import type { GameInstance, GameModule, GameModuleCtx } from "../../contract";
import {
  deviationOf,
  FLASH_AIMING_DEFAULTS,
  type FlashAimingParams,
  type Point,
  randomCenter,
  randomLifetime,
  randomSize,
  type SpawnArea,
} from "./logic";

// Timer accounting follows the engine's pause law: deadlines are parked as
// remaining-ms at pause and re-armed at resume, so a pause never burns a
// target's lifetime or the spawn cadence.
interface LiveTarget {
  readonly id: string;
  readonly el: HTMLElement;
  readonly center: Point;
  readonly size: number;
  readonly spawnAt: number;
  /** Absolute expiry (ctx.now base); kept fresh by every re-arm. */
  expiresAt: number;
  /** Lifetime left when parked by a pause; null while the timer runs. */
  parkedMs: number | null;
  expiryTimer: ReturnType<typeof setTimeout> | null;
  clicked: boolean;
}

function spawnAreaOf(el: HTMLElement): SpawnArea {
  const rect = el.getBoundingClientRect();
  return { width: rect.width, height: rect.height };
}

export const flashAimingModule: GameModule = {
  code: "flash_aiming",
  create(ctx: GameModuleCtx): GameInstance {
    const params: FlashAimingParams = {
      spawnIntervalMs: ctx.params.num("spawn_interval_ms", FLASH_AIMING_DEFAULTS.spawnIntervalMs),
      sizeMinPx: ctx.params.num("target_size_min_px", FLASH_AIMING_DEFAULTS.sizeMinPx),
      sizeSpanPx: ctx.params.num("target_size_span_px", FLASH_AIMING_DEFAULTS.sizeSpanPx),
      lifetimeMinMs: ctx.params.num("target_lifetime_min_ms", FLASH_AIMING_DEFAULTS.lifetimeMinMs),
      lifetimeSpanMs: ctx.params.num(
        "target_lifetime_span_ms",
        FLASH_AIMING_DEFAULTS.lifetimeSpanMs,
      ),
      centerHitDeviationPx: ctx.params.num(
        "center_hit_deviation_px",
        FLASH_AIMING_DEFAULTS.centerHitDeviationPx,
      ),
    };

    const stage = document.createElement("div");
    stage.className = "h-full w-full touch-none overflow-hidden relative select-none";
    let started = false;
    let paused = false;
    let destroyed = false;
    let misses = 0;
    let targetSeq = 0;
    let visible: LiveTarget | null = null;
    const pending = new Set<LiveTarget>();
    let spawnTimer: ReturnType<typeof setTimeout> | null = null;
    let spawnDeadline = 0;
    let spawnParkedMs: number | null = null;

    function clearTimer(timer: ReturnType<typeof setTimeout> | null): null {
      if (timer !== null) {
        clearTimeout(timer);
      }
      return null;
    }

    function dropFromView(target: LiveTarget): void {
      if (visible === target) {
        visible = null;
      }
      target.el.remove();
    }

    function onExpiry(target: LiveTarget): void {
      target.expiryTimer = null;
      target.parkedMs = null;
      pending.delete(target);
      dropFromView(target);
      if (!target.clicked) {
        // An unclicked target is a miss the moment its lifetime ends —
        // legacy law, replacement never rescues it.
        misses += 1;
        ctx.stat("misses", misses);
      }
    }

    function armExpiry(target: LiveTarget): void {
      if (paused || destroyed) {
        return;
      }
      const remaining = Math.max(0, target.parkedMs ?? target.expiresAt - ctx.now());
      target.parkedMs = null;
      target.expiresAt = ctx.now() + remaining;
      target.expiryTimer = setTimeout(() => onExpiry(target), remaining);
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
      const size = randomSize(params);
      const pos = randomCenter(spawnAreaOf(stage), size);
      targetSeq += 1;
      const el = document.createElement("div");
      el.className = "absolute rounded-full bg-danger cursor-pointer";
      el.style.width = `${size}px`;
      el.style.height = `${size}px`;
      el.style.left = `${pos.x - size / 2}px`;
      el.style.top = `${pos.y - size / 2}px`;
      const target: LiveTarget = {
        id: String(targetSeq),
        el,
        center: pos,
        size,
        spawnAt: ctx.now(),
        expiresAt: ctx.now() + randomLifetime(params),
        parkedMs: null,
        expiryTimer: null,
        clicked: false,
      };
      el.addEventListener("pointerdown", (event) => onTargetTap(event, target));
      // Only the newest target stays visible (legacy law: one at a time).
      if (visible !== null) {
        dropFromView(visible);
      }
      visible = target;
      stage.appendChild(el);
      pending.add(target);
      armExpiry(target);
      armSpawn(params.spawnIntervalMs);
    }

    function onTargetTap(event: PointerEvent, target: LiveTarget): void {
      if (destroyed || !started || paused || target.clicked) {
        return;
      }
      event.stopPropagation();
      target.clicked = true;
      target.expiryTimer = clearTimer(target.expiryTimer);
      pending.delete(target);
      dropFromView(target);
      const rect = stage.getBoundingClientRect();
      const click = { x: event.clientX - rect.left, y: event.clientY - rect.top };
      ctx.collect({
        targetId: target.id,
        target: { x: target.center.x, y: target.center.y },
        targetSize: target.size,
        click,
        deviation: deviationOf(target.center, click),
        aimTimeMs: Math.max(0, Math.round(ctx.now() - target.spawnAt)),
      });
      // A click does not pull the next spawn forward — the cadence owns it.
    }

    return {
      start(): void {
        if (destroyed || started) {
          return;
        }
        started = true;
        ctx.mount.appendChild(stage);
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
        for (const target of pending) {
          if (target.expiryTimer !== null) {
            target.parkedMs = Math.max(0, target.expiresAt - ctx.now());
            target.expiryTimer = clearTimer(target.expiryTimer);
          }
        }
      },
      resume(): void {
        if (!paused || destroyed) {
          return;
        }
        paused = false;
        armSpawn(spawnParkedMs ?? params.spawnIntervalMs);
        spawnParkedMs = null;
        for (const target of pending) {
          armExpiry(target);
        }
      },
      destroy(): void {
        if (destroyed) {
          return;
        }
        destroyed = true;
        spawnTimer = clearTimer(spawnTimer);
        for (const target of pending) {
          target.expiryTimer = clearTimer(target.expiryTimer);
          dropFromView(target);
        }
        pending.clear();
        visible = null;
        stage.remove();
      },
    };
  },
};
