import { useNavigate, useSearch } from "@tanstack/react-router";
import { useState } from "react";
import { Button, Form, Input, Label, TextField } from "react-aria-components";
import { m } from "#/paraglide/messages";
import { loginFn } from "./login.fn";
import { beginWechatLoginFn } from "./wechat.fn";

interface LoginSearch {
  next?: string;
}

function safeNext(next: string | undefined): string {
  return typeof next === "string" && next.startsWith("/") ? next : "/profile";
}

export function LoginPage() {
  const navigate = useNavigate();
  const search = useSearch({ from: "/login/" }) as LoginSearch;
  const [pending, setPending] = useState(false);
  const [wechatPending, setWechatPending] = useState(false);
  const [error, setError] = useState<"credentials" | "generic" | null>(null);

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    setPending(true);
    setError(null);
    try {
      await loginFn({
        data: {
          account: String(form.get("account") ?? ""),
          password: String(form.get("password") ?? ""),
        },
      });
      await navigate({ to: safeNext(search.next) });
    } catch {
      setError("credentials");
    } finally {
      setPending(false);
    }
  }

  async function onWechat() {
    setWechatPending(true);
    setError(null);
    try {
      // No query on the redirect_uri: wechat's authorize validation rejects
      // embedded queries, so the post-login hop is always /profile.
      const redirectUri = `${window.location.origin}/login/wechat/callback`;
      const res = await beginWechatLoginFn({ data: { redirectUri } });
      window.location.href = res.authorizeUrl;
    } catch {
      setWechatPending(false);
      setError("generic");
    }
  }

  return (
    <main className="safe-px safe-pt flex min-h-app flex-col items-center justify-center">
      <div className="flex w-full max-w-sm flex-col gap-4 rounded-lg border border-line bg-surface-card p-5">
        <h1 className="font-bold text-ink text-xl">{m.login_title()}</h1>
        <Button
          className="rounded-sm bg-[#07c160] px-4 py-3 font-semibold text-white disabled:opacity-50"
          isDisabled={wechatPending}
          onPress={onWechat}
        >
          {wechatPending ? m.login_pending() : m.login_wechat_button()}
        </Button>
        <div className="flex items-center gap-3">
          <span className="flex-1 border-line-soft border-t" />
          <span className="text-ink-tertiary text-xs">{m.login_divider()}</span>
          <span className="flex-1 border-line-soft border-t" />
        </div>
        <Form className="flex flex-col gap-4" onSubmit={onSubmit}>
          <TextField autoComplete="username" isRequired name="account" type="text">
            <Label className="mb-1 block text-ink-secondary text-sm">
              {m.login_account_label()}
            </Label>
            <Input className="w-full rounded-sm border border-line bg-surface-input px-3 py-2.5 text-ink outline-none focus:border-focus" />
          </TextField>
          <TextField autoComplete="current-password" isRequired name="password" type="password">
            <Label className="mb-1 block text-ink-secondary text-sm">
              {m.login_password_label()}
            </Label>
            <Input className="w-full rounded-sm border border-line bg-surface-input px-3 py-2.5 text-ink outline-none focus:border-focus" />
          </TextField>
          {error !== null && (
            <p className="text-danger-text text-sm">
              {error === "credentials" ? m.login_error_bad_credentials() : m.login_error_generic()}
            </p>
          )}
          <Button
            className="rounded-sm bg-linear-to-r from-[#2563eb] to-[#06b6d4] px-4 py-3 font-semibold text-white disabled:opacity-50"
            isDisabled={pending}
            type="submit"
          >
            {pending ? m.login_pending() : m.login_submit()}
          </Button>
        </Form>
      </div>
    </main>
  );
}
