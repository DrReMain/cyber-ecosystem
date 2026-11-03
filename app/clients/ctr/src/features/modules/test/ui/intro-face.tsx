import { ChevronLeft } from "lucide-react";
import { useEffect, useState } from "react";
import { Button } from "react-aria-components";
import { abilityWord } from "#/features/modules/match/vocabulary";
import { m } from "#/paraglide/messages";
import type { StepContext } from "../contract";
import { ProgStrip } from "./prog-strip";

/* The intro blob rides the ability's family color (prototype test-intro). */
const abilityBlob: Record<string, string> = {
  reactionSpeed: "bg-[#7f1d1d]",
  aimingAccuracy: "bg-[#ea580c]",
  trackingStability: "bg-[#0e7490]",
  skillReleaseAccuracy: "bg-[#7e22ce]",
  resourceAllocation: "bg-[#15803d]",
  decisionEfficiency: "bg-[#1d4ed8]",
};

interface IntroFaceProps {
  readonly testName: string;
  readonly introduction: string;
  readonly unitLabel: string;
  readonly step: StepContext | null;
  /** The played receipt (prototype #prev): this game's session best and
   * cross-session best; null while no scored line exists. */
  readonly prevBest: { sessionBest: number; historyBest: number | null } | null;
  readonly onBack: () => void;
  readonly onStart: () => void;
}

/**
 * The pre-run face (prototype test-intro): the lineup strip, the step's
 * identity card over its how-to card, the session's two mini receipts,
 * and the ready CTA into a 3-2-1 countdown (the countdown is an addition;
 * the prototype jumps straight into the run). Practice runs — no session
 * binding — drop the strip, the receipts, and the step numbering.
 */
