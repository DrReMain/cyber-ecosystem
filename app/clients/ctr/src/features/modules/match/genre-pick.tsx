import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import { listGames } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/game-GameService_connectquery";
import { getProfile } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile-ProfileService_connectquery";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { ChevronRight, FolderOpen } from "lucide-react";
import { toast } from "sonner";
import type { AmbientOrb } from "#/features/app/ambient/ambient-backdrop";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { PROFILE_STALE_MS } from "#/features/auth/profile-cache";
import { m } from "#/paraglide/messages";
import { FlowHeader } from "./ui/flow-header";
import { abilityWord, defaultGenre, GENRE_ORDER, type Genre, genreLabel } from "./vocabulary";

/* The page's own orb pair (prototype match-genre-pick). */
const pickOrbs: readonly AmbientOrb[] = [
  { className: "top-[-60px] left-[-50px] size-[min(59vw,220px)]", tone: "violet" },
  { className: "bottom-[-50px] right-[-50px] size-[min(53vw,200px)]", tone: "purple" },
];

/**
 * The manual genre pick (prototype match-genre-pick): tie face, plain
 * face, or the empty-pool face (?why=empty). Choosing rebuilds the
 * lineup by seeding the session creation with that genre (the URL param)
 * and returns to the guide. The 推荐 corner badge is an enhancement
 * beyond the prototype's pick face — it carries the preference majority
 * signal the guide's 默认推荐 tag also shows.
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
      <AmbientLayout orbs={pickOrbs}>
        <div className="flex w-full flex-col gap-3 pb-6">
          <FlowHeader
            onBack={() => void navigate({ to: "/" })}
            title={m.match_pick_empty_title()}
          />
          <div className="border-line bg-surface-card rounded-card border px-4 py-[22px] text-center">
            <FolderOpen
              aria-hidden
              className="text-ink-tertiary mx-auto"
              size={30}
              strokeWidth={1.5}
            />
            <div className="text-ink mt-2 text-sm font-extrabold">
              {m.match_pick_empty_card_title()}
            </div>
            <p className="text-ink-muted mt-2 text-[12.5px] leading-[1.8]">
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
            className="bg-surface-input text-ink-secondary rounded-control w-full border-[1.5px] border-[#4b5563] px-4 py-[14px] text-[15px] font-medium active:opacity-60"
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
    <AmbientLayout orbs={pickOrbs}>
      <div className="flex w-full flex-col gap-3 pb-6">
        <FlowHeader onBack={() => void navigate({ to: "/" })} title={m.match_pick_title()} />
        {search.why === "tie" ? (
          <div className="bg-surface-content text-status-pending rounded-control border border-[rgba(245,158,11,0.45)] px-3.5 py-[13px] text-[12px] leading-[1.85]">
            {m.match_pick_tie_lead()}
            <span className="font-bold">{m.match_pick_tie_tied()}</span>
            {m.match_pick_tie_tail()}
          </div>
        ) : (
          <div className="border-line bg-surface-content text-ink-muted rounded-control border px-3.5 py-[13px] text-[12px] leading-[1.85]">
            {m.match_pick_normal_lead()}
            <span className="font-bold">{m.match_pick_normal_session()}</span>
            {m.match_pick_normal_tail()}
          </div>
        )}
        {(["fps", "moba"] as const).map((genre) => (
          <button
            className="border-line bg-surface-card rounded-card relative border px-4 py-[15px] text-left active:opacity-60"
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
                <div className="text-ink-tertiary mt-[3px] text-[11px]">{m.match_pick_sub()}</div>
              </div>
              <ChevronRight aria-hidden className="text-ink-tertiary" size={16} />
            </div>
            {GENRE_ORDER[genre].map((ability) => {
              const word = abilityWord(ability);
              return (
                <div className="mt-[5px] flex items-center gap-[7px]" key={ability}>
                  <span
                    className={`${word.gradient} flex h-[22px] w-[22px] items-center justify-center rounded-full text-[9px] font-bold text-white`}
                  >
                    {word.axis}
                  </span>
                  <span className="text-ink-muted text-[12px]">{word.cn}</span>
                </div>
              );
            })}
          </button>
        ))}
      </div>
    </AmbientLayout>
  );
}
