// offense_defense — a decision card per round: the scenario's field
// report, three stances, and a window. The fact stream carries one
// decision per round (the pick with its round-relative time, or the
// timeout flag with no choice); correctness is the server's judgment
// against the seeded key, the per-round feedback card reads the bank
// mirror's optimal only to teach, and a timeout spends its round with
// no feedback pause. Rounds fit the config clock by construction:
// worst case rounds × (window + feedback) stays under the 30-second
// run. A pause parks the loop, and every window is measured on the net
// clock, so pausing stretches wall time, never decision time.

import { m } from "#/paraglide/messages";
import type { GameInstance, GameModule, GameModuleCtx } from "../../contract";
import { createFrameLoop } from "../../frame-loop";
import { createLucideIcon, type IconNode } from "../lucide-icon";
import { OFFENSE_DEFENSE_DEFAULTS, pickScenarios } from "./logic";
import { type OdChoice, type OdComplexity, type OdScenario, OFFENSE_SCENARIOS } from "./scenarios";

// lucide check / x node trees, verbatim geometry — the feedback verdict
// renders as icons, never text glyphs.
const CHECK: IconNode = [["path", { d: "M20 6 9 17l-5-5" }]];
const X_MARK: IconNode = [
  ["path", { d: "M18 6 6 18" }],
  ["path", { d: "m6 6 12 12" }],
];

const CHOICE_STYLE: Record<OdChoice, { readonly face: string; readonly label: () => string }> = {
  attack: { face: "bg-blue-600 active:bg-blue-500", label: m.test_od_choice_attack },
  defense: { face: "bg-amber-500 active:bg-amber-400", label: m.test_od_choice_defense },
  retreat: { face: "bg-red-600 active:bg-red-500", label: m.test_od_choice_retreat },
};

const COMPLEXITY_LABEL: Record<OdComplexity, () => string> = {
  simple: m.test_od_diff_simple,
  medium: m.test_od_diff_medium,
  complex: m.test_od_diff_complex,
};

type Phase = "decide" | "feedback";

interface FaceParts {
  readonly stage: HTMLElement;
  readonly chipRound: HTMLElement;
  readonly chipTime: HTMLElement;
  readonly chipCorrect: HTMLElement;
  readonly card: HTMLElement;
  readonly badgeCategory: HTMLElement;
  readonly badgeComplexity: HTMLElement;
  readonly description: HTMLElement;
  readonly buttons: Readonly<Record<OdChoice, HTMLElement>>;
  readonly feedback: HTMLElement;
  readonly feedbackIcon: HTMLElement;
  readonly feedbackTitle: HTMLElement;
  readonly feedbackBest: HTMLElement;
  readonly feedbackText: HTMLElement;
}

