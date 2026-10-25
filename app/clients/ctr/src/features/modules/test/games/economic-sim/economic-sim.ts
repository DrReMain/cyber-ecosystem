// economic_sim — one minion on the block at a time, last-hitting under a
// stepped decay. The tap lands in one of four windows read from the
// minion's remaining health; perfect and good confirm the last hit, early
// pays a sliver, and a too-early tap bleeds the player's own health — the
// economy's only death. A perfect tap adds a second to the run (the
// clock's bonus channel), a recalled purchase spends gold on the
// three-rung ladder and restores health to full, and an untended minion
// that bleeds out counts as missed. Every tap and purchase mints its fact
// at the moment it happens with the health and the floored gold the
// server recomputes — the per-tap cross-check passes by construction —
// and the terminal health stat is written at every change so a clock-out
// reads the true final value.

import { m } from "#/paraglide/messages";
import type { GameInstance, GameModule, GameModuleCtx } from "../../contract";
import { createFrameLoop } from "../../frame-loop";
import {
  confirmsHit,
  ECONOMIC_SIM_DEFAULTS,
  type EsMinionType,
  type EsParams,
  goldOf,
  healthAt,
  nextMinionType,
  windowOf,
} from "./logic";

const MINION_LABEL: Record<EsMinionType, () => string> = {
  melee: m.test_es_minion_melee,
  ranged: m.test_es_minion_ranged,
  cannon: m.test_es_minion_cannon,
};

interface FaceParts {
  readonly stage: HTMLElement;
  readonly chipGold: HTMLElement;
  readonly chipCs: HTMLElement;
  readonly chipItems: HTMLElement;
  readonly chipHealth: HTMLElement;
  readonly chipMissed: HTMLElement;
  readonly minionLabel: HTMLElement;
  readonly healthFill: HTMLElement;
  readonly lastHitButton: HTMLElement;
  readonly recallButton: HTMLElement;
  readonly bonusToast: HTMLElement;
}

function chipSpan(): HTMLElement {
  const span = document.createElement("span");
  span.className = "text-[10px] font-semibold text-white/90";
  return span;
}

function buildFace(): FaceParts {
  const stage = document.createElement("div");
  stage.className = "relative flex h-full w-full flex-col overflow-hidden select-none";
  const chipRow = document.createElement("div");
  chipRow.className =
    "pointer-events-none absolute top-3 left-1/2 flex w-[92%] -translate-x-1/2 flex-wrap items-center justify-center gap-x-3 gap-y-1 rounded-2xl bg-black/60 px-3 py-1.5 backdrop-blur-sm";
  const chipGold = chipSpan();
  const chipCs = chipSpan();
  const chipItems = chipSpan();
  const chipHealth = chipSpan();
  const chipMissed = chipSpan();
  chipMissed.className = "text-[10px] font-semibold text-white/40";
  chipRow.append(chipGold, chipCs, chipItems, chipHealth, chipMissed);
  const center = document.createElement("div");
  center.className = "flex flex-1 flex-col items-center justify-center gap-4 px-6";
  const minionLabel = document.createElement("div");
  minionLabel.className = "text-sm font-semibold text-white/80";
  const healthTrack = document.createElement("div");
  healthTrack.className =
    "h-10 w-full max-w-md overflow-hidden rounded-2xl border-2 border-white/15 bg-black/40";
  const healthFill = document.createElement("div");
  healthFill.className =
    "h-full bg-gradient-to-r from-blue-600 to-blue-400 transition-[width] duration-100";
  healthTrack.appendChild(healthFill);
  center.append(minionLabel, healthTrack);
  const actions = document.createElement("div");
  actions.className = "flex shrink-0 gap-3 border-t border-white/10 px-4 pt-3 pb-5";
  const lastHitButton = document.createElement("div");
  lastHitButton.className =
    "flex flex-1 items-center justify-center rounded-2xl bg-cyan-500 py-4 text-base font-bold text-black transition-all active:scale-95";
  const recallButton = document.createElement("div");
  recallButton.className =
    "flex flex-1 items-center justify-center rounded-2xl border border-amber-400/60 bg-amber-500/20 py-4 text-sm font-bold text-amber-200 transition-all active:scale-95";
  actions.append(lastHitButton, recallButton);
  const bonusToast = document.createElement("div");
  bonusToast.className =
    "pointer-events-none absolute top-16 left-1/2 -translate-x-1/2 rounded-full bg-status-ok/20 px-3 py-1 text-[11px] font-bold text-status-ok opacity-0 transition-opacity";
  stage.append(chipRow, center, actions, bonusToast);
  return {
    stage,
    chipGold,
    chipCs,
    chipItems,
    chipHealth,
    chipMissed,
    minionLabel,
    healthFill,
    lastHitButton,
    recallButton,
    bonusToast,
  };
}

