// dual_zone — two lanes, one bar each: left-zone obstacles fall toward a
// horizontal bar that slides along x, right-zone obstacles fly right
// toward a vertical bar that slides along y (either bar follows any
// touch inside its zone on its single axis). Every resolved obstacle
// mints exactly one behavior fact — a dodge once it fully crosses the
// bar's plane, a collision the frame its box overlaps the bar — and a
// collision ends the run on the spot: the run's natural end, so the
// wall anchor lands on the fatal frame and the net clock and the event
// agree. Stamps ride the clock, so a pause freezes both motion and the
// fact timeline. A shrunk zone re-clamps its bar and the in-flight
// obstacles' cross axis in-bounds.

import { m } from "#/paraglide/messages";
import type { GameInstance, GameModule, GameModuleCtx } from "../../contract";
import { createFrameLoop } from "../../frame-loop";
import {
  aabbOverlap,
  barRect,
  clampBarPos,
  clampObstacle,
  DUAL_ZONE_DEFAULTS,
  nextSpawnDelay,
  outOfZone,
  type Point,
  passedBar,
  round1,
  spawnObstacle,
  stepObstacle,
  type Zone,
  type ZoneSize,
} from "./logic";

// Visual variety only — obstacle looks cycle by id; no fact rides them.
const OBSTACLE_FILLS = [
  "fill-red-500/90",
  "fill-orange-500/90",
  "fill-green-500/90",
  "fill-blue-500/90",
  "fill-purple-500/90",
];

const ZONE_HINT: Record<Zone, () => string> = {
  left: m.test_dz_zone_left,
  right: m.test_dz_zone_right,
};

interface Obstacle {
  readonly el: HTMLElement;
  readonly id: number;
  readonly zone: Zone;
  center: Point;
  resolved: boolean;
}

function place(el: HTMLElement, p: Point): void {
  el.style.left = `${round1(p.x)}px`;
  el.style.top = `${round1(p.y)}px`;
}

function makeObstacleEl(id: number, sizePx: number): HTMLElement {
  const fill = OBSTACLE_FILLS[(id - 1) % OBSTACLE_FILLS.length] ?? "fill-white/80";
  const root = document.createElement("div");
  root.className = "pointer-events-none absolute select-none";
  root.style.width = `${sizePx}px`;
  root.style.height = `${sizePx}px`;
  root.style.transform = "translate(-50%, -50%)";
  const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  svg.setAttribute("class", "h-full w-full");
  svg.setAttribute("viewBox", `0 0 ${sizePx} ${sizePx}`);
  const half = sizePx / 2;
  const shape = document.createElementNS("http://www.w3.org/2000/svg", "polygon");
  shape.setAttribute("class", `${fill} stroke-white/40`);
  shape.setAttribute("stroke-width", "1.5");
  const kind = (id - 1) % 3;
  if (kind === 0) {
    const circle = document.createElementNS("http://www.w3.org/2000/svg", "circle");
    circle.setAttribute("cx", String(half));
    circle.setAttribute("cy", String(half));
    circle.setAttribute("r", String(half));
    circle.setAttribute("class", `${fill} stroke-white/40`);
    circle.setAttribute("stroke-width", "1.5");
    svg.appendChild(circle);
  } else {
    // An axis-aligned square or an upward triangle, in viewBox units.
    const points =
      kind === 1
        ? `${half * 0.3},${half * 0.3} ${half * 1.7},${half * 0.3} ${half * 1.7},${half * 1.7} ${half * 0.3},${half * 1.7}`
        : `${half},${half * 0.2} ${half * 1.8},${half * 1.7} ${half * 0.2},${half * 1.7}`;
    shape.setAttribute("points", points);
    svg.appendChild(shape);
  }
  root.appendChild(svg);
  return root;
}

