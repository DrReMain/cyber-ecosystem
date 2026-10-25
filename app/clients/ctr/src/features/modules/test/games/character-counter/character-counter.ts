// character_counter — counter-pick against a drawn enemy lineup. Each
// round draws five distinct roles from the six-role roster, the player
// picks their own five, and the confirm stamps the round's fact: both
// lineups as English role tokens (the wire vocabulary the server's bank
// judges) with the round's net-clock spend. A confirm ends the feedback
// window and deals the next enemy; the third confirm is the run's natural
// end. The clock-out path mints nothing for an unfinished round — its
// lineup may not be full, and a partial lineup must not reach the gate.
// The hit pairs shown between rounds read the relation mirror (the
// feedback face's only judgment — the scoring key stays server-side).

import { m } from "#/paraglide/messages";
import type { GameInstance, GameModule, GameModuleCtx } from "../../contract";
import { createFrameLoop } from "../../frame-loop";
import {
  CC_ROSTER,
  type CcRole,
  CHARACTER_COUNTER_DEFAULTS,
  counterScore,
  drawEnemyLineup,
  hitPairs,
} from "./logic";

const ROLE_LABEL: Record<CcRole, () => string> = {
  Tank: m.test_cc_role_tank,
  Fighter: m.test_cc_role_fighter,
  Mage: m.test_cc_role_mage,
  Marksman: m.test_cc_role_marksman,
  Support: m.test_cc_role_support,
  Assassin: m.test_cc_role_assassin,
};

const ROLE_DESC: Record<CcRole, () => string> = {
  Tank: m.test_cc_role_desc_tank,
  Fighter: m.test_cc_role_desc_fighter,
  Mage: m.test_cc_role_desc_mage,
  Marksman: m.test_cc_role_desc_marksman,
  Support: m.test_cc_role_desc_support,
  Assassin: m.test_cc_role_desc_assassin,
};

type Phase = "pick" | "feedback";

interface FaceParts {
  readonly stage: HTMLElement;
  readonly chipRound: HTMLElement;
  readonly chipTime: HTMLElement;
  readonly board: HTMLElement;
  readonly enemyRow: HTMLElement;
  readonly pickGrid: HTMLElement;
  readonly chipSelected: HTMLElement;
  readonly confirm: HTMLElement;
  readonly feedback: HTMLElement;
  readonly feedbackTitle: HTMLElement;
  readonly feedbackPairs: HTMLElement;
}

function sectionLabel(text: () => string): HTMLElement {
  const label = document.createElement("div");
  label.className = "text-[11px] font-semibold tracking-wide text-white/50";
  label.textContent = text();
  return label;
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
  const chipSelected = document.createElement("span");
  chipSelected.className = "text-[10px] font-semibold text-status-ok";
  chip.append(chipRound, chipTime, chipSelected);
  const board = document.createElement("div");
  board.className = "flex flex-1 flex-col gap-3 overflow-y-auto px-4 pt-12 pb-4";
  const enemyBlock = document.createElement("div");
  enemyBlock.className = "flex flex-col gap-2";
  enemyBlock.appendChild(sectionLabel(m.test_cc_enemy));
  const enemyRow = document.createElement("div");
  enemyRow.className = "flex flex-wrap gap-2";
  enemyBlock.appendChild(enemyRow);
  const pickBlock = document.createElement("div");
  pickBlock.className = "flex flex-1 flex-col gap-2";
  pickBlock.appendChild(sectionLabel(m.test_cc_pick));
  const pickGrid = document.createElement("div");
  pickGrid.className = "grid flex-1 grid-cols-2 content-start gap-2";
  pickBlock.appendChild(pickGrid);
  board.append(enemyBlock, pickBlock);
  const footer = document.createElement("div");
  footer.className =
    "flex shrink-0 items-center justify-center border-t border-white/10 px-4 pt-3 pb-5";
  const confirm = document.createElement("div");
  confirm.className =
    "flex items-center justify-center rounded-2xl px-8 py-3 text-sm font-bold text-white transition-all active:scale-95";
  footer.appendChild(confirm);
  const feedback = document.createElement("div");
  feedback.className =
    "absolute inset-x-4 top-1/2 z-10 hidden -translate-y-1/2 flex-col gap-2 rounded-2xl border border-white/10 bg-deep/95 p-4 backdrop-blur";
  const feedbackTitle = document.createElement("div");
  feedbackTitle.className = "text-sm font-bold text-white";
  const feedbackPairs = document.createElement("div");
  feedbackPairs.className = "flex flex-col gap-1";
  feedback.append(feedbackTitle, feedbackPairs);
  stage.append(chip, board, footer, feedback);
  return {
    stage,
    chipRound,
    chipTime,
    chipSelected,
    board,
    enemyRow,
    pickGrid,
    confirm,
    feedback,
    feedbackTitle,
    feedbackPairs,
  };
}

