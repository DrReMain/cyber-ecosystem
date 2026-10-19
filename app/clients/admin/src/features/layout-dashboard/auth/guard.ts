import { makeEnvelope } from "@cyber-ecosystem/shared-error";
import type { QueryClient } from "@tanstack/react-query";
import { redirect } from "@tanstack/react-router";

import { LOGIN_PATH } from "../area";
import { sessionQuery } from "./session.fn";

type Session = Awaited<ReturnType<typeof sessionQuery.queryFn>>;
type AuthedSession = Extract<Session, { status: "authed" }>;

export type SessionContext = Pick<AuthedSession, "user" | "permissions">;

export async function sessionGuard(args: {
  context: { queryClient: QueryClient };
  location: { href: string };
}): Promise<SessionContext> {
  const session = await args.context.queryClient.query(sessionQuery);
  if (session.status === "unreachable") {
    args.context.queryClient.removeQueries({ queryKey: sessionQuery.queryKey });
    throw makeEnvelope("network", "session backend unreachable");
  }
  if (session.status !== "authed") {
    const search = new URLSearchParams({ redirect: args.location.href });
    if (session.status === "expired") search.set("expired", "true");
    throw redirect({ href: `${LOGIN_PATH}?${search.toString()}` });
  }
  return { user: session.user, permissions: session.permissions };
}
