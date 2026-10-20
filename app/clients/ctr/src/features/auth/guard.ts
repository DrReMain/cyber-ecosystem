import { makeEnvelope } from "@cyber-ecosystem/shared-error";
import type { QueryClient } from "@tanstack/react-query";
import { redirect } from "@tanstack/react-router";

import { sessionQuery } from "./session.fn";

// The completeness law behind the onboarding gate: nickname plus a complete
// birth pair. Sex stays out until P8; a band-only entry has no contract
// (birth is the only age input UpdateProfile accepts), so birth is the gate.
export function profileComplete(profile: {
  birthMonth?: number | null;
  birthYear?: number | null;
  nickname?: string | null;
}): boolean {
  return profile.nickname != null && profile.birthYear != null && profile.birthMonth != null;
}

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
    // Only a dead session is "expired" — a fresh anonymous visit logs in
    // without the notice.
    throw redirect({
      search:
        session.status === "expired"
          ? { expired: true, next: args.location.href }
          : { next: args.location.href },
      to: "/login",
    });
  }
}

// The flow-entry guard: authed (sessionGuard) plus a complete profile —
// an incomplete member lands on the wizard's nickname step. Sits on the
// flow layouts (match now; the test runtime later), the same law the
// post-login router enforces.
export async function profileFlowGuard(args: {
  context: { queryClient: QueryClient };
  location: { href: string };
}): Promise<void> {
  await sessionGuard(args);
  const session = await args.context.queryClient.query(sessionQuery);
  if (session.status === "authed" && session.profile && !profileComplete(session.profile)) {
    throw redirect({ to: "/profile/setup/nickname" });
  }
}
