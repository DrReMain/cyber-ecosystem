// tracking_moving — the joystick steers, the button fires, and the dot
// waits to be outplayed. Two frame kinds ride one stream: every shot
// carries its hit verdict and the dot's spawn-to-kill lifetime (the
// clock-based lifetime, so a pause between spawn and kill never counts
// as aim time), while the cadence-sampled movement frames carry only
// the player polyline — hit presence is the discriminator the frozen
// contract judges shape by. A fire into the respawn gap is local
// feedback only: the shot frame cannot name a missing dot, so no fact
// is minted from air (the blank-tap family law). A shrunk stage
// repositions a stranded dot in-bounds keeping its spawn stamp, so the
// lifetime clock keeps running truthfully.

import { m } from "#/paraglide/messages";
import type { GameInstance, GameModule, GameModuleCtx } from "../../contract";
import { createFrameLoop } from "../../frame-loop";
import { createLucideIcon, type IconNode } from "../lucide-icon";
import {
  distance,
  isHit,
  joystickVector,
  outsidePaddedArea,
  type Point,
  randomCenter,
  round1,
  type SpawnArea,
  stepCrosshair,
  TRACKING_MOVING_DEFAULTS,
} from "./logic";

// lucide "target" node tree, verbatim geometry — the fire button's mark.
const TARGET: IconNode = [
  ["circle", { cx: "12", cy: "12", r: "10" }],
  ["circle", { cx: "12", cy: "12", r: "6" }],
  ["circle", { cx: "12", cy: "12", r: "2" }],
];

const BASE_RADIUS_PX = 48;

function spawnAreaOf(el: HTMLElement): SpawnArea {
  const rect = el.getBoundingClientRect();
  return { width: rect.width, height: rect.height };
}

interface DotState {
  readonly el: HTMLElement;
  center: Point;
  /** Clock elapsed at spawn — the lifetime's zero point. */
  readonly spawnElapsedMs: number;
}

interface FaceParts {
  readonly stage: HTMLElement;
  readonly chipValue: HTMLElement;
  readonly dotEl: HTMLElement;
  readonly crosshair: HTMLElement;
  readonly stick: HTMLElement;
  readonly knob: HTMLElement;
  readonly fireBtn: HTMLElement;
}

