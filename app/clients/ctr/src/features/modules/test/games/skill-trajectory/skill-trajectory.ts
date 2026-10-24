// skill_trajectory — three shapes appear, hide, and the player drags the
// palette copies back where memory says they stood. Each confirmed round
// reports one representative fact: the closest same-shape same-color
// pair within the match distance, carrying the pair's exact deviation
// (original, placement, and deviation agree geometrically — the triple a
// board-reproducing server can verify from random_seed), or the 999
// sentinel with a null placement when nothing matched. The board's
// shape, color, and position draws all flow from the seeded PRNG, so one
// seed redraws one board sequence. A shrunk stage re-clamps live targets
// and placed pieces in-bounds — display and fact stay one truth.

import { m } from "#/paraglide/messages";
import type { GameInstance, GameModule, GameModuleCtx } from "../../contract";
import { createFrameLoop } from "../../frame-loop";
import {
  bestPair,
  clampInside,
  colorAt,
  decimatePath,
  generateBoard,
  MISS_DEV_SENTINEL,
  mulberry32,
  type Point,
  round1,
  SKILL_TRAJECTORY_DEFAULTS,
  type StColor,
  type StShape,
  shapeAt,
  shapeVertices,
} from "./logic";

// Semantic color token → SVG fill class; the fact carries the token, the
// view carries the color.
const COLOR_FILL: Record<StColor, string> = {
  red: "fill-red-500/90",
  orange: "fill-orange-500/90",
  green: "fill-green-500/90",
  blue: "fill-blue-500/90",
  purple: "fill-purple-500/90",
};

const FEEDBACK_MS = 1500;

type Phase = "memorize" | "recall" | "feedback";

interface LiveEntity {
  readonly el: HTMLElement;
  readonly shape: StShape;
  readonly color: StColor;
  center: Point;
}

interface Piece extends LiveEntity {
  path: Point[];
  placed: boolean;
}

interface DragCtl {
  id: number | null;
}

function place(el: HTMLElement, p: Point): void {
  el.style.left = `${round1(p.x)}px`;
  el.style.top = `${round1(p.y)}px`;
}

// One shape as a positioned SVG (circle renders as <circle>, the rest as
// polygons from the shared vertex law).
function makeShapeEl(shape: StShape, color: StColor, sizePx: number): HTMLElement {
  const root = document.createElement("div");
  root.className = "absolute select-none";
  root.style.width = `${sizePx}px`;
  root.style.height = `${sizePx}px`;
  root.style.transform = "translate(-50%, -50%)";
  const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  svg.setAttribute("class", "h-full w-full");
  svg.setAttribute("viewBox", `0 0 ${sizePx} ${sizePx}`);
  if (shape === "circle") {
    const circle = document.createElementNS("http://www.w3.org/2000/svg", "circle");
    circle.setAttribute("cx", String(sizePx / 2));
    circle.setAttribute("cy", String(sizePx / 2));
    circle.setAttribute("r", String(sizePx / 2));
    circle.setAttribute("class", `${COLOR_FILL[color]} stroke-white/50`);
    circle.setAttribute("stroke-width", "1.5");
    svg.appendChild(circle);
  } else {
    const polygon = document.createElementNS("http://www.w3.org/2000/svg", "polygon");
    polygon.setAttribute(
      "points",
      shapeVertices(shape, { x: sizePx / 2, y: sizePx / 2 }, sizePx / 2)
        .map((p) => `${round1(p.x)},${round1(p.y)}`)
        .join(" "),
    );
    polygon.setAttribute("class", `${COLOR_FILL[color]} stroke-white/50`);
    polygon.setAttribute("stroke-width", "1.5");
    svg.appendChild(polygon);
  }
  root.appendChild(svg);
  return root;
}

function makeEntity(
  parent: HTMLElement,
  shape: StShape,
  color: StColor,
  center: Point,
  sizePx: number,
): LiveEntity {
  const el = makeShapeEl(shape, color, sizePx);
  place(el, center);
  parent.appendChild(el);
  return { el, shape, color, center };
}

