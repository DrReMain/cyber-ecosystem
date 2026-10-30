import { createClient } from "@connectrpc/connect";
import {
  MergePreviewStatus,
  ProfileService,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile_pb";
import { AuthService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/auth_pb";
import { UserService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/user_pb";
import { createServerFn } from "@tanstack/react-start";
import { z } from "zod";
import { normalize } from "#/domains/error";
import { loadSession, sessionCookieHeader } from "#/features/auth/custody.server";
import { createCtrTransport } from "#/services/connect/transport";

const phoneSchema = z.string().regex(/^1[3-9]\d{9}$/);
const codeSchema = z.string().regex(/^\d{6}$/);

function authedClient() {
  const token = loadSession();
  if (token === null) {
    throw new Error("no session");
  }
  return {
    client: createClient(AuthService, createCtrTransport()),
    headers: { cookie: sessionCookieHeader(token) },
  };
}

// The offer verdict crosses the server-fn boundary as transport JSON (the
// toJson anti-corruption layer): enums arrive as their proto names and WKT
// timestamps as strings. This DTO owns that shape — pages never guess.
export interface MergeSideDTO {
  userId: string;
  phoneMask?: string;
  wechat: boolean;
  sessions: number;
  memberUntil?: string;
  unlocked: boolean;
}

export type MergeStatusDTO = "FREE" | "SELF" | "KEEP_CALLER" | "KEEP_OTHER" | "BLOCKED";

export interface MergeOfferDTO {
  status: MergeStatusDTO;
  caller?: MergeSideDTO;
  other?: MergeSideDTO;
}

interface RawSide {
  userId?: string;
  phoneMask?: string | null;
  wechat?: boolean;
  sessions?: number;
  memberUntil?: string | null;
  unlocked?: boolean;
}

function sideDTO(raw: RawSide | undefined | null): MergeSideDTO | undefined {
  if (raw === undefined || raw === null) {
    return undefined;
  }
  return {
    userId: raw.userId ?? "",
    phoneMask: raw.phoneMask ?? undefined,
    wechat: raw.wechat ?? false,
    sessions: raw.sessions ?? 0,
    memberUntil: raw.memberUntil ?? undefined,
    unlocked: raw.unlocked ?? false,
  };
}

function statusDTO(raw: unknown): MergeStatusDTO {
  let name = String(raw);
  if (typeof raw === "number") {
    name =
      Object.keys(MergePreviewStatus).find(
        (k) => MergePreviewStatus[k as keyof typeof MergePreviewStatus] === raw,
      ) ?? "";
  }
  return name.replace("MERGE_PREVIEW_STATUS_", "") as MergeStatusDTO;
}

function offerDTO(raw: {
  status?: unknown;
  caller?: RawSide | null;
  other?: RawSide | null;
}): MergeOfferDTO {
  return { status: statusDTO(raw.status), caller: sideDTO(raw.caller), other: sideDTO(raw.other) };
}

const sendBind = z.object({ phone: phoneSchema });

// The bind twin of the login send: same body, the rung is the difference —
// the minted state carries this session, which BindCredential's cash
// enforces.
export const sendBindSmsCodeFn = createServerFn({ method: "POST" })
  .validator((raw) => sendBind.parse(raw))
  .handler(async ({ data }) => {
    const { client, headers } = authedClient();
    try {
      const res = await client.sendBindSmsCode({ phone: data.phone, provider: "sms" }, { headers });
      return { state: res.state, status: "ok" } as const;
    } catch (error) {
      return { reason: normalize(error).reason, status: "rejected" } as const;
    }
  });

const bind = z.object({ code: codeSchema, phone: phoneSchema, state: z.string().min(1) });

export const bindSmsCredentialFn = createServerFn({ method: "POST" })
  .validator((raw) => bind.parse(raw))
  .handler(async ({ data }) => {
    const { client, headers } = authedClient();
    try {
      await client.bindCredential(
        {
          payload: { case: "sms", value: { code: data.code, phone: data.phone } },
          provider: "sms",
          state: data.state,
        },
        { headers },
      );
      return { status: "ok" } as const;
    } catch (error) {
      return { reason: normalize(error).reason, status: "rejected" } as const;
    }
  });

const begin = z.object({ redirectUri: z.string().min(1) });

// The bind leg's front door: mint the session-bound state and hand back the
// authorize URL the browser must visit. The callback path differs per step
// (bind, merge execution), so the caller names it.
export const beginBindWechatFn = createServerFn({ method: "POST" })
  .validator((raw) => begin.parse(raw))
  .handler(async ({ data }) => {
    const { client, headers } = authedClient();
    try {
      const res = await client.beginBindCredential(
        { provider: "wechat", redirectUri: data.redirectUri },
        { headers },
      );
      return { authorizeUrl: res.authorizeUrl, status: "ok" } as const;
    } catch (error) {
      return { reason: normalize(error).reason, status: "rejected" } as const;
    }
  });

const wechatProof = z.object({ code: z.string().min(1), state: z.string().min(1) });

export const bindWechatCredentialFn = createServerFn({ method: "POST" })
  .validator((raw) => wechatProof.parse(raw))
  .handler(async ({ data }) => {
    const { client, headers } = authedClient();
    try {
      const res = await client.bindCredential(
        {
          payload: { case: "wechat", value: { code: data.code } },
          provider: "wechat",
          state: data.state,
        },
        { headers },
      );
      // The person behind the openid arrives with the binding — the wizard
      // pill's whole point is bind-and-prefill in one exchange. The
      // already-bound-to-self branch answers it too, so a stale pill click
      // still prefills instead of dead-ending.
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
      return { reason: normalize(error).reason, status: "rejected" } as const;
    }
  });

// The offer face, phone branch: occupancy + emptiness verdict before a code
// is ever requested (the prototype's "occupied — no send" law).
const previewPhone = z.object({ phone: phoneSchema });

export const previewMergePhoneFn = createServerFn({ method: "POST" })
  .validator((raw) => previewPhone.parse(raw))
  .handler(async ({ data }) => {
    const token = loadSession();
    if (token === null) {
      return { reason: "GENERAL_ERROR_UNAUTHENTICATED", status: "rejected" } as const;
    }
    try {
      const res = await createClient(ProfileService, createCtrTransport()).previewMerge(
        { credential: { case: "phone", value: data.phone } },
        { headers: { cookie: sessionCookieHeader(token) } },
      );
      return { offer: offerDTO(res as never), status: "ok" } as const;
    } catch (error) {
      return { reason: normalize(error).reason, status: "rejected" } as const;
    }
  });

// The offer face, wechat branch: proof-carrying — the state BeginBindCredential
// minted plus the redirect code; the exchange burns both.
export const previewMergeWechatFn = createServerFn({ method: "POST" })
  .validator((raw) => wechatProof.parse(raw))
  .handler(async ({ data }) => {
    const token = loadSession();
    if (token === null) {
      return { reason: "GENERAL_ERROR_UNAUTHENTICATED", status: "rejected" } as const;
    }
    try {
      const res = await createClient(ProfileService, createCtrTransport()).previewMerge(
        { credential: { case: "wechat", value: { code: data.code } }, state: data.state },
        { headers: { cookie: sessionCookieHeader(token) } },
      );
      return { offer: offerDTO(res as never), status: "ok" } as const;
    } catch (error) {
      return { reason: normalize(error).reason, status: "rejected" } as const;
    }
  });

const mergeSms = z.object({ code: codeSchema, phone: phoneSchema, state: z.string().min(1) });

// The executor, sms proof: the caller is the surviving account, the code
// proves control of the absorbed side's phone.
export const mergeSmsFn = createServerFn({ method: "POST" })
  .validator((raw) => mergeSms.parse(raw))
  .handler(async ({ data }) => {
    const { client, headers } = authedClient();
    try {
      await client.mergeAccounts(
        {
          payload: { case: "sms", value: { code: data.code, phone: data.phone } },
          provider: "sms",
          state: data.state,
        },
        { headers },
      );
      return { status: "ok" } as const;
    } catch (error) {
      return { reason: normalize(error).reason, status: "rejected" } as const;
    }
  });

// The executor, wechat proof: a fresh OAuth leg minted on this session.
export const mergeWechatFn = createServerFn({ method: "POST" })
  .validator((raw) => wechatProof.parse(raw))
  .handler(async ({ data }) => {
    const { client, headers } = authedClient();
    try {
      await client.mergeAccounts(
        {
          payload: { case: "wechat", value: { code: data.code } },
          provider: "wechat",
          state: data.state,
        },
        { headers },
      );
      return { status: "ok" } as const;
    } catch (error) {
      return { reason: normalize(error).reason, status: "rejected" } as const;
    }
  });

// The switched-session merge reuses the login family's own legs verbatim:
// beginWechatLoginFn (custom redirect) and wechatLoginFn (exchange +
// cookie save) in features/modules/login/wechat.fn.ts.

const reset = z.object({
  code: codeSchema,
  newPassword: z.string().min(6).max(128),
  phone: phoneSchema,
  state: z.string().min(1),
});

// One setter for both password-less legs (forgot on the login page, first
// set in account security): the code to the account's own phone is the
// whole proof.
export const resetPasswordFn = createServerFn({ method: "POST" })
  .validator((raw) => reset.parse(raw))
  .handler(async ({ data }) => {
    try {
      await createClient(AuthService, createCtrTransport()).resetPassword({
        newPassword: data.newPassword,
        payload: { case: "sms", value: { code: data.code, phone: data.phone } },
        provider: "sms",
        state: data.state,
      });
      return { status: "ok" } as const;
    } catch (error) {
      return { reason: normalize(error).reason, status: "rejected" } as const;
    }
  });

const change = z.object({
  newPassword: z.string().min(6).max(128),
  oldPassword: z.string().min(1),
});

export const changePasswordFn = createServerFn({ method: "POST" })
  .validator((raw) => change.parse(raw))
  .handler(async ({ data }) => {
    const token = loadSession();
    if (token === null) {
      return { reason: "GENERAL_ERROR_UNAUTHENTICATED", status: "rejected" } as const;
    }
    try {
      await createClient(UserService, createCtrTransport()).changePassword(
        { newPassword: data.newPassword, oldPassword: data.oldPassword },
        { headers: { cookie: sessionCookieHeader(token) } },
      );
      return { status: "ok" } as const;
    } catch (error) {
      return { reason: normalize(error).reason, status: "rejected" } as const;
    }
  });
