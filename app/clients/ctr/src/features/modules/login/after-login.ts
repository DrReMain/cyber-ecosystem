import type { Transport } from "@connectrpc/connect";
import { createQueryOptions } from "@connectrpc/connect-query";
import { getProfile } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile-ProfileService_connectquery";
import type { QueryClient } from "@tanstack/react-query";
import type { useNavigate } from "@tanstack/react-router";
import { profileComplete } from "#/features/auth/guard";
import { sessionQuery } from "#/features/auth/session.fn";

type Navigate = ReturnType<typeof useNavigate>;

export { profileComplete };

// Everything the shared post-login router needs; callers assemble it with
// the hooks they already own (useQueryClient, useTransport, useNavigate).
export interface AfterLoginDeps {
  queryClient: QueryClient;
  transport: Transport;
  navigate: Navigate;
}

// Open-redirect hygiene (admin's redirect.ts law, inline): same-origin
// paths only — protocol-relative "//host" is the one shape a plain
// startsWith("/") check lets through.
function isSafeNext(value: string): boolean {
  return value.startsWith("/") && !value.startsWith("//");
}

// Shared post-login router: an explicit ?next= wins (guard bounce-back);
// otherwise completeness decides — incomplete profiles land on the wizard's
// nickname step, complete ones on home.
export async function routeAfterLogin(
  { queryClient, transport, navigate }: AfterLoginDeps,
  next: string | undefined,
): Promise<void> {
  // The login just minted the session cookie: drop any cached anonymous
  // verdict before any navigation — a guarded landing target reads the same
  // cache (60s staleTime) and would bounce straight back to login.
  queryClient.removeQueries({ queryKey: sessionQuery.queryKey });
  if (typeof next === "string" && isSafeNext(next) && next !== "/profile" && next !== "/") {
    await navigate({ href: next });
    return;
  }
  let complete = false;
  try {
    const data = await queryClient.ensureQueryData(
      createQueryOptions(getProfile, {}, { transport }),
    );
    complete = data.profile !== undefined && profileComplete(data.profile);
  } catch {
    // A profile fetch failure never blocks the login itself — the overview
    // page owns surfacing profile errors.
    complete = false;
  }
  await navigate({ to: complete ? "/" : "/profile/setup/nickname" });
}
