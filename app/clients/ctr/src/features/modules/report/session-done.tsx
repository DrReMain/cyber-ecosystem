import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import { getSessionReport } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/report-ReportService_connectquery";
import { useNavigate, useParams } from "@tanstack/react-router";
import { Check } from "lucide-react";
import { Button } from "react-aria-components";
import { toast } from "sonner";
import { loginOrbs } from "#/features/app/ambient/ambient-backdrop";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { StateScreen } from "#/features/app/states/state-screen";
import { FlowHeader } from "#/features/modules/match/ui/flow-header";
import { abilityWord } from "#/features/modules/match/vocabulary";
import { TierChip } from "#/features/modules/test/ui/tier-chip";
import { m } from "#/paraglide/messages";
import { elapsedSecondsOf, type WireAttempt } from "./attempts";

// The transport flattens the WKT wrappers on this face's lines: the
// optional numbers arrive as plain scalars-or-undefined, narrowed once.
interface WireAbility {
  ability: string;
  bestScore: number;
  bestGrade: string | undefined;
  percentile: number | undefined;
}

const timeTextOf = (seconds: number | undefined): string | undefined =>
  seconds === undefined
    ? undefined
    : seconds < 60
      ? m.report_session_time_sec({ sec: seconds })
      : m.report_session_time_min_sec({
          min: Math.floor(seconds / 60),
          sec: seconds % 60,
        });

/** The session-done page (prototype test-done): the closing ceremony a
 * finished lineup earns — the check hero with its N/N receipt, the
 * strongest-ability gold card, the ability grid, then the share/profile
 * CTAs. A session with no scored ability keeps the page honest: the
 * empty board states it, never invents a winner. */
