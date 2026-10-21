// multi_target — priority firing under value tiers. Legacy semantic
// anchors: a spawn cadence drawn once per run, four tiers by weighted lot
// with their values printed on the face (the trap shows its −50), every
// target living its lifetime then vanishing, at most eight on screen with
// the oldest crowded out, an expired or crowded-out high-value target
// counting as a miss, picks recorded at the run clock's elapsed, and blank
// air taps counted locally but never collected. Positions are
// container-relative. The run itself is owned by the clock — this class
// never ends it.
import type { GameInstance, GameModule, GameModuleCtx } from "../../contract";
import {
  drawTier,
  MULTI_TARGET_DEFAULTS,
  type MultiTargetParams,
  type Point,
  randomCenter,
  type SpawnArea,
  spawnInterval,
  type TierSpec,
} from "./logic";

// Timer accounting follows the engine's pause law: deadlines are parked as
// remaining-ms at pause and re-armed at resume, so a pause never burns a
// target's lifetime or the spawn cadence.
interface LiveTarget {
  readonly id: string;
  readonly el: HTMLElement;
  readonly center: Point;
  readonly spec: TierSpec;
  readonly spawnAt: number;
  /** Absolute expiry (ctx.now base); kept fresh by every re-arm. */
  expiresAt: number;
  /** Lifetime left when parked by a pause; null while the timer runs. */
  parkedMs: number | null;
  expiryTimer: ReturnType<typeof setTimeout> | null;
}

function spawnAreaOf(el: HTMLElement): SpawnArea {
  const rect = el.getBoundingClientRect();
  return { width: rect.width, height: rect.height };
}

// Rings follow the legacy target look: white/20 on value tiers, a bright
// danger ring on the trap (legacy border-red-500, flattened to the danger fill).
const TIER_BG: Record<string, string> = {
  high: "bg-tier-high border-white/20",
  medium: "bg-tier-medium border-white/20",
  low: "bg-tier-low border-white/20",
  trap: "bg-tier-trap border-danger",
};

export const multiTargetModule: GameModule = {
  code: "multi_target",
  create(ctx: GameModuleCtx): GameInstance {
    const params: MultiTargetParams = {
      spawnIntervalMinMs: ctx.params.num(
        "spawn_interval_min_ms",
        MULTI_TARGET_DEFAULTS.spawnIntervalMinMs,
      ),
      spawnIntervalSpanMs: ctx.params.num(
        "spawn_interval_span_ms",
        MULTI_TARGET_DEFAULTS.spawnIntervalSpanMs,
      ),
      targetLifetimeMs: ctx.params.num(
        "target_lifetime_ms",
        MULTI_TARGET_DEFAULTS.targetLifetimeMs,
      ),
      maxOnScreen: ctx.params.num("max_on_screen", MULTI_TARGET_DEFAULTS.maxOnScreen),
    };
    // Drawn once per run — the legacy interval's delay was fixed at creation.
    const cadenceMs = spawnInterval(params);

    const stage = document.createElement("div");
    stage.className = "h-full w-full touch-none overflow-hidden relative select-none";
    let started = false;
    let paused = false;
    let destroyed = false;
    let missedHigh = 0;
    let incorrectClicks = 0;
    let targetSeq = 0;
    const targets: LiveTarget[] = [];
    let spawnTimer: ReturnType<typeof setTimeout> | null = null;
    let spawnDeadline = 0;
    let spawnParkedMs: number | null = null;

    function clearTimer(timer: ReturnType<typeof setTimeout> | null): null {
      if (timer !== null) {
        clearTimeout(timer);
      }
      return null;
    }

    function removeTarget(target: LiveTarget): void {
      const index = targets.indexOf(target);
      if (index >= 0) {
        targets.splice(index, 1);
      }
      target.expiryTimer = clearTimer(target.expiryTimer);
      target.el.remove();
    }

    function countMissedHigh(target: LiveTarget): void {
      if (target.spec.tier === "high") {
        // Only a high-value target's loss is a miss — legacy law.
        missedHigh += 1;
        ctx.stat("missedHigh", missedHigh);
      }
    }

    function onExpiry(target: LiveTarget): void {
      target.expiryTimer = null;
      target.parkedMs = null;
      removeTarget(target);
      countMissedHigh(target);
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

    function armSpawn(): void {
      if (paused || destroyed || spawnTimer !== null) {
        return;
      }
      const remaining = Math.max(0, spawnParkedMs ?? cadenceMs);
      spawnParkedMs = null;
      spawnDeadline = ctx.now() + remaining;
      spawnTimer = setTimeout(onSpawnTick, remaining);
    }

    function onSpawnTick(): void {
      spawnTimer = null;
      if (paused || destroyed) {
        return;
      }
      const spec = drawTier();
      const area = spawnAreaOf(stage);
      const pos = randomCenter(area, spec.radiusPx);
      targetSeq += 1;
      const el = document.createElement("div");
      const size = `${spec.radiusPx * 2}px`;
      el.className = `absolute flex cursor-pointer items-center justify-center rounded-full border-2 font-bold text-ink text-sm select-none ${TIER_BG[spec.tier]}`;
      el.style.width = size;
      el.style.height = size;
      el.style.left = `${pos.x - spec.radiusPx}px`;
      el.style.top = `${pos.y - spec.radiusPx}px`;
      el.textContent = String(spec.value);
      const target: LiveTarget = {
        id: String(targetSeq),
        el,
        center: pos,
        spec,
        spawnAt: ctx.now(),
        expiresAt: ctx.now() + params.targetLifetimeMs,
        parkedMs: null,
        expiryTimer: null,
      };
      el.addEventListener("pointerdown", (event) => onTargetTap(event, target));
      // Crowd the oldest out at the cap — its loss counts like an expiry.
      while (targets.length >= params.maxOnScreen) {
        const oldest = targets[0];
        if (oldest === undefined) {
          break;
        }
        removeTarget(oldest);
        countMissedHigh(oldest);
      }
      targets.push(target);
      stage.appendChild(el);
      armExpiry(target);
      armSpawn();
    }

    function onTargetTap(event: PointerEvent, target: LiveTarget): void {
      if (destroyed || !started || paused) {
        return;
      }
      event.stopPropagation();
      event.preventDefault();
      removeTarget(target);
      ctx.collect({
        targetId: target.id,
        tier: target.spec.tier,
        isHit: true,
        elapsedMs: Math.max(0, Math.round(ctx.clock.elapsedMs())),
      });
    }

    function onStageTap(event: PointerEvent): void {
      if (destroyed || !started || paused) {
        return;
      }
      // Legacy law: a blank tap counts as an incorrect click only while
      // targets are on screen; non-primary mouse buttons are ignored.
      if (event.pointerType === "mouse" && event.button !== 0) {
        return;
      }
      if (targets.length === 0) {
        return;
      }
      incorrectClicks += 1;
      ctx.stat("incorrectClicks", incorrectClicks);
    }

    return {
      start(): void {
        if (destroyed || started) {
          return;
        }
        started = true;
        ctx.mount.appendChild(stage);
        stage.addEventListener("pointerdown", onStageTap);
        armSpawn();
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
        for (const target of targets) {
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
        armSpawn();
        for (const target of targets) {
          armExpiry(target);
        }
      },
      destroy(): void {
        if (destroyed) {
          return;
        }
        destroyed = true;
        spawnTimer = clearTimer(spawnTimer);
        for (const target of [...targets]) {
          removeTarget(target);
        }
        stage.removeEventListener("pointerdown", onStageTap);
        stage.remove();
      },
    };
  },
};
