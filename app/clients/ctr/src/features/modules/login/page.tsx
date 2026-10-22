import { useTransport } from "@connectrpc/connect-query";
import { System } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/error_reason_pb";
import { useQueryClient } from "@tanstack/react-query";
import { Link, useNavigate, useSearch } from "@tanstack/react-router";
import { Zap } from "lucide-react";
import { type SyntheticEvent, useEffect, useRef, useState } from "react";
import {
  Button,
  CheckboxButton,
  CheckboxField,
  Form,
  Input,
  Label,
  TextField,
} from "react-aria-components";
import { toast } from "sonner";
import { systemReason } from "#/domains/error";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { m } from "#/paraglide/messages";
import { getLocale } from "#/paraglide/runtime";
import { routeAfterLogin } from "./after-login";
import { loginFn } from "./login.fn";
import { AgreeBox } from "./ui/agree-box";
import { beginWechatLoginFn } from "./wechat.fn";

type FieldError = "short" | "generic" | null;

// The login verdict branches per reason — wrong password, no seat at this
// client's door, a disabled account, a disabled tenant are different stories
// with different next actions; none of them collapses into "bad credentials".
function failureCopy(error: FieldError, reason: string | undefined): string | null {
  if (error === "short") return m.login_password_short();
  if (error === "generic") return m.login_error_generic();
  switch (systemReason(reason)) {
    case System.LOGIN_FAILED:
      return m.login_error_bad_credentials();
    case System.LOGIN_SEAT_REQUIRED:
      return m.error_SYSTEM_LOGIN_SEAT_REQUIRED();
    case System.USER_DISABLED:
      return m.error_SYSTEM_USER_DISABLED();
    case System.TENANT_DISABLED:
      return m.error_SYSTEM_TENANT_DISABLED();
    case undefined:
      return null;
    default:
      return m.login_error_generic();
  }
}

const inputClass =
  "w-full rounded-sm border border-line bg-surface-input px-3 py-3 text-ink outline-none focus:border-focus";

