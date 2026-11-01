import type { Transport } from "@connectrpc/connect";
import { createClient } from "@connectrpc/connect";
import { createQueryOptions } from "@connectrpc/connect-query";
import {
  type AgreementLoginMethod,
  AgreementService,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/agreement_pb";
import { getProfile } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile-ProfileService_connectquery";
import {
  AuthService,
  type GetSecurityStatusResponse,
} from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/auth_pb";
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

// The corpus's post-login bind prompt ("after login, first time,
// skippable"): once per account, a wechat-born member without a phone is
// walked to the bind-phone page; a phone-born member in the wechat browser
// without a wechat, to bind-wechat. Skip marks itself — the flag is the
// "don't nag" memory; the pages themselves offer the same exits.
async function bindPromptTarget(
  transport: Transport,
  userID: string | undefined,
): Promise<string | null> {
  if (userID === undefined || localStorage.getItem(`ctr.bindprompt.${userID}`) === "1") {
    return null;
  }
  let status: GetSecurityStatusResponse | undefined;
  try {
    status = await createClient(AuthService, transport).getSecurityStatus({});
  } catch {
    return null;
  }
  // toJson elides empty repeated fields — a bare account has no credentials key.
  const credentials = status.credentials ?? [];
  const phone = credentials.some((c) => c.kind === "sms");
  const wechat = credentials.some((c) => c.kind === "wechat_mp");
  const inWechat = /MicroMessenger/i.test(navigator.userAgent);
  localStorage.setItem(`ctr.bindprompt.${userID}`, "1");
  if (!phone) {
    return "/profile/security/bind-phone?from=login";
  }
  if (!wechat && inWechat) {
    return "/profile/security/bind-wechat?from=login";
  }
  return null;
}

// Shared post-login router: an explicit ?next= wins (guard bounce-back);
// otherwise completeness decides — incomplete profiles land on the wizard's
// nickname step, complete ones get the one-time bind prompt before home.
// It files the login's consent evidence in the background — one call
// records every current document on the versions the server resolves; the
// checked boxes at the gate are the claim, this is the receipt. Never
// awaited: the receipt must not add its round-trip to the login's own
// navigation (SPA routing keeps the in-flight write alive), and a failed
// write never blocks a login that already succeeded.
export async function routeAfterLogin(
  { queryClient, transport, navigate }: AfterLoginDeps,
  next: string | undefined,
  loginMethod: AgreementLoginMethod,
): Promise<void> {
  void createClient(AgreementService, transport)
    .recordAgreementConsent({ loginMethod })
    .catch(() => {
      // The next login re-records.
    });
  // The login just minted the session cookie: re-verify the session query
  // before any navigation — a guarded landing target reads the same cache
  // (60s staleTime) and would bounce straight back to login. Invalidate,
  // never remove: removing the query strands the watcher's observer on the
  // dead instance (removal does not re-subscribe observers), and every
  // later invalidation — the dead-session re-verification — would miss it.
  void queryClient.invalidateQueries({ queryKey: sessionQuery.queryKey });
  if (typeof next === "string" && isSafeNext(next) && next !== "/profile" && next !== "/") {
    await navigate({ href: next });
    return;
  }
  let complete = false;
  let userID: string | undefined;
  try {
    const data = await queryClient.query({
      ...createQueryOptions(getProfile, {}, { transport }),
      // A different member may sit down within PROFILE_STALE_MS — the
      // login seam always reads the wire, never the previous seat's cache.
      staleTime: 0,
    });
    complete = data.profile !== undefined && profileComplete(data.profile);
    userID = data.profile?.id ?? undefined;
  } catch {
    // A profile fetch failure never blocks the login itself — the overview
    // page owns surfacing profile errors.
    complete = false;
  }
  if (!complete) {
    await navigate({ to: "/profile/setup/nickname" });
    return;
  }
  const prompt = await bindPromptTarget(transport, userID);
  if (prompt !== null) {
    await navigate({ href: prompt });
    return;
  }
  await navigate({ to: "/" });
}
