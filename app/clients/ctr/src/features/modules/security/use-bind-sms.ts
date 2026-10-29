import { useEffect, useState } from "react";
import { toast } from "sonner";
import { mainlandMobile, phoneLiveInvalid } from "#/features/modules/login/sms";
import { m } from "#/paraglide/messages";
import { bindSmsCredentialFn, sendBindSmsCodeFn } from "./security.fn";

// The bind-phone leg's state machine, the useSmsLogin sibling: the same
// field verdicts and 60s window mirror, but the send is preview-gated (an
// occupied phone never spends an SMS) and the submit binds instead of
// logging in. The verdict is always data — never a thrown error. The
// optional seed carries a phone the flow already knows (the merge confirm
// page re-uses the number the bind attempt typed).
export function useBindSms(seed = "") {
  const [phone, setPhoneValue] = useState(seed);
  const [phoneErr, setPhoneErr] = useState(false);
  const [code, setCodeValue] = useState("");
  const [codeErr, setCodeErr] = useState(false);
  const [state, setState] = useState("");
  const [countdown, setCountdown] = useState(0);
  const [sending, setSending] = useState(false);
  const [pending, setPending] = useState(false);
  const [reason, setReason] = useState<string | undefined>(undefined);
  const [failed, setFailed] = useState(false);

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

  // The caller gates the send on the preview verdict (occupied phones never
  // reach here), so this send is only ever the plain-bind path's.
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
      const res = await sendBindSmsCodeFn({ data: { phone } });
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

  async function submit(): Promise<"ok" | "rejected" | "failed"> {
    if (!mainlandMobile.test(phone)) {
      setPhoneErr(true);
      return "rejected";
    }
    if (!/^\d{6}$/.test(code)) {
      setCodeErr(true);
      return "rejected";
    }
    setPending(true);
    setReason(undefined);
    setFailed(false);
    try {
      const res = await bindSmsCredentialFn({ data: { code, phone, state } });
      if (res.status === "rejected") {
        setReason(res.reason);
        return "rejected";
      }
      return "ok";
    } catch {
      setFailed(true);
      return "failed";
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
    send,
    sending,
    setCode,
    setPhone,
    setState,
    state,
    submit,
  };
}

export type BindSms = ReturnType<typeof useBindSms>;