export function LoginPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const transport = useTransport();
  const { expired, next } = useSearch({ from: "/login/" });
  const [pending, setPending] = useState(false);
  const [wechatPending, setWechatPending] = useState(false);
  const [wechatEnv, setWechatEnv] = useState(false);
  const [error, setError] = useState<FieldError>(null);
  const [serverReason, setServerReason] = useState<string | undefined>(undefined);
  const formRef = useRef<HTMLFormElement>(null);
  const [agree, setAgree] = useState({ privacy: false, user: false });
  const [missing, setMissing] = useState({ privacy: false, user: false });

  // WeChat-only login env: snsapi authorize needs the WeChat browser. The
  // check runs post-mount — SSR renders the same dimmed state, so hydration
  // never mismatches.
  useEffect(() => {
    setWechatEnv(/MicroMessenger/i.test(navigator.userAgent));
  }, []);

  // A session killed server-side (ban, revoke) lands here with the notice;
  // the param strips once fired so a refresh or shared link never re-toasts.
  useEffect(() => {
    if (!expired) {
      return;
    }
    toast(m.login_session_expired(), { id: "session-expired" });
    void navigate({
      replace: true,
      search: (prev) => ({ next: prev.next }),
      to: "/login",
      viewTransition: false,
    });
  }, [expired, navigate]);

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

  async function onSubmit(event: SyntheticEvent<HTMLFormElement>) {
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
    setServerReason(undefined);
    try {
      const res = await loginFn({
        data: {
          account: String(form.get("account") ?? ""),
          password,
        },
      });
      if (res.status === "rejected") {
        setServerReason(res.reason);
        return;
      }
      toast.success(m.login_success());
      await routeAfterLogin({ navigate, queryClient, transport }, next);
    } catch {
      setError("generic");
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

  const errorCopy = failureCopy(error, serverReason);

  return (
    <AmbientLayout dense>
      <div className="flex w-full flex-col gap-4">
        <div className="flex items-center gap-3">
          <div className="bg-cta-primary flex h-11 w-11 items-center justify-center rounded-xl">
            <Zap aria-hidden className="text-white" size={22} />
          </div>
          <div>
            <div className="text-ink text-xl font-extrabold">{m.login_brand_name()}</div>
            <div className="text-ink-tertiary mt-0.5 text-[11.5px]">{m.login_brand_tagline()}</div>
          </div>
        </div>

        {/* Segment keeps the prototype's two-mode shape; the SMS tab is a
            disabled slot until the system SMS contract exists. */}
        <div className="border-line flex overflow-hidden rounded-sm border text-sm">
          <div className="bg-surface-content text-ink flex-1 py-2.5 text-center font-bold">
            {m.login_tab_password()}
          </div>
          <button
            className="text-ink-tertiary/60 flex-1 py-2.5 text-center active:opacity-60"
            onClick={() => toast.info(m.common_soon())}
            type="button"
          >
            {m.login_tab_code()}
          </button>
        </div>

        <Form className="flex flex-col gap-4" onSubmit={onSubmit} ref={formRef}>
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
            <div className="flex items-start justify-between gap-2" role="alert">
              <p className="text-danger-text text-sm leading-snug">{errorCopy}</p>
              {systemReason(serverReason) === System.LOGIN_SEAT_REQUIRED && (
                <button
                  className="tap-zone text-ink-secondary shrink-0 font-bold underline active:opacity-60"
                  onClick={() => {
                    formRef.current?.reset();
                    setServerReason(undefined);
                    formRef.current
                      ?.querySelector<HTMLInputElement>("input[name='account']")
                      ?.focus();
                  }}
                  type="button"
                >
                  {m.login_error_switch_account()}
                </button>
              )}
            </div>
          )}

          <div className="flex flex-col gap-1.5">
            <CheckboxField
              isSelected={agree.user}
              onChange={(selected) => {
                setAgree((prev) => ({ ...prev, user: selected }));
                setMissing((prev) => ({ ...prev, user: false }));
              }}
            >
              <CheckboxButton>
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
              </CheckboxButton>
            </CheckboxField>
            <CheckboxField
              isSelected={agree.privacy}
              onChange={(selected) => {
                setAgree((prev) => ({ ...prev, privacy: selected }));
                setMissing((prev) => ({ ...prev, privacy: false }));
              }}
            >
              <CheckboxButton>
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
              </CheckboxButton>
            </CheckboxField>
            {/* Select-all rides the user row's line: one tap covers both. */}
            <div className="flex items-start justify-between gap-2.5">
              <span className="text-ink-tertiary ml-6 text-[10.5px] leading-snug">
                {m.login_agree_minor_note()}
              </span>
              <CheckboxField
                isIndeterminate={agree.user !== agree.privacy}
                isSelected={agree.user && agree.privacy}
                onChange={(selected) => {
                  setAgree({ privacy: selected, user: selected });
                  setMissing({ privacy: false, user: false });
                }}
              >
                <CheckboxButton>
                  <div className="text-ink-secondary flex items-center gap-2 font-bold whitespace-nowrap">
                    <AgreeBox
                      checked={agree.user && agree.privacy}
                      indeterminate={agree.user !== agree.privacy}
                    />
                    <span className="text-sm">{m.login_agree_all()}</span>
                  </div>
                </CheckboxButton>
              </CheckboxField>
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
          className={`bg-wechat rounded-sm px-4 py-3 font-semibold text-white disabled:opacity-50 ${
            wechatEnv ? "" : "opacity-55"
          }`}
          isDisabled={wechatPending}
          onPress={onWechat}
        >
          {wechatPending ? m.login_pending() : m.login_wechat_button()}
        </Button>

        {/* Forgot-password stays contract-gated (SMS provider). */}
        <div className="text-ink-tertiary text-center text-[10.5px] leading-relaxed">
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
        <div className="text-ink-tertiary text-center text-[10.5px] leading-loose">
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
