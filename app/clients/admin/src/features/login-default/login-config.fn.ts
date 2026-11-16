import { createClient } from "@connectrpc/connect";
import { AuthService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/auth_pb";
import { createServerFn } from "@tanstack/react-start";
import { createKratosTransport } from "#/services/connect";

// The login page's anonymous discovery face: which identity providers this
// deployment wired. The answer is a boot-time fact, so the query caches for
// the session — the page renders the account form alone on a closed
// deployment instead of a dead button.
export const getLoginConfigFn = createServerFn({ method: "GET" }).handler(async () => {
  const res = await createClient(AuthService, createKratosTransport()).getLoginConfig({});
  return {
    providers: (res.providers ?? []).map((provider) => ({
      key: provider.key,
      kind: provider.kind,
    })),
  } as const;
});