// A palette piece's pointer law: grab anywhere on it, follow the finger
// clamped to the play band, and the drop leaves it standing. The path
// restarts per grab — the reported trajectory is the drag that produced
// the final position.
function wirePiece(
  piece: Piece,
  layer: HTMLElement,
  padding: number,
  drag: DragCtl,
  canDrag: () => boolean,
  onDrop: () => void,
): void {
  const el = piece.el;
  el.classList.add("cursor-grab", "active:cursor-grabbing");
  const move = (event: PointerEvent): void => {
    const rect = layer.getBoundingClientRect();
    const clamped = clampInside(
      { x: event.clientX - rect.left, y: event.clientY - rect.top },
      { width: rect.width, height: rect.height },
      padding,
    );
    piece.center = clamped;
    place(el, clamped);
    piece.path.push(clamped);
  };
  el.addEventListener("pointerdown", (event) => {
    if (!canDrag()) {
      return;
    }
    // The drag's identity is claimed before capture: capture keeps a
    // real finger's moves flowing to the piece when the pointer outruns
    // it, but synthetic pointers (automation harnesses) carry no
    // active-pointer entry, so the call throws and must stay
    // best-effort — the drag law cannot die with it.
    drag.id = event.pointerId;
    try {
      el.setPointerCapture(event.pointerId);
    } catch {
      // Moving without capture works while the pointer stays on the
      // piece, which the piece following each move keeps true.
    }
    piece.path = [];
    move(event);
  });
  el.addEventListener("pointermove", (event) => {
    if (drag.id === event.pointerId) {
      move(event);
    }
  });
  const drop = (event: PointerEvent): void => {
    if (drag.id !== event.pointerId) {
      return;
    }
    drag.id = null;
    piece.placed = true;
    onDrop();
  };
  el.addEventListener("pointerup", drop);
  el.addEventListener("pointercancel", drop);
}

// The round's one fact — the representative pair's exact geometry, or
// the sentinel shape when nothing matched (an unmatched round still
// names a real target from the board as its representative).
function buildRoundFact(
  round: number,
  match: { target: LiveEntity; placement: Piece; deviation: number } | null,
  fallback: LiveEntity | undefined,
  targetSizePx: number,
  placementTimeMs: number,
): Record<string, unknown> {
  const repr = match?.target ?? fallback;
  return {
    roundNumber: round,
    graphicId: `round-${round}`,
    shape: repr?.shape ?? shapeAt(0),
    graphicSize: targetSizePx,
    colorName: repr?.color ?? colorAt(0),
    originalPosition: { x: round1(repr?.center.x ?? 0), y: round1(repr?.center.y ?? 0) },
    trajectory:
      match === null
        ? []
        : decimatePath(match.placement.path, 512).map((p) => ({
            x: round1(p.x),
            y: round1(p.y),
          })),
    ...(match !== null
      ? {
          playerPlacement: {
            x: round1(match.placement.center.x),
            y: round1(match.placement.center.y),
          },
        }
      : {}),
    deviation: match ? round1(match.deviation) : MISS_DEV_SENTINEL,
    placementTimeMs,
    selected: match !== null,
    placed: match !== null,
  };
}

interface FaceParts {
  readonly stage: HTMLElement;
  readonly playLayer: HTMLElement;
  readonly chipRound: HTMLElement;
  readonly chipHint: HTMLElement;
  readonly palette: HTMLElement;
  readonly slots: readonly HTMLElement[];
  readonly confirm: HTMLElement;
}

