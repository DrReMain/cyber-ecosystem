import { System } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/error_reason_pb";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { systemReason } from "#/domains/error";
import { m } from "#/paraglide/messages";
import { sendSmsCodeFn, smsLoginFn } from "./sms.fn";

export const mainlandMobile = /^1[3-9]\d{9}$/;

// The prototype's two-gate live check: a wrong carrier prefix is verdictable
// from the third digit on; the full shape re-checks at eleven. A short but
// prefix-plausible number only fails on blur or submit.
export function phoneLiveInvalid(value: string): boolean {
  if (value.length >= 3 && !/^1[3-9]/.test(value)) {
    return true;
  }
  return value.length === 11 && !mainlandMobile.test(value);
}

// The sms leg's reason branches: a wrong or burned code and a consumed state
// both mean "send again", the send limiter has its own copy, and the seat
// stories the password leg already tells stay shared. A reason outside the
// System family (an infra-class channel verdict, e.g. a rejected message
// shape) must still say something — never vanish as "no error".
export function smsFailureCopy(reason: string | undefined): string | null {
  const sys = systemReason(reason);
  if (sys === undefined) {
    return reason === undefined ? null : m.login_error_generic();
  }
  switch (sys) {
    case System.SMS_CODE_INVALID:
      return m.login_error_sms_code();
    case System.IDENTITY_EXCHANGE_FAILED:
      return m.login_error_sms_expired();
    case System.SMS_RATE_LIMITED:
      return m.login_error_sms_rate();
    case System.LOGIN_SEAT_REQUIRED:
      return m.error_SYSTEM_LOGIN_SEAT_REQUIRED();
    case System.USER_DISABLED:
      return m.error_SYSTEM_USER_DISABLED();
    default:
      return m.login_error_generic();
  }
}

// The sms tab's whole state machine: phone/code field verdicts, the send
// countdown (the client's mirror of the server's 60s window — it starts only
// on an accepted send), the one-shot state riding between send and exchange,
// and the submit. The agree gate and post-login routing stay with the page;
// onOk receives them as one callback.
export function useSmsLogin(onOk: () => Promise<void>) {
  const [phone, setPhoneValue] = useState("");
  const [phoneErr, setPhoneErr] = useState(false);
  const [code, setCodeValue] = useState("");
  const [codeErr, setCodeErr] = useState(false);
  const [state, setState] = useState("");
  const [countdown, setCountdown] = useState(0);
  const [sending, setSending] = useState(false);
  const [pending, setPending] = useState(false);
  const [reason, setReason] = useState<string | undefined>(undefined);
  const [failed, setFailed] = useState(false);

  // One timeout per tick, cleaned up on change and unmount — the updater
  // stays pure, no interval lives outside the render cycle.
  useEffect(() => {
    if (countdown <= 0) {
      return;
    }
    const id = window.setTimeout(() => setCountdown((left) => left - 1), 1000);
    return () => window.clearTimeout(id);
  }, [countdown]);

  function setPhone(raw: string) {
    const digits = raw.replace(/\D/g, "").slice(0, 11);
    setPhoneValue(digits);
    setPhoneErr(phoneLiveInvalid(digits));
    // The countdown mirrors the current phone's server window; a different
    // complete number is a different window — the server is the real
    // limiter, the client never holds a new number hostage to the old one's
    // countdown.
    if (digits.length === 11 && digits !== phone) {
      setCountdown(0);
    }
  }

  function blurPhone() {
    if (phone !== "") {
      setPhoneErr(!mainlandMobile.test(phone));
    }
  }

  function setCode(raw: string) {
    setCodeValue(raw.replace(/\D/g, "").slice(0, 6));
    setCodeErr(false);
  }

  function resetVerdicts() {
    setPhoneErr(false);
    setCodeErr(false);
    setReason(undefined);
    setFailed(false);
  }

  async function send() {
    if (countdown > 0 || sending) {
      return;
    }
    if (!mainlandMobile.test(phone)) {
      setPhoneErr(true);
      return;
    }
    setSending(true);
    setReason(undefined);
    setFailed(false);
    try {
      const res = await sendSmsCodeFn({ data: { phone } });
      if (res.status === "rejected") {
        setReason(res.reason);
        return;
      }
      setState(res.state);
      toast.success(m.login_code_sent());
      setCountdown(60);
    } catch {
      setFailed(true);
    } finally {
      setSending(false);
    }
  }

  async function submit() {
    if (!mainlandMobile.test(phone)) {
      setPhoneErr(true);
      return;
    }
    if (!/^\d{6}$/.test(code)) {
      setCodeErr(true);
      return;
    }
    setPending(true);
    setReason(undefined);
    setFailed(false);
    try {
      const res = await smsLoginFn({ data: { code, phone, state } });
      if (res.status === "rejected") {
        setReason(res.reason);
        return;
      }
      await onOk();
    } catch {
      setFailed(true);
    } finally {
      setPending(false);
    }
  }

  return {
    blurPhone,
    code,
    codeErr,
    countdown,
    failed,
    pending,
    phone,
    phoneErr,
    reason,
    resetVerdicts,
    send,
    sending,
    setCode,
    setPhone,
    state,
    submit,
  };
}

export type SmsLogin = ReturnType<typeof useSmsLogin>;
