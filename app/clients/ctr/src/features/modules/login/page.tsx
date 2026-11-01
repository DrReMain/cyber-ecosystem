import { useTransport } from "@connectrpc/connect-query";
import { AgreementLoginMethod } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/agreement_pb";
import { System } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/error_reason_pb";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { type SyntheticEvent, useEffect, useRef, useState } from "react";
import { Button, Form } from "react-aria-components";
import { toast } from "sonner";
import { systemReason } from "#/domains/error";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { m } from "#/paraglide/messages";
import { getLocale } from "#/paraglide/runtime";
import { routeAfterLogin } from "./after-login";
import { loginFn } from "./login.fn";
import { smsFailureCopy, useSmsLogin } from "./sms";
import { AgreeGate } from "./ui/agree-gate";
import { BrandLockup } from "./ui/brand-lockup";
import { FormErrorLine } from "./ui/form-error-line";
import { LoginFooter } from "./ui/login-footer";
import { ModeSegment } from "./ui/mode-segment";
import { PasswordFields } from "./ui/password-fields";
import { SmsFields } from "./ui/sms-fields";
import { WechatButton } from "./ui/wechat-button";

type FieldError = "short" | "generic" | null;

// The password leg's verdict branches per reason — wrong password, no seat
// at this client's door, a disabled account, a disabled tenant are different
// stories with different next actions; none of them collapses into "bad
// credentials".
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

export function LoginPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const transport = useTransport();
  const { expired, next } = useSearch({ from: "/login/" });
  const [pending, setPending] = useState(false);
  const [wechatEnv, setWechatEnv] = useState(false);
  const [error, setError] = useState<FieldError>(null);
  const [serverReason, setServerReason] = useState<string | undefined>(undefined);
  const formRef = useRef<HTMLFormElement>(null);
  const [agree, setAgree] = useState({ privacy: false, user: false });
  const [missing, setMissing] = useState({ privacy: false, user: false });
  const [mode, setMode] = useState<"code" | "password">("code");

  const sms = useSmsLogin(async () => {
    toast.success(m.login_success());
    await routeAfterLogin(
      { navigate, queryClient, transport },
      next,
      AgreementLoginMethod.SMS_CODE,
    );
  });

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

  // The consent gate's verdict — every leg (submit, code send, wechat)
  // refuses to fire while a box is still empty.
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

  function switchMode(target: "code" | "password") {
    setMode(target);
    setError(null);
    setServerReason(undefined);
    sms.resetVerdicts();
  }

  // The send also walks the agree gate (the prototype's rule: 发送验证码也
  // 校验); the hook owns the phone verdict and the server's 60s window.
  function onSendCode() {
    if (!agreeOK()) {
      return;
    }
    void sms.send();
  }

  async function submitCode() {
    if (sms.state === "") {
      toast(m.login_code_send_first());
      return;
    }
    await sms.submit();
  }

  async function onSubmit(event: SyntheticEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!agreeOK()) {
      return;
    }
    if (mode === "code") {
      await submitCode();
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
      await routeAfterLogin(
        { navigate, queryClient, transport },
        next,
        AgreementLoginMethod.PASSWORD,
      );
    } catch {
      setError("generic");
    } finally {
      setPending(false);
    }
  }

  function codeModeCopy(): string | null {
    if (sms.failed) {
      return m.login_error_generic();
    }
    if (sms.phoneErr) {
      return m.login_error_sms_phone();
    }
    if (sms.codeErr) {
      return m.login_error_sms_code_short();
    }
    return smsFailureCopy(sms.reason);
  }

  const errorCopy = mode === "code" ? codeModeCopy() : failureCopy(error, serverReason);
  const busy = pending || sms.pending || sms.sending;

  return (
    <AmbientLayout dense>
      <div className="flex w-full flex-col gap-3.5 pt-7">
        <BrandLockup />

        <ModeSegment mode={mode} onSwitch={switchMode} />

        <Form className="flex flex-col gap-4" onSubmit={onSubmit} ref={formRef}>
          {mode === "code" ? <SmsFields onSend={onSendCode} sms={sms} /> : <PasswordFields />}
          <FormErrorLine
            copy={errorCopy}
            mode={mode}
            onSwitchAccount={() => {
              formRef.current?.reset();
              setServerReason(undefined);
              formRef.current?.querySelector<HTMLInputElement>("input[name='account']")?.focus();
            }}
            serverReason={serverReason}
          />

          <AgreeGate
            agree={agree}
            missing={missing}
            onAll={(selected) => {
              setAgree({ privacy: selected, user: selected });
              setMissing({ privacy: false, user: false });
            }}
            onPrivacy={(selected) => {
              setAgree((prev) => ({ ...prev, privacy: selected }));
              setMissing((prev) => ({ ...prev, privacy: false }));
            }}
            onUser={(selected) => {
              setAgree((prev) => ({ ...prev, user: selected }));
              setMissing((prev) => ({ ...prev, user: false }));
            }}
          />

          <Button
            className="bg-cta-primary shadow-cta-primary rounded-control px-4 py-3.5 text-[15px] font-bold text-white disabled:opacity-50"
            isDisabled={busy}
            type="submit"
          >
            {busy ? m.login_pending() : m.login_submit()}
          </Button>
        </Form>

        <WechatButton agreeOK={agreeOK} onError={() => setError("generic")} wechatEnv={wechatEnv} />

        {/* Forgot-password stays contract-gated (SMS provider). */}
        <LoginFooter />
      </div>
    </AmbientLayout>
  );
}
