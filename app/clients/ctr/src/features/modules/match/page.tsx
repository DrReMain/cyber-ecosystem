import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import { getMatchSession } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/match-MatchService_connectquery";
import { listTests } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test-TestService_connectquery";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { Check, Crosshair, Hexagon } from "lucide-react";
import { useEffect } from "react";
import type { AmbientOrb } from "#/features/app/ambient/ambient-backdrop";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { StateScreen } from "#/features/app/states/state-screen";
import { getGameModule } from "#/features/modules/test/registry";
import { m } from "#/paraglide/messages";
import { reasonGenreUndecided } from "./session";
import { FlowHeader } from "./ui/flow-header";
import { abilityWord, type Genre, genreLabel, genreMeta, narrowGenre } from "./vocabulary";

/* The guide page's own orb pair (prototype match-guide). */
const guideOrbs: readonly AmbientOrb[] = [
  { className: "top-[-70px] left-[-50px] size-[min(62vw,240px)]", tone: "blue" },
  { className: "bottom-[-60px] right-[-70px] size-[min(54vw,210px)]", tone: "cyan" },
];

interface StepRowProps {
  readonly index: number;
  readonly ability: string;
  readonly name: string;
  readonly left: number;
  readonly ready: boolean;
  readonly done: boolean;
  readonly onEnter: () => void;
}

// One lineup row: the ability face, the drawn test's name, and the retake
// receipt (the "N/3" face, zero reads as spent). A step whose engine has
// not shipped yet renders dimmed with a coming-soon tag instead of the
// retake receipt — the flow's start button skips it, so the tag is what
// keeps the row honest. A done step wears the prototype's green border and
// check; a playable one is tappable into its intro (the prototype's step
// rows navigate; a done step's score read-face waits on the report slice).
function StepRow({ index, ability, name, left, ready, done, onEnter }: Readonly<StepRowProps>) {
  const word = abilityWord(ability);
  const playable = ready && left > 0;
  return (
    <button
      className={`${done ? "border-[rgba(34,197,94,0.55)]" : "border-line"} bg-surface-card shadow-card rounded-card flex w-full items-center gap-[11px] border px-[13px] py-[11px] text-start ${ready ? "" : "opacity-60"} ${playable ? "cursor-pointer active:opacity-80" : "cursor-default"}`}
      disabled={!playable}
      onClick={playable ? onEnter : undefined}
      type="button"
    >
      <div
        className={`${word.gradient} flex h-[38px] w-[38px] shrink-0 items-center justify-center rounded-[12px] text-[15px] font-bold text-white`}
      >
        {index + 1}
      </div>
      <div className="min-w-0 flex-1">
        <div className="text-ink text-sm font-extrabold">
          {word.cn}
          {done ? (
            <Check aria-hidden className="text-status-ok ml-1 inline h-3 w-3" strokeWidth={3} />
          ) : null}
        </div>
        <div className="text-ink-tertiary mt-0.5 text-[9.5px] tracking-[0.9px]">{word.en}</div>
      </div>
      <div className="shrink-0 text-right">
        <div className="text-ink-secondary text-[12px]">{name}</div>
        {ready ? (
          <div className="text-ink-tertiary mt-0.5 flex items-center justify-end gap-1 text-[10px]">
            <span>{m.match_step_time()}</span>
            <span aria-hidden>·</span>
            {left > 0 ? (
              <span className="text-status-ok font-bold">{m.match_step_left({ n: left })}</span>
            ) : (
              <span>{m.match_step_left_zero()}</span>
            )}
          </div>
        ) : (
          <div className="text-ink-tertiary mt-0.5 text-[10px]">{m.match_step_engine_soon()}</div>
        )}
      </div>
    </button>
  );
}

interface GenreHeaderCardProps {
  readonly genre: Genre;
  readonly switched: boolean;
}

// The genre banner: icon, bilingual name, and the origin tag (a genre
// param in the URL marks an explicitly chosen session).
function GenreHeaderCard({ genre, switched }: Readonly<GenreHeaderCardProps>) {
  const meta = genreMeta(genre);
  const Icon = genre === "moba" ? Hexagon : Crosshair;
  return (
    <div className={`${meta.gradient} rounded-card flex items-center gap-3.25 overflow-hidden p-4`}>
      <div className="flex h-11.5 w-11.5 shrink-0 items-center justify-center rounded-xl bg-[rgba(0,0,0,0.22)] text-xl text-white">
        <Icon aria-hidden size={20} />
      </div>
      <div className="min-w-0 flex-1">
        <div className="text-[17px] leading-tight font-black text-white">{meta.cn}</div>
        <div className="mt-0.5 text-[10.5px] tracking-[1.6px] text-white/80">{meta.en}</div>
      </div>
      <span className="shrink-0 rounded-full bg-[rgba(255,255,255,0.18)] px-2.5 py-1 text-[10.5px] font-extrabold text-white">
        {switched ? m.match_tag_switched() : m.match_tag_default()}
      </span>
    </div>
  );
}

interface SessionGateProps {
  readonly pending: boolean;
  readonly error: boolean;
  readonly undecided: boolean;
  readonly onRetry: () => void;
}

