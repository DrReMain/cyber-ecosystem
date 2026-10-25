// resource_match — a collecting field: tap or drag to steer the hero,
// pick up minions, monsters, and dragons, and keep clear of the red
// wanderers. Every resolved event mints exactly one fact — a collect
// with its kind and net-clock stamp, a collision on the frame contact
// is made (re-contact requires separation first, so staying overlapped
// counts once), and an audit trajectory point on a fixed sample beat.
// The field spawns in waves: the on-field count stays small, every
// pickup respawns from the remaining per-kind budget, and each spawn
// keeps the min-gap law — so no two pickups can ever land inside the
// 80 ms collect floor, and a run's per-kind totals can never exceed
// the spawn caps the server judges against. Total health is the death
// economy (collision cap × 10): the collision that empties it ends the
// run on the spot, the fatal fact the wall anchor lands on. Stamps ride
// the net clock, so a pause freezes motion and facts together. A shrunk
// field re-clamps the hero, the wanderers, and every standing resource
// in-bounds.

import { m } from "#/paraglide/messages";
import type { GameInstance, GameModule, GameModuleCtx } from "../../contract";
import { createFrameLoop } from "../../frame-loop";
import {
  type Area,
  clampPoint,
  distance,
  type Point,
  RESOURCE_MATCH_DEFAULTS,
  type RmKind,
  round1,
  spawnPoint,
  stepEnemy,
  stepToward,
} from "./logic";

const KIND_STYLE: Record<RmKind, { readonly face: string; readonly legend: () => string }> = {
  minion: { face: "rounded-full bg-blue-500/90", legend: m.test_rm_legend_minion },
  monster: { face: "rounded-md bg-orange-500/90", legend: m.test_rm_legend_monster },
  dragon: { face: "rotate-45 rounded-[4px] bg-purple-500/90", legend: m.test_rm_legend_dragon },
};

const HIT_PENALTY = 10;

interface LiveResource {
  readonly el: HTMLElement;
  readonly kind: RmKind;
  center: Point;
}

interface LiveEnemy {
  readonly el: HTMLElement;
  center: Point;
  dir: number;
}

function place(el: HTMLElement, p: Point): void {
  el.style.left = `${round1(p.x)}px`;
  el.style.top = `${round1(p.y)}px`;
}

// Tap-or-drag steering on the field: any pointer down or drag move
// retargets the hero, clamped to the padded bounds. Capture keeps a
// real finger steering when it outruns the field; synthetic pointers
// carry no active-pointer entry, so the call throws and stays
// best-effort — the tap law holds without it.
function wireSteering(
  field: HTMLElement,
  isLive: () => boolean,
  onTarget: (p: Point) => void,
  halfSizePx: number,
): void {
  let dragId: number | null = null;
  const onPoint = (event: PointerEvent): void => {
    if (!isLive()) {
      return;
    }
    const rect = field.getBoundingClientRect();
    onTarget(
      clampPoint(
        { x: event.clientX - rect.left, y: event.clientY - rect.top },
        { width: rect.width, height: rect.height },
        halfSizePx,
      ),
    );
  };
  field.addEventListener("pointerdown", (event) => {
    dragId = event.pointerId;
    try {
      field.setPointerCapture(event.pointerId);
    } catch {
      // Without capture the field keeps receiving moves while the
      // finger stays inside it, which tap-steering satisfies.
    }
    onPoint(event);
  });
  field.addEventListener("pointermove", (event) => {
    if (dragId === event.pointerId) {
      onPoint(event);
    }
  });
}

// A sized disc centered on its point via translate — the shape's class
// carries the kind's identity, geometry stays uniform.
function makeDisc(sizePx: number, face: string): HTMLElement {
  const el = document.createElement("div");
  el.className = `pointer-events-none absolute select-none ${face}`;
  el.style.width = `${sizePx}px`;
  el.style.height = `${sizePx}px`;
  el.style.transform = "translate(-50%, -50%)";
  return el;
}

interface FaceParts {
  readonly stage: HTMLElement;
  readonly field: HTMLElement;
  readonly chipCollected: HTMLElement;
  readonly healthFill: HTMLElement;
}

