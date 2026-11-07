import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import { getCareerReport } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/report-ReportService_connectquery";
import { useNavigate } from "@tanstack/react-router";
import { Target } from "lucide-react";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { StateScreen } from "#/features/app/states/state-screen";
import { FlowHeader } from "#/features/modules/match/ui/flow-header";
import { m } from "#/paraglide/messages";
import { wireCareerOf } from "./career";
import { CareerAreaCard } from "./ui/career-area-card";
import { CareerImprovementCard } from "./ui/career-improvement-card";

// The career aggregate's page. The prototype carries only its closed
// state, so the closed face is the prototype verbatim while the open
// face renders what the seeded library rows carry — the areas, the gap
// board, the library's own facts — and invents nothing beyond them.
export function CareerPage() {
  const navigate = useNavigate();
  const careerQuery = useConnectQuery(getCareerReport, {}, { throwOnError: false });

  if (careerQuery.isPending) {
    return (
      <AmbientLayout>
        <StateScreen variant="loading" />
      </AmbientLayout>
    );
  }
  if (careerQuery.isError) {
    return (
      <AmbientLayout>
        <StateScreen
          action={
            <button
              className="bg-cta-primary rounded-control cursor-pointer px-5 py-2.5 text-sm font-semibold text-white active:opacity-80"
              onClick={() => void careerQuery.refetch()}
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

  const r = wireCareerOf(careerQuery.data);

  return (
    <AmbientLayout>
      <div className="flex w-full flex-col gap-3 pb-6">
        <FlowHeader
          onBack={() => void navigate({ to: "/report/center" })}
          title={m.report_career_title()}
        />

        {!r.access.allowed ? (
          <div className="border-line bg-surface-card shadow-card rounded-card border px-4 py-6 text-center backdrop-blur-md">
            <Target aria-hidden className="text-danger-text" size={26} />
            <div className="text-ink mt-2 text-[14px] font-bold">
              {m.report_career_closed_title()}
            </div>
            <div className="text-ink-tertiary mt-1.5 text-[12px] leading-relaxed">
              {m.report_career_closed_desc()}
            </div>
            <div className="text-ink-tertiary text-[12px] leading-relaxed">
              {m.report_career_closed_future()}
            </div>
            <div className="border-line rounded-control bg-surface-card mt-4 border px-3 py-2 text-start">
              <div className="text-danger-text text-[11px] font-bold">
                {m.report_career_note_title()}
              </div>
              <div className="text-ink-tertiary mt-1 text-[10.5px] leading-relaxed">
                {m.report_career_note()}
              </div>
            </div>
            <button
              className="bg-cta-primary rounded-control mt-4 px-5 py-2.5 text-sm font-semibold text-white active:opacity-80"
              onClick={() => void navigate({ to: "/report/center" })}
              type="button"
            >
              {m.report_career_back()}
            </button>
          </div>
        ) : (
          <>
            {r.recommendations.length === 0 && r.improvements.length === 0 ? (
              <StateScreen
                description={m.report_career_empty_desc()}
                title={m.report_career_empty_title()}
                variant="empty"
              />
            ) : null}
            {r.recommendations.map((rec) => (
              <CareerAreaCard key={rec.careerArea} rec={rec} />
            ))}
            {r.improvements.length > 0 ? (
              <>
                <div className="text-ink px-0.5 pt-1 text-[13px] font-bold">
                  {m.report_career_improvement_title()}
                </div>
                {r.improvements.map((imp) => (
                  <CareerImprovementCard imp={imp} key={imp.dimension} />
                ))}
              </>
            ) : null}
          </>
        )}
      </div>
    </AmbientLayout>
  );
}