export const characterCounterModule: GameModule = {
  code: "character_counter",
  create(ctx: GameModuleCtx): GameInstance {
    const params = {
      roundCount: ctx.params.num("round_count", CHARACTER_COUNTER_DEFAULTS.roundCount),
      lineupSize: ctx.params.num("lineup_size", CHARACTER_COUNTER_DEFAULTS.lineupSize),
    };

    const face = buildFace();

    let started = false;
    let paused = false;
    let destroyed = false;
    let finished = false;
    let phase: Phase = "pick";
    let round = 0;
    let roundStartMs = 0;
    let feedbackStartMs = 0;
    let enemy: CcRole[] = [];
    let selected: CcRole[] = [];

    function setChip(): void {
      const remaining = Math.max(0, ctx.durationMs - ctx.clock.elapsedMs());
      face.chipRound.textContent = m.test_cc_round({ n: round, total: params.roundCount });
      face.chipTime.textContent = m.test_od_time_left({ s: (remaining / 1000).toFixed(1) });
      face.chipSelected.textContent = m.test_cc_selected({
        n: selected.length,
        total: params.lineupSize,
      });
    }

    function setConfirm(): void {
      const full = selected.length === params.lineupSize;
      face.confirm.className = full
        ? "flex items-center justify-center rounded-2xl bg-cyan-500 px-8 py-3 text-sm font-bold text-black transition-all active:scale-95"
        : "flex items-center justify-center rounded-2xl bg-white/10 px-8 py-3 text-sm font-bold text-white/40 transition-all";
      face.confirm.textContent = m.test_cc_confirm();
    }

    function toggle(role: CcRole): void {
      if (selected.includes(role)) {
        selected = selected.filter((entry) => entry !== role);
      } else if (selected.length < params.lineupSize) {
        selected = [...selected, role];
      }
      setConfirm();
      setChip();
      renderPick();
    }

    function renderPick(): void {
      face.pickGrid.replaceChildren();
      for (const role of CC_ROSTER) {
        const picked = selected.includes(role);
        const card = document.createElement("div");
        card.className = picked
          ? "flex flex-col gap-1 rounded-xl border-2 border-cyan-400 bg-cyan-400/15 p-3"
          : "flex flex-col gap-1 rounded-xl border border-white/15 bg-white/5 p-3";
        const name = document.createElement("span");
        name.className = "text-sm font-bold text-white";
        name.textContent = ROLE_LABEL[role]();
        const desc = document.createElement("span");
        desc.className = "text-[10px] leading-snug text-white/50";
        desc.textContent = ROLE_DESC[role]();
        card.append(name, desc);
        card.addEventListener("pointerdown", () => {
          if (phase === "pick") {
            toggle(role);
          }
        });
        face.pickGrid.appendChild(card);
      }
    }

    function openRound(): void {
      phase = "pick";
      face.feedback.classList.add("hidden");
      face.feedback.classList.remove("flex");
      round += 1;
      roundStartMs = ctx.clock.elapsedMs();
      enemy = drawEnemyLineup(params.lineupSize, Math.random);
      selected = [];
      face.enemyRow.replaceChildren();
      for (const role of enemy) {
        const chip = document.createElement("span");
        chip.className =
          "rounded-full border border-danger/40 bg-danger/15 px-3 py-1.5 text-xs font-bold text-white";
        chip.textContent = ROLE_LABEL[role]();
        face.enemyRow.appendChild(chip);
      }
      renderPick();
      setConfirm();
      setChip();
    }

    function showFeedback(): void {
      const hits = hitPairs(selected, enemy);
      face.feedbackTitle.textContent = m.test_cc_round_hits({
        n: hits.length,
        score: counterScore(hits),
      });
      face.feedbackPairs.replaceChildren();
      for (const hit of hits) {
        const pair = document.createElement("span");
        pair.className = "text-xs font-semibold text-status-ok";
        pair.textContent = `${ROLE_LABEL[hit.counter]()} → ${ROLE_LABEL[hit.countered]()} · ${hit.score}`;
        face.feedbackPairs.appendChild(pair);
      }
      face.feedback.classList.remove("hidden");
      face.feedback.classList.add("flex");
    }

    function confirmRound(): void {
      if (
        destroyed ||
        finished ||
        paused ||
        phase !== "pick" ||
        selected.length !== params.lineupSize
      ) {
        return;
      }
      phase = "feedback";
      feedbackStartMs = ctx.clock.elapsedMs();
      const spent = Math.max(0, Math.round(feedbackStartMs - roundStartMs));
      ctx.collect({
        roundNumber: round,
        enemyLineup: [...enemy],
        selectedLineup: [...selected],
        timeSpentMs: spent,
      });
      showFeedback();
      setChip();
    }

    // The feedback window rides the net clock — a pause freezes it with
    // the run, so the round hand-off can never stall behind a wall timer.
    const loop = createFrameLoop((): void => {
      if (destroyed || finished || paused) {
        return;
      }
      if (phase === "feedback" && ctx.clock.elapsedMs() - feedbackStartMs >= 900) {
        if (round >= params.roundCount) {
          finished = true;
          ctx.end("completed");
          return;
        }
        openRound();
      }
      setChip();
    });

    return {
      start(): void {
        if (destroyed || started) {
          return;
        }
        started = true;
        ctx.mount.appendChild(face.stage);
        face.confirm.addEventListener("pointerdown", confirmRound);
        openRound();
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
        face.stage.remove();
      },
    };
  },
};
