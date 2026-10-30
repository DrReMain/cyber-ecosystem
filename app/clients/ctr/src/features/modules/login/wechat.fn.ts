import { createClient } from "@connectrpc/connect";
import { AuthService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/auth_pb";
import { createServerFn } from "@tanstack/react-start";
import { z } from "zod";
import { normalize } from "#/domains/error";
import { save } from "#/features/auth/custody.server";
import { createCtrTransport } from "#/services/connect/transport";

const redirect = z.object({ redirectUri: z.string().min(1) });

// The wechat leg never touches a secret: system mints the authorize URL
// (state included) and later exchanges the code server-side; the browser
// only ever sees a redirect and a session cookie.
export const beginWechatLoginFn = createServerFn({ method: "POST" })
  .validator((raw) => redirect.parse(raw))
  .handler(async ({ data }) => {
    const res = await createClient(AuthService, createCtrTransport()).beginIdentityLogin({
      provider: "wechat",
      redirectUri: data.redirectUri,
    });
    return { authorizeUrl: res.authorizeUrl } as const;
  });

const exchange = z.object({ code: z.string().min(1), state: z.string().min(1) });

export const wechatLoginFn = createServerFn({ method: "POST" })
  .validator((raw) => exchange.parse(raw))
  .handler(async ({ data }) => {
    try {
      const res = await createClient(AuthService, createCtrTransport()).loginByIdentity({
        provider: "wechat",
        payload: { case: "wechat", value: { code: data.code } },
        state: data.state,
      });
      save(res.sessionToken);
      // Same verdict-as-data law as the password leg: the rejection reason
      // survives the serialization boundary only as data. The person the
      // userinfo leg learned rides along verbatim — the wizard prefill's
      // only source; absent when that leg degraded.
      return {
        status: "ok" as const,
        wechatProfile:
          res.wechatProfile == null
            ? undefined
            : {
                nickname: res.wechatProfile.nickname ?? "",
                avatarUrl: res.wechatProfile.avatarUrl ?? "",
              },
      };
    } catch (error) {
      return { status: "rejected", reason: normalize(error).reason } as const;
    }
  });
