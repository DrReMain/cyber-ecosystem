import { Repeat, Save, Sun, Timer, Trophy } from "lucide-react";
import { toast } from "sonner";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { m } from "#/paraglide/messages";
import { FlowHeader } from "./ui/flow-header";

/**
 * The pre-assessment notice (prototype test-guide): first/second-session
 * head, the five rules, the scoring-basis card, and the ready CTA. The
 * first/second distinction rides the match session — every session is the
 * first one until then.
 */
export function TestGuidePage() {
  // isFirstSession arrives with the match-session law; until then the face
  // is permanently the first-assessment one.
  const first = true;
  const count = 4;

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

        <div className="rounded-lg border border-line bg-surface-card p-4">
          <div className="flex items-center gap-2.5">
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-tile bg-cta-practice">
              <Save aria-hidden className="text-white" size={18} />
            </div>
            <div>
              <div className="font-extrabold text-[14.5px] text-ink">
                {first ? m.match_guide_first() : m.match_guide_second()}
              </div>
              <div className="mt-0.5 text-[11px] text-ink-tertiary">
                {first
                  ? m.match_guide_first_sub({ n: count })
                  : m.match_guide_second_sub({ n: count })}
              </div>
            </div>
          </div>
          <div className="mt-2">
            {rules.map(({ Icon, key, title, desc }) => (
              <div
                className="flex gap-2.5 border-line-soft border-b py-2 last:border-b-0"
                key={key}
              >
                <Icon aria-hidden className="mt-0.5 shrink-0 text-ink-secondary" size={17} />
                <div>
                  <div className="font-bold text-[12.5px] text-ink">{title}</div>
                  <div className="mt-0.5 text-[10.5px] text-ink-tertiary leading-relaxed">
                    {desc}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="rounded-lg border border-line bg-surface-deep p-3.5 text-[11.5px] text-ink-tertiary leading-relaxed">
          <span className="font-bold text-ink-secondary">{m.match_guide_score_t()}</span>：
          {m.match_guide_score_d()}
        </div>

        <button
          className="w-full rounded-sm bg-cta-primary px-4 py-3 font-semibold text-white active:opacity-80"
          onClick={() => toast.info(m.match_flow_soon())}
          type="button"
        >
          {m.match_guide_ready()}
        </button>
        <div className="text-center text-[11px] text-ink-tertiary">{m.match_guide_keep()}</div>
      </div>
    </AmbientLayout>
  );
}
