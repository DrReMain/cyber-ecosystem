import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import { peekMatchSession } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/match-MatchService_connectquery";
import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { Bell } from "lucide-react";
import { toast } from "sonner";
import { SoonCard } from "#/components/soon-card";
import type { AmbientOrb } from "#/features/app/ambient/ambient-backdrop";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { TabBar } from "#/features/app/tabbar";
import { sessionQuery } from "#/features/auth/session.fn";
import { m } from "#/paraglide/messages";
import { CtaCard } from "./ui/cta-card";
import { ResumeCard } from "./ui/resume-card";

/* The home page's orb pair (prototype home). */
const homeOrbs: readonly AmbientOrb[] = [
  { className: "top-[-60px] left-[-70px] size-[240px]", tone: "blue" },
  { className: "top-[280px] right-[-70px] size-[200px]", tone: "purple" },
];

export function HomePage() {
  const navigate = useNavigate();
  const session = useQuery({ ...sessionQuery, throwOnError: false });
  // One session truth: the probe carries the profile (it proves the
  // credential with getProfile), so the greet, the nickname, and the games
  // gate all read it — profile writes invalidate it, and the dehydrated
  // SSR verdict delivers the final state on first paint.
  const authedSession = session.data?.status === "authed" ? session.data : undefined;
  const authed = authedSession !== undefined;
  const profile = authedSession?.profile;
  const hasGames = (profile?.games ?? []).length > 0;
  // The resume card's truth (prototype home): the read-only peek reports
  // the active in-window session — never mints one, so a home visit can't
  // hijack a later explicit genre choice. Guests skip the call entirely.
  const peekQuery = useConnectQuery(peekMatchSession, {}, { enabled: authed });
  const resume = peekQuery.data?.session ?? null;
  const resumeSteps = (resume?.steps ?? []).length;
  const resumeDone = (resume?.done ?? []).length;
  const showResume = resume !== null && resumeSteps > 0 && resumeDone < resumeSteps;

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
              {authed ? m.home_greet_logged() : m.home_greet_guest()}
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

        <SoonCard description={m.home_soon_desc()} title={m.home_soon_title()} />
      </div>
    </AmbientLayout>
  );
}
