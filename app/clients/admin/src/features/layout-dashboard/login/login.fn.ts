import { createClient } from "@connectrpc/connect";
import { AuthService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/auth_pb";
import { createServerFn } from "@tanstack/react-start";
import { z } from "zod";
import { zodValidator } from "#/libs";
import { createKratosTransport } from "#/services/connect";
import { save } from "../auth/custody.server";

export const loginFn = createServerFn({ method: "POST" })
  .validator(zodValidator(z.object({ email: z.string().min(1), password: z.string().min(1) })))
  .handler(async ({ data }) => {
    const res = await createClient(AuthService, createKratosTransport()).login({
      email: data.email,
      password: data.password,
    });
    save(res.sessionToken);
    return { ok: true };
  });
