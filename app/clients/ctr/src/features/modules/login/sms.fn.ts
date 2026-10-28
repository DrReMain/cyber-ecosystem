import { createClient } from "@connectrpc/connect";
import { AuthService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/auth_pb";
import { createServerFn } from "@tanstack/react-start";
import { z } from "zod";
import { normalize } from "#/domains/error";
import { save } from "#/features/auth/custody.server";
import { createCtrTransport } from "#/services/connect/transport";

const send = z.object({ phone: z.string().regex(/^1[3-9]\d{9}$/) });

// The state is a flow token, not a secret — it proves a send happened, the
// way the wechat state rides the callback URL — so it crosses to the browser
// and rides back with the exchange. The code itself never does: only the
// phone's SMS inbox sees it.
export const sendSmsCodeFn = createServerFn({ method: "POST" })
  .validator((raw) => send.parse(raw))
  .handler(async ({ data }) => {
    try {
      const res = await createClient(AuthService, createCtrTransport()).sendSmsCode({
        phone: data.phone,
        provider: "sms",
      });
      return { state: res.state, status: "ok" } as const;
    } catch (error) {
      return { reason: normalize(error).reason, status: "rejected" } as const;
    }
  });

const exchange = z.object({
  code: z.string().regex(/^\d{6}$/),
  phone: z.string().regex(/^1[3-9]\d{9}$/),
  state: z.string().min(1),
});

// Same verdict-as-data law as the other two legs: the rejection reason
// survives the serialization boundary only as data, and the cookie mints
// server-side.
export const smsLoginFn = createServerFn({ method: "POST" })
  .validator((raw) => exchange.parse(raw))
  .handler(async ({ data }) => {
    try {
      const res = await createClient(AuthService, createCtrTransport()).loginByIdentity({
        payload: { case: "sms", value: { code: data.code, phone: data.phone } },
        provider: "sms",
        state: data.state,
      });
      save(res.sessionToken);
      return { status: "ok" } as const;
    } catch (error) {
      return { reason: normalize(error).reason, status: "rejected" } as const;
    }
  });