// The game's face: a read-out chip, the scenario card, the three stance
// buttons, and a feedback card that fills only between rounds — built
// once, rounds only rebind its content.
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
  const chipCorrect = document.createElement("span");
  chipCorrect.className = "text-[10px] font-semibold text-status-ok";
  chip.append(chipRound, chipTime, chipCorrect);
  const body = document.createElement("div");
  body.className = "flex flex-1 flex-col justify-center gap-4 overflow-y-auto px-5 py-14";
  const card = document.createElement("div");
  card.className =
    "flex flex-col gap-3 rounded-2xl border border-white/10 bg-white/5 p-4 backdrop-blur-sm";
  const badges = document.createElement("div");
  badges.className = "flex items-center gap-2";
  const badgeCategory = document.createElement("span");
  badgeCategory.className =
    "rounded-full border border-white/15 px-2 py-0.5 text-[10px] font-semibold text-white/60";
  const badgeComplexity = document.createElement("span");
  badgeComplexity.className =
    "rounded-full border border-white/15 px-2 py-0.5 text-[10px] font-semibold text-white/60";
  badges.append(badgeCategory, badgeComplexity);
  const description = document.createElement("p");
  description.className = "text-[15px] leading-relaxed font-medium text-white";
  const feedback = document.createElement("div");
  feedback.className = "hidden flex-col gap-1 rounded-xl p-3";
  const feedbackHead = document.createElement("div");
  feedbackHead.className = "flex items-center gap-2";
  const feedbackIcon = document.createElement("span");
  feedbackIcon.className = "flex h-5 w-5 items-center justify-center";
  const feedbackTitle = document.createElement("span");
  feedbackTitle.className = "text-sm font-bold";
  feedbackHead.append(feedbackIcon, feedbackTitle);
  const feedbackBest = document.createElement("span");
  feedbackBest.className = "text-[11px] font-semibold text-white/80";
  const feedbackText = document.createElement("span");
  feedbackText.className = "text-[11px] leading-relaxed text-white/60";
  feedback.append(feedbackHead, feedbackBest, feedbackText);
  card.append(badges, description, feedback);
  body.appendChild(card);
  const buttonRow = document.createElement("div");
  buttonRow.className = "flex shrink-0 gap-2.5 border-t border-white/10 px-4 pt-3 pb-5";
  const buttons = {} as Record<OdChoice, HTMLElement>;
  for (const choice of ["attack", "defense", "retreat"] as const) {
    const button = document.createElement("div");
    button.className = `flex flex-1 flex-col items-center gap-1 rounded-2xl py-3.5 font-bold text-white ${CHOICE_STYLE[choice].face}`;
    const label = document.createElement("span");
    label.className = "text-sm";
    label.textContent = CHOICE_STYLE[choice].label();
    button.appendChild(label);
    buttons[choice] = button;
    buttonRow.appendChild(button);
  }
  stage.append(chip, body, buttonRow);
  return {
    stage,
    chipRound,
    chipTime,
    chipCorrect,
    card,
    badgeCategory,
    badgeComplexity,
    description,
    buttons,
    feedback,
    feedbackIcon,
    feedbackTitle,
    feedbackBest,
    feedbackText,
  };
}