export function IntroFace({
  testName,
  introduction,
  unitLabel,
  step,
  prevBest,
  onBack,
  onStart,
}: Readonly<IntroFaceProps>) {
  const [countdown, setCountdown] = useState<number | null>(null);

  useEffect(() => {
    if (countdown === null) {
      return;
    }
    if (countdown === 0) {
      onStart();
      return;
    }
    const id = setTimeout(() => setCountdown(countdown - 1), 1000);
    return () => clearTimeout(id);
  }, [countdown, onStart]);

  const word = step !== null ? abilityWord(step.ability) : null;
  const blob = (step !== null ? abilityBlob[step.ability] : undefined) ?? "bg-[#2563eb]";
  // The category was a pool word (universal/fps/moba) — internal
  // vocabulary the prototype's identity card never shows (Q1-⑦); the
  // unit label renders only when the test row authored one.
  const catUnit = unitLabel;
  const title = step !== null ? m.test_intro_step_title({ n: step.index + 1 }) : testName;
  // A bound step with its retake budget spent must not start a run — the
  // 30 seconds would only die at submit (ESPORTS_MATCH_RETEST_EXHAUSTED).
  // Practice runs have no budget to spend and stay open.
  const spent = step !== null && step.retestLeft <= 0;

  // overflow-x-hidden: the orb bleeds past the right edge by design, and
  // a lone overflow-y-auto computes overflow-x to auto (the CSS
  // visible-with-non-visible rule) — the bleed then buys the face a
  // horizontal scrollbar. x stays clipped, y scrolls.
  return (
    <div className="bg-canvas absolute inset-0 isolate z-40 overflow-x-hidden overflow-y-auto">
      <div
        aria-hidden
        className={`pointer-events-none absolute top-[-60px] right-[-70px] size-[240px] rounded-full opacity-26 blur-[70px] ${blob}`}
      />
      <div className="relative w-full">
        <div className="flex items-center gap-3 px-4 pt-5 pb-2.5">
          <button
            aria-label={m.common_back()}
            className="bg-surface-input text-ink-secondary flex h-10 w-10 shrink-0 items-center justify-center rounded-full active:opacity-60"
            onClick={onBack}
            type="button"
          >
            <ChevronLeft aria-hidden size={20} />
          </button>
          <h1 className="text-ink text-[19px] font-bold">{title}</h1>
        </div>

        {step !== null ? (
          <div className="mb-[14px] px-4">
            <ProgStrip currentIndex={step.index} doneCodes={step.doneCodes} lineup={step.lineup} />
          </div>
        ) : null}

        <div className="flex flex-col gap-3 px-4 pb-6">
          <div className="border-line bg-surface-card shadow-card rounded-card border px-4 py-[18px] backdrop-blur-md">
            <div className="flex items-center gap-3">
              {step !== null ? (
                <div
                  className={`${word?.gradient} shadow-inset-ring flex h-[52px] w-[52px] shrink-0 items-center justify-center rounded-[12px] text-[22px] font-bold text-white`}
                >
                  {step.index + 1}
                </div>
              ) : null}
              <div className="min-w-0">
                {word !== null ? (
                  <div className="text-ink-tertiary text-[11px]">{word.cn}</div>
                ) : null}
                <div className="text-ink mt-[3px] text-[17px] font-extrabold">{testName}</div>
                {catUnit !== "" ? (
                  <div className="text-ink-tertiary mt-[3px] text-[10.5px]">{catUnit}</div>
                ) : null}
              </div>
            </div>
            {/* The how-to card rides the server's introduction (the seed
                has not authored it yet — Q1-⑥); an absent text leaves the
                identity card clean instead of an empty bordered box. */}
            {introduction !== "" ? (
              <>
                <div aria-hidden className="h-[14px]" />
                <div className="border-line bg-surface-content text-ink-muted rounded-control border px-[13px] py-3 text-[12px] leading-[1.85]">
                  {introduction}
                </div>
              </>
            ) : null}
          </div>

          {step !== null ? (
            <div className="flex gap-2.5">
              <div className="border-line bg-surface-content rounded-control flex-1 border px-3 py-[11px]">
                <div className="text-ink-tertiary text-[10.5px]">
                  {m.test_intro_progress_label()}
                </div>
                <div className="text-ink mt-[3px] text-[15px] font-extrabold tabular-nums">
                  {step.doneCodes.length} / {step.lineup.length}
                </div>
              </div>
              <div className="border-line bg-surface-content rounded-control flex-1 border px-3 py-[11px]">
                <div className="text-ink-tertiary text-[10.5px]">{m.test_intro_retest_label()}</div>
                <div className="text-ink mt-[3px] text-[15px] font-extrabold tabular-nums">
                  {m.test_intro_retest_value({ n: step.retestLeft })}
                </div>
              </div>
            </div>
          ) : null}

          {prevBest !== null ? (
            <div className="border-line bg-surface-content text-ink-muted rounded-control border px-[13px] py-[11px] text-[11.5px] leading-[1.7]">
              {m.test_intro_prev_prefix()}{" "}
              <b className="font-bold text-[#FDE047] tabular-nums">{prevBest.sessionBest}</b>
              {m.test_intro_prev_history({ n: prevBest.historyBest ?? "—" })}
            </div>
          ) : null}

          <Button
            className={`bg-cta-primary shadow-cta-primary rounded-control w-full cursor-pointer px-4 py-[14px] text-[15px] font-bold text-white active:opacity-80 ${
              spent ? "pointer-events-none opacity-40" : ""
            }`}
            isDisabled={spent}
            onPress={() => setCountdown(3)}
          >
            {m.test_intro_go()}
          </Button>
          {spent ? (
            <div className="text-ink-tertiary px-0.5 text-center text-[11px] leading-relaxed">
              {m.test_intro_blocked()}
            </div>
          ) : null}
        </div>
      </div>

      {countdown !== null ? (
        <div className="bg-canvas absolute inset-0 z-50 flex flex-col items-center justify-center gap-3">
          <div className="text-ink text-7xl font-extrabold tabular-nums">{countdown}</div>
          <div className="text-ink-secondary text-sm">{m.test_countdown_ready()}</div>
          <button
            className="text-ink-tertiary mt-4 cursor-pointer text-sm underline active:opacity-60"
            onClick={onBack}
            type="button"
          >
            {m.test_countdown_cancel()}
          </button>
        </div>
      ) : null}
    </div>
  );
}
