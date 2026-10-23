import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import { listGames } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/game-GameService_connectquery";
import { getProfile } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile-ProfileService_connectquery";
import { listTests } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test-TestService_connectquery";
import { PROFILE_STALE_MS } from "#/features/auth/profile-cache";
import { type ArrangedStep, arrangePlayable, defaultGenre, type Genre } from "./arranger";

export interface ArrangedSteps {
  /** The genre the lineup renders: the caller's override, else the
   * preference majority (null = tie or no classified picks → pick page). */
  readonly genre: Genre | null;
  /** The preference-majority genre regardless of override — the switched
   * tag and the switch CTA compare against it. */
  readonly fallback: Genre | null;
  /** The walkable lineup for `genre` — engine-ready steps only. */
  readonly steps: readonly ArrangedStep[];
  /** True while any catalog input is still in flight. */
  readonly pending: boolean;
}

/**
 * The shared lineup resolver for the match flow pages: catalog + games +
 * profile queries, the preference-majority genre, and the engine-ready
 * arrangement in one place, so the lineup page and the pre-assessment
 * notice page walk the same first step (prototype law: both CTAs enter the
 * current step's intro, never a dead end).
 */
export function useArrangedSteps(override?: Genre): ArrangedSteps {
  const testsQuery = useConnectQuery(listTests, {});
  const gamesQuery = useConnectQuery(listGames, {});
  const profileQuery = useConnectQuery(getProfile, {}, { staleTime: PROFILE_STALE_MS });

  const tests = testsQuery.data?.tests ?? [];
  const games = gamesQuery.data?.games ?? [];
  const preferenceCodes = (profileQuery.data?.profile?.games ?? [])
    .map((g) => g.gameId)
    .filter((id): id is string => id != null);
  const fallback = defaultGenre(preferenceCodes, games);
  const genre = override ?? fallback;
  return {
    genre,
    fallback,
    steps: genre !== null ? arrangePlayable(tests, genre) : [],
    pending: testsQuery.isPending || gamesQuery.isPending || profileQuery.isPending,
  };
}
