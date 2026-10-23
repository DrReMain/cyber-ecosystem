import { useNavigate } from "@tanstack/react-router";
import { Repeat, Save, Sun, Timer, Trophy } from "lucide-react";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { m } from "#/paraglide/messages";
import { FlowHeader } from "./ui/flow-header";
import { useArrangedSteps } from "./use-arranged-steps";

/**
 * The pre-assessment notice (prototype test-guide): first/second-session
 * head, the five rules, the scoring-basis card, and the ready CTA — which
 * enters the current step's intro (prototype law), not a dead end. The
 * first/second distinction rides the match session — every session is the
 * first one until then.
 */
export function TestGuidePage() {
  const navigate = useNavigate();
  const { steps, pending } = useArrangedSteps();
  // isFirstSession arrives with the match-session law; until then the face
  // is permanently the first-assessment one.
  const first = true;
  const count = steps.length > 0 ? steps.length : 4;

  function onReady() {
    const firstStep = steps[0];
    if (firstStep !== undefined) {
      void navigate({ to: "/test/$code", params: { code: firstStep.code } });
      return;
    }
    // Nothing playable resolves (transient catalog state) — the lineup
    // page owns that face.
    void navigate({ to: "/match" });
  }

  const rules = [
    { Icon: Sun, key: "1", title: m.match_guide_rule1_t(), desc: m.match_guide_rule1_d() },
    {
      Icon: Timer,
      key: "2",
      title: m.match_guide_rule2_t({ n: count }),
      desc: m.match_guide_rule2_d({ n: count }),
    },
    {
      Icon: Repeat,
      key: "3",
      title: first ? m.match_guide_rule3_t_first() : m.match_guide_rule3_t_second(),
      desc: m.match_guide_rule3_d(),
    },
    { Icon: Trophy, key: "4", title: m.match_guide_rule4_t(), desc: m.match_guide_rule4_d() },
    { Icon: Save, key: "5", title: m.match_guide_rule5_t(), desc: m.match_guide_rule5_d() },
  ];

  return (
    <AmbientLayout>
      <div className="flex w-full flex-col gap-3 pb-6">
        <FlowHeader onBack={() => void history.back()} title={m.match_guide_title()} />

        <div className="border-line bg-surface-card rounded-card border p-4">
          <div className="flex items-center gap-2.5">
            <div className="rounded-tile bg-cta-practice flex h-10 w-10 shrink-0 items-center justify-center">
              <Save aria-hidden className="text-white" size={18} />
            </div>
            <div>
              <div className="text-ink text-[14.5px] font-extrabold">
                {first ? m.match_guide_first() : m.match_guide_second()}
              </div>
              <div className="text-ink-tertiary mt-0.5 text-[11px]">
                {first
                  ? m.match_guide_first_sub({ n: count })
                  : m.match_guide_second_sub({ n: count })}
              </div>
            </div>
          </div>
          <div className="mt-2">
            {rules.map(({ Icon, key, title, desc }) => (
              <div
                className="border-line-soft flex gap-2.5 border-b py-2 last:border-b-0"
                key={key}
              >
                <Icon aria-hidden className="text-ink-secondary mt-0.5 shrink-0" size={17} />
                <div>
                  <div className="text-ink text-[12.5px] font-bold">{title}</div>
                  <div className="text-ink-tertiary mt-0.5 text-[10.5px] leading-relaxed">
                    {desc}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="border-line bg-surface-deep text-ink-tertiary rounded-control border p-3.5 text-[11.5px] leading-relaxed">
          <span className="text-ink-secondary font-bold">{m.match_guide_score_t()}</span>：
          {m.match_guide_score_d()}
        </div>

        <button
          className="bg-cta-primary rounded-control w-full px-4 py-[14px] text-[15px] font-bold text-white active:opacity-80"
          disabled={pending}
          onClick={onReady}
          type="button"
        >
          {m.match_guide_ready()}
        </button>
        <div className="text-ink-tertiary text-center text-[11px]">{m.match_guide_keep()}</div>
      </div>
    </AmbientLayout>
  );
}
