import { createQueryOptions } from "@connectrpc/connect-query";
import { getProfile } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile-ProfileService_connectquery";
import { connectTransport } from "#/services/connect";

// The getProfile query's shared cache policy. Profile freshness is owned by
// its write seam — UpdateProfile invalidates this key and the session probe
// (use-profile-update) — so mounts read warm cache instead of refetching on
// every visit (T-RTE-5). The window bounds only admin-side drift. Two seams
// must bypass it: after-login forces staleTime 0 (a different member may sit
// down at the same browser), and the session-kick seam drops the cache
// outright so no seat ever reads the previous one.
export const PROFILE_STALE_MS = 30_000;

// Key parity with every consumer — the pages' useQuery and the loaders'
// prefetch both ride the static-keyed transport singleton (§3), so
// invalidation here reaches every cache entry.
export const profileQueryKey = createQueryOptions(
  getProfile,
  {},
  { transport: connectTransport },
).queryKey;