export const economicSimModule: GameModule = {
  code: "economic_sim",
  create(ctx: GameModuleCtx): GameInstance {
    // One complete shape, every field from its boot key — the logic
    // layer's functions all take it whole.
    const p: EsParams = {
      decayIntervalMs: ctx.params.num("decay_interval_ms", ECONOMIC_SIM_DEFAULTS.decayIntervalMs),
      decayMelee: ctx.params.num("decay_melee", ECONOMIC_SIM_DEFAULTS.decayMelee),
      decayRanged: ctx.params.num("decay_ranged", ECONOMIC_SIM_DEFAULTS.decayRanged),
      decayCannon: ctx.params.num("decay_cannon", ECONOMIC_SIM_DEFAULTS.decayCannon),
      respawnDelayMs: ctx.params.num("respawn_delay_ms", ECONOMIC_SIM_DEFAULTS.respawnDelayMs),
      perfectWindowMelee: ctx.params.num(
        "perfect_window_melee",
        ECONOMIC_SIM_DEFAULTS.perfectWindowMelee,
      ),
      perfectWindowCannon: ctx.params.num(
        "perfect_window_cannon",
        ECONOMIC_SIM_DEFAULTS.perfectWindowCannon,
      ),
      goodWindowMelee: ctx.params.num("good_window_melee", ECONOMIC_SIM_DEFAULTS.goodWindowMelee),
      goodWindowCannon: ctx.params.num(
        "good_window_cannon",
        ECONOMIC_SIM_DEFAULTS.goodWindowCannon,
      ),
      goldMelee: ctx.params.num("gold_melee", ECONOMIC_SIM_DEFAULTS.goldMelee),
      goldCannon: ctx.params.num("gold_cannon", ECONOMIC_SIM_DEFAULTS.goldCannon),
      earlyHealthCap: ctx.params.num("early_health_cap", ECONOMIC_SIM_DEFAULTS.earlyHealthCap),
      healthPenaltyPct: ctx.params.num(
        "health_penalty_pct",
        ECONOMIC_SIM_DEFAULTS.healthPenaltyPct,
      ),
      itemCost1: ctx.params.num("item_cost_1", ECONOMIC_SIM_DEFAULTS.itemCost1),
      itemCost2: ctx.params.num("item_cost_2", ECONOMIC_SIM_DEFAULTS.itemCost2),
      itemCost3: ctx.params.num("item_cost_3", ECONOMIC_SIM_DEFAULTS.itemCost3),
      maxItems: ctx.params.num("max_items", ECONOMIC_SIM_DEFAULTS.maxItems),
      perfectBonusMs: ctx.params.num("perfect_bonus_ms", ECONOMIC_SIM_DEFAULTS.perfectBonusMs),
    };

    const face = buildFace();

    let started = false;
    let paused = false;
    let destroyed = false;
    let finished = false;
    let gold = 0;
    let cs = 0;
    let perfectCs = 0;
    let items = 0;
    let healthPct = 100;
    let missed = 0;
    let minionType: EsMinionType = "melee";
    let spawnNetMs = 0;
    let deadNetMs: number | null = null;
    let toastTimer: ReturnType<typeof setTimeout> | null = null;

    function itemCosts(): number[] {
      return [p.itemCost1, p.itemCost2, p.itemCost3];
    }

    function nextCost(): number | null {
      if (items >= p.maxItems || items >= itemCosts().length) {
        return null;
      }
      return itemCosts()[items] ?? null;
    }

    function writeHealthStat(): void {
      // 15-per-percent on the wire scale — written at every change so a
      // clock-out reads the true terminal value whatever ended the run.
      ctx.stat("final_health_hp", Math.max(0, Math.round(healthPct * 15)));
    }

    function liveHealth(): number {
      return healthAt(ctx.clock.elapsedMs() - spawnNetMs, minionType, p);
    }

    function spawnMinion(): void {
      minionType = nextMinionType(Math.random);
      spawnNetMs = ctx.clock.elapsedMs();
      deadNetMs = null;
    }

    function flashBonus(): void {
      face.bonusToast.classList.remove("opacity-0");
      if (toastTimer !== null) {
        clearTimeout(toastTimer);
      }
      toastTimer = setTimeout(() => face.bonusToast.classList.add("opacity-0"), 800);
    }

    function render(): void {
      const health = deadNetMs === null ? liveHealth() : 0;
      face.chipGold.textContent = m.test_es_gold({ n: gold });
      face.chipCs.textContent = m.test_es_cs({ n: cs });
      face.chipItems.textContent = m.test_es_items({ n: items, total: p.maxItems });
      face.chipHealth.textContent = m.test_es_health({ n: healthPct });
      face.chipMissed.textContent = m.test_es_missed({ n: missed });
      face.minionLabel.textContent = deadNetMs === null ? MINION_LABEL[minionType]() : "···";
      face.healthFill.style.width = `${health}%`;
      face.healthFill.className = `h-full transition-[width] duration-100 ${
        health > 50
          ? "bg-gradient-to-r from-blue-600 to-blue-400"
          : health > 20
            ? "bg-gradient-to-r from-amber-500 to-amber-300"
            : "bg-gradient-to-r from-emerald-500 to-emerald-300"
      }`;
      face.lastHitButton.textContent = m.test_es_last_hit();
      const cost = nextCost();
      face.recallButton.textContent =
        cost !== null ? m.test_es_recall({ cost }) : m.test_es_recall_full();
      const canRecall = cost !== null && gold >= cost;
      face.recallButton.classList.toggle("opacity-40", !canRecall);
    }

    function endRun(): void {
      finished = true;
      writeHealthStat();
      ctx.end("completed");
    }

    function tapLastHit(): void {
      if (destroyed || finished || paused || deadNetMs !== null) {
        return;
      }
      const now = ctx.clock.elapsedMs();
      const health = healthAt(now - spawnNetMs, minionType, p);
      if (health <= 0) {
        // The frame loop owns the death hand-off; a tap into a spent
        // minion mints nothing.
        return;
      }
      const window = windowOf(health, minionType, p);
      const goldEarned = goldOf(window, minionType, p);
      const confirmed = confirmsHit(window);
      ctx.collect({
        type: "hit",
        elapsedMs: Math.round(now),
        isLastHit: confirmed,
        goldEarned,
        minionType,
        minionHealth: Math.min(100, Math.max(1, health)),
      });
      gold += goldEarned;
      if (confirmed) {
        cs += 1;
      }
      if (window === "perfect") {
        perfectCs += 1;
        ctx.stat("perfect_cs", perfectCs);
        ctx.clock.addTime(p.perfectBonusMs);
        flashBonus();
      } else if (window === "too_early") {
        healthPct = Math.max(0, healthPct - p.healthPenaltyPct);
        writeHealthStat();
        if (healthPct <= 0) {
          endRun();
          return;
        }
      }
      spawnMinion();
      render();
    }

    function recall(): void {
      if (destroyed || finished || paused) {
        return;
      }
      const cost = nextCost();
      if (cost === null || gold < cost) {
        return;
      }
      const now = ctx.clock.elapsedMs();
      items += 1;
      gold -= cost;
      healthPct = 100;
      writeHealthStat();
      ctx.collect({
        type: "purchase",
        elapsedMs: Math.round(now),
        itemName: `item-${items}`,
        cost,
      });
      render();
    }

    const loop = createFrameLoop((): void => {
      if (destroyed || finished || paused) {
        return;
      }
      const now = ctx.clock.elapsedMs();
      if (deadNetMs === null && liveHealth() <= 0) {
        deadNetMs = now;
        missed += 1;
        ctx.stat("missed_minions", missed);
      }
      if (deadNetMs !== null && now - deadNetMs >= p.respawnDelayMs) {
        spawnMinion();
      }
      render();
    });

    return {
      start(): void {
        if (destroyed || started) {
          return;
        }
        started = true;
        ctx.mount.appendChild(face.stage);
        face.lastHitButton.addEventListener("pointerdown", tapLastHit);
        face.recallButton.addEventListener("pointerdown", recall);
        writeHealthStat();
        render();
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
        if (toastTimer !== null) {
          clearTimeout(toastTimer);
        }
        face.stage.remove();
      },
    };
  },
};