// The game's face: read-out chip, play layer, palette strip with slot
// ghosts, and the confirm button — built once; rounds only fill it.
function buildFace(): FaceParts {
  const stage = document.createElement("div");
  stage.className = "relative flex h-full w-full touch-none flex-col overflow-hidden select-none";
  const chip = document.createElement("div");
  chip.className =
    "pointer-events-none absolute top-3 left-1/2 flex -translate-x-1/2 items-center gap-3 rounded-full bg-black/60 px-3.5 py-1.5 backdrop-blur-sm";
  const chipRound = document.createElement("span");
  chipRound.className = "text-[10px] font-semibold text-white/90";
  const chipHint = document.createElement("span");
  chipHint.className = "text-[10px] font-semibold text-white/90";
  chip.append(chipRound, chipHint);
  const playLayer = document.createElement("div");
  playLayer.className = "relative flex-1";
  const palette = document.createElement("div");
  palette.className = "flex h-20 items-center gap-4 border-t border-white/15 px-5";
  const slots: HTMLElement[] = [];
  for (let i = 0; i < 3; i += 1) {
    const slot = document.createElement("div");
    slot.className = "h-12 w-12 shrink-0 rounded-lg border border-dashed border-white/25";
    slots.push(slot);
    palette.appendChild(slot);
  }
  const confirm = document.createElement("div");
  confirm.className =
    "ml-auto flex cursor-pointer items-center rounded-control bg-cta-primary px-5 py-2.5 text-sm font-semibold text-white opacity-40 active:opacity-80";
  confirm.textContent = m.test_st_confirm();
  palette.appendChild(confirm);
  stage.append(chip, playLayer, palette);
  return { stage, playLayer, chipRound, chipHint, palette, slots, confirm };
}

