// target_recognition — friend/foe discrimination. Legacy semantic anchors:
// one target visible at a time drawn by type lot (enemy 60% / friendly 25%
// / neutral 15%), replaced on a 700 ms cadence, the first target appearing
// immediately at start; taps are hit-tested geometrically at the stage
// level (the slop zone reaches past the small visual), a hit records the
// click with its type and reaction — correctness is the server's judgment,
// never asserted here — and a blank tap counts locally without a fact
// (the contract's target_id cannot name air). Positions are
// container-relative. The run itself is owned by the clock.

import { m } from "#/paraglide/messages";
import type { GameInstance, GameModule, GameModuleCtx } from "../../contract";
import { createLucideIcon, type IconNode } from "../lucide-icon";
import {
  distance,
  drawType,
  isHit,
  outsidePaddedArea,
  type Point,
  randomCenter,
  type SpawnArea,
  TARGET_RECOGNITION_DEFAULTS,
  type TargetRecognitionParams,
  type TargetType,
  TYPE_SPECS,
  type TypeSpec,
} from "./logic";

// D-AG B1: the in-target form is the type's lucide icon (lucide 1.49
// skull / shield / user node trees, verbatim geometry) — shape coding
// reads without color and the value never rides the target; the bottom
// legend bar teaches the reading. The legacy emoji were placeholder art
// and stay unbuilt (zero-emoji ruling).
const SKULL: IconNode = [
  ["path", { d: "m12.5 17-.5-1-.5 1h1z" }],
  [
    "path",
    {
      d: "M15 22a1 1 0 0 0 1-1v-1a2 2 0 0 0 1.56-3.25 8 8 0 1 0-11.12 0A2 2 0 0 0 8 20v1a1 1 0 0 0 1 1z",
    },
  ],
  ["circle", { cx: "15", cy: "12", r: "1" }],
  ["circle", { cx: "9", cy: "12", r: "1" }],
];
const SHIELD: IconNode = [
  [
    "path",
    {
      d: "M20 13c0 5-3.5 7.5-7.66 8.95a1 1 0 0 1-.67-.01C7.5 20.5 4 18 4 13V6a1 1 0 0 1 1-1c2 0 4.5-1.2 6.24-2.72a1.17 1.17 0 0 1 1.52 0C14.51 3.81 17 5 19 5a1 1 0 0 1 1 1z",
    },
  ],
];
const USER: IconNode = [
  ["path", { d: "M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2" }],
  ["circle", { cx: "12", cy: "7", r: "4" }],
];
const TYPE_ICONS: Record<TargetType, IconNode> = {
  enemy: SKULL,
  friendly: SHIELD,
  neutral: USER,
};

const TYPE_BG: Record<string, string> = {
  enemy: "bg-rec-enemy",
  friendly: "bg-rec-friendly",
  neutral: "bg-rec-neutral",
};

const LEGEND_LABEL: Record<TargetType, () => string> = {
  enemy: m.test_rec_legend_enemy,
  friendly: m.test_rec_legend_friendly,
  neutral: m.test_rec_legend_neutral,
};

// Timer accounting follows the engine's pause law: deadlines are parked as
// remaining-ms at pause and re-armed at resume.
interface LiveTarget {
  readonly id: string;
  readonly el: HTMLElement;
  center: Point;
  readonly spec: TypeSpec;
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
    // D-AG B1: the reading lives in the bottom legend bar (legacy form:
    // black 60% + blur, one entry per type — color dot + name + value) so
    // the target face stays a spatial field of skull/shield/user shapes.
    const legend = document.createElement("div");
    legend.className =
      "pointer-events-none absolute bottom-3 left-1/2 flex -translate-x-1/2 items-center gap-3 rounded-full bg-black/60 px-3.5 py-1.5 backdrop-blur-sm";
    for (const spec of TYPE_SPECS) {
      const entry = document.createElement("span");
      entry.className = "flex items-center gap-1.5";
      const dot = document.createElement("span");
      dot.className = `h-2 w-2 rounded-full ${TYPE_BG[spec.type]}`;
      const label = document.createElement("span");
      label.className = "text-[10px] font-semibold text-white/90";
      label.textContent = `${LEGEND_LABEL[spec.type]()} ${
        spec.value > 0 ? `+${spec.value}` : `−${Math.abs(spec.value)}`
      }`;
      entry.append(dot, label);
      legend.appendChild(entry);
    }
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
    let observer: ResizeObserver | null = null;

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
      // One anchored wrapper carries both visuals: the target disc and
      // the faint detection ring — D-AG A1, the legacy "subtle
      // detection-range indicator" that honestly shows the tap slop the
      // geometric hit test actually grants (diameter = detection range
      // × 2, pure visual layer, zero judgment/telemetry impact).
      const el = document.createElement("div");
      el.className = "pointer-events-none absolute select-none";
      el.style.left = `${pos.x}px`;
      el.style.top = `${pos.y}px`;
      el.style.transform = "translate(-50%, -50%)";
      const ring = document.createElement("div");
      const ringSize = `${params.clickDetectionRangePx * 2}px`;
      ring.className = "absolute rounded-full border border-white/20";
      ring.style.width = ringSize;
      ring.style.height = ringSize;
      ring.style.left = "50%";
      ring.style.top = "50%";
      ring.style.transform = "translate(-50%, -50%)";
      const disc = document.createElement("div");
      const size = `${spec.diameterPx}px`;
      disc.className = `absolute flex items-center justify-center rounded-full border-2 border-white/20 text-white ${TYPE_BG[type]}`;
      disc.style.width = size;
      disc.style.height = size;
      disc.style.left = "50%";
      disc.style.top = "50%";
      disc.style.transform = "translate(-50%, -50%)";
      disc.appendChild(createLucideIcon(TYPE_ICONS[type], Math.round(spec.diameterPx * 0.55)));
      el.append(ring, disc);
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

    function onResize(): void {
      if (destroyed || target === null) {
        return;
      }
      // A shrunk stage strands the target off-screen where the hit test
      // can never be satisfied — move the same target (same id and spawn,
      // so the reaction clock keeps running truthfully) to a fresh
      // in-bounds spot.
      if (outsidePaddedArea(target.center, spawnAreaOf(stage), params.spawnPaddingPx)) {
        const pos = randomCenter(spawnAreaOf(stage), params.spawnPaddingPx);
        target.center = pos;
        target.el.style.left = `${pos.x}px`;
        target.el.style.top = `${pos.y}px`;
      }
    }

    return {
      start(): void {
        if (destroyed || started) {
          return;
        }
        started = true;
        ctx.mount.appendChild(stage);
        stage.appendChild(legend);
        stage.addEventListener("pointerdown", onStageTap);
        observer = new ResizeObserver(onResize);
        observer.observe(stage);
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
        observer?.disconnect();
        observer = null;
        stage.removeEventListener("pointerdown", onStageTap);
        stage.remove();
      },
    };
  },
};
