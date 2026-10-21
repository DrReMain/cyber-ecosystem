import { useTransport } from "@connectrpc/connect-query";
import { System } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/error_reason_pb";
import { useQueryClient } from "@tanstack/react-query";
import { Link, useNavigate, useSearch } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";
import { systemReason } from "#/domains/error";
import { m } from "#/paraglide/messages";
import { routeAfterLogin } from "./after-login";
import { wechatLoginFn } from "./wechat.fn";

// The OAuth landing: wechat redirects here with code+state. The exchange
// and the cookie minting both happen in the server fn; this page only
// drives the states, then hands routing to the shared completeness-aware
// post-login router. search.next is validated by the route but never
// arrives today — the redirect_uri carries no query (wechat rejects it),
// so the post-login hop always falls through to the completeness branch;
// the seat stays for a future state-relay.
export function WechatCallbackPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const transport = useTransport();
  const search = useSearch({ from: "/login/wechat/callback" });
  const [failed, setFailed] = useState(false);
  const [serverReason, setServerReason] = useState<string | undefined>(undefined);
  const started = useRef(false);

  useEffect(() => {
    if (started.current) {
      return;
    }
    started.current = true;
    const { code, state } = search;
    if (code === undefined || state === undefined) {
      setFailed(true);
      return;
    }
    wechatLoginFn({ data: { code, state } })
      .then((res) => {
        if (res.status === "rejected") {
          setServerReason(res.reason);
          return;
        }
        return routeAfterLogin({ navigate, queryClient, transport }, search.next);
      })
      .catch(() => {
        setFailed(true);
      });
  }, [navigate, queryClient, search, transport]);

  if (failed || serverReason !== undefined) {
    // The verdict branches per reason like the password leg — a disabled
    // account is a different story from a failed exchange.
    const copy = (() => {
      switch (systemReason(serverReason)) {
        case System.USER_DISABLED:
          return m.error_SYSTEM_USER_DISABLED();
        case System.TENANT_DISABLED:
          return m.error_SYSTEM_TENANT_DISABLED();
        default:
          return m.login_error_wechat();
      }
    })();
    return (
      <main className="safe-px safe-pt flex min-h-app flex-col items-center justify-center gap-3">
        <p className="text-danger-text">{copy}</p>
        <Link className="text-ink-secondary text-sm underline" to="/login">
          {m.login_back()}
        </Link>
      </main>
    );
  }

  return (
    <main className="safe-px safe-pt flex min-h-app items-center justify-center">
      <p className="text-ink-tertiary">{m.login_callback_pending()}</p>
    </main>
  );
}
