// number_stars — an endless string of rounds on one clock: five numbers
// spread on the field, tap them in ascending order. A correct tap removes
// its number; a wrong tap voids the round on the spot (the number set is
// redrawn, the round counter advances — the legacy restart, minus its
// reused round numbers the server's strictly-increasing gate must refuse).
// Every fact is minted the moment it happens — the round's number set at
// open, each tap with its round-relative stamp and the number's position —
// so a clock-out mid-round still carries a complete, judgeable history.
// The opening waits for a real layout (a degenerate rect would collapse
// the board), and the tap gate keeps consecutive stamps a frame apart so
// no same-frame double-tap can mint an equal stamp the gate rejects.

import { m } from "#/paraglide/messages";
import type { GameInstance, GameModule, GameModuleCtx } from "../../contract";
import { createFrameLoop } from "../../frame-loop";
import {
  ascending,
  drawNumbers,
  type NsArea,
  type NsPoint,
  NUMBER_STARS_DEFAULTS,
  spawnNumberPoints,
} from "./logic";

interface FaceParts {
  readonly stage: HTMLElement;
  readonly field: HTMLElement;
  readonly chipRound: HTMLElement;
  readonly chipTime: HTMLElement;
  readonly chipCompleted: HTMLElement;
}

function buildFace(): FaceParts {
  const stage = document.createElement("div");
  stage.className = "relative flex h-full w-full flex-col overflow-hidden select-none";
  const chip = document.createElement("div");
  chip.className =
    "pointer-events-none absolute top-3 left-1/2 flex -translate-x-1/2 items-center gap-3 rounded-full bg-black/60 px-3.5 py-1.5 backdrop-blur-sm";
  const chipRound = document.createElement("span");
  chipRound.className = "text-[10px] font-semibold text-white/90";
  const chipTime = document.createElement("span");
  chipTime.className = "text-[10px] font-semibold text-white/90";
  const chipCompleted = document.createElement("span");
  chipCompleted.className = "text-[10px] font-semibold text-status-ok";
  chip.append(chipRound, chipTime, chipCompleted);
  const field = document.createElement("div");
  field.className = "relative flex-1 overflow-hidden";
  stage.append(chip, field);
  return { stage, field, chipRound, chipTime, chipCompleted };
}