export const offenseDefenseModule: GameModule = {
  code: "offense_defense",
  create(ctx: GameModuleCtx): GameInstance {
    const params = {
      totalRounds: ctx.params.num("total_rounds", OFFENSE_DEFENSE_DEFAULTS.totalRounds),
      timePerRoundMs: ctx.params.num("time_per_round_ms", OFFENSE_DEFENSE_DEFAULTS.timePerRoundMs),
      feedbackMs: ctx.params.num("feedback_ms", OFFENSE_DEFENSE_DEFAULTS.feedbackMs),
    };
    const lineup = pickScenarios(OFFENSE_SCENARIOS, params.totalRounds, Math.random);

    const face = buildFace();

    let started = false;
    let paused = false;
    let destroyed = false;
    let finished = false;
    let phase: Phase = "decide";
    let round = 0;
    let correct = 0;
    let roundStartMs = 0;
    let phaseStartMs = 0;
    let current: OdScenario | undefined;

    function setChip(): void {
      face.chipRound.textContent = m.test_od_round({ n: round + 1, total: lineup.length });
      const left = Math.max(0, params.timePerRoundMs - (ctx.clock.elapsedMs() - roundStartMs));
      face.chipTime.textContent = m.test_od_time_left({ s: (left / 1000).toFixed(1) });
      face.chipCorrect.textContent = m.test_od_correct({ n: correct });
    }

    function setButtonsLocked(locked: boolean): void {
      for (const button of Object.values(face.buttons)) {
        button.classList.toggle("pointer-events-none", locked);
        button.classList.toggle("opacity-45", locked);
      }
    }

    function showFeedback(scenario: OdScenario, picked: OdChoice): void {
      const right = picked === scenario.optimal;
      if (right) {
        correct += 1;
      }
      face.feedback.classList.remove("hidden");
      face.feedback.classList.add("flex");
      face.feedback.classList.add(
        "border",
        right ? "border-status-ok/60" : "border-danger/60",
        right ? "bg-status-ok/10" : "bg-danger/10",
      );
      face.feedbackIcon.replaceChildren(createLucideIcon(right ? CHECK : X_MARK, 16));
      face.feedbackIcon.className = `flex h-5 w-5 items-center justify-center ${right ? "text-status-ok" : "text-danger"}`;
      face.feedbackTitle.textContent = right ? m.test_od_right() : m.test_od_wrong();
      face.feedbackTitle.className = `text-sm font-bold ${right ? "text-status-ok" : "text-danger"}`;
      face.feedbackBest.textContent = m.test_od_best({
        choice: CHOICE_STYLE[scenario.optimal].label(),
      });
      const optimalText = scenario.choices.find((entry) => entry.choice === scenario.optimal)?.text;
      face.feedbackText.textContent = optimalText ?? "";
    }

    function hideFeedback(): void {
      face.feedback.classList.add("hidden");
      face.feedback.classList.remove("flex");
      face.feedback.classList.remove(
        "border",
        "border-status-ok/60",
        "border-danger/60",
        "bg-status-ok/10",
        "bg-danger/10",
      );
    }

    function startRound(): void {
      current = lineup[round];
      if (current === undefined) {
        return;
      }
      phase = "decide";
      hideFeedback();
      setButtonsLocked(false);
      for (const choice of ["attack", "defense", "retreat"] as const) {
        face.buttons[choice].classList.remove("ring-2", "ring-white/80");
      }
      face.badgeCategory.textContent = current.category;
      face.badgeComplexity.textContent = COMPLEXITY_LABEL[current.complexity]();
      face.description.textContent = current.description;
      roundStartMs = ctx.clock.elapsedMs();
      phaseStartMs = roundStartMs;
      setChip();
    }

    function decide(choice: OdChoice): void {
      if (destroyed || finished || paused || phase !== "decide" || current === undefined) {
        return;
      }
      phase = "feedback";
      phaseStartMs = ctx.clock.elapsedMs();
      // The pick with its round-relative time — the server judges the
      // choice against the key; a pick faster than any reading can
      // produce fails the reaction floor there.
      ctx.collect({
        scenarioId: current.code,
        choice,
        decisionTimeMs: Math.max(0, Math.round(phaseStartMs - roundStartMs)),
        timeout: false,
      });
      setButtonsLocked(true);
      face.buttons[choice].classList.add("ring-2", "ring-white/80");
      showFeedback(current, choice);
      setChip();
    }

    function timeOut(): void {
      if (current === undefined) {
        return;
      }
      phase = "feedback";
      phaseStartMs = ctx.clock.elapsedMs();
      // An unanswered round: the choice stays absent on the wire, the
      // timeout flag rides, and the round's own window is the time of
      // record. No feedback pause — a timeout spends its window already.
      ctx.collect({
        scenarioId: current.code,
        decisionTimeMs: Math.round(Math.max(phaseStartMs - roundStartMs, params.timePerRoundMs)),
        timeout: true,
      });
      advance();
    }

    function advance(): void {
      round += 1;
      if (round >= lineup.length) {
        finished = true;
        loop.stop();
        ctx.end("completed");
        return;
      }
      startRound();
    }

    const loop = createFrameLoop((): void => {
      if (destroyed || finished || paused) {
        return;
      }
      const elapsed = ctx.clock.elapsedMs();
      if (phase === "decide") {
        if (elapsed - roundStartMs >= params.timePerRoundMs) {
          timeOut();
        } else {
          setChip();
        }
      } else if (elapsed - phaseStartMs >= params.feedbackMs) {
        advance();
      }
    });

    return {
      start(): void {
        if (destroyed || started) {
          return;
        }
        started = true;
        ctx.mount.appendChild(face.stage);
        for (const choice of ["attack", "defense", "retreat"] as const) {
          face.buttons[choice].addEventListener("pointerdown", () => decide(choice));
        }
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
        face.stage.remove();
      },
    };
  },
};
