import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import { listGames } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/game-GameService_connectquery";
import { getProfile } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile-ProfileService_connectquery";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { ChevronRight, FolderOpen } from "lucide-react";
import { toast } from "sonner";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { PROFILE_STALE_MS } from "#/features/auth/profile-cache";
import { m } from "#/paraglide/messages";
import { defaultGenre, GENRE_ORDER, type Genre } from "./arranger";
import { FlowHeader } from "./ui/flow-header";
import { abilityWord, genreLabel } from "./vocabulary";

/**
 * The manual genre pick (prototype match-genre-pick): tie face, plain
 * face, or the empty-pool face (?why=empty). Choosing rebuilds the
 * lineup via the genre URL param and returns to the guide. The card the
 * preference majority points at carries the prototype's 推荐 corner badge.
 */
export function GenrePickPage() {
  const navigate = useNavigate();
  const search = useSearch({ from: "/match/pick" }) as { why?: "tie" | "empty" };
  const empty = search.why === "empty";
  const profileQuery = useConnectQuery(getProfile, {}, { staleTime: PROFILE_STALE_MS });
  const gamesQuery = useConnectQuery(listGames, {});
  const preferenceCodes = (profileQuery.data?.profile?.games ?? [])
    .map((g) => g.gameId)
    .filter((id): id is string => id != null);
  const recommended = defaultGenre(preferenceCodes, gamesQuery.data?.games ?? []);

  function choose(genre: Genre) {
    toast.success(m.match_pick_toast({ type: genreLabel(genre) }));
    void navigate({ to: "/match", search: { genre } });
  }

  if (empty) {
    return (
      <AmbientLayout>
        <div className="flex w-full flex-col gap-3 pb-6">
          <FlowHeader
            onBack={() => void navigate({ to: "/" })}
            title={m.match_pick_empty_title()}
          />
          <div className="border-line bg-surface-card rounded-card border p-5 text-center">
            <FolderOpen
              aria-hidden
              className="text-ink-tertiary mx-auto"
              size={30}
              strokeWidth={1.5}
            />
            <div className="text-ink mt-2 text-sm font-extrabold">{m.match_pick_empty_title()}</div>
            <p className="text-ink-tertiary mt-2 text-[12.5px] leading-relaxed">
              {m.match_pick_empty_body()}
            </p>
          </div>
          <button
            className="bg-cta-primary rounded-control w-full px-4 py-[14px] text-[15px] font-bold text-white active:opacity-80"
            onClick={() => void navigate({ to: "/match/pick" })}
            type="button"
          >
            {m.match_pick_empty_retry()}
          </button>
          <button
            className="border-line bg-surface-content text-ink rounded-control w-full border px-4 py-[14px] text-[15px] font-bold active:opacity-60"
            onClick={() => void navigate({ to: "/" })}
            type="button"
          >
            {m.match_pick_empty_home()}
          </button>
        </div>
      </AmbientLayout>
    );
  }

  return (
    <AmbientLayout>
      <div className="flex w-full flex-col gap-3 pb-6">
        <FlowHeader onBack={() => void navigate({ to: "/" })} title={m.match_pick_title()} />
        {search.why === "tie" ? (
          <div className="border-status-pending/45 bg-surface-deep text-status-pending rounded-control border p-3.5 text-[12px] leading-relaxed">
            {m.match_pick_tie_head()}
          </div>
        ) : (
          <div className="border-line bg-surface-deep text-ink-tertiary rounded-control border p-3.5 text-[12px] leading-relaxed">
            {m.match_pick_normal_head()}
          </div>
        )}
        {(["fps", "moba"] as const).map((genre) => (
          <button
            className="border-line bg-surface-card rounded-card relative border p-4 text-left active:opacity-60"
            key={genre}
            onClick={() => choose(genre)}
            type="button"
          >
            {recommended === genre ? (
              <span className="bg-ambient-blue absolute top-3 right-3 rounded-full px-2 py-0.5 text-[10px] font-bold text-white">
                {m.match_pick_recommended()}
              </span>
            ) : null}
            <div className="flex items-center">
              <div className="flex-1">
                <div className="text-ink text-[15px] font-extrabold">
                  {genre === "fps" ? m.match_pick_fps() : m.match_pick_moba()}
                </div>
                <div className="text-ink-tertiary mt-0.5 text-[11px]">{m.match_pick_sub()}</div>
              </div>
              <ChevronRight aria-hidden className="text-ink-tertiary" size={16} />
            </div>
            <div className="mt-1.5">
              {GENRE_ORDER[genre].map((ability) => {
                const word = abilityWord(ability);
                return (
                  <div className="mt-1 flex items-center gap-2" key={ability}>
                    <span
                      className={`${word.gradient} flex h-5 w-5 items-center justify-center rounded-sm text-[9px] font-bold text-white`}
                    >
                      {word.axis}
                    </span>
                    <span className="text-ink-secondary text-[12px]">{word.cn}</span>
                  </div>
                );
              })}
            </div>
          </button>
        ))}
      </div>
    </AmbientLayout>
  );
}
