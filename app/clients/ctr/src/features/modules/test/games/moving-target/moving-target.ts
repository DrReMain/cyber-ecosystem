// moving_target — the player's pointer is the crosshair: hold it on the
// roaming target while a cadence-sampled stream records the deviation.
// Sampling starts at the first pointer contact (a crosshair that never
// engaged must not mint tracking time), each sample stamps the live
// pattern and the crosshair-inside-the-disc fact, and the run itself is
// owned by the clock. Resizes need no stranded-entity guard here: the
// target's position is recomputed from the live stage bounds every
// frame, so a shrink re-clamps it on the very next tick.

import { m } from "#/paraglide/messages";
import type { GameInstance, GameModule, GameModuleCtx } from "../../contract";
import { createFrameLoop } from "../../frame-loop";
import {
  clampInside,
  distance,
  holdingBand,
  MOVING_TARGET_DEFAULTS,
  PATTERN_TICK_MS,
  type Point,
  patternAt,
  patternIndexAt,
  patternOffset,
  patternPhase,
  randomWalkStep,
  round1,
  type SpawnArea,
  type TrackPattern,
} from "./logic";

const PATTERN_LABEL: Record<string, () => string> = {
  straight: m.test_mt_pattern_straight,
  curve: m.test_mt_pattern_curve,
  zigzag: m.test_mt_pattern_zigzag,
  circle: m.test_mt_pattern_circle,
  random: m.test_mt_pattern_random,
  variable: m.test_mt_pattern_variable,
};

function spawnAreaOf(el: HTMLElement): SpawnArea {
  const rect = el.getBoundingClientRect();
  return { width: rect.width, height: rect.height };
}