interface RmParams {
  readonly counts: Record<RmKind, number>;
  readonly collisionCap: number;
  readonly playerSizePx: number;
  readonly enemySizePx: number;
  readonly resourceSizePx: number;
  readonly collectRadiusPx: number;
  readonly playerSpeedPxS: number;
  readonly enemySpeedPxS: number;
  readonly enemyCount: number;
  readonly onFieldCount: number;
  readonly spawnMinGapPx: number;
  readonly trajectoryIntervalMs: number;
}

// The run's opening: the hero at center, the first resource wave, and
// the wanderers — every spawn keeps its distance from the hero's start,
// so the run opens fair and the first collect cannot beat the span
// floor physics already forbids. Returns the hero's start point.
function openField(
  field: HTMLElement,
  playerEl: HTMLElement,
  size: Area,
  params: RmParams,
  spawnOne: () => void,
  enemies: LiveEnemy[],
): Point {
  const player = { x: size.width / 2, y: size.height / 2 };
  place(playerEl, player);
  field.appendChild(playerEl);
  const wanted = Math.min(
    params.onFieldCount,
    params.counts.minion + params.counts.monster + params.counts.dragon,
  );
  for (let i = 0; i < wanted; i += 1) {
    spawnOne();
  }
  for (let i = 0; i < params.enemyCount; i += 1) {
    const center = spawnPoint(
      size,
      [player],
      params.spawnMinGapPx,
      params.enemySizePx / 2,
      Math.random,
    );
    const el = makeDisc(params.enemySizePx, "rounded-full border border-white/30 bg-red-500");
    place(el, center);
    field.appendChild(el);
    enemies.push({ el, center, dir: Math.random() * Math.PI * 2 });
  }
  return player;
}

function buildFace(): FaceParts {
  const stage = document.createElement("div");
  stage.className = "relative flex h-full w-full touch-none flex-col overflow-hidden select-none";
  const chip = document.createElement("div");
  chip.className =
    "pointer-events-none absolute top-3 left-1/2 flex w-44 -translate-x-1/2 items-center gap-2.5 rounded-full bg-black/60 px-3.5 py-1.5 backdrop-blur-sm";
  const chipCollected = document.createElement("span");
  chipCollected.className = "shrink-0 text-[10px] font-semibold text-white/90";
  const healthTrack = document.createElement("span");
  healthTrack.className = "h-1.5 flex-1 overflow-hidden rounded-full bg-white/15";
  const healthFill = document.createElement("span");
  healthFill.className = "block h-full rounded-full bg-status-ok";
  healthTrack.appendChild(healthFill);
  chip.append(chipCollected, healthTrack);
  const field = document.createElement("div");
  field.className = "relative flex-1 overflow-hidden";
  const legend = document.createElement("div");
  legend.className =
    "flex shrink-0 items-center justify-center gap-4 border-t border-white/10 px-4 py-2.5";
  const entries: [RmKind | "enemy", string, () => string][] = [
    ["minion", "h-2.5 w-2.5 rounded-full bg-blue-500/90", KIND_STYLE.minion.legend],
    ["monster", "h-2.5 w-2.5 rounded-[2px] bg-orange-500/90", KIND_STYLE.monster.legend],
    ["dragon", "h-2.5 w-2.5 rotate-45 rounded-[1px] bg-purple-500/90", KIND_STYLE.dragon.legend],
    ["enemy", "h-2.5 w-2.5 rounded-full bg-red-500/90", m.test_rm_legend_enemy],
  ];
  for (const [, face, label] of entries) {
    const item = document.createElement("span");
    item.className = "flex items-center gap-1.5 text-[10px] text-white/50";
    const dot = document.createElement("span");
    dot.className = face;
    item.append(dot, label());
    legend.appendChild(item);
  }
  stage.append(chip, field, legend);
  return { stage, field, chipCollected, healthFill };
}

