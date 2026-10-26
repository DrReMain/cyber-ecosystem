import { createQueryOptions } from "@connectrpc/connect-query";
import {
  getMatchSession,
  peekMatchSession,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/match-MatchService_connectquery";
import type { QueryClient } from "@tanstack/react-query";
import { connectTransport } from "#/services/connect";

// The match session's client seam. Consumers query with exactly two input
// shapes — the bare resume call and the genre-seeded creation call the pick
// page produces — so the keys built here plus this module's invalidation
// reach every cache entry the UI can create. The test module reads the
// same query to bind submits, mirroring the server's one-way match→test
// edge; the test flow never writes this aggregate.

// The server's verdict when the preference vote cannot seed a creation —
// the pick page owns the manual choice then.
export const reasonGenreUndecided = "ESPORTS_MATCH_GENRE_UNDECIDED";

export function matchSessionKey(genre?: string) {
  const input = genre === undefined ? {} : { genre };
  return createQueryOptions(getMatchSession, input, { transport: connectTransport }).queryKey;
}

// The read-only peek (home's resume card): one input shape, and it must
// ride the same invalidation so a switch/expiry the user just performed
// never leaves a stale card on the next home visit.
export function matchPeekKey() {
  return createQueryOptions(peekMatchSession, {}, { transport: connectTransport }).queryKey;
}

// The pick page's manual choice is the only genre-seeded call and its
// vocabulary is the two seeded genres — three keys cover every entry.
export function invalidateMatchSession(queryClient: QueryClient): void {
  const keys = [matchSessionKey(), matchSessionKey("fps"), matchSessionKey("moba"), matchPeekKey()];
  for (const key of keys) {
    void queryClient.invalidateQueries({ queryKey: key });
  }
}
