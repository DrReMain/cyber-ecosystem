import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import { listGames } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/game-GameService_connectquery";
import { getProfile } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile-ProfileService_connectquery";
import { listTests } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test-TestService_connectquery";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { Crosshair, Hexagon } from "lucide-react";
import { useEffect } from "react";
import { toast } from "sonner";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { PROFILE_STALE_MS } from "#/features/auth/profile-cache";
import { m } from "#/paraglide/messages";
import { arrange, defaultGenre, type Genre } from "./arranger";
import { FlowHeader } from "./ui/flow-header";
import { abilityWord, genreLabel, genreMeta } from "./vocabulary";

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
    <AmbientLayout>
      <div className="flex w-full flex-col gap-3 pb-6">
        <FlowHeader onBack={() => void navigate({ to: "/" })} title={m.match_title()} />

        {meta !== null && genre !== null ? (
          <>
            <div
              className={`${meta.gradient} flex items-center gap-3 overflow-hidden rounded-card border border-line p-4`}
            >
              <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-black/25 text-white text-xl">
                <Icon aria-hidden size={22} />
              </div>
              <div className="min-w-0 flex-1">
                <div className="font-black text-[17px] text-white leading-tight">{meta.cn}</div>
                <div className="mt-0.5 text-[10.5px] text-white/80 tracking-[1.6px]">{meta.en}</div>
              </div>
              <span className="shrink-0 rounded-full bg-white/20 px-2.5 py-1 font-extrabold text-[10.5px] text-white">
                {switched ? m.match_tag_switched() : m.match_tag_default()}
              </span>
            </div>

            <div className="mb-1 flex items-baseline justify-between px-0.5">
              <span className="text-[11.5px] text-ink-tertiary">
                {m.match_sub({ n: steps.length })}
              </span>
              <span className="font-extrabold text-[12.5px] text-status-ok">
                {m.match_done({ done: 0, total: steps.length })}
              </span>
            </div>

            <div className="flex flex-col gap-2">
              {steps.map((step, i) => {
                const word = abilityWord(step.ability);
                return (
                  <div
                    className="flex items-center gap-2.5 rounded-lg border border-line bg-surface-card p-3"
                    key={step.code}
                  >
                    <div
                      className={`${word.gradient} flex h-9 w-9 shrink-0 items-center justify-center rounded-tile font-bold text-[15px] text-white`}
                    >
                      {i + 1}
                    </div>
                    <div className="min-w-0 flex-1">
                      <div className="font-extrabold text-ink text-sm">{word.cn}</div>
                      <div className="mt-0.5 text-[9.5px] text-ink-tertiary tracking-[0.9px]">
                        {word.en}
                      </div>
                    </div>
                    <div className="shrink-0 text-right">
                      <div className="text-[12px] text-ink-secondary">{step.name}</div>
                      <div className="mt-0.5 text-[10px] text-ink-tertiary">
                        {m.match_step_time()}
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>

            <button
              className="mt-2 w-full rounded-sm bg-cta-primary px-4 py-3 font-semibold text-white active:opacity-80"
              onClick={onStart}
              type="button"
            >
              {m.match_go()}
            </button>
            <div className="flex gap-2">
              <button
                className="flex-1 rounded-sm border border-line bg-surface-content px-3 py-2.5 font-semibold text-[12.5px] text-ink active:opacity-60"
                onClick={() => void navigate({ to: "/match/guide" })}
                type="button"
              >
                {m.match_btn_guide()}
              </button>
              <button
                className="flex-1 rounded-sm border border-line bg-surface-content px-3 py-2.5 font-semibold text-[12.5px] text-ink active:opacity-60"
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
