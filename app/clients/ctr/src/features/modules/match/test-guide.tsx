import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import { getMatchSession } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/match-MatchService_connectquery";
import { useNavigate } from "@tanstack/react-router";
import { Repeat, Save, Sun, Timer, TriangleAlert, Trophy } from "lucide-react";
import type { AmbientOrb } from "#/features/app/ambient/ambient-backdrop";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { getGameModule } from "#/features/modules/test/registry";
import { m } from "#/paraglide/messages";
import { FlowHeader } from "./ui/flow-header";

/* The page's own orb (prototype test-guide): a single practice-orange blob. */
const guideOrbs: readonly AmbientOrb[] = [
  { className: "top-[-60px] right-[-60px] size-[220px]", tone: "orange" },
];

/**
 * The pre-assessment notice (prototype test-guide): first/second-session
 * head, the five rules, the scoring-basis card, and the ready CTA into the
 * flow. The first/second distinction and the step count ride the live
 * session; the face degrades to the first-assessment defaults until the
 * read lands.
 */
export function TestGuidePage() {
  const navigate = useNavigate();
  const sessionQuery = useConnectQuery(getMatchSession, {});
  const session = sessionQuery.data?.session ?? null;
  const steps = session?.steps ?? [];
  const first = (session?.firstSession as boolean | undefined) ?? true;
  const count = steps.length > 0 ? steps.length : 4;

  // The ready CTA enters the first undone engine-backed step (prototype:
  // the first not-done step's intro). With nothing left it falls back to
  // the match guide — the report face the prototype routes to there is a
  // later slice.
  const done = session?.done ?? [];
  const start = steps.find(
    (s) => s.testCode != null && !done.includes(s.testCode) && getGameModule(s.testCode) !== null,
  );

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
    <AmbientLayout orbs={guideOrbs}>
      <div className="flex w-full flex-col gap-3 pb-6">
        <FlowHeader onBack={() => void history.back()} title={m.match_guide_title()} />

        <div className="border-line bg-surface-card shadow-card rounded-card border p-4 backdrop-blur-md">
          <div className="flex items-center gap-2.5">
            <div className="bg-cta-practice shadow-inset-ring rounded-control flex h-11 w-11 shrink-0 items-center justify-center">
              <TriangleAlert aria-hidden className="text-white" size={20} />
            </div>
            <div>
              <div className="text-ink text-[14.5px] font-extrabold">
                {first ? m.match_guide_first() : m.match_guide_second()}
              </div>
              {/* The retake count wears the prototype's bold; the spaces
                  live in the lead/tail messages — zh and en space
                  differently around the number. */}
              <div className="text-ink-tertiary mt-0.5 text-[11px]">
                {m.match_guide_sub_lead()}
                <span className="font-bold">
                  {first ? m.match_guide_sub_first() : m.match_guide_sub_second()}
                </span>
                {m.match_guide_sub_tail({ n: count })}
              </div>
            </div>
          </div>
          <div className="mt-2">
            {rules.map(({ Icon, key, title, desc }) => (
              <div
                className="flex gap-[11px] border-b border-[rgba(55,65,81,0.45)] py-[9px] last:border-b-0"
                key={key}
              >
                <Icon aria-hidden className="text-ink-secondary mt-0.5 shrink-0" size={17} />
                <div>
                  <div className="text-ink text-[12.5px] font-bold">{title}</div>
                  <div className="text-ink-tertiary mt-0.5 text-[10.5px] leading-[1.55]">
                    {desc}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="border-line bg-surface-content text-ink-muted rounded-control border px-3.5 py-[13px] text-[11.5px] leading-[1.85]">
          <span className="font-bold text-[#e5e7eb]">{m.match_guide_score_t()}</span>
          {m.match_guide_score_lead()}
          <span className="font-bold">{m.match_guide_score_best()}</span>
          {m.match_guide_score_mid()}
          <span className="font-bold">{m.match_guide_score_cross()}</span>
          {m.match_guide_score_tail()}
        </div>

        <button
          className="bg-cta-primary rounded-control w-full px-4 py-[14px] text-[15px] font-bold text-white active:opacity-80"
          onClick={() => {
            if (start?.testCode != null) {
              void navigate({ to: "/test/$code", params: { code: start.testCode } });
            } else {
              void navigate({ to: "/match" });
            }
          }}
          type="button"
        >
          {m.match_guide_ready()}
        </button>
        <div className="text-ink-tertiary text-center text-[11px]">{m.match_guide_keep()}</div>
      </div>
    </AmbientLayout>
  );
}
