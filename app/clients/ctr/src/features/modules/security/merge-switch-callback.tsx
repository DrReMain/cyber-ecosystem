import { System } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/error_reason_pb";
import { useQueryClient } from "@tanstack/react-query";
import { Link, useNavigate, useSearch } from "@tanstack/react-router";
import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";
import { systemReason } from "#/domains/error";
import { sessionQuery } from "#/features/auth/session.fn";
import { inputClass } from "#/features/modules/login/ui/sms-fields";
import { wechatLoginFn } from "#/features/modules/login/wechat.fn";
import { m } from "#/paraglide/messages";
import { mergeFailureCopy } from "./merge-failure";
import { mergeSmsFn, sendBindSmsCodeFn } from "./security.fn";

const sendBtnClass =
  "border-focus text-link bg-transparent rounded-control w-28 shrink-0 border px-3 py-3.25 text-[13px] font-bold active:opacity-60 disabled:opacity-60";

// The switch flow's tail: the wechat login switched the session to the rich
// account; this page collects the poor side's phone code and fires the
// merge from the surviving session. The poor phone number is typed, never
// carried — the client only ever held the masked form.
export function MergeSwitchCallbackPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const search = useSearch({ from: "/profile/security/merge/wechat-callback" });
  const started = useRef(false);
  const [phase, setPhase] = useState<"switching" | "proving" | "failed">("switching");
  const [phone, setPhone] = useState("");
  const [code, setCode] = useState("");
  const [state, setState] = useState("");
  const [countdown, setCountdown] = useState(0);
  const [sending, setSending] = useState(false);
  const [merging, setMerging] = useState(false);
  const [mergeReason, setMergeReason] = useState<string | undefined>(undefined);

  useEffect(() => {
    if (countdown <= 0) {
      return;
    }
    const id = window.setTimeout(() => setCountdown((left) => left - 1), 1000);
    return () => window.clearTimeout(id);
  }, [countdown]);

  useEffect(() => {
    if (started.current) {
      return;
    }
    started.current = true;
    const { code: wxCode, state: wxState } = search;
    if (wxCode === undefined || wxState === undefined) {
      setPhase("failed");
      return;
    }
    void wechatLoginFn({ data: { code: wxCode, state: wxState } }).then(async (res) => {
      if (res.status === "rejected") {
        setPhase("failed");
        return;
      }
      // The session identity changed: every cached read of the old seat is
      // stale, the same law the post-login router enforces.
      await queryClient.invalidateQueries();
      void queryClient.refetchQueries({ queryKey: sessionQuery.queryKey });
      setPhase("proving");
    });
  }, [queryClient, search]);

  async function onSend() {
    if (!/^1[3-9]\d{9}$/.test(phone) || countdown > 0 || sending) {
      return;
    }
    setSending(true);
    try {
      const res = await sendBindSmsCodeFn({ data: { phone } });
      if (res.status === "rejected") {
        toast.error(
          systemReason(res.reason) === System.SMS_RATE_LIMITED
            ? m.error_SYSTEM_SMS_RATE_LIMITED()
            : m.security_error_generic(),
        );
        return;
      }
      setState(res.state);
      toast.success(m.security_code_sent());
      setCountdown(60);
    } catch {
      toast.error(m.security_error_generic());
    } finally {
      setSending(false);
    }
  }

  async function onMerge() {
    if (!(/^1[3-9]\d{9}$/.test(phone) && /^\d{6}$/.test(code))) {
      return;
    }
    setMerging(true);
    const res = await mergeSmsFn({ data: { code, phone, state } });
    setMerging(false);
    if (res.status === "rejected") {
      setMergeReason(res.reason);
      return;
    }
    await queryClient.invalidateQueries();
    await navigate({ to: "/profile/security/merge/success", viewTransition: false });
  }

  if (phase === "failed") {
    return (
      <main className="safe-px safe-pt min-h-app flex flex-col items-center justify-center gap-3">
        <p className="text-danger-text">{m.security_merge_switch_failed()}</p>
        <Link className="text-ink-secondary text-sm underline" to="/login">
          {m.security_back()}
        </Link>
      </main>
    );
  }

  if (phase === "switching") {
    return (
      <main className="safe-px safe-pt min-h-app flex items-center justify-center">
        <p className="text-ink-tertiary">{m.security_merge_pending()}</p>
      </main>
    );
  }

  return (
    <main className="safe-px safe-pt min-h-app mx-auto flex w-full max-w-105 flex-col gap-3 px-5 pt-6">
      <div className="border-line-accent-blue bg-identity rounded-card border px-4 py-4">
        <div className="text-ink text-[15px] font-extrabold">{m.security_merge_tail_title()}</div>
        <p className="text-ink-secondary mt-1.5 text-[12px] leading-[1.8]">
          {m.security_merge_tail_pitch()}
        </p>
      </div>
      <input
        aria-label={m.security_phone_label()}
        className={inputClass}
        inputMode="numeric"
        maxLength={11}
        onChange={(event) => setPhone(event.target.value.replace(/\D/g, "").slice(0, 11))}
        placeholder={m.security_bind_phone_placeholder()}
        value={phone}
      />
      <div className="flex items-stretch gap-3">
        <input
          aria-label={m.security_code_label()}
          className={`${inputClass} flex-1 tracking-[4px]`}
          inputMode="numeric"
          maxLength={6}
          onChange={(event) => setCode(event.target.value.replace(/\D/g, "").slice(0, 6))}
          placeholder={m.security_bind_phone_code_placeholder()}
          value={code}
        />
        <button
          className={sendBtnClass}
          disabled={countdown > 0 || sending}
          onClick={() => void onSend()}
          type="button"
        >
          {countdown > 0
            ? m.security_code_resend_in({ seconds: countdown })
            : m.security_code_send()}
        </button>
      </div>
      {mergeReason !== undefined && (
        <p className="text-danger-text text-[11.5px]">{mergeFailureCopy(mergeReason)}</p>
      )}
      <button
        className="rounded-control bg-[linear-gradient(to_right,#EA580C,#F59E0B)] px-4 py-3.5 text-[15px] font-bold text-white shadow-[0_10px_24px_rgba(234,88,12,0.35)] disabled:opacity-50"
        disabled={merging || code.length !== 6 || phone.length !== 11}
        onClick={() => void onMerge()}
        type="button"
      >
        {merging ? m.security_pending() : m.security_merge_confirm_cta()}
      </button>
    </main>
  );
}