// The game's face: stage, read-out chip, dot, crosshair, joystick, and
// fire button — built once per attempt; the run's behavior wires onto
// these parts and never rebuilds them.
function buildFace(dotSizePx: number): FaceParts {
  const stage = document.createElement("div");
  stage.className = "h-full w-full touch-none overflow-hidden relative select-none";
  // The read-out strip: hits out of shots fired.
  const chip = document.createElement("div");
  chip.className =
    "pointer-events-none absolute top-3 left-1/2 flex -translate-x-1/2 items-center gap-1.5 rounded-full bg-black/60 px-3.5 py-1.5 backdrop-blur-sm";
  const chipLabel = document.createElement("span");
  chipLabel.className = "text-[10px] font-semibold text-white/90";
  chipLabel.textContent = m.test_tm_hits();
  const chipValue = document.createElement("span");
  chipValue.className = "text-status-ok text-[10px] font-bold";
  chipValue.textContent = "0/0";
  chip.append(chipLabel, chipValue);
  stage.appendChild(chip);
  // The dot: a glow layer behind a green disc, anchored at its center.
  const dotEl = document.createElement("div");
  dotEl.className = "pointer-events-none absolute select-none";
  dotEl.style.transform = "translate(-50%, -50%)";
  const glow = document.createElement("div");
  glow.className =
    "absolute left-1/2 top-1/2 h-full w-full -translate-x-1/2 -translate-y-1/2 rounded-full bg-green-400/40 blur-xl";
  const disc = document.createElement("div");
  disc.className =
    "absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-green-200 bg-gradient-to-br from-green-300 to-green-500";
  const dotSize = `${dotSizePx}px`;
  disc.style.width = dotSize;
  disc.style.height = dotSize;
  dotEl.append(glow, disc);
  // The crosshair: a currentColor cross that flashes red on a miss.
  const crosshair = document.createElement("div");
  crosshair.className = "pointer-events-none absolute h-10 w-10 text-status-ok left-0 top-0";
  const hBar = document.createElement("div");
  hBar.className = "absolute top-1/2 left-0 h-0.5 w-full -translate-y-1/2 bg-current";
  const vBar = document.createElement("div");
  vBar.className = "absolute left-1/2 top-0 h-full w-0.5 -translate-x-1/2 bg-current";
  const center = document.createElement("div");
  center.className =
    "absolute left-1/2 top-1/2 h-1.5 w-1.5 -translate-x-1/2 -translate-y-1/2 rounded-full bg-current";
  crosshair.append(hBar, vBar, center);
  stage.appendChild(crosshair);
  // The joystick: bottom-left base with a knob that follows the finger.
  const stickBox = document.createElement("div");
  stickBox.className = "absolute bottom-6 left-5 flex flex-col items-center gap-1";
  const stick = document.createElement("div");
  stick.className =
    "relative h-24 w-24 rounded-full border-2 border-white/40 bg-black/50 backdrop-blur-sm";
  const knob = document.createElement("div");
  knob.className =
    "absolute left-1/2 top-1/2 h-8 w-8 -translate-x-1/2 -translate-y-1/2 rounded-full bg-white/80";
  stick.appendChild(knob);
  const stickCaption = document.createElement("span");
  stickCaption.className = "text-ink-tertiary text-[10px]";
  stickCaption.textContent = m.test_tm_move();
  stickBox.append(stick, stickCaption);
  stage.appendChild(stickBox);
  // The fire button: bottom-right, pressed not clicked (latency is aim).
  const fireBox = document.createElement("div");
  fireBox.className = "absolute right-5 bottom-6 flex flex-col items-center gap-1";
  const fireBtn = document.createElement("div");
  fireBtn.className =
    "flex h-[88px] w-[88px] cursor-pointer items-center justify-center rounded-full border-4 border-white/50 bg-danger shadow-2xl active:scale-90";
  fireBtn.appendChild(createLucideIcon(TARGET, 40));
  const fireCaption = document.createElement("span");
  fireCaption.className = "text-ink-tertiary text-[10px]";
  fireCaption.textContent = m.test_tm_fire();
  fireBox.append(fireBtn, fireCaption);
  stage.appendChild(fireBox);
  return { stage, chipValue, dotEl, crosshair, stick, knob, fireBtn };
}