export function SessionDonePage() {
  const { sessionId } = useParams({ from: "/report/session/$sessionId" });
  const navigate = useNavigate();
  // The retry face below owns a failed report read; the global throw
  // default would hand it to the route boundary instead.
  const reportQuery = useConnectQuery(getSessionReport, { sessionId }, { throwOnError: false });

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
  const total = data.totalDimensions ?? 0;

  // The receipt counts what was PLAYED, not what a calibration curve has
  // graded (the match guide's own count law — one session, one truth).
  // attemptDetails is the full attempt list, pending rows included, in
  // submit order: distinct abilities = played.
  const attempts = (data.attemptDetails ?? []) as WireAttempt[];
  const played = new Set(attempts.map((a) => a.ability)).size;
  const timeText = timeTextOf(elapsedSecondsOf(attempts));

  // The grid rows ride the lineup's own order; the gold card's winner is
  // the best score among them. A row exists only for abilities this
  // lineup actually scored, so an omitted best (proto3 implicit presence
  // drops the zero) narrows to the honest zero — never NaN.
  const abilities = ((data.abilities ?? []) as WireAbility[]).map((a) => ({
    ...a,
    bestScore: a.bestScore ?? 0,
  }));
  // One strongest criterion page-wide (prototype: highest percentile
  // ranks both the grid and the gold card); with no cohort answer
  // anywhere, the best raw score holds the card.
  const ranked = [...abilities].sort((a, b) => (b.percentile ?? -1) - (a.percentile ?? -1));
  const leader = ranked[0];
  const top =
    leader !== undefined && leader.percentile !== undefined
      ? leader
      : abilities.reduce<WireAbility | undefined>(
          (best, a) => (best === undefined || a.bestScore > best.bestScore ? a : best),
          undefined,
        );

  return (
    // The prototype's test-done trio (test-done.html:3-5) is exactly the
    // shared login trio — passed explicitly to art-direct the page.
    <AmbientLayout orbs={loginOrbs}>
      <div className="flex w-full flex-col gap-3 pb-6">
        {/* The page header (prototype: back fab + title + the N/N done
            tag, design sheet 27c). */}
        <FlowHeader
          onBack={() => void navigate({ to: "/" })}
          title={m.report_session_title()}
          trailing={
            <span className="rounded-full bg-[rgba(34,197,94,0.15)] px-2.5 py-0.75 text-[11px] font-semibold text-[#6EE7B7]">
              {m.report_session_done_tag({ done: played, total })}
            </span>
          }
        />

        {/* ① The check hero (prototype 27c): the ceremony's own frame. */}
        <div className="relative flex w-full flex-col items-center pt-4 pb-1 text-center">
          <div
            aria-hidden
            className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_50%_0%,rgba(6,182,212,0.3),transparent_65%)]"
          />
          <div className="relative flex size-[88px] items-center justify-center rounded-full bg-[radial-gradient(circle,rgba(6,182,212,0.35),transparent_70%)]">
            <Check aria-hidden className="text-[#22D3EE]" size={40} strokeWidth={2.5} />
          </div>
          <div className="text-ink relative mt-2.5 text-2xl font-extrabold">
            {m.report_session_hero()}
          </div>
          <div className="text-ink-tertiary relative mt-1 text-[12px]">
            {m.report_session_line({ done: played, total })}
            {timeText !== undefined ? ` · ${timeText}` : ""}
          </div>
        </div>

        {top !== undefined ? (
          <>
            {/* ② The strongest-ability gold card — the prototype's pad is
                text-centered throughout, this card included. */}
            <div className="border-line bg-surface-card shadow-card rounded-card relative w-full overflow-hidden border border-[rgba(234,179,8,0.45)] px-4 py-5 text-center backdrop-blur-md">
              <div
                aria-hidden
                className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_50%_0%,rgba(234,179,8,0.16),transparent_70%)]"
              />
              <div className="text-ink-tertiary relative text-[12px]">
                {m.report_session_top_label()}
              </div>
              <div className="text-grade-s relative mt-2 text-3xl font-black tracking-tight">
                {m.report_session_top_score({
                  ability: abilityWord(top.ability).cn,
                  score: Math.round(top.bestScore * 10) / 10,
                })}
              </div>
              <div className="relative mt-3 flex flex-wrap justify-center gap-2">
                <TierChip grade={top.bestGrade} />
                {top.percentile !== undefined ? (
                  <span className="rounded-full bg-[rgba(34,197,94,0.15)] px-2.5 py-0.75 text-[11px] font-semibold text-[#6EE7B7]">
                    {m.report_session_top_beat({ pct: Math.round(top.percentile) })}
                  </span>
                ) : null}
              </div>
            </div>

            {/* ③ The ability grid (prototype 2×2; one column when the
                lineup says fewer): percentile-ranked, strongest first,
                the leader crowned with the gold border; a pending row
                carries no percentile line. */}
            <div className="grid w-full grid-cols-2 gap-2.5">
              {ranked.map((a, i) => (
                <div
                  className="border-line bg-surface-card shadow-card rounded-card border p-3.5 text-start backdrop-blur-md"
                  key={a.ability}
                  style={
                    i === 0 && a.percentile !== undefined
                      ? { borderColor: "rgba(234,179,8,0.5)" }
                      : undefined
                  }
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
                      {m.report_session_card_score({ score: Math.round(a.bestScore * 10) / 10 })}
                    </span>
                    <TierChip grade={a.bestGrade} />
                  </div>
                  {a.percentile !== undefined ? (
                    <div className="text-ink-tertiary mt-1 text-[10px]">
                      {m.report_session_card_beat({ pct: Math.round(a.percentile) })}
                    </div>
                  ) : null}
                </div>
              ))}
            </div>
          </>
        ) : (
          /* The honest empty board: no scored ability, no invented winner. */
          <div className="border-line bg-surface-card shadow-card rounded-card border px-4 py-5 text-center backdrop-blur-md">
            <div className="text-ink text-sm font-extrabold">{m.report_session_empty_title()}</div>
            <div className="text-ink-tertiary mt-1.5 text-[11.5px] leading-[1.7]">
              {m.report_session_empty_desc()}
            </div>
          </div>
        )}

        {/* ④ The closing CTAs (prototype 27c): share leads, profile
            second — the report entry deliberately stays off this page
            (test-done.html:69; it lives in the report center/history).
            The invite face is MSG-2's to build, so until then the share
            button announces instead of linking — the G-202 gated
            placeholder precedent. */}
        <div className="flex flex-col gap-3">
          <Button
            className="bg-cta-success shadow-cta-success rounded-control w-full cursor-pointer px-4 py-3.5 text-[15px] font-bold text-white active:opacity-80"
            onPress={() => toast.info(m.common_soon())}
          >
            {m.report_session_cta_share()}
          </Button>
          <Button
            className="border-line text-ink-secondary rounded-control flex w-full cursor-pointer items-center justify-center border-[1.5px] px-4 py-3 text-[14px] font-bold active:opacity-80"
            onPress={() => void navigate({ to: "/profile" })}
          >
            {m.report_session_cta_profile()}
          </Button>
          <div className="text-ink-tertiary text-center text-[10px]">
            {m.report_session_share_note()}
          </div>
        </div>
      </div>
    </AmbientLayout>
  );
}
