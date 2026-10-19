import { useTransport } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { Link, useNavigate, useSearch } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";
import { m } from "#/paraglide/messages";
import { routeAfterLogin } from "./after-login";
import { wechatLoginFn } from "./wechat.fn";

// The OAuth landing: wechat redirects here with code+state. The exchange
// and the cookie minting both happen in the server fn; this page only
// drives the states, then hands routing to the shared completeness-aware
// post-login router. search.next is validated by the route but never
// arrives today — the redirect_uri carries no query (wechat rejects it),
// so the post-login hop always falls through to the completeness branch;
// the seat stays for a future state-relay (roadmap: Review 遗留转交).
export function WechatCallbackPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const transport = useTransport();
  const search = useSearch({ from: "/login/wechat/callback" });
  const [failed, setFailed] = useState(false);
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
      .then(() => routeAfterLogin({ navigate, queryClient, transport }, search.next))
      .catch(() => {
        setFailed(true);
      });
  }, [navigate, queryClient, search, transport]);

  if (failed) {
    return (
      <main className="safe-px safe-pt flex min-h-app flex-col items-center justify-center gap-3">
        <p className="text-danger-text">{m.login_error_wechat()}</p>
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