export const movingTargetModule: GameModule = {
  code: "moving_target",
  create(ctx: GameModuleCtx): GameInstance {
    const params = {
      patternDurationMs: ctx.params.num(
        "pattern_duration_ms",
        MOVING_TARGET_DEFAULTS.patternDurationMs,
      ),
      sampleIntervalMs: ctx.params.num(
        "sample_interval_ms",
        MOVING_TARGET_DEFAULTS.sampleIntervalMs,
      ),
      targetSizePx: ctx.params.num("target_size_px", MOVING_TARGET_DEFAULTS.targetSizePx),
      marginPx: ctx.params.num("margin_px", MOVING_TARGET_DEFAULTS.marginPx),
      perfectDeviationPx: ctx.params.num(
        "perfect_deviation_px",
        MOVING_TARGET_DEFAULTS.perfectDeviationPx,
      ),
      goodDeviationPx: ctx.params.num("good_deviation_px", MOVING_TARGET_DEFAULTS.goodDeviationPx),
    };

    const stage = document.createElement("div");
    stage.className = "h-full w-full touch-none overflow-hidden relative select-none";
    // The read-out strip: the live pattern name and the current deviation,
    // colored by the same bands the server judges with.
    const chip = document.createElement("div");
    chip.className =
      "pointer-events-none absolute top-3 left-1/2 flex -translate-x-1/2 items-center gap-3 rounded-full bg-black/60 px-3.5 py-1.5 backdrop-blur-sm";
    const patternSpan = document.createElement("span");
    patternSpan.className = "text-[10px] font-semibold text-white/90";
    const devSpan = document.createElement("span");
    devSpan.className = "text-[10px] font-semibold text-white/90";
    chip.append(patternSpan, devSpan);
    stage.appendChild(chip);
    // One anchored wrapper carries the disc and both honest rings: the
    // perfect and good radii mirror the thresholds the server classifies
    // by, so the feedback geometry is never a second opinion. The rings
    // paint above the disc (legacy's layering) — under it the perfect
    // ring folds inside the disc and the good ring hides behind its own
    // border, both bands invisible.
    const targetEl = document.createElement("div");
    targetEl.className = "pointer-events-none absolute select-none";
    targetEl.style.transform = "translate(-50%, -50%)";
    const perfectRing = document.createElement("div");
    perfectRing.className =
      "absolute rounded-full border-2 border-status-ok/40 left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2";
    const goodRing = document.createElement("div");
    goodRing.className =
      "absolute rounded-full border-2 border-status-pending/30 left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2";
    const disc = document.createElement("div");
    disc.className =
      "absolute rounded-full border-2 border-white/40 bg-gradient-to-br from-cyan-400 to-cyan-600 shadow-[0_0_24px_rgba(34,211,238,0.6)] left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2";
    const size = `${params.targetSizePx}px`;
    disc.style.width = size;
    disc.style.height = size;
    // The legacy disc's 🎯 center mark, drawn as canvas geometry (the
    // games are framework-free — no emoji, no icon fonts): the icon
    // family's concentric target, 10/6 stroke rings and a 2 dot.
    const centerMark = document.createElementNS("http://www.w3.org/2000/svg", "svg");
    centerMark.setAttribute("viewBox", "0 0 24 24");
    centerMark.setAttribute(
      "class",
      "absolute left-1/2 top-1/2 h-6 w-6 -translate-x-1/2 -translate-y-1/2 text-white/80",
    );
    for (const radius of [10, 6]) {
      const ring = document.createElementNS("http://www.w3.org/2000/svg", "circle");
      ring.setAttribute("cx", "12");
      ring.setAttribute("cy", "12");
      ring.setAttribute("r", String(radius));
      ring.setAttribute("fill", "none");
      ring.setAttribute("stroke", "currentColor");
      ring.setAttribute("stroke-width", "2");
      centerMark.appendChild(ring);
    }
    const centerDot = document.createElementNS("http://www.w3.org/2000/svg", "circle");
    centerDot.setAttribute("cx", "12");
    centerDot.setAttribute("cy", "12");
    centerDot.setAttribute("r", "2");
    centerDot.setAttribute("fill", "currentColor");
    centerMark.appendChild(centerDot);
    disc.appendChild(centerMark);
    const perfectSize = `${params.perfectDeviationPx * 2}px`;
    perfectRing.style.width = perfectSize;
    perfectRing.style.height = perfectSize;
    const goodSize = `${params.goodDeviationPx * 2}px`;
    goodRing.style.width = goodSize;
    goodRing.style.height = goodSize;
    targetEl.append(disc, perfectRing, goodRing);
    stage.appendChild(targetEl);
    // The dashed tether shows the gap the score is about.
    const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
    // SVGElement's className is a read-only SVGAnimatedString — the class
    // goes through the attribute.
    svg.setAttribute("class", "pointer-events-none absolute inset-0 h-full w-full");
    const line = document.createElementNS("http://www.w3.org/2000/svg", "line");
    line.setAttribute("stroke", "rgba(255,255,255,0.5)");
    line.setAttribute("stroke-width", "2");
    line.setAttribute("stroke-dasharray", "5,5");
    // The tether rides at half opacity (legacy's law) over whatever band
    // color the run paints on it.
    line.setAttribute("opacity", "0.5");
    svg.appendChild(line);
    svg.style.display = "none";
    stage.appendChild(svg);
    const crosshair = document.createElement("div");
    // The -translate pair centers the 40px box on the aim point — without
    // it the drawn cross sits at the box's top-left, 20px down-right of
    // the player's actual finger (a perfect-band-sized lie).
    crosshair.className =
      "pointer-events-none absolute top-0 left-0 h-10 w-10 -translate-x-1/2 -translate-y-1/2 text-white/90";
    crosshair.style.display = "none";
    const hBar = document.createElement("div");
    hBar.className = "absolute top-1/2 left-0 h-0.5 w-full -translate-y-1/2 bg-current";
    const vBar = document.createElement("div");
    vBar.className = "absolute left-1/2 top-0 h-full w-0.5 -translate-x-1/2 bg-current";
    const dot = document.createElement("div");
    dot.className =
      "absolute left-1/2 top-1/2 h-1.5 w-1.5 -translate-x-1/2 -translate-y-1/2 rounded-full bg-current";
    crosshair.append(hBar, vBar, dot);
    stage.appendChild(crosshair);

    let started = false;
    let paused = false;
    let destroyed = false;
    // Null until the first pointer contact — the engagement gate that
    // keeps a never-touched run an honest empty stream.
    let aim: Point | null = null;
    let sampleCount = 0;
    // Run elapsed when the sampling grid began; null before engagement.
    let gridStartMs: number | null = null;
    let patternIdx = -1;
    let lastTarget: Point = { x: 0, y: 0 };
    let walkPos: Point = lastTarget;
    let walkTicks = 0;

    function place(el: HTMLElement, p: Point): void {
      el.style.left = `${round1(p.x)}px`;
      el.style.top = `${round1(p.y)}px`;
    }

    function setChip(pattern: string, deviation: number): void {
      patternSpan.textContent = PATTERN_LABEL[pattern]?.() ?? pattern;
      devSpan.textContent = `${m.test_mt_deviation()} ${deviation.toFixed(1)}px`;
      const band = holdingBand(deviation, params);
      devSpan.className = `text-[10px] font-semibold ${
        band === "perfect" ? "text-status-ok" : band === "good" ? "text-white/90" : "text-danger"
      }`;
    }

    const loop = createFrameLoop((): void => step());

    // Where the target sits right now: the rotation index decides the
    // pattern, the pattern's age drives its formula, and the random walk
    // inherits the previous pattern's last spot (legacy law).
    function roamTarget(elapsed: number, area: SpawnArea): { pattern: TrackPattern; point: Point } {
      const idx = patternIndexAt(elapsed, params.patternDurationMs);
      if (idx !== patternIdx) {
        patternIdx = idx;
        walkTicks = 0;
        walkPos = lastTarget;
      }
      const pattern = patternAt(idx);
      const age = elapsed - idx * params.patternDurationMs;
      if (pattern === "random") {
        const ticks = Math.floor(age / PATTERN_TICK_MS);
        while (walkTicks < ticks) {
          walkPos = randomWalkStep(walkPos, area, params.marginPx, () => Math.random() * 2 - 1);
          walkTicks += 1;
        }
        return { pattern, point: walkPos };
      }
      const center = { x: area.width / 2, y: area.height / 2 };
      const offset = patternOffset(pattern, patternPhase(age), area, params.marginPx);
      const point = clampInside(
        { x: center.x + offset.x, y: center.y + offset.y },
        area,
        params.marginPx,
      );
      return { pattern, point };
    }

    function step(): void {
      if (destroyed || paused) {
        return;
      }
      const elapsed = ctx.clock.elapsedMs();
      const area = spawnAreaOf(stage);
      const { pattern, point: target } = roamTarget(elapsed, area);
      lastTarget = target;
      place(targetEl, target);
      if (aim !== null) {
        place(crosshair, aim);
        const deviation = distance(target, aim);
        line.setAttribute("x1", String(round1(target.x)));
        line.setAttribute("y1", String(round1(target.y)));
        line.setAttribute("x2", String(round1(aim.x)));
        line.setAttribute("y2", String(round1(aim.y)));
        // The disc's glow carries the band (legacy's three-state law):
        // cyan while perfect, amber within good, red beyond — the same
        // holdingBand the chip and the rings classify by, never a second
        // opinion. The tether and the crosshair's center dot ride the
        // same band through the chip's ok/pending/danger vocabulary
        // (legacy's getDeviationColor consumers). Before the first touch
        // the static cyan stays: no player state, no verdict color.
        const band = holdingBand(deviation, params);
        disc.style.boxShadow =
          band === "perfect"
            ? "0 0 30px rgba(34, 211, 238, 0.8)"
            : band === "good"
              ? "0 0 20px rgba(251, 191, 36, 0.6)"
              : "0 0 10px rgba(239, 68, 68, 0.4)";
        const bandColor =
          band === "perfect"
            ? "var(--color-status-ok)"
            : band === "good"
              ? "var(--color-status-pending)"
              : "var(--color-danger)";
        line.setAttribute("stroke", bandColor);
        dot.style.backgroundColor = bandColor;
        // The sampling grid anchors at engagement: one sample per
        // interval of run time since the crosshair first touched the
        // stage (the clock freezes through pauses, so the grid does
        // too). Anchoring matters — counting from run start would
        // back-fill the whole pre-engagement span in one frame burst
        // and the server's density gate would refuse the stream.
        if (gridStartMs === null) {
          gridStartMs = elapsed;
        }
        if (elapsed - gridStartMs >= sampleCount * params.sampleIntervalMs) {
          ctx.collect({
            elapsedMs: Math.round(elapsed),
            target: { x: round1(target.x), y: round1(target.y) },
            crosshair: { x: round1(aim.x), y: round1(aim.y) },
            deviation: round1(deviation),
            hit: deviation <= params.goodDeviationPx,
            pattern,
          });
          sampleCount += 1;
          setChip(pattern, deviation);
        }
      }
    }

    function onPointer(event: PointerEvent): void {
      if (destroyed || !started || paused) {
        return;
      }
      if (event.pointerType === "mouse" && event.button !== 0) {
        return;
      }
      const rect = stage.getBoundingClientRect();
      aim = { x: event.clientX - rect.left, y: event.clientY - rect.top };
      crosshair.style.display = "";
      svg.style.display = "";
    }

    return {
      start(): void {
        if (destroyed || started) {
          return;
        }
        started = true;
        ctx.mount.appendChild(stage);
        stage.addEventListener("pointerdown", onPointer);
        stage.addEventListener("pointermove", onPointer);
        loop.start();
      },
      pause(): void {
        if (paused || destroyed) {
          return;
        }
        paused = true;
        loop.stop();
      },
      resume(): void {
        if (!paused || destroyed) {
          return;
        }
        paused = false;
        loop.start();
      },
      destroy(): void {
        if (destroyed) {
          return;
        }
        destroyed = true;
        loop.stop();
        stage.removeEventListener("pointerdown", onPointer);
        stage.removeEventListener("pointermove", onPointer);
        stage.remove();
      },
    };
  },
};
