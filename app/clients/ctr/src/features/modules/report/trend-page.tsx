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
            <div className="flex items-center justify-end px-0.5">
              <span
                className={`rounded-full border px-2 py-0.5 text-[10px] font-bold ${
                  r.advancedAccess.allowed
                    ? "border-[#22D3EE]/40 text-[#22D3EE]"
                    : "border-line text-ink-tertiary"
                }`}
              >
                {r.advancedAccess.allowed
                  ? m.report_trend_advanced_on()
                  : m.report_trend_advanced_off()}
              </span>
            </div>
            {/* The summary banner rides the prototype's own law: the count
                line states itself, and a history mixing the two genres
                earns the not-directly-comparable warning (decision ②) —
                a single-genre history carries no banner. */}
            <div className="border-line bg-surface-card rounded-card border px-3.5 py-3 backdrop-blur-md">
              <div className="text-ink-secondary text-[11.5px] leading-[1.85]">
                {m.report_trend_sessions({ n: r.sessionsTotal, k: r.abilities.length })}
              </div>
              {mixedGenres ? (
                <div className="mt-1 text-[11.5px] leading-[1.85] text-[#FCD34D]">
                  {m.report_trend_mixed()}
                </div>
              ) : null}
            </div>
            <div className="border-line bg-surface-card shadow-card rounded-card border px-4 py-3.5 backdrop-blur-md">
              <div className="mb-1 flex items-center justify-between">
                <span className="text-ink text-[13px] font-bold">
                  {m.report_trend_base_title()}
                </span>
                <span className="text-status-ok rounded-full bg-[rgba(34,197,94,0.15)] px-2 py-0.5 text-[10px] font-bold">
                  {m.report_trend_base_tag()}
                </span>
              </div>
              {r.abilities.map((s) => (
                <TrendChart key={s.ability} series={s} />
              ))}
            </div>
          </>
        )}
      </div>
    </AmbientLayout>
  );
}
