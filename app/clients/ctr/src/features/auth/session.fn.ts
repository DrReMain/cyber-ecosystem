import { createClient } from "@connectrpc/connect";
import { ProfileService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile_pb";
import { createServerFn } from "@tanstack/react-start";
import { normalize } from "#/domains/error";
import { createCtrTransport } from "#/services/connect/transport";
import { loadSession } from "./custody.server";

const TERMINAL_AUTH_REASON = "GENERAL_ERROR_UNAUTHENTICATED";

// The member session probe, admin's session.fn shape: auth outcomes land as
// status values, never as errors — a 401 is state, so the error domain (and
// its throwOnError escalation) never sees it.
export const sessionFn = createServerFn({ method: "GET" }).handler(async () => {
  const token = loadSession();
  if (token === null) {
    return { status: "anonymous" } as const;
  }
  try {
    const res = await createClient(ProfileService, createCtrTransport()).getProfile({});
    return {
      status: "authed",
      profile: res.profile,
      // Transport-flattened: the Timestamp arrives as its ISO string —
      // the generated type claims a Message the runtime never delivers.
      memberUntil: (res.memberUntil ?? null) as string | null,
      lastMemberUntil: (res.lastMemberUntil ?? null) as string | null,
    } as const;
  } catch (error) {
    const { kind, reason } = normalize(error);
    if (reason === TERMINAL_AUTH_REASON) {
      return { status: "expired" } as const;
    }
    if (kind === "network") {
      return { status: "unreachable" } as const;
    }
    throw error;
  }
});

export const sessionQuery = {
  queryKey: ["auth", "session"] as const,
  queryFn: () => sessionFn(),
  staleTime: 60_000,
};
