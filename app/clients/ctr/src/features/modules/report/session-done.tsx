import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import { getSessionReport } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/report-ReportService_connectquery";
import { useNavigate, useParams } from "@tanstack/react-router";
import { Check, House } from "lucide-react";
import { Button } from "react-aria-components";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { StateScreen } from "#/features/app/states/state-screen";
import { abilityWord } from "#/features/modules/match/vocabulary";
import { TierChip } from "#/features/modules/test/ui/tier-chip";
import { m } from "#/paraglide/messages";

// The transport flattens the WKT wrappers on this face's lines: the
// optional numbers arrive as plain scalars-or-undefined, narrowed once.
interface WireAbility {
  ability: string;
  bestScore: number;
  bestGrade: string | undefined;
  percentile: number | undefined;
}

/** The session-done page (prototype test-done): the closing ceremony a
 * finished lineup earns — the check hero with its N/N receipt, the
 * strongest-ability gold card, the ability grid, then the report entry.
 * A session with no scored ability keeps the page honest: the empty
 * board states it, never invents a winner. */
export function SessionDonePage() {
  const { sessionId } = useParams({ from: "/report/session/$sessionId" });
  const navigate = useNavigate();
  const reportQuery = useConnectQuery(getSessionReport, { sessionId });

  if (reportQuery.isPending) {
    return (
      <AmbientLayout>
        <StateScreen variant="loading" />
      </AmbientLayout>
    );
  }
  if (reportQuery.isError) {
    return (
      <AmbientLayout>
        <StateScreen
          action={
            <button
              className="bg-cta-primary rounded-control cursor-pointer px-5 py-2.5 text-sm font-semibold text-white active:opacity-80"
              onClick={() => void reportQuery.refetch()}
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

  const data = reportQuery.data;
  if (data === undefined) {
    return null;
  }
  // The transport flattens the wrappers: the anchor and composite ride
  // as plain scalars-or-undefined.
  const anchor = (data as { anchorRecordId?: string }).anchorRecordId;
  const tested = data.testedDimensions ?? 0;
  const total = data.totalDimensions ?? 0;

  // The grid rows ride the lineup's own order; the gold card's winner is
  // the best score among them.
  const abilities = (data.abilities ?? []) as WireAbility[];
  const top = abilities.reduce<WireAbility | undefined>(
    (best, a) => (best === undefined || a.bestScore > best.bestScore ? a : best),
    undefined,
  );

  return (
    <AmbientLayout>
      <div className="flex w-full flex-col gap-3 pb-6">
        {/* ① The check hero (prototype 27c): the ceremony's own frame. */}
        <div className="relative flex w-full flex-col items-center pt-4 pb-1 text-center">
          <div
            aria-hidden
            className="bg-ambient-cyan pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_50%_0%,rgba(6,182,212,0.3),transparent_65%)]"
          />
          <div className="relative flex size-[88px] items-center justify-center rounded-full bg-[radial-gradient(circle,rgba(6,182,212,0.35),transparent_70%)]">
            <Check aria-hidden className="text-[#22D3EE]" size={40} strokeWidth={2.5} />
          </div>
          <div className="text-ink relative mt-2.5 text-2xl font-extrabold">
            {m.session_done_hero()}
          </div>
          <div className="text-ink-tertiary relative mt-1 text-[12px]">
            {m.session_done_line({ done: tested, total })}
          </div>
        </div>

        {top !== undefined ? (
          <>
            {/* ② The strongest-ability gold card. */}
            <div className="border-line bg-surface-card shadow-card rounded-card relative w-full overflow-hidden border border-[rgba(234,179,8,0.45)] px-4 py-5 backdrop-blur-md">
              <div
                aria-hidden
                className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_50%_0%,rgba(234,179,8,0.16),transparent_70%)]"
              />
              <div className="text-ink-tertiary relative text-[12px]">
                {m.session_done_top_label()}
              </div>
              <div className="text-grade-s relative mt-2 text-3xl font-black tracking-tight">
                {m.session_done_top_score({
                  ability: abilityWord(top.ability).cn,
                  score: Math.round(top.bestScore * 10) / 10,
                })}
              </div>
              <div className="relative mt-3 flex flex-wrap justify-center gap-2">
                <TierChip grade={top.bestGrade} />
                {top.percentile !== undefined ? (
                  <span className="rounded-full bg-[rgba(34,197,94,0.15)] px-2.5 py-0.75 text-[11px] font-semibold text-[#6EE7B7]">
                    {m.session_done_top_beat({ pct: Math.round(top.percentile) })}
                  </span>
                ) : null}
              </div>
            </div>

            {/* ③ The ability grid (prototype 2×2; one column when the
                lineup says fewer). */}
            <div className="grid w-full grid-cols-2 gap-2.5">
              {abilities.map((a) => (
                <div
                  className="border-line bg-surface-card shadow-card rounded-card border p-3.5 text-start backdrop-blur-md"
                  key={a.ability}
                >
                  <div className="flex items-center gap-2">
                    <div
                      aria-hidden
                      className={`${abilityWord(a.ability).gradient} shadow-inset-ring flex h-[34px] w-[34px] shrink-0 items-center justify-center rounded-[10px] text-[13px] font-bold text-white`}
                    >
                      {abilityWord(a.ability).cn.slice(0, 2)}
                    </div>
                    <span className="text-ink text-[12.5px] font-semibold">
                      {abilityWord(a.ability).cn}
                    </span>
                  </div>
                  <div className="mt-2 flex items-center justify-between">
                    <span className="text-ink text-xl font-extrabold tabular-nums">
                      {Math.round(a.bestScore * 10) / 10}
                    </span>
                    <TierChip grade={a.bestGrade} />
                  </div>
                </div>
              ))}
            </div>
          </>
        ) : (
          /* The honest empty board: no scored ability, no invented winner. */
          <div className="border-line bg-surface-card shadow-card rounded-card border px-4 py-5 text-center backdrop-blur-md">
            <div className="text-ink text-sm font-extrabold">{m.session_done_empty_title()}</div>
            <div className="text-ink-tertiary mt-1.5 text-[11.5px] leading-[1.7]">
              {m.session_done_empty_desc()}
            </div>
          </div>
        )}

        {/* ④ The closing CTAs: the report entry when the session scored,
            home always. */}
        <div className="flex flex-col gap-3">
          {anchor !== undefined ? (
            <Button
              className="bg-cta-primary shadow-cta-primary rounded-control w-full cursor-pointer px-4 py-3.5 text-[15px] font-bold text-white active:opacity-80"
              onPress={() =>
                void navigate({ to: "/report/$recordId", params: { recordId: anchor } })
              }
            >
              {m.test_result_report()}
            </Button>
          ) : null}
          <Button
            className="border-line text-ink-secondary rounded-control flex w-full cursor-pointer items-center justify-center gap-2 border px-4 py-3 text-[14px] font-bold active:opacity-80"
            onPress={() => void navigate({ to: "/" })}
          >
            <House aria-hidden size={16} />
            {m.session_done_cta_home()}
          </Button>
        </div>
      </div>
    </AmbientLayout>
  );
}
