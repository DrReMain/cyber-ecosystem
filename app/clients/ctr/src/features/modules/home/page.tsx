import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import { peekMatchSession } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/match-MatchService_connectquery";
import {
  getAbilityDigest,
  getSessionReport,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/report-ReportService_connectquery";
import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { Bell, ChevronRight, FileText } from "lucide-react";
import { toast } from "sonner";
import type { AmbientOrb } from "#/features/app/ambient/ambient-backdrop";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { TabBar } from "#/features/app/tabbar";
import { sessionQuery } from "#/features/auth/session.fn";
import { latestScoredOf, type WireDigest, wireDateOf } from "#/features/modules/report/digest";
import { m } from "#/paraglide/messages";
import {
  authedDigestOf,
  type RecentWire,
  recentGenreWordOf,
  recentItemsOf,
  recentTierLetterOf,
  recentWireOf,
} from "./recent";
import { CtaCard } from "./ui/cta-card";
import { RecentCard } from "./ui/recent-card";
import { ResumeCard } from "./ui/resume-card";
import { StatsCard } from "./ui/stats-card";

/* The home page's orb pair (prototype home). */
const homeOrbs: readonly AmbientOrb[] = [
  { className: "top-[-60px] left-[-70px] size-[240px]", tone: "blue" },
  { className: "top-[280px] right-[-70px] size-[200px]", tone: "purple" },
];

// The read-back cards' own gates: each renders only when its read
// answered — a card's absence is the honest no-data face, so the page
// body stays a flat section list.
function RecentSection({ recent }: Readonly<{ recent: RecentWire | null }>) {
  if (recent === null || recent.anchorRecordId === undefined) {
    return null;
  }
  return (
    <RecentCard
      anchorId={recent.anchorRecordId}
      composite={recent.compositePercentile}
      count={(recent.abilities ?? []).length}
      date={wireDateOf(recent.startedAt)}
      genreWord={recentGenreWordOf(recent)}
      items={recentItemsOf(recent)}
      tierLetter={recentTierLetterOf(recent)}
    />
  );
}

function StatsSection({ digest }: Readonly<{ digest: WireDigest | null }>) {
  if (digest === null) {
    return null;
  }
  return (
    <StatsCard
      covered={digest.abilities.length}
      reactionBest={digest.abilities.find((a) => a.ability === "reactionSpeed")?.bestScore}
      total={digest.sessionsTotal}
    />
  );
}

export function HomePage() {
  const navigate = useNavigate();
  const session = useQuery({ ...sessionQuery, throwOnError: false });
  // One session truth: the probe carries the profile and the membership
  // liveness (it proves the credential with getProfile), so the greet, the
  // nickname, and the games gate all read it — profile writes invalidate
  // it, and the dehydrated SSR verdict delivers the final state on first
  // paint.
  const authedSession = session.data?.status === "authed" ? session.data : undefined;
  const authed = authedSession !== undefined;
  const profile = authedSession?.profile;
  // The server already collapsed liveness to presence (nil = none standing
  // or lapsed-only), so the greet renders presence and never re-derives
  // expiry.
  const memberUntil = authedSession?.memberUntil ?? null;
  const hasGames = (profile?.games ?? []).length > 0;
  // The resume card's truth (prototype home): the read-only peek reports
  // the active in-window session — never mints one, so a home visit can't
  // hijack a later explicit genre choice. Guests skip the call entirely.
  const peekQuery = useConnectQuery(peekMatchSession, {}, { enabled: authed });
  const resume = peekQuery.data?.session ?? null;
  const resumeSteps = (resume?.steps ?? []).length;
  const resumeDone = (resume?.done ?? []).length;
  const showResume = resume !== null && resumeSteps > 0 && resumeDone < resumeSteps;

  // The read-back overview line (prototype home's 最近成绩 + 数据统计):
  // the digest names the latest scored session and carries the counts,
  // that session's report answers the verdict face. Guests skip the
  // reads; a failed read leaves the cards absent rather than blocking
  // the start flow beside them.
  const digestQuery = useConnectQuery(
    getAbilityDigest,
    {},
    {
      enabled: authed,
      throwOnError: false,
    },
  );
  const digest = authedDigestOf(authed, digestQuery);
  const latestScored = digest !== null ? latestScoredOf(digest) : undefined;
  const recentQuery = useConnectQuery(
    getSessionReport,
    { sessionId: latestScored?.sessionId ?? "" },
    { enabled: latestScored !== undefined, throwOnError: false },
  );
  const recent = recentWireOf(recentQuery.data);

  function onStart() {
    if (!authed) {
      void navigate({ to: "/login", search: { next: "/" } });
      return;
    }
    if (!hasGames) {
      void navigate({ to: "/profile/setup/games" });
      return;
    }
    void navigate({ to: "/match" });
  }

  return (
    <AmbientLayout chrome={<TabBar active="home" />} orbs={homeOrbs}>
      <div className="flex w-full flex-col gap-3 pb-20 md:pb-0">
        <div className="flex items-start justify-between">
          <div>
            <div className="text-ink-tertiary text-[11px]">
              {authed
                ? memberUntil != null
                  ? m.home_greet_member()
                  : m.home_greet_logged()
                : m.home_greet_guest()}
            </div>
            <h1 className="text-ink mt-0.5 text-xl font-black">
              {authed && profile?.nickname != null ? profile.nickname : m.home_guest_name()}
            </h1>
          </div>
          <button
            aria-label={m.profile_msg()}
            className="bg-surface-input text-ink-secondary flex h-10 w-10 items-center justify-center rounded-full active:opacity-60"
            onClick={() => toast.info(m.common_soon())}
            type="button"
          >
            <Bell aria-hidden size={20} />
          </button>
        </div>

        {showResume && resume !== null && (
          <ResumeCard
            createdAt={resume.createdAt}
            done={resumeDone}
            onOpen={() => void navigate({ to: "/match" })}
            total={resumeSteps}
          />
        )}

        <CtaCard authed={authed} hasGames={hasGames} onStart={onStart} />

        <RecentSection recent={recent} />

        <StatsSection digest={digest} />

        {/* The report-center entry (prototype home's 报告中心): the four
            aggregate faces now live, so the row routes into the center
            that composes them. */}
        <button
          className="border-line-accent-blue bg-identity tap-zone rounded-card flex w-full items-center gap-3 border px-4 py-3.5 text-start active:opacity-60"
          onClick={() => void navigate({ to: "/report/center" })}
          type="button"
        >
          <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-[rgba(37,99,235,0.16)]">
            <FileText aria-hidden className="text-accent-info" size={16} />
          </div>
          <div className="min-w-0 flex-1">
            <div className="text-ink text-[13.5px] font-bold">{m.home_center_title()}</div>
            <div className="text-ink-tertiary mt-0.5 text-[11px]">{m.home_center_desc()}</div>
          </div>
          <ChevronRight aria-hidden className="text-ink-tertiary shrink-0" size={16} />
        </button>
      </div>
    </AmbientLayout>
  );
}
