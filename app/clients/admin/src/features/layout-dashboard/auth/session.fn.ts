import { createClient } from "@connectrpc/connect";
import { AuthService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/auth_pb";
import { createServerFn } from "@tanstack/react-start";
import { normalize } from "#/domains/error";
import { createKratosTransport } from "#/services/connect";
import { loadSession } from "./custody.server";

const TERMINAL_AUTH_REASON = "GENERAL_ERROR_UNAUTHENTICATED";

export const getCurrentUserFn = createServerFn({ method: "GET" }).handler(async () => {
  const token = loadSession();
  if (token === null) {
    return { status: "anonymous" } as const;
  }
  try {
    const res = await createClient(AuthService, createKratosTransport()).getCurrentUser({});
    return {
      status: "authed",
      user: { id: res.user?.id ?? "", email: res.user?.email ?? "", avatar: res.user?.avatar },
      permissions: [...(res.permissions ?? [])],
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
  queryFn: () => getCurrentUserFn(),
  staleTime: 60_000,
};