// Write the bar's geometry onto its element — the logic layer owns the
// rect math, this only translates it to style.
function applyBar(
  bar: HTMLElement,
  zone: Zone,
  size: ZoneSize,
  pos: number,
  thicknessPx: number,
): void {
  const rect = barRect(zone, size, pos, { ...DUAL_ZONE_DEFAULTS, barThicknessPx: thicknessPx });
  bar.style.left = `${round1(rect.left)}px`;
  bar.style.top = `${round1(rect.top)}px`;
  if (zone === "left") {
    bar.style.width = `${round1(rect.right - rect.left)}px`;
    bar.style.height = `${thicknessPx}px`;
  } else {
    bar.style.width = `${thicknessPx}px`;
    bar.style.height = `${round1(rect.bottom - rect.top)}px`;
  }
}

// Both lanes' spawn cadence with pause parking: each zone arms one
// wall-clock deadline; a pause banks the remainder and a resume re-arms.
interface SpawnClock {
  arm(zone: Zone, delayMs: number): void;
  park(): void;
  resume(): void;
  clear(): void;
}

function createSpawnClock(now: () => number, onFire: (zone: Zone) => void): SpawnClock {
  const timers: Record<Zone, ReturnType<typeof setTimeout> | null> = { left: null, right: null };
  const deadlines: Record<Zone, number> = { left: 0, right: 0 };
  const parked: Record<Zone, number | null> = { left: null, right: null };
  return {
    arm(zone: Zone, delayMs: number): void {
      const wait = Math.max(0, delayMs);
      deadlines[zone] = now() + wait;
      timers[zone] = setTimeout(() => {
        timers[zone] = null;
        parked[zone] = null;
        onFire(zone);
      }, wait);
    },
    park(): void {
      for (const zone of ["left", "right"] as const) {
        if (timers[zone] !== null) {
          parked[zone] = Math.max(0, deadlines[zone] - now());
          clearTimeout(timers[zone]);
          timers[zone] = null;
        }
      }
    },
    resume(): void {
      for (const zone of ["left", "right"] as const) {
        const remaining = parked[zone];
        if (remaining !== null) {
          parked[zone] = null;
          this.arm(zone, remaining);
        }
      }
    },
    clear(): void {
      for (const zone of ["left", "right"] as const) {
        if (timers[zone] !== null) {
          clearTimeout(timers[zone]);
          timers[zone] = null;
        }
      }
    },
  };
}

interface FaceParts {
  readonly stage: HTMLElement;
  readonly zones: Readonly<Record<Zone, HTMLElement>>;
  readonly bars: Readonly<Record<Zone, HTMLElement>>;
  readonly chipDodges: HTMLElement;
  readonly chipSurvived: HTMLElement;
}

// The game's face: two lanes split by a divider, each with its hint and
// its blue dodge bar, and a read-out chip over the middle. The bars are
// purely visual — the lane owns the pointer and the bar follows.
function buildFace(): FaceParts {
  const stage = document.createElement("div");
  stage.className = "relative flex h-full w-full touch-none overflow-hidden select-none";
  const chip = document.createElement("div");
  chip.className =
    "pointer-events-none absolute top-3 left-1/2 flex -translate-x-1/2 items-center gap-3 rounded-full bg-black/60 px-3.5 py-1.5 backdrop-blur-sm";
  const chipDodges = document.createElement("span");
  chipDodges.className = "text-[10px] font-semibold text-white/90";
  const chipSurvived = document.createElement("span");
  chipSurvived.className = "text-[10px] font-semibold text-white/90";
  chip.append(chipDodges, chipSurvived);
  const zones: Partial<Record<Zone, HTMLElement>> = {};
  const bars: Partial<Record<Zone, HTMLElement>> = {};
  for (const zone of ["left", "right"] as const) {
    const lane = document.createElement("div");
    lane.className =
      zone === "left"
        ? "relative flex-1 overflow-hidden border-r border-white/15 bg-blue-500/5"
        : "relative flex-1 overflow-hidden bg-green-500/5";
    const hint = document.createElement("span");
    hint.className = "pointer-events-none absolute top-2 left-2 text-[10px] text-white/50";
    hint.textContent = ZONE_HINT[zone]();
    const bar = document.createElement("div");
    bar.className = "absolute rounded-md bg-blue-500 shadow-lg";
    lane.append(hint, bar);
    zones[zone] = lane;
    bars[zone] = bar;
    stage.appendChild(lane);
  }
  stage.appendChild(chip);
  return {
    stage,
    zones: zones as Readonly<Record<Zone, HTMLElement>>,
    bars: bars as Readonly<Record<Zone, HTMLElement>>,
    chipDodges,
    chipSurvived,
  };
}

