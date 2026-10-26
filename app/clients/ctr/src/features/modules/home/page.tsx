import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { Bell } from "lucide-react";
import { toast } from "sonner";
import { peekMatchSession } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/match-MatchService_connectquery";
import { SoonCard } from "#/components/soon-card";
import type { AmbientOrb } from "#/features/app/ambient/ambient-backdrop";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { TabBar } from "#/features/app/tabbar";
import { sessionQuery } from "#/features/auth/session.fn";
import { m } from "#/paraglide/messages";

/* The home page's orb pair (prototype home). */
const homeOrbs: readonly AmbientOrb[] = [
  { className: "top-[-60px] left-[-70px] size-[min(62vw,240px)]", tone: "blue" },
  { className: "top-[280px] right-[-70px] size-[min(51vw,200px)]", tone: "purple" },
];

/* The resume card's desc vars: the progress counts plus the age face —
 * the prototype's inclusive day count (started today reads 已过 1 天) and
 * a local-date stamp; createdAt is already the proto Timestamp string. */
function resumeDescVars(createdAt: string | undefined, done: number, total: number) {
  const created = createdAt !== undefined ? new Date(createdAt).getTime() : Date.now();
  const days = Math.max(1, Math.floor((Date.now() - created) / 86_400_000) + 1);
  const d = new Date(created);
  const pad = (n: number) => String(n).padStart(2, "0");
  const date = `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
  return { done, total, date, days };
}

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
          <div className="border-line bg-surface-card shadow-card rounded-card flex items-center gap-3 border p-3.5">
            <div className="min-w-0 flex-1">
              <div className="text-ink text-[13.5px] font-extrabold">{m.home_resume_title()}</div>
              <div className="text-ink-tertiary mt-[3px] text-[11.5px] leading-[1.7]">
                {m.home_resume_desc(
                  resumeDescVars(resume.createdAt, resumeDone, resumeSteps),
                )}
              </div>
            </div>
            <button
              className="bg-cta-practice shadow-cta-practice rounded-control shrink-0 px-3.5 py-2 text-[12.5px] font-bold text-white active:opacity-80"
              onClick={() => void navigate({ to: "/match" })}
              type="button"
            >
              {m.home_resume_cta()}
            </button>
          </div>
        )}

        <div className="border-line bg-surface-card shadow-card rounded-card border px-4 py-[18px]">
          <div className="text-ink text-base font-extrabold">{m.home_cta_title()}</div>
          <p className="text-ink-tertiary mt-[5px] text-[11.5px] leading-[1.7]">
            {/* The desc follows the games state alone (prototype law): a
                guest has no preferences yet, so it reads the pick-games
                line — the login-first ordering rides on the button and
                hint, not the desc. */}
            {!hasGames ? m.home_cta_desc_need_games() : m.home_cta_desc_ready()}
          </p>
          <button
            className="bg-cta-primary shadow-cta-primary rounded-control mt-3.5 w-full px-4 py-[14px] text-[15px] font-bold text-white active:opacity-80"
            onClick={onStart}
            type="button"
          >
            {!authed ? m.home_cta_login() : !hasGames ? m.home_cta_pick() : m.home_cta_start()}
          </button>
          {!(authed && hasGames) && (
            <p className="text-ink-tertiary mt-[9px] text-center text-[11px]">
              {!authed ? m.home_hint_login() : m.home_hint_games()}
            </p>
          )}
        </div>

        <SoonCard description={m.home_soon_desc()} title={m.home_soon_title()} />
      </div>
    </AmbientLayout>
  );
}
