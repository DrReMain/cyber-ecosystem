import { useEffect, useState } from "react";
import { toast } from "sonner";
import { mainlandMobile } from "#/features/modules/login/sms";
import { sendSmsCodeFn } from "#/features/modules/login/sms.fn";
import { m } from "#/paraglide/messages";

// The bare send-code half of the login hook: mint the public one-shot
// state and hold the 60s window mirror. Consumers own what the exchange
// proves (password reset here, login on the login page).
export function useSmsCode(seedPhone = "") {
  const [phone, setPhoneValue] = useState(seedPhone);
  const [code, setCodeValue] = useState("");
  const [state, setState] = useState("");
  const [countdown, setCountdown] = useState(0);
  const [sending, setSending] = useState(false);
  const [failed, setFailed] = useState(false);

  useEffect(() => {
    if (countdown <= 0) {
      return;
    }
    const id = window.setTimeout(() => setCountdown((left) => left - 1), 1000);
    return () => window.clearTimeout(id);
  }, [countdown]);

  async function send() {
    if (countdown > 0 || sending) {
      return;
    }
    if (!mainlandMobile.test(phone)) {
      return;
    }
    setSending(true);
    setFailed(false);
    try {
      const res = await sendSmsCodeFn({ data: { phone } });
      if (res.status === "rejected") {
        toast.error(
          res.reason === "SYSTEM_SMS_RATE_LIMITED"
            ? m.error_SYSTEM_SMS_RATE_LIMITED()
            : m.security_error_generic(),
        );
        return;
      }
      setState(res.state);
      toast.success(m.security_code_sent());
      setCountdown(60);
    } catch {
      setFailed(true);
    } finally {
      setSending(false);
    }
  }

  return {
    code,
    countdown,
    failed,
    phone,
    send,
    sending,
    setCode: (raw: string) => setCodeValue(raw.replace(/\D/g, "").slice(0, 6)),
    setPhone: (raw: string) => setPhoneValue(raw.replace(/\D/g, "").slice(0, 11)),
    setState,
    state,
  };
}