export const resourceMatchModule: GameModule = {
  code: "resource_match",
  create(ctx: GameModuleCtx): GameInstance {
    const params: RmParams = {
      counts: {
        minion: ctx.params.num("minion_count", RESOURCE_MATCH_DEFAULTS.minionCount),
        monster: ctx.params.num("monster_count", RESOURCE_MATCH_DEFAULTS.monsterCount),
        dragon: ctx.params.num("dragon_count", RESOURCE_MATCH_DEFAULTS.dragonCount),
      },
      collisionCap: ctx.params.num("collision_cap", RESOURCE_MATCH_DEFAULTS.collisionCap),
      playerSizePx: ctx.params.num("player_size_px", RESOURCE_MATCH_DEFAULTS.playerSizePx),
      enemySizePx: ctx.params.num("enemy_size_px", RESOURCE_MATCH_DEFAULTS.enemySizePx),
      resourceSizePx: ctx.params.num("resource_size_px", RESOURCE_MATCH_DEFAULTS.resourceSizePx),
      collectRadiusPx: ctx.params.num("collect_radius_px", RESOURCE_MATCH_DEFAULTS.collectRadiusPx),
      playerSpeedPxS: ctx.params.num("player_speed_px_s", RESOURCE_MATCH_DEFAULTS.playerSpeedPxS),
      enemySpeedPxS: ctx.params.num("enemy_speed_px_s", RESOURCE_MATCH_DEFAULTS.enemySpeedPxS),
      enemyCount: ctx.params.num("enemy_count", RESOURCE_MATCH_DEFAULTS.enemyCount),
      onFieldCount: ctx.params.num("on_field_count", RESOURCE_MATCH_DEFAULTS.onFieldCount),
      spawnMinGapPx: ctx.params.num("spawn_min_gap_px", RESOURCE_MATCH_DEFAULTS.spawnMinGapPx),
      trajectoryIntervalMs: ctx.params.num(
        "trajectory_interval_ms",
        RESOURCE_MATCH_DEFAULTS.trajectoryIntervalMs,
      ),
    };

    const { stage, field, chipCollected, healthFill } = buildFace();

    let started = false;
    let paused = false;
    let destroyed = false;
    let finished = false;
    // The opening wave waits for a real layout: at start() the field may
    // not be laid out yet (a degenerate rect collapses every spawn
    // candidate onto one point, and a same-frame multi-collect mints
    // equal stamps the scorer's strictly-increasing gate must reject) —
    // so the opening is idempotent and defers to the observer's first
    // real-size callback.
    let opened = false;
    // The spawn economy: budgets count down at mint, so the field's
    // totals can never exceed the caps the server judges against.
    const budget: Record<RmKind, number> = { ...params.counts };
    let collected = 0;
    let health = params.collisionCap * HIT_PENALTY;
    let player: Point = { x: 0, y: 0 };
    let target: Point | null = null;
    let inContact = false;
    let lastTrajectoryMs = -Infinity;
    let resources: LiveResource[] = [];
    let enemies: LiveEnemy[] = [];
    const playerEl = makeDisc(
      params.playerSizePx,
      "rounded-full border-2 border-white/70 bg-cyan-400 shadow-lg",
    );
    let observer: ResizeObserver | null = null;

    function area(): Area {
      const rect = field.getBoundingClientRect();
      return { width: rect.width, height: rect.height };
    }

    function occupiedPoints(): Point[] {
      return [...resources.map((resource) => resource.center), player];
    }

    function setChip(): void {
      chipCollected.textContent = m.test_rm_collected({ n: collected });
      healthFill.style.width = `${Math.max(0, Math.min(100, (health / (params.collisionCap * HIT_PENALTY)) * 100))}%`;
      healthFill.className = `block h-full rounded-full ${health <= HIT_PENALTY * 2 ? "bg-danger" : "bg-status-ok"}`;
    }

    function spawnResource(): void {
      const kinds = (Object.keys(budget) as RmKind[]).filter((entry) => budget[entry] > 0);
      if (kinds.length === 0) {
        return;
      }
      const kind = kinds[Math.floor(Math.random() * kinds.length)] as RmKind;
      budget[kind] -= 1;
      const center = spawnPoint(
        area(),
        occupiedPoints(),
        params.spawnMinGapPx,
        params.resourceSizePx / 2,
        Math.random,
      );
      const el = makeDisc(params.resourceSizePx, KIND_STYLE[kind].face);
      place(el, center);
      field.appendChild(el);
      resources.push({ el, kind, center });
    }

    function collectResource(resource: LiveResource): void {
      ctx.collect({
        type: "collect",
        kind: resource.kind,
        elapsedMs: Math.round(ctx.clock.elapsedMs()),
      });
      resource.el.remove();
      resources = resources.filter((item) => item !== resource);
      collected += 1;
      setChip();
      spawnResource();
    }

    function fatalBleed(): void {
      ctx.collect({
        type: "collision",
        elapsedMs: Math.round(ctx.clock.elapsedMs()),
      });
      health -= HIT_PENALTY;
      setChip();
      playerEl.classList.add("ring-4", "ring-danger");
      setTimeout(() => playerEl.classList.remove("ring-4", "ring-danger"), 300);
      if (health <= 0) {
        // The death economy's last allowed collision — the run's natural
        // end, with the fatal fact the stream's final stamp.
        finished = true;
        loop.stop();
        ctx.end("completed");
      }
    }

    function advanceField(dtSec: number, size: Area): void {
      if (target !== null) {
        const moved = stepToward(player, target, params.playerSpeedPxS, dtSec);
        player = moved.point;
        if (moved.arrived) {
          target = null;
        }
        place(playerEl, player);
      }
      for (const enemy of enemies) {
        const stepped = stepEnemy(
          enemy.center,
          enemy.dir,
          params.enemySpeedPxS,
          dtSec,
          size,
          params.enemySizePx / 2,
        );
        enemy.center = stepped.point;
        enemy.dir = stepped.dir;
        place(enemy.el, enemy.center);
      }
    }

    function resolveField(): void {
      const collectBand = params.collectRadiusPx + params.resourceSizePx / 2;
      for (const resource of [...resources]) {
        if (distance(player, resource.center) < collectBand) {
          collectResource(resource);
        }
      }
      const contactBand = (params.playerSizePx + params.enemySizePx) / 2;
      const contact = enemies.some((enemy) => distance(player, enemy.center) < contactBand);
      if (contact && !inContact) {
        fatalBleed();
      }
      inContact = contact;
      const nowMs = ctx.clock.elapsedMs();
      if (nowMs - lastTrajectoryMs >= params.trajectoryIntervalMs) {
        lastTrajectoryMs = nowMs;
        ctx.collect({
          type: "trajectory",
          x: round1(player.x),
          y: round1(player.y),
          elapsedMs: Math.round(nowMs),
        });
      }
    }

    const loop = createFrameLoop((dtSec: number): void => {
      if (destroyed || finished || paused) {
        return;
      }
      advanceField(dtSec, area());
      resolveField();
    });

    function ensureOpened(): void {
      if (opened || destroyed || !started) {
        return;
      }
      const size = area();
      if (size.width <= 0 || size.height <= 0) {
        return;
      }
      opened = true;
      player = openField(field, playerEl, size, params, spawnResource, enemies);
    }

    function onResize(): void {
      if (destroyed) {
        return;
      }
      ensureOpened();
      if (!opened) {
        return;
      }
      const size = area();
      player = clampPoint(player, size, params.playerSizePx / 2);
      place(playerEl, player);
      for (const resource of resources) {
        resource.center = clampPoint(resource.center, size, params.resourceSizePx / 2);
        place(resource.el, resource.center);
      }
      for (const enemy of enemies) {
        enemy.center = clampPoint(enemy.center, size, params.enemySizePx / 2);
        place(enemy.el, enemy.center);
      }
    }

    return {
      start(): void {
        if (destroyed || started) {
          return;
        }
        started = true;
        ctx.mount.appendChild(stage);
        ensureOpened();
        wireSteering(
          field,
          () => !(destroyed || finished || paused),
          (point) => {
            target = point;
          },
          params.playerSizePx / 2,
        );
        observer = new ResizeObserver(onResize);
        observer.observe(field);
        setChip();
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
        for (const resource of resources) {
          resource.el.remove();
        }
        resources = [];
        for (const enemy of enemies) {
          enemy.el.remove();
        }
        enemies = [];
        stage.remove();
      },
    };
  },
};
