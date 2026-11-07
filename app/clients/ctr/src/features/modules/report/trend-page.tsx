import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import { getTrendReport } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/report-ReportService_connectquery";
import { useNavigate } from "@tanstack/react-router";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { StateScreen } from "#/features/app/states/state-screen";
import { FlowHeader } from "#/features/modules/match/ui/flow-header";
import { narrowGenre } from "#/features/modules/match/vocabulary";
import { m } from "#/paraglide/messages";
import { wireTrendOf } from "./trend";
import { TrendChart } from "./ui/trend-chart";

// The chronology page: one series per ability, each point the
// session-anchor's cohort answer with its genre on the marker. The base
// tier is free (measured members see it); the advanced verdict rides the
// same response, and an unmeasured member gets the prototype's own
// honest empty state rather than an empty chart.
export function TrendPage() {
  const navigate = useNavigate();
  const trendQuery = useConnectQuery(getTrendReport, {}, { throwOnError: false });

  if (trendQuery.isPending) {
    return (
      <AmbientLayout>
        <StateScreen variant="loading" />
      </AmbientLayout>
    );
  }
  if (trendQuery.isError) {
    return (
      <AmbientLayout>
        <StateScreen
          action={
            <button
              className="bg-cta-primary rounded-control cursor-pointer px-5 py-2.5 text-sm font-semibold text-white active:opacity-80"
              onClick={() => void trendQuery.refetch()}
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

  const r = wireTrendOf(trendQuery.data);
  const genres = new Set(r.abilities.flatMap((s) => s.points.map((p) => narrowGenre(p.genre))));
  const mixedGenres = genres.has("fps") && genres.has("moba");

  return (
    <AmbientLayout>
      <div className="flex w-full flex-col gap-3 pb-6">
        <FlowHeader
          onBack={() => void navigate({ to: "/report/center" })}
          title={m.report_trend_title()}
        />

        {r.sessionsTotal === 0 || r.abilities.length === 0 ? (
          <StateScreen
            action={
              <button
                className="bg-cta-primary rounded-control cursor-pointer px-5 py-2.5 text-sm font-semibold text-white active:opacity-80"
                onClick={() => void navigate({ to: "/match" })}
                type="button"
              >
                {m.report_trend_go_test()}
              </button>
            }
            description={m.report_trend_empty_desc()}
            title={m.report_trend_empty_title()}
            variant="empty"
          />
        ) : (
          <>
            {/* The summary banner rides the prototype's own law: the count
                line states itself, and a history mixing the two genres
                earns the not-directly-comparable warning (decision ②) —
                a single-genre history carries no banner. */}
            <div className="border-line bg-surface-card rounded-card border px-3.5 py-3 backdrop-blur-md">
              <div className="text-ink-secondary text-[11.5px] leading-[1.85]">
                {m.report_trend_sessions({ n: r.sessionsTotal, k: r.abilities.length })}
              </div>
              {mixedGenres ? (
                <div className="text-status-pending mt-1 text-[11.5px] leading-[1.85]">
                  {m.report_trend_mixed()}
                </div>
              ) : null}
            </div>
            <div className="border-line bg-surface-card shadow-card rounded-card border px-4 py-3.5 backdrop-blur-md">
              <div className="mb-1 flex items-center justify-between">
                <span className="text-ink text-[13px] font-bold">
                  {m.report_trend_base_title()}
                </span>
                <span className="text-status-ok bg-status-ok-bg rounded-full px-2 py-0.5 text-[10px] font-bold">
                  {m.report_trend_base_tag()}
                </span>
              </div>
              {r.abilities.map((s) => (
                <TrendChart key={s.ability} series={s} />
              ))}
            </div>
            {/* The advanced module (prototype's own two states): the
                member card lists what the tier buys; a locked visitor
                gets the one-line pitch, the membership CTA, and the
                peek at what membership includes. The CTA's pay leg is
                the PAY family's own slice; this face carries the
                presence and the copy. */}
            <div className="bg-surface-card shadow-card rounded-card border-ability-skill/50 border px-4 py-3.5">
              {r.advancedAccess.allowed ? (
                <>
                  <div className="mb-1 flex items-center justify-between">
                    <span className="text-ink text-[13px] font-bold">
                      {m.report_trend_advanced_title()}
                    </span>
                    <span className="text-status-soon rounded-full bg-[rgba(124,58,237,0.15)] px-2 py-0.5 text-[10px] font-bold">
                      {m.report_trend_advanced_member()}
                    </span>
                  </div>
                  <div className="text-ink-tertiary text-[12.5px] leading-[1.9]">
                    <div>· {m.report_trend_advanced_compare()}</div>
                    <div>· {m.report_trend_advanced_plan()}</div>
                    <div>· {m.report_trend_advanced_percentile()}</div>
                  </div>
                </>
              ) : (
                <>
                  <div className="flex items-center">
                    <div className="min-w-0 flex-1">
                      <div className="text-ink text-[13.5px] font-extrabold">
                        {m.report_trend_advanced_title()}
                      </div>
                      <div className="text-ink-tertiary mt-1 text-[11.5px] leading-[1.65]">
                        {m.report_trend_advanced_desc()}
                      </div>
                    </div>
                    <span className="text-status-soon rounded-full bg-[rgba(124,58,237,0.15)] px-2 py-0.5 text-[10px] font-bold">
                      {m.report_trend_advanced_member()}
                    </span>
                  </div>
                  <button
                    className="rounded-control mt-3.5 w-full bg-linear-to-r from-[#7C3AED] to-[#A855F7] px-4 py-3 text-[15px] font-bold text-white active:opacity-80"
                    onClick={() => void navigate({ to: "/pay", search: { from: "/report/trend" } })}
                    type="button"
                  >
                    {m.report_trend_advanced_cta()}
                  </button>
                  <button
                    className="rounded-control border-line text-ink-secondary mt-2 block w-full border px-4 py-2.5 text-[13px] font-bold active:opacity-70"
                    onClick={() => void navigate({ to: "/member" })}
                    type="button"
                  >
                    {m.report_trend_advanced_peek()}
                  </button>
                </>
              )}
            </div>
          </>
        )}
      </div>
    </AmbientLayout>
  );
}