// The session read's non-faces: loading, the brief wait while the tie
// verdict's redirect is in flight, and the retryable error face.
function SessionGate({ pending, error, undecided, onRetry }: Readonly<SessionGateProps>) {
  if (pending || (error && undecided)) {
    return (
      <AmbientLayout orbs={guideOrbs}>
        <StateScreen variant="loading" />
      </AmbientLayout>
    );
  }
  return (
    <AmbientLayout orbs={guideOrbs}>
      <StateScreen
        action={
          <button
            className="bg-cta-primary rounded-control cursor-pointer px-5 py-2.5 text-sm font-semibold text-white active:opacity-80"
            onClick={onRetry}
            type="button"
          >
            {m.common_retry()}
          </button>
        }
        description={m.match_error_desc()}
        title={m.match_error_title()}
        variant="empty"
      />
    </AmbientLayout>
  );
}

/**
 * The match guide (prototype match-guide): genre header card, the
 * server-drawn four-ability lineup, and the flow actions. The session is
 * server truth — the genre param only seeds a creation (a resume ignores
 * it), the tie verdict routes to the pick page, and the start button
 * enters the first step with retakes left.
 */
export function MatchPage() {
  const navigate = useNavigate();
  const search = useSearch({ from: "/match/" }) as { genre?: Genre };
  const testsQuery = useConnectQuery(listTests, {});
  const sessionQuery = useConnectQuery(
    getMatchSession,
    search.genre ? { genre: search.genre } : {},
  );

  const tests = testsQuery.data?.tests ?? [];
  const session = sessionQuery.data?.session ?? null;
  const steps = session?.steps ?? [];
  const leftOf = (code: string): number => session?.retestLeft?.[code] ?? 0;
  const nameOf = (code: string) => tests.find((t) => t.code === code)?.name ?? code;
  const engineReady = (code: string): boolean => getGameModule(code) !== null;
  const doneCount = (session?.done ?? []).length;
  // An untested step is skippable by law, and a step whose engine has not
  // shipped must be skipped — the start button walks the engine-backed
  // steps only, never parking the flow on an unplayable one.
  const next = steps.find(
    (s) => s.testCode != null && leftOf(s.testCode) > 0 && engineReady(s.testCode),
  );
  const undecided =
    (sessionQuery.error as { reason?: string } | null)?.reason === reasonGenreUndecided;
  // The session's genre is an open word on the wire; faces narrow it to
  // the two-genre vocabulary this UI speaks.
  const genre = narrowGenre(session?.genre);
  const switched = search.genre !== undefined;

  useEffect(() => {
    if (undecided) {
      void navigate({ to: "/match/pick", search: { why: "tie" } });
    }
  }, [undecided, navigate]);

  if (sessionQuery.isPending || sessionQuery.isError) {
    return (
      <SessionGate
        error={sessionQuery.isError}
        onRetry={() => void sessionQuery.refetch()}
        pending={sessionQuery.isPending}
        undecided={undecided}
      />
    );
  }

  return (
    <AmbientLayout orbs={guideOrbs}>
      <div className="flex w-full flex-col gap-3 pb-6">
        <FlowHeader onBack={() => void navigate({ to: "/" })} title={m.match_title()} />

        {genre !== null ? (
          <>
            <GenreHeaderCard genre={genre} switched={switched} />

            <div className="mt-3 mb-2 flex items-baseline justify-between px-0.5">
              <span className="text-ink-tertiary text-[11.5px]">
                {m.match_sub({ n: steps.length })}
              </span>
              <span className="text-status-ok text-[12.5px] font-extrabold">
                {m.match_done({ done: doneCount, total: steps.length })}
              </span>
            </div>

            <div className="flex flex-col gap-[9px]">
              {steps.map((step, i) => {
                const code = step.testCode ?? "";
                return (
                  <StepRow
                    ability={step.ability ?? ""}
                    done={(session?.done ?? []).includes(code)}
                    index={i}
                    key={code || i}
                    left={leftOf(code)}
                    name={nameOf(code)}
                    onEnter={() => void navigate({ to: "/test/$code", params: { code } })}
                    ready={engineReady(code)}
                  />
                );
              })}
            </div>

            <button
              className="bg-cta-primary shadow-cta-primary rounded-control mt-3 w-full px-4 py-[14px] text-[15px] font-bold text-white active:opacity-80 disabled:opacity-40"
              disabled={next?.testCode == null}
              onClick={() => {
                if (next?.testCode != null) {
                  void navigate({ to: "/test/$code", params: { code: next.testCode } });
                }
              }}
              type="button"
            >
              {m.match_go()}
            </button>
            {/* A dimmed start button with no explanation reads as a bug; the
                note names the two reasons it can park (retakes spent, engines
                unshipped) and the way out (switching starts a fresh session). */}
            {next?.testCode == null && (
              <div className="text-ink-tertiary mt-1.5 px-0.5 text-center text-[11px] leading-relaxed">
                {m.match_go_blocked()}
              </div>
            )}
            <div className="mt-[9px] flex gap-[9px]">
              <button
                className="bg-surface-input text-ink-secondary rounded-control flex-1 border-[1.5px] border-[#4b5563] py-[11px] text-center text-[12.5px] font-medium active:opacity-60"
                onClick={() => void navigate({ to: "/match/guide" })}
                type="button"
              >
                {m.match_btn_guide()}
              </button>
              <button
                className="bg-surface-input text-ink-secondary rounded-control flex-1 border-[1.5px] border-[#4b5563] py-[11px] text-center text-[12.5px] font-medium active:opacity-60"
                onClick={() => void navigate({ to: "/match/switch" })}
                type="button"
              >
                {m.match_btn_switch({
                  type: genre === "fps" ? genreLabel("moba") : genreLabel("fps"),
                })}
              </button>
            </div>
          </>
        ) : null}
      </div>
    </AmbientLayout>
  );
}
