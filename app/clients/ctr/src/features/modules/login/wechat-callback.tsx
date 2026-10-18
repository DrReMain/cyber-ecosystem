import { useNavigate, useSearch } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";
import { m } from "#/paraglide/messages";
import { wechatLoginFn } from "./wechat.fn";

interface CallbackSearch {
  code?: string;
  state?: string;
  next?: string;
}

function safeNext(next: string | undefined): string {
  return typeof next === "string" && next.startsWith("/") ? next : "/profile";
}

// The OAuth landing: wechat redirects here with code+state (plus the next
// hop carried through the redirect_uri query). The exchange and the cookie
// minting both happen in the server fn; this page only drives the states.
export function WechatCallbackPage() {
  const navigate = useNavigate();
  const search = useSearch({ from: "/login/wechat/callback" }) as CallbackSearch;
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
      .then(() => {
        void navigate({ to: safeNext(search.next) });
      })
      .catch(() => {
        setFailed(true);
      });
  }, [navigate, search]);

  if (failed) {
    return (
      <main className="safe-px safe-pt flex min-h-app flex-col items-center justify-center gap-3">
        <p className="text-danger-text">{m.login_error_wechat()}</p>
        <a className="text-ink-secondary text-sm underline" href="/login">
          {m.login_back()}
        </a>
      </main>
    );
  }

  return (
    <main className="safe-px safe-pt flex min-h-app items-center justify-center">
      <p className="text-ink-tertiary">{m.login_callback_pending()}</p>
    </main>
  );
}
