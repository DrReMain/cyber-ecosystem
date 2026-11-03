import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import {
  getAbilityDigest,
  getCareerReport,
  getCognitiveReport,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/report-ReportService_connectquery";
import { useNavigate } from "@tanstack/react-router";
import { toast } from "sonner";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { StateScreen } from "#/features/app/states/state-screen";
import { FlowHeader } from "#/features/modules/match/ui/flow-header";
import { m } from "#/paraglide/messages";
import { wireCareerOf } from "./career";
import { wireCognitiveOf } from "./cognitive";
import { wireDigestOf } from "./digest";

// The aggregate reports' front door — a composition face, no RPC of its
// own: the digest answers coverage, the two access verdicts answer the
// cards' own gates, and every card routes into its page where the gate
// itself renders (a badge states the state; the page carries the door).
export function CenterPage() {
  const navigate = useNavigate();
  const digestQuery = useConnectQuery(getAbilityDigest, {}, { throwOnError: false });
  const cogQuery = useConnectQuery(getCognitiveReport, {}, { throwOnError: false });
  const careerQuery = useConnectQuery(getCareerReport, {}, { throwOnError: false });

  if (digestQuery.isPending || cogQuery.isPending || careerQuery.isPending) {
    return (
      <AmbientLayout>
        <StateScreen variant="loading" />
      </AmbientLayout>
    );
  }
  if (digestQuery.isError || cogQuery.isError || careerQuery.isError) {
    return (
      <AmbientLayout>
        <StateScreen
          action={
            <button
              className="bg-cta-primary rounded-control cursor-pointer px-5 py-2.5 text-sm font-semibold text-white active:opacity-80"
              onClick={() => {
                void digestQuery.refetch();
                void cogQuery.refetch();
                void careerQuery.refetch();
              }}
              type="button"
            >
              {m.common_retry()}
            </button>
          }
          description={m.report_error_desc()}
          title={m.report_error_title()}
          variant="empty"
        />
      </AmbientLayout>
    );
  }

  const digest = wireDigestOf(digestQuery.data);
  const cog = wireCognitiveOf(cogQuery.data);
  const career = wireCareerOf(careerQuery.data);
  const covered = digest.abilities.length;
  const latestAnchor = digest.sessions.find((s) => s.anchorRecordId !== undefined)?.anchorRecordId;

  const goDigest = () => {
    if (latestAnchor !== undefined) {
      void navigate({ to: "/report/$recordId", params: { recordId: latestAnchor } });
    } else {
      toast.info(m.report_center_digest_empty());
    }
  };

  const cards = [
    {
      desc: m.report_center_card_digest_desc(),
      icon: "📊",
      state: covered > 0 ? m.report_center_status_view() : m.report_center_status_locked(),
      title: m.report_center_card_digest_title(),
      onOpen: goDigest,
    },
    {
      // The covered banner's own law: at full coverage the copy switches
      // to the complete-report sentence instead of the fill-it nudge
      // (prototype report-center, both the banner and this card's line).
      desc:
        covered >= 6
          ? m.report_center_cognitive_full({ covered })
          : m.report_center_card_cognitive_desc({ covered }),
      icon: "🧠",
      state: cog.access.allowed ? m.report_center_status_view() : m.report_center_status_locked(),
      title: m.report_center_card_cognitive_title(),
      onOpen: () => void navigate({ to: "/report/cognitive" }),
    },
    {
      desc: m.report_center_card_career_desc(),
      icon: "🎯",
      state: career.access.allowed
        ? m.report_center_status_open()
        : m.report_center_status_closed(),
      title: m.report_center_card_career_title(),
      onOpen: () => void navigate({ to: "/report/career" }),
    },
    {
      desc: m.report_center_card_trend_desc(),
      icon: "📈",
      state:
        digest.sessionsTotal > 0
          ? m.report_center_status_view()
          : m.report_center_status_untested(),
      title: m.report_center_card_trend_title(),
      onOpen: () => void navigate({ to: "/report/trend" }),
    },
  ];

  return (
    <AmbientLayout>
      <div className="flex w-full flex-col gap-3 pb-6">
        <FlowHeader onBack={() => void navigate({ to: "/" })} title={m.report_center_title()} />

        <div className="border-line bg-surface-card rounded-card border px-4 py-3 backdrop-blur-md">
          <div className="text-ink-secondary text-[11.5px] leading-relaxed">
            {covered >= 6
              ? m.report_center_covered_full({ covered, total: 6 })
              : m.report_center_covered({ covered, total: 6 })}
          </div>
          <div className="bg-line mt-2 h-1.5 w-full overflow-hidden rounded-full">
            <div
              className="bg-identity h-full rounded-full"
              style={{ width: `${(covered / 6) * 100}%` }}
            />
          </div>
        </div>

        {cards.map((card) => (
          <button
            className="border-line bg-surface-card shadow-card rounded-card flex w-full items-center gap-3 border px-4 py-3.5 text-start active:opacity-60"
            key={card.title}
            onClick={card.onOpen}
            type="button"
          >
            <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-[rgba(37,99,235,0.16)] text-[16px]">
              {card.icon}
            </div>
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2">
                <span className="text-ink text-[13.5px] font-bold">{card.title}</span>
                <span className="border-line text-ink-tertiary rounded-full border px-1.5 py-0.5 text-[9.5px] font-bold">
                  {card.state}
                </span>
              </div>
              <div className="text-ink-tertiary mt-0.5 text-[11px] leading-relaxed">
                {card.desc}
              </div>
            </div>
            <span className="text-ink-tertiary shrink-0 text-[11px]">›</span>
          </button>
        ))}
      </div>
    </AmbientLayout>
  );
}
