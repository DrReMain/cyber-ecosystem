import { createClient } from "@connectrpc/connect";
import { AuthService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/auth_pb";
import { createServerFn } from "@tanstack/react-start";
import { createKratosTransport } from "#/services/connect";
import { clear } from "./custody.server";

export const logoutFn = createServerFn({ method: "POST" }).handler(async () => {
  try {
    // The transport forwards the session cookie, so the backend revokes the
    // key; a rejection here is common (the session is often already dead) -
    // clear() below ends the local session either way.
    await createClient(AuthService, createKratosTransport()).logout({});
  } catch {
    // Swallowed on purpose.
  }
  clear();
  return { ok: true };
});