export const numberStarsModule: GameModule = {
  code: "number_stars",
  create(ctx: GameModuleCtx): GameInstance {
    const params = {
      numbersPerRound: ctx.params.num("numbers_per_round", NUMBER_STARS_DEFAULTS.numbersPerRound),
      minNumber: ctx.params.num("min_number", NUMBER_STARS_DEFAULTS.minNumber),
      maxNumber: ctx.params.num("max_number", NUMBER_STARS_DEFAULTS.maxNumber),
      spawnPaddingPx: ctx.params.num("spawn_padding_px", NUMBER_STARS_DEFAULTS.spawnPaddingPx),
      spawnMinGapPx: ctx.params.num("spawn_min_gap_px", NUMBER_STARS_DEFAULTS.spawnMinGapPx),
    };

    const { stage, field, chipRound, chipTime, chipCompleted } = buildFace();

    let started = false;
    let paused = false;
    let destroyed = false;
    // The opening round waits for a real layout — a degenerate rect
    // collapses every spawn candidate onto one point.
    let opened = false;
    let roundSeq = 0;
    let completed = 0;
    let sorted: number[] = [];
    let progress = 0;
    let roundStartMs = 0;
    let lastClickMs = -Infinity;
    let els: HTMLElement[] = [];
    let observer: ResizeObserver | null = null;

    function area(): NsArea {
      const rect = field.getBoundingClientRect();
      return { width: rect.width, height: rect.height };
    }

    function fadeOut(el: HTMLElement, wrong: boolean): void {
      if (wrong) {
        el.classList.add("border-danger", "bg-danger", "text-white");
      }
      el.classList.add("opacity-0", "scale-50", "pointer-events-none");
      setTimeout(() => el.remove(), 350);
    }

    function setChip(): void {
      const remaining = Math.max(0, ctx.durationMs - ctx.clock.elapsedMs());
      chipRound.textContent = m.test_ns_round({ n: roundSeq });
      chipTime.textContent = m.test_od_time_left({ s: (remaining / 1000).toFixed(1) });
      chipCompleted.textContent = m.test_ns_completed({ n: completed });
    }

    function openRound(): void {
      roundSeq += 1;
      const numbers = drawNumbers(
        params.minNumber,
        params.maxNumber,
        params.numbersPerRound,
        Math.random,
      );
      const points = spawnNumberPoints(
        area(),
        numbers.length,
        params.spawnPaddingPx,
        params.spawnMinGapPx,
        Math.random,
      );
      for (const el of els) {
        fadeOut(el, false);
      }
      els = [];
      sorted = ascending(numbers);
      progress = 0;
      roundStartMs = ctx.clock.elapsedMs();
      ctx.collect({ type: "round", roundNumber: roundSeq, numbers });
      numbers.forEach((number, i) => {
        const point = points[i];
        if (point === undefined) {
          return;
        }
        const el = document.createElement("div");
        el.className =
          "absolute flex h-12 w-12 -translate-x-1/2 -translate-y-1/2 items-center justify-center rounded-full border border-white/25 bg-white/10 text-lg font-bold text-white backdrop-blur-sm transition-all duration-200 active:scale-90";
        el.style.left = `${point.x}px`;
        el.style.top = `${point.y}px`;
        el.dataset.number = String(number);
        el.textContent = String(number);
        el.addEventListener("pointerdown", () => tap(number, point));
        field.appendChild(el);
        els.push(el);
      });
      setChip();
    }

    function tap(number: number, point: NsPoint): void {
      if (destroyed || paused || !opened) {
        return;
      }
      // Only standing numbers count — a fading element from a voided board
      // still holds its listener, but its round is closed.
      if (!els.some((candidate) => candidate.dataset.number === String(number))) {
        return;
      }
      const now = ctx.clock.elapsedMs();
      // Consecutive stamps must stay strictly increasing on the net clock —
      // a same-frame double-tap would mint an equal stamp the gate rejects.
      if (now - lastClickMs < 100) {
        return;
      }
      lastClickMs = now;
      applyTap(number, point, now);
    }

    function applyTap(number: number, point: NsPoint, now: number): void {
      ctx.collect({
        type: "click",
        number,
        elapsedMs: Math.max(0, Math.round(now - roundStartMs)),
        position: { x: Math.round(point.x), y: Math.round(point.y) },
      });
      const expected = sorted[progress];
      const el = els.find((candidate) => candidate.dataset.number === String(number));
      if (expected !== undefined && number === expected) {
        progress += 1;
        if (el !== undefined) {
          fadeOut(el, false);
          els = els.filter((candidate) => candidate !== el);
        }
        if (progress >= sorted.length) {
          completed += 1;
          openRound();
        }
      } else {
        // A wrong tap voids the round — red flash, redraw, next number.
        if (el !== undefined) {
          fadeOut(el, true);
        }
        openRound();
      }
      setChip();
    }

    function ensureOpened(): void {
      if (opened || destroyed || !started) {
        return;
      }
      const size = area();
      if (size.width <= 0 || size.height <= 0) {
        return;
      }
      opened = true;
      openRound();
    }

    function onResize(): void {
      if (destroyed || !opened) {
        return;
      }
      // A shrunk field re-clamps standing numbers in-bounds; positions on
      // the wire stay the click-time truth.
      const size = area();
      for (const el of els) {
        const left = Number.parseFloat(el.style.left);
        const top = Number.parseFloat(el.style.top);
        el.style.left = `${Math.min(Math.max(left, 30), Math.max(30, size.width - 30))}px`;
        el.style.top = `${Math.min(Math.max(top, 30), Math.max(30, size.height - 30))}px`;
      }
    }

    const loop = createFrameLoop((): void => {
      if (destroyed || paused) {
        return;
      }
      setChip();
    });

    return {
      start(): void {
        if (destroyed || started) {
          return;
        }
        started = true;
        ctx.mount.appendChild(stage);
        ensureOpened();
        observer = new ResizeObserver(() => {
          ensureOpened();
          onResize();
        });
        observer.observe(field);
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
        observer?.disconnect();
        observer = null;
        stage.remove();
      },
    };
  },
};
