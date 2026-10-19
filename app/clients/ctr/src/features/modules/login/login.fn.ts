import { createClient } from "@connectrpc/connect";
import { AuthService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/auth_pb";
import { createServerFn } from "@tanstack/react-start";
import { z } from "zod";
import { save } from "#/features/auth/custody.server";
import { createCtrTransport } from "#/services/connect/transport";

const credentials = z.object({ account: z.string().min(1), password: z.string().min(1) });

// The load-bearing wall: the token never reaches the browser — the server
// fn exchanges credentials and mints the HttpOnly session cookie itself.
export const loginFn = createServerFn({ method: "POST" })
  .validator((raw) => credentials.parse(raw))
  .handler(async ({ data }) => {
    const res = await createClient(AuthService, createCtrTransport()).login({
      account: data.account,
      password: data.password,
    });
    save(res.sessionToken);
    return { ok: true } as const;
  });
