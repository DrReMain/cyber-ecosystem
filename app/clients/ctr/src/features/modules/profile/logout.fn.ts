import { createClient } from "@connectrpc/connect";
import { AuthService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/auth_pb";
import { createServerFn } from "@tanstack/react-start";
import { clear, loadSession, sessionCookieHeader } from "#/features/auth/custody.server";
import { createCtrTransport } from "#/services/connect/transport";

export const logoutFn = createServerFn({ method: "POST" }).handler(async () => {
  const token = loadSession();
  if (token !== null) {
    try {
      await createClient(AuthService, createCtrTransport()).logout(
        {},
        { headers: { cookie: sessionCookieHeader(token) } },
      );
    } catch {
      // Revocation failure must not strand the cookie.
    }
  }
  clear();
  return { ok: true } as const;
});