export const dualZoneModule: GameModule = {
  code: "dual_zone",
  create(ctx: GameModuleCtx): GameInstance {
    const params = {
      shapeSizePx: ctx.params.num("shape_size_px", DUAL_ZONE_DEFAULTS.shapeSizePx),
      barThicknessPx: ctx.params.num("bar_thickness_px", DUAL_ZONE_DEFAULTS.barThicknessPx),
      barRatio: ctx.params.num("bar_ratio", DUAL_ZONE_DEFAULTS.barRatio),
      obstacleSpeedPxS: ctx.params.num("obstacle_speed_px_s", DUAL_ZONE_DEFAULTS.obstacleSpeedPxS),
      spawnIntervalMinMs: ctx.params.num(
        "spawn_interval_min_ms",
        DUAL_ZONE_DEFAULTS.spawnIntervalMinMs,
      ),
      spawnIntervalSpanMs: ctx.params.num(
        "spawn_interval_span_ms",
        DUAL_ZONE_DEFAULTS.spawnIntervalSpanMs,
      ),
    };

    const { stage, zones, bars, chipDodges, chipSurvived } = buildFace();

    let started = false;
    let paused = false;
    let destroyed = false;
    let finished = false;
    let nextId = 1;
    let dodges = 0;
    let obstacles: Obstacle[] = [];
    // One bar scalar per zone: the left bar's center x, the right bar's
    // center y.
    const barPos: Record<Zone, number> = { left: 0, right: 0 };
    let observer: ResizeObserver | null = null;

    function zoneSize(zone: Zone): ZoneSize {
      const rect = zones[zone].getBoundingClientRect();
      return { width: rect.width, height: rect.height };
    }

    function setChip(): void {
      chipDodges.textContent = m.test_dz_dodges({ n: dodges });
      chipSurvived.textContent = m.test_dz_survived({
        s: (ctx.clock.elapsedMs() / 1000).toFixed(1),
      });
    }

    function spawn(zone: Zone): void {
      const center = spawnObstacle(zone, zoneSize(zone), params.shapeSizePx, Math.random);
      const el = makeObstacleEl(nextId, params.shapeSizePx);
      place(el, center);
      zones[zone].appendChild(el);
      obstacles.push({ el, id: nextId, zone, center, resolved: false });
      nextId += 1;
    }

    const clock = createSpawnClock(ctx.now, (zone) => {
      spawn(zone);
      clock.arm(zone, nextSpawnDelay(params, Math.random));
    });

    function fatalCollision(obstacle: Obstacle): void {
      ctx.collect({
        elapsedMs: Math.round(ctx.clock.elapsedMs()),
        zone: obstacle.zone,
        event: "collision",
        position: { x: round1(obstacle.center.x), y: round1(obstacle.center.y) },
        obstacleId: obstacle.id,
      });
      bars[obstacle.zone].classList.add("bg-danger");
      finished = true;
      loop.stop();
      clock.clear();
      ctx.end("completed");
    }

    function resolveObstacle(obstacle: Obstacle, dtSec: number): boolean {
      const size = zoneSize(obstacle.zone);
      obstacle.center = stepObstacle(
        obstacle.zone,
        obstacle.center,
        params.obstacleSpeedPxS,
        dtSec,
      );
      place(obstacle.el, obstacle.center);
      const bar = barRect(obstacle.zone, size, barPos[obstacle.zone], params);
      if (aabbOverlap(obstacle.center, params.shapeSizePx, bar)) {
        fatalCollision(obstacle);
        return true;
      }
      if (
        !obstacle.resolved &&
        passedBar(obstacle.zone, obstacle.center, params.shapeSizePx, bar)
      ) {
        obstacle.resolved = true;
        dodges += 1;
        ctx.collect({
          elapsedMs: Math.round(ctx.clock.elapsedMs()),
          zone: obstacle.zone,
          event: "dodge",
          obstacleId: obstacle.id,
        });
      }
      if (outOfZone(obstacle.zone, obstacle.center, params.shapeSizePx, size)) {
        obstacle.el.remove();
        obstacles = obstacles.filter((o) => o !== obstacle);
      }
      return false;
    }

    const loop = createFrameLoop((dtSec: number): void => {
      if (destroyed || finished || paused) {
        return;
      }
      for (const obstacle of [...obstacles]) {
        if (resolveObstacle(obstacle, dtSec)) {
          return;
        }
      }
      setChip();
    });

    function wireLane(zone: Zone): void {
      const lane = zones[zone];
      const onPoint = (event: PointerEvent): void => {
        if (destroyed || finished || paused) {
          return;
        }
        const rect = lane.getBoundingClientRect();
        const raw = zone === "left" ? event.clientX - rect.left : event.clientY - rect.top;
        barPos[zone] = clampBarPos(zone, { width: rect.width, height: rect.height }, raw, params);
        applyBar(
          bars[zone],
          zone,
          { width: rect.width, height: rect.height },
          barPos[zone],
          params.barThicknessPx,
        );
      };
      lane.addEventListener("pointerdown", (event) => {
        // Capture keeps a real finger's moves flowing to the lane when
        // the pointer outruns it; synthetic pointers (automation
        // harnesses) carry no active-pointer entry, so the call throws
        // and must stay best-effort — the bar still follows every move.
        try {
          lane.setPointerCapture(event.pointerId);
        } catch {
          // Without capture the lane keeps receiving moves while the
          // pointer stays inside it, which is the common case.
        }
        onPoint(event);
      });
      lane.addEventListener("pointermove", onPoint);
    }

    function onResize(): void {
      if (destroyed) {
        return;
      }
      for (const zone of ["left", "right"] as const) {
        const size = zoneSize(zone);
        barPos[zone] = clampBarPos(zone, size, barPos[zone], params);
        applyBar(bars[zone], zone, size, barPos[zone], params.barThicknessPx);
      }
      for (const obstacle of obstacles) {
        obstacle.center = clampObstacle(
          obstacle.zone,
          obstacle.center,
          zoneSize(obstacle.zone),
          params.shapeSizePx,
        );
        place(obstacle.el, obstacle.center);
      }
    }

    return {
      start(): void {
        if (destroyed || started) {
          return;
        }
        started = true;
        ctx.mount.appendChild(stage);
        for (const zone of ["left", "right"] as const) {
          const size = zoneSize(zone);
          barPos[zone] = clampBarPos(
            zone,
            size,
            zone === "left" ? size.width / 2 : size.height / 2,
            params,
          );
          applyBar(bars[zone], zone, size, barPos[zone], params.barThicknessPx);
          wireLane(zone);
        }
        observer = new ResizeObserver(onResize);
        observer.observe(stage);
        setChip();
        // The first obstacle of each lane arrives without a warm-up gap;
        // the cadence governs from the second on.
        spawn("left");
        spawn("right");
        clock.arm("left", nextSpawnDelay(params, Math.random));
        clock.arm("right", nextSpawnDelay(params, Math.random));
        loop.start();
      },
      pause(): void {
        if (paused || destroyed || finished) {
          return;
        }
        paused = true;
        loop.stop();
        clock.park();
      },
      resume(): void {
        if (!paused || destroyed || finished) {
          return;
        }
        paused = false;
        loop.start();
        clock.resume();
      },
      destroy(): void {
        if (destroyed) {
          return;
        }
        destroyed = true;
        loop.stop();
        clock.clear();
        observer?.disconnect();
        observer = null;
        for (const obstacle of obstacles) {
          obstacle.el.remove();
        }
        obstacles = [];
        stage.remove();
      },
    };
  },
};
