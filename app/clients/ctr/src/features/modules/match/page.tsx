import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import { listGames } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/game-GameService_connectquery";
import { getProfile } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile-ProfileService_connectquery";
import { listTests } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test-TestService_connectquery";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { Crosshair, Hexagon } from "lucide-react";
import { useEffect } from "react";
import { toast } from "sonner";
import type { AmbientOrb } from "#/features/app/ambient/ambient-backdrop";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { PROFILE_STALE_MS } from "#/features/auth/profile-cache";
import { m } from "#/paraglide/messages";
import { arrange, defaultGenre, type Genre } from "./arranger";
import { FlowHeader } from "./ui/flow-header";
import { abilityWord, genreLabel, genreMeta } from "./vocabulary";

/* The guide page's own orb pair (prototype match-guide). */
const guideOrbs: readonly AmbientOrb[] = [
  { className: "top-[-70px] left-[-50px] size-[min(62vw,240px)]", tone: "blue" },
  { className: "bottom-[-60px] right-[-70px] size-[min(54vw,210px)]", tone: "cyan" },
];

/**
 * The match guide (prototype match-guide): genre header card, the
 * four-ability lineup, and the flow actions. The genre resolves from the
 * URL override, else the preference majority vote; a tie routes to the
 * pick page. Progress marks (done/best) arrive with the match session — the
 * lineup renders the fresh-session face until then.
 */
export function MatchPage() {
  const navigate = useNavigate();
  const search = useSearch({ from: "/match/" }) as { genre?: Genre };
  const testsQuery = useConnectQuery(listTests, {});
  const gamesQuery = useConnectQuery(listGames, {});
  const profileQuery = useConnectQuery(getProfile, {}, { staleTime: PROFILE_STALE_MS });

  const tests = testsQuery.data?.tests ?? [];
  const games = gamesQuery.data?.games ?? [];
  const preferenceCodes = (profileQuery.data?.profile?.games ?? [])
    .map((g) => g.gameId)
    .filter((id): id is string => id != null);
  const fallback = defaultGenre(preferenceCodes, games);
  const genre = search.genre ?? fallback;
  // A tie (or no classified picks) hands the choice to the pick page; an
  // empty pool would route to the same page's empty face. Both verdicts
  // wait for the catalog queries to land before navigating.
  useEffect(() => {
    if (testsQuery.isPending || gamesQuery.isPending || profileQuery.isPending) {
      return;
    }
    if (genre === null) {
      void navigate({ to: "/match/pick", search: { why: "tie" } });
      return;
    }
    if (arrange(tests, genre).length === 0) {
      void navigate({ to: "/match/pick", search: { why: "empty" } });
    }
  }, [genre, gamesQuery.isPending, navigate, profileQuery.isPending, tests, testsQuery.isPending]);
  const steps = genre ? arrange(tests, genre) : [];
  const meta = genre ? genreMeta(genre) : null;
  const switched = search.genre !== undefined && search.genre !== fallback;
  const Icon = genre === "moba" ? Hexagon : Crosshair;

  function onStart() {
    toast.info(m.match_flow_soon());
  }

  return (
    <AmbientLayout orbs={guideOrbs}>
      <div className="flex w-full flex-col gap-3 pb-6">
        <FlowHeader onBack={() => void navigate({ to: "/" })} title={m.match_title()} />

        {meta !== null && genre !== null ? (
          <>
            <div
              className={`${meta.gradient} rounded-card flex items-center gap-[13px] overflow-hidden p-4`}
            >
              <div className="flex h-[46px] w-[46px] shrink-0 items-center justify-center rounded-xl bg-[rgba(0,0,0,0.22)] text-xl text-white">
                <Icon aria-hidden size={20} />
              </div>
              <div className="min-w-0 flex-1">
                <div className="text-[17px] leading-tight font-black text-white">{meta.cn}</div>
                <div className="mt-0.5 text-[10.5px] tracking-[1.6px] text-white/80">{meta.en}</div>
              </div>
              <span className="shrink-0 rounded-full bg-white/20 px-2.5 py-1 text-[10.5px] font-extrabold text-white">
                {switched ? m.match_tag_switched() : m.match_tag_default()}
              </span>
            </div>

            <div className="mt-3 mb-2 flex items-baseline justify-between px-0.5">
              <span className="text-ink-tertiary text-[11.5px]">
                {m.match_sub({ n: steps.length })}
              </span>
              <span className="text-status-ok text-[12.5px] font-extrabold">
                {m.match_done({ done: 0, total: steps.length })}
              </span>
            </div>

            <div className="flex flex-col gap-[9px]">
              {steps.map((step, i) => {
                const word = abilityWord(step.ability);
                return (
                  <div
                    className="border-line bg-surface-card shadow-card rounded-control flex items-center gap-[11px] border px-[13px] py-[11px]"
                    key={step.code}
                  >
                    <div
                      className={`${word.gradient} rounded-control flex h-[38px] w-[38px] shrink-0 items-center justify-center text-[15px] font-bold text-white`}
                    >
                      {i + 1}
                    </div>
                    <div className="min-w-0 flex-1">
                      <div className="text-ink text-sm font-extrabold">{word.cn}</div>
                      <div className="text-ink-tertiary mt-0.5 text-[9.5px] tracking-[0.9px]">
                        {word.en}
                      </div>
                    </div>
                    <div className="shrink-0 text-right">
                      <div className="text-ink-secondary text-[12px]">{step.name}</div>
                      <div className="text-ink-tertiary mt-0.5 text-[10px]">
                        {m.match_step_time()}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>

            <button
              className="bg-cta-primary shadow-cta-primary rounded-control mt-3 w-full px-4 py-[14px] text-[15px] font-bold text-white active:opacity-80"
              onClick={onStart}
              type="button"
            >
              {m.match_go()}
            </button>
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
                onClick={() =>
                  void navigate({ to: "/match/switch", search: genre ? { from: genre } : {} })
                }
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
