import { makeEnvelope } from "@cyber-ecosystem/shared-error";
import type { QueryClient } from "@tanstack/react-query";
import { redirect } from "@tanstack/react-router";

import { sessionQuery } from "./session.fn";

// Admin's sessionGuard shape on ctr's login convention: anonymous and
// expired both hop to the login page carrying the origin href as next;
// unreachable is a backend outage, not an auth verdict — drop the cached
// probe and surface the network error face instead.
export async function sessionGuard(args: {
  context: { queryClient: QueryClient };
  location: { href: string };
}): Promise<void> {
  const session = await args.context.queryClient.query(sessionQuery);
  if (session.status === "unreachable") {
    args.context.queryClient.removeQueries({ queryKey: sessionQuery.queryKey });
    throw makeEnvelope("network", "session backend unreachable");
  }
  if (session.status !== "authed") {
    throw redirect({ to: "/login", search: { next: args.location.href } });
  }
}