export const trackingMovingModule: GameModule = {
  code: "tracking_moving",
  create(ctx: GameModuleCtx): GameInstance {
    const params = {
      hitThresholdPx: ctx.params.num("hit_threshold_px", TRACKING_MOVING_DEFAULTS.hitThresholdPx),
      dotSizePx: ctx.params.num("dot_size_px", TRACKING_MOVING_DEFAULTS.dotSizePx),
      spawnPaddingPx: ctx.params.num("spawn_padding_px", TRACKING_MOVING_DEFAULTS.spawnPaddingPx),
      crosshairSpeedPxS: ctx.params.num(
        "crosshair_speed_px_s",
        TRACKING_MOVING_DEFAULTS.crosshairSpeedPxS,
      ),
      movementSampleIntervalMs: ctx.params.num(
        "movement_sample_interval_ms",
        TRACKING_MOVING_DEFAULTS.movementSampleIntervalMs,
      ),
      respawnDelayMaxMs: ctx.params.num(
        "respawn_delay_max_ms",
        TRACKING_MOVING_DEFAULTS.respawnDelayMaxMs,
      ),
    };

    const { stage, chipValue, dotEl, crosshair, stick, knob, fireBtn } = buildFace(
      params.dotSizePx,
    );

    let started = false;
    let paused = false;
    let destroyed = false;
    let dot: DotState | null = null;
    let aim: Point = { x: 0, y: 0 };
    let vector: Point = { x: 0, y: 0 };
    let sampleCount = 0;
    // Run elapsed when the movement grid began; null until the first frame.
    let gridStartMs: number | null = null;
    let hits = 0;
    let shots = 0;
    let flashTimer: ReturnType<typeof setTimeout> | null = null;
    let respawnTimer: ReturnType<typeof setTimeout> | null = null;
    let respawnParkedMs: number | null = null;
    let respawnDeadline = 0;
    let observer: ResizeObserver | null = null;

    function place(el: HTMLElement, p: Point): void {
      el.style.left = `${round1(p.x)}px`;
      el.style.top = `${round1(p.y)}px`;
    }

    function setChip(): void {
      chipValue.textContent = `${hits}/${shots}`;
    }

    function flashHit(miss: boolean): void {
      crosshair.classList.toggle("text-danger", miss);
      if (flashTimer !== null) {
        clearTimeout(flashTimer);
      }
      flashTimer = setTimeout(() => {
        flashTimer = null;
        crosshair.classList.remove("text-danger");
      }, 200);
    }

    function spawnDot(): void {
      respawnTimer = null;
      respawnParkedMs = null;
      const area = spawnAreaOf(stage);
      const pos = randomCenter(area, params.spawnPaddingPx, Math.random);
      dot = {
        el: dotEl,
        center: pos,
        spawnElapsedMs: ctx.clock.elapsedMs(),
      };
      place(dotEl, pos);
      stage.appendChild(dotEl);
    }

    function removeDot(): void {
      dot?.el.remove();
      dot = null;
    }

    function armRespawn(delayMs: number): void {
      if (paused || destroyed) {
        return;
      }
      respawnDeadline = ctx.now() + Math.max(0, delayMs);
      respawnTimer = setTimeout(spawnDot, Math.max(0, delayMs));
    }

    function onFire(): void {
      if (destroyed || !started || paused) {
        return;
      }
      if (dot === null) {
        flashHit(true);
        return;
      }
      shots += 1;
      const deviation = distance(dot.center, aim);
      const hit = isHit(deviation, params);
      ctx.collect({
        elapsedMs: Math.round(ctx.clock.elapsedMs()),
        target: { x: round1(dot.center.x), y: round1(dot.center.y) },
        crosshair: { x: round1(aim.x), y: round1(aim.y) },
        deviation: round1(deviation),
        player: { x: round1(aim.x), y: round1(aim.y) },
        hit,
        dotLifetimeMs: Math.max(0, Math.round(ctx.clock.elapsedMs() - dot.spawnElapsedMs)),
      });
      if (hit) {
        hits += 1;
        removeDot();
        armRespawn(Math.random() * params.respawnDelayMaxMs);
      }
      flashHit(!hit);
      setChip();
    }

    function knobOffset(event: PointerEvent): Point {
      const rect = stick.getBoundingClientRect();
      return {
        x: event.clientX - (rect.left + rect.width / 2),
        y: event.clientY - (rect.top + rect.height / 2),
      };
    }

    function onStickStart(event: PointerEvent): void {
      if (destroyed || !started || paused) {
        return;
      }
      stick.setPointerCapture(event.pointerId);
      onStickMove(event);
    }

    function onStickMove(event: PointerEvent): void {
      if (destroyed || !started || paused) {
        return;
      }
      const offset = knobOffset(event);
      vector = joystickVector(offset, BASE_RADIUS_PX);
      const len = Math.hypot(offset.x, offset.y);
      const scale = len > BASE_RADIUS_PX ? BASE_RADIUS_PX / len : 1;
      knob.style.transform = `translate(calc(-50% + ${round1(offset.x * scale)}px), calc(-50% + ${round1(offset.y * scale)}px))`;
    }

    function onStickEnd(event: PointerEvent): void {
      vector = { x: 0, y: 0 };
      knob.style.transform = "translate(-50%, -50%)";
      if (stick.hasPointerCapture(event.pointerId)) {
        stick.releasePointerCapture(event.pointerId);
      }
    }

    const loop = createFrameLoop((dtSec: number): void => {
      if (destroyed || paused) {
        return;
      }
      const area = spawnAreaOf(stage);
      aim = stepCrosshair(aim, vector, params.crosshairSpeedPxS, dtSec, area);
      place(crosshair, aim);
      const elapsed = ctx.clock.elapsedMs();
      // The movement grid anchors at the first frame: one polyline sample
      // per interval of run time since the stream began (the clock
      // freezes through pauses, so the grid does too). Anchoring at the
      // stream's start — not the run's — keeps the count×interval law
      // the server's density gate checks.
      if (gridStartMs === null) {
        gridStartMs = elapsed;
      }
      if (elapsed - gridStartMs >= sampleCount * params.movementSampleIntervalMs) {
        if (dot !== null) {
          ctx.collect({
            elapsedMs: Math.round(elapsed),
            target: { x: round1(dot.center.x), y: round1(dot.center.y) },
            crosshair: { x: round1(aim.x), y: round1(aim.y) },
            deviation: round1(distance(dot.center, aim)),
            player: { x: round1(aim.x), y: round1(aim.y) },
          });
        } else {
          ctx.collect({
            elapsedMs: Math.round(elapsed),
            crosshair: { x: round1(aim.x), y: round1(aim.y) },
            player: { x: round1(aim.x), y: round1(aim.y) },
          });
        }
        sampleCount += 1;
      }
    });

    function onResize(): void {
      if (destroyed || dot === null) {
        return;
      }
      // A shrunk stage strands the dot where no crosshair can legally
      // reach it — move the same dot (same spawn stamp, so its lifetime
      // keeps running truthfully) to a fresh in-bounds spot.
      if (outsidePaddedArea(dot.center, spawnAreaOf(stage), params.spawnPaddingPx)) {
        const pos = randomCenter(spawnAreaOf(stage), params.spawnPaddingPx, Math.random);
        dot.center = pos;
        place(dotEl, pos);
      }
    }

    return {
      start(): void {
        if (destroyed || started) {
          return;
        }
        started = true;
        const area = spawnAreaOf(stage);
        aim = { x: area.width / 2, y: area.height / 2 };
        place(crosshair, aim);
        ctx.mount.appendChild(stage);
        stick.addEventListener("pointerdown", onStickStart);
        stick.addEventListener("pointermove", onStickMove);
        stick.addEventListener("pointerup", onStickEnd);
        stick.addEventListener("pointercancel", onStickEnd);
        fireBtn.addEventListener("pointerdown", onFire);
        observer = new ResizeObserver(onResize);
        observer.observe(stage);
        spawnDot();
        loop.start();
      },
      pause(): void {
        if (paused || destroyed) {
          return;
        }
        paused = true;
        loop.stop();
        if (respawnTimer !== null) {
          respawnParkedMs = Math.max(0, respawnDeadline - ctx.now());
          clearTimeout(respawnTimer);
          respawnTimer = null;
        }
      },
      resume(): void {
        if (!paused || destroyed) {
          return;
        }
        paused = false;
        loop.start();
        if (dot === null && respawnParkedMs !== null) {
          const parked = respawnParkedMs;
          respawnParkedMs = null;
          armRespawn(parked);
        }
      },
      destroy(): void {
        if (destroyed) {
          return;
        }
        destroyed = true;
        loop.stop();
        if (respawnTimer !== null) {
          clearTimeout(respawnTimer);
          respawnTimer = null;
        }
        if (flashTimer !== null) {
          clearTimeout(flashTimer);
          flashTimer = null;
        }
        observer?.disconnect();
        observer = null;
        stick.removeEventListener("pointerdown", onStickStart);
        stick.removeEventListener("pointermove", onStickMove);
        stick.removeEventListener("pointerup", onStickEnd);
        stick.removeEventListener("pointercancel", onStickEnd);
        fireBtn.removeEventListener("pointerdown", onFire);
        removeDot();
        stage.remove();
      },
    };
  },
};