export const skillTrajectoryModule: GameModule = {
  code: "skill_trajectory",
  create(ctx: GameModuleCtx): GameInstance {
    const params = {
      totalRounds: ctx.params.num("total_rounds", SKILL_TRAJECTORY_DEFAULTS.totalRounds),
      displayTimeMs: ctx.params.num("display_time_ms", SKILL_TRAJECTORY_DEFAULTS.displayTimeMs),
      targetSizePx: ctx.params.num("target_size_px", SKILL_TRAJECTORY_DEFAULTS.targetSizePx),
      paletteSizePx: ctx.params.num("palette_size_px", SKILL_TRAJECTORY_DEFAULTS.paletteSizePx),
      matchDistancePx: ctx.params.num(
        "match_distance_px",
        SKILL_TRAJECTORY_DEFAULTS.matchDistancePx,
      ),
    };
    // The board's single input; the server reproduces boards from it.
    const seed = Math.floor(Math.random() * 1_000_000);
    const rng = mulberry32(seed);
    ctx.stat("random_seed", seed);

    const { stage, playLayer, chipRound, chipHint, palette, slots, confirm } = buildFace();
    const drag: DragCtl = { id: null };

    let started = false;
    let paused = false;
    let destroyed = false;
    let finished = false;
    let phase: Phase = "memorize";
    let round = 0;
    let phaseStartMs = 0;
    let recallStartMs = 0;
    let lastMatchDeviation: number | null = null;
    let targets: LiveEntity[] = [];
    let pieces: Piece[] = [];
    let observer: ResizeObserver | null = null;

    function layerSize(): { width: number; height: number } {
      const rect = playLayer.getBoundingClientRect();
      return { width: rect.width, height: rect.height };
    }

    function slotCenter(index: number): Point {
      const layer = playLayer.getBoundingClientRect();
      const slot = slots[index]?.getBoundingClientRect();
      if (slot === undefined) {
        return { x: layer.width / 2, y: layer.height - 40 };
      }
      return {
        x: slot.left + slot.width / 2 - layer.left,
        y: slot.top + slot.height / 2 - layer.top,
      };
    }

    function setChip(): void {
      chipRound.textContent = m.test_st_round({ n: round, total: params.totalRounds });
      if (phase === "memorize") {
        chipHint.textContent = m.test_st_memorize();
      } else if (phase === "recall") {
        chipHint.textContent = m.test_st_recall();
      } else {
        chipHint.textContent =
          lastMatchDeviation === null
            ? m.test_st_missed()
            : m.test_st_matched({ d: lastMatchDeviation.toFixed(1) });
      }
    }

    function setVisible(list: readonly LiveEntity[], visible: boolean): void {
      for (const item of list) {
        item.el.style.display = visible ? "" : "none";
      }
    }

    function updateConfirm(): void {
      confirm.classList.toggle("opacity-40", !pieces.every((piece) => piece.placed));
      confirm.dataset.ready = pieces.every((piece) => piece.placed) ? "1" : "0";
    }

    function startRound(): void {
      round += 1;
      for (const item of [...targets, ...pieces]) {
        item.el.remove();
      }
      drag.id = null;
      lastMatchDeviation = null;
      const board = generateBoard(rng, layerSize(), params);
      targets = board.map((t) =>
        makeEntity(playLayer, t.shape, t.color, t.center, params.targetSizePx),
      );
      pieces = board.map((t, i) => {
        const piece: Piece = {
          ...makeEntity(playLayer, t.shape, t.color, slotCenter(i), params.paletteSizePx),
          path: [],
          placed: false,
        };
        wirePiece(
          piece,
          playLayer,
          params.paletteSizePx / 2,
          drag,
          () => !(destroyed || finished || paused) && phase === "recall",
          updateConfirm,
        );
        return piece;
      });
      phase = "memorize";
      phaseStartMs = ctx.clock.elapsedMs();
      palette.style.display = "none";
      setVisible(pieces, false);
      setVisible(targets, true);
      setChip();
    }

    function confirmRound(): void {
      const elapsed = ctx.clock.elapsedMs();
      const match = bestPair(targets, pieces, params.matchDistancePx);
      ctx.collect(
        buildRoundFact(
          round,
          match,
          targets[0],
          params.targetSizePx,
          Math.max(0, Math.round(elapsed - recallStartMs)),
        ),
      );
      lastMatchDeviation = match === null ? null : round1(match.deviation);
      phase = "feedback";
      phaseStartMs = elapsed;
      palette.style.display = "none";
      // Feedback reveals the truth: the round's targets and the player's
      // placements side by side.
      setVisible(targets, true);
      setChip();
    }

    const loop = createFrameLoop((): void => {
      if (destroyed || finished || paused) {
        return;
      }
      const elapsed = ctx.clock.elapsedMs();
      if (phase === "memorize" && elapsed - phaseStartMs >= params.displayTimeMs) {
        phase = "recall";
        recallStartMs = elapsed;
        setVisible(targets, false);
        setVisible(pieces, true);
        palette.style.display = "";
        updateConfirm();
        setChip();
      } else if (phase === "feedback" && elapsed - phaseStartMs >= FEEDBACK_MS) {
        if (round >= params.totalRounds) {
          finished = true;
          loop.stop();
          ctx.end("completed");
        } else {
          startRound();
        }
      }
    });

    function onResize(): void {
      if (destroyed) {
        return;
      }
      const area = layerSize();
      for (const target of targets) {
        target.center = clampInside(target.center, area, params.targetSizePx);
        place(target.el, target.center);
      }
      for (const [i, piece] of pieces.entries()) {
        // A never-grabbed piece stays glued to its slot; a placed one
        // re-clamps into the band.
        piece.center = piece.placed
          ? clampInside(piece.center, area, params.paletteSizePx / 2)
          : slotCenter(i);
        place(piece.el, piece.center);
      }
    }

    return {
      start(): void {
        if (destroyed || started) {
          return;
        }
        started = true;
        ctx.mount.appendChild(stage);
        confirm.addEventListener("pointerdown", () => {
          if (
            !(destroyed || finished || paused) &&
            phase === "recall" &&
            confirm.dataset.ready === "1"
          ) {
            confirmRound();
          }
        });
        observer = new ResizeObserver(onResize);
        observer.observe(playLayer);
        startRound();
        loop.start();
      },
      pause(): void {
        if (paused || destroyed || finished) {
          return;
        }
        paused = true;
        loop.stop();
      },
      resume(): void {
        if (!paused || destroyed || finished) {
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
        for (const item of [...targets, ...pieces]) {
          item.el.remove();
        }
        stage.remove();
      },
    };
  },
};
