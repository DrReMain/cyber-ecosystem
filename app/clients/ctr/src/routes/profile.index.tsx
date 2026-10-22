import { createQueryOptions } from "@connectrpc/connect-query";
import { getProfile } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile-ProfileService_connectquery";
import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { PROFILE_STALE_MS } from "#/features/auth/profile-cache";
import { sessionQuery } from "#/features/auth/session.fn";
import { ProfilePage } from "#/features/modules/profile/page";
import { m } from "#/paraglide/messages";
import { connectTransport } from "#/services/connect/transport";

export const Route = createFileRoute("/profile/")({
  head: () => ({
    meta: [{ title: pageTitle(m.profile_title()) }],
  }),
  loader: async ({ context }) => {
    const session = await context.queryClient.query(sessionQuery);
    if (session.status !== "authed") return;
    await context.queryClient
      .query({
        ...createQueryOptions(getProfile, {}, { transport: connectTransport }),
        // The profile is write-seam fresh (PROFILE_STALE_MS); a stale-time-0
        // loader fetch would race the route swap and get cancelled with the
        // departing page's observers.
        staleTime: PROFILE_STALE_MS,
      })
      .catch(() => undefined);
  },
  component: ProfilePage,
});
