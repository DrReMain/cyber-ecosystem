import { useTransport } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { Link, useNavigate, useSearch } from "@tanstack/react-router";
import { Zap } from "lucide-react";
import { useEffect, useState } from "react";
import { Button, Checkbox, Form, Input, Label, TextField } from "react-aria-components";
import { toast } from "sonner";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { m } from "#/paraglide/messages";
import { getLocale } from "#/paraglide/runtime";
import { routeAfterLogin } from "./after-login";
import { loginFn } from "./login.fn";
import { AgreeBox } from "./ui/agree-box";
import { beginWechatLoginFn } from "./wechat.fn";

type FieldError = "bad" | "short" | "generic" | null;

const inputClass =
  "w-full rounded-sm border border-line bg-surface-input px-3 py-3 text-ink outline-none focus:border-focus";

export function LoginPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const transport = useTransport();
  const search = useSearch({ from: "/login/" });
  const [pending, setPending] = useState(false);
  const [wechatPending, setWechatPending] = useState(false);
  const [wechatEnv, setWechatEnv] = useState(false);
  const [error, setError] = useState<FieldError>(null);
  const [agree, setAgree] = useState({ privacy: false, user: false });
  const [missing, setMissing] = useState({ privacy: false, user: false });

  // WeChat-only login env: snsapi authorize needs the WeChat browser. The
  // check runs post-mount — SSR renders the same dimmed state, so hydration
  // never mismatches.
  useEffect(() => {
    setWechatEnv(/MicroMessenger/i.test(navigator.userAgent));
  }, []);

  function agreeOK(): boolean {
    const miss = {
      privacy: !agree.privacy,
      user: !agree.user,
    };
    setMissing(miss);
    if (miss.privacy || miss.user) {
      const joiner = getLocale() === "zh-CN" ? "和" : "and";
      const names = [
        miss.user ? m.login_agree_user_name() : null,
        miss.privacy ? m.login_agree_privacy_name() : null,
      ]
        .filter((name) => name !== null)
        .join(` ${joiner} `);
      toast.error(m.login_agree_missing({ names }));
      return false;
    }
    return true;
  }

  function agreeRowClass(field: "user" | "privacy"): string {
    return `flex items-start gap-2 py-1.5 ${missing[field] ? "text-danger-text" : "text-ink-secondary"}`;
  }

  async function onSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!agreeOK()) {
      return;
    }
    const form = new FormData(event.currentTarget);
    const password = String(form.get("password") ?? "");
    if (password.length < 6) {
      setError("short");
      return;
    }
    setPending(true);
    setError(null);
    try {
      await loginFn({
        data: {
          account: String(form.get("account") ?? ""),
          password,
        },
      });
      toast.success(m.login_success());
      await routeAfterLogin({ navigate, queryClient, transport }, search.next);
    } catch {
      setError("bad");
    } finally {
      setPending(false);
    }
  }

  async function onWechat() {
    if (!wechatEnv) {
      toast.error(m.login_error_wechat_env());
      return;
    }
    if (!agreeOK()) {
      return;
    }
    setWechatPending(true);
    setError(null);
    try {
      // No query on the redirect_uri: WeChat's authorize validation rejects
      // embedded queries, so the post-login hop is always derived afresh.
      const redirectUri = `${window.location.origin}/login/wechat/callback`;
      const res = await beginWechatLoginFn({ data: { redirectUri } });
      window.location.href = res.authorizeUrl;
    } catch {
      setWechatPending(false);
      setError("generic");
    }
  }

  const errorCopy =
    error === "bad"
      ? m.login_error_bad_credentials()
      : error === "short"
        ? m.login_password_short()
        : error === "generic"
          ? m.login_error_generic()
          : null;

  return (
    <AmbientLayout dense>
      <div className="flex w-full flex-col gap-4">
        <div className="flex items-center gap-3">
          <div className="flex h-11 w-11 items-center justify-center rounded-xl bg-cta-primary">
            <Zap aria-hidden className="text-white" size={22} />
          </div>
          <div>
            <div className="font-extrabold text-ink text-xl">{m.login_brand_name()}</div>
            <div className="mt-0.5 text-[11.5px] text-ink-tertiary">{m.login_brand_tagline()}</div>
          </div>
        </div>

        {/* Segment keeps the prototype's two-mode shape; the SMS tab is a
            disabled slot until the system SMS contract exists. */}
        <div className="flex overflow-hidden rounded-sm border border-line text-sm">
          <div className="flex-1 bg-surface-content py-2.5 text-center font-bold text-ink">
            {m.login_tab_password()}
          </div>
          <button
            className="flex-1 py-2.5 text-center text-ink-tertiary/60 active:opacity-60"
            onClick={() => toast.info(m.common_soon())}
            type="button"
          >
            {m.login_tab_code()}
          </button>
        </div>

        <Form className="flex flex-col gap-4" onSubmit={onSubmit}>
          <TextField autoComplete="username" isRequired name="account" type="text">
            <Label className="sr-only">{m.login_account_label()}</Label>
            <Input className={inputClass} placeholder={m.login_account_placeholder()} />
          </TextField>
          <TextField autoComplete="current-password" isRequired name="password" type="password">
            <Label className="sr-only">{m.login_password_label()}</Label>
            <Input
              className={inputClass}
              maxLength={20}
              placeholder={m.login_password_placeholder()}
            />
          </TextField>
          {errorCopy !== null && (
            <p className="text-danger-text text-sm" role="alert">
              {errorCopy}
            </p>
          )}

          <div className="flex flex-col gap-1.5">
            <Checkbox
              isSelected={agree.user}
              onChange={(selected) => {
                setAgree((prev) => ({ ...prev, user: selected }));
                setMissing((prev) => ({ ...prev, user: false }));
              }}
            >
              {() => (
                <div className={agreeRowClass("user")}>
                  <AgreeBox checked={agree.user} />
                  <span className="text-sm leading-snug">
                    {m.login_agree_prefix()}
                    {/* The tap must navigate without flipping the checkbox. */}
                    <Link
                      className="underline active:opacity-60"
                      onClick={(event) => event.stopPropagation()}
                      to="/agreement/user"
                    >
                      {m.login_agree_user_name()}
                    </Link>
                  </span>
                </div>
              )}
            </Checkbox>
            <Checkbox
              isSelected={agree.privacy}
              onChange={(selected) => {
                setAgree((prev) => ({ ...prev, privacy: selected }));
                setMissing((prev) => ({ ...prev, privacy: false }));
              }}
            >
              {() => (
                <div className={agreeRowClass("privacy")}>
                  <AgreeBox checked={agree.privacy} />
                  <span className="text-sm leading-snug">
                    {m.login_agree_prefix()}
                    <Link
                      className="underline active:opacity-60"
                      onClick={(event) => event.stopPropagation()}
                      to="/agreement/privacy"
                    >
                      {m.login_agree_privacy_name()}
                    </Link>
                  </span>
                </div>
              )}
            </Checkbox>
            {/* Select-all rides the user row's line: one tap covers both. */}
            <div className="flex items-start justify-between gap-2.5">
              <span className="ml-6 text-[10.5px] text-ink-tertiary leading-snug">
                {m.login_agree_minor_note()}
              </span>
              <Checkbox
                isIndeterminate={agree.user !== agree.privacy}
                isSelected={agree.user && agree.privacy}
                onChange={(selected) => {
                  setAgree({ privacy: selected, user: selected });
                  setMissing({ privacy: false, user: false });
                }}
              >
                {() => (
                  <div className="flex items-center gap-2 whitespace-nowrap font-bold text-ink-secondary">
                    <AgreeBox
                      checked={agree.user && agree.privacy}
                      indeterminate={agree.user !== agree.privacy}
                    />
                    <span className="text-sm">{m.login_agree_all()}</span>
                  </div>
                )}
              </Checkbox>
            </div>
          </div>

          <Button
            className="rounded-sm bg-linear-to-r from-[#2563eb] to-[#06b6d4] px-4 py-3 font-semibold text-white disabled:opacity-50"
            isDisabled={pending}
            type="submit"
          >
            {pending ? m.login_pending() : m.login_submit()}
          </Button>
        </Form>

        <Button
          className={`rounded-sm bg-wechat px-4 py-3 font-semibold text-white disabled:opacity-50 ${
            wechatEnv ? "" : "opacity-55"
          }`}
          isDisabled={wechatPending}
          onPress={onWechat}
        >
          {wechatPending ? m.login_pending() : m.login_wechat_button()}
        </Button>
        {wechatEnv || (
          <p className="text-center text-[11px] text-ink-tertiary">{m.login_wechat_env_note()}</p>
        )}

        {/* Forgot-password stays contract-gated (SMS provider). */}
        <div className="text-center text-[10.5px] text-ink-tertiary leading-relaxed">
          <button
            className="tap-zone py-1 underline active:opacity-60"
            onClick={() => toast.info(m.common_soon())}
            type="button"
          >
            {m.login_forgot()}
          </button>
          <span className="mx-1.5">·</span>
          <Link className="tap-zone py-1 underline active:opacity-60" to="/agreement/user">
            {m.login_agree_user_name()}
          </Link>
          <span className="mx-1.5">·</span>
          <Link className="tap-zone py-1 underline active:opacity-60" to="/agreement/privacy">
            {m.login_agree_privacy_name()}
          </Link>
        </div>
        <div className="text-center text-[10.5px] text-ink-tertiary leading-loose">
          {m.login_footer_company()}
          <br />
          {m.login_footer_icp()}
          <br />
          {m.login_footer_email()}
        </div>
      </div>
    </AmbientLayout>
  );
}
