import { createClient } from "@connectrpc/connect";
import { AuthService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/auth_pb";
import { createServerFn } from "@tanstack/react-start";
import { z } from "zod";
import { normalize } from "#/domains/error";
import { env } from "#/env";
import { save } from "#/features/auth/custody.server";
import { createCtrTransport } from "#/services/connect/transport";

const credentials = z.object({ account: z.string().min(1), password: z.string().min(1) });

// The load-bearing wall: the token never reaches the browser — the server
// fn exchanges credentials and mints the HttpOnly session cookie itself.
// The verdict crosses the serialization boundary as data: a thrown
// ConnectError loses its reason on the wire, and with it the honest
// stratification the login surface branches on.
export const loginFn = createServerFn({ method: "POST" })
  .validator((raw) => credentials.parse(raw))
  .handler(async ({ data }) => {
    try {
      const res = await createClient(AuthService, createCtrTransport()).login({
        account: data.account,
        password: data.password,
        population: env.VITE_POPULATION,
      });
      save(res.sessionToken);
      // The discriminant is a string tag, never a boolean: the framed
      // serialization boundary mangles booleans into truthy strings.
      return { status: "ok" } as const;
    } catch (error) {
      // The transport throws the error domain's envelope (never a raw
      // ConnectError), and normalize passes envelopes through — one call
      // covers both shapes.
      return { status: "rejected", reason: normalize(error).reason } as const;
    }
  });
