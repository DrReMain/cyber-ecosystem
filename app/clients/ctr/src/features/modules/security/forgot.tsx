import { System } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/error_reason_pb";
import { useNavigate } from "@tanstack/react-router";
import { useState } from "react";
import { toast } from "sonner";
import { systemReason } from "#/domains/error";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { inputClass } from "#/features/modules/login/ui/sms-fields";
import { FlowHeader } from "#/features/modules/match/ui/flow-header";
import { m } from "#/paraglide/messages";
import { resetPasswordFn } from "./security.fn";
import { useSmsCode } from "./use-sms-code";

// Forgot password: the public reset leg. The SMS code to the
// account's own phone is the whole proof; success walks back to the login
// page for the promised password login.
export function ForgotPasswordPage() {
  const navigate = useNavigate();
  const sms = useSmsCode();
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [reason, setReason] = useState<string | undefined>(undefined);
  const [pending, setPending] = useState(false);

  async function onSubmit() {
    setError(null);
    setReason(undefined);
    if (!/^1[3-9]\d{9}$/.test(sms.phone)) {
      setError(m.security_error_sms_phone());
      return;
    }
    if (!/^\d{6}$/.test(sms.code)) {
      setError(m.security_error_sms_code_short());
      return;
    }
    if (newPassword.length < 6) {
      setError(m.security_password_short());
      return;
    }
    if (newPassword !== confirmPassword) {
      setError(m.security_pw_error_mismatch());
      return;
    }
    setPending(true);
    try {
      const res = await resetPasswordFn({
        data: { code: sms.code, newPassword, phone: sms.phone, state: sms.state },
      });
      if (res.status === "rejected") {
        setReason(res.reason);
        return;
      }
      toast.success(m.security_forgot_done());
      await navigate({ to: "/login", viewTransition: false });
    } catch {
      setError(m.security_error_generic());
    } finally {
      setPending(false);
    }
  }

  const failureCopy =
    systemReason(reason) === System.SMS_CODE_INVALID
      ? m.error_SYSTEM_SMS_CODE_INVALID()
      : systemReason(reason) === System.USER_NOT_FOUND
        ? m.error_SYSTEM_USER_NOT_FOUND()
        : systemReason(reason) === System.SMS_RATE_LIMITED
          ? m.error_SYSTEM_SMS_RATE_LIMITED()
          : m.security_error_generic();

  return (
    <AmbientLayout>
      <FlowHeader
        onBack={() => void navigate({ to: "/login" })}
        title={m.security_forgot_title()}
      />
      <div className="flex w-full flex-col gap-3.5 pt-3">
        <div className="text-ink-tertiary text-center text-xs leading-[1.6]">
          {m.security_forgot_pitch()}
        </div>
        <div className="border-line-accent-blue bg-identity rounded-card border px-4 py-5">
          <div className="mb-3.5">
            <div className="text-ink-secondary mb-1.5 text-xs">
              {m.security_forgot_phone_label()}
            </div>
            <input
              aria-label={m.security_phone_label()}
              className={inputClass}
              inputMode="numeric"
              maxLength={11}
              onChange={(event) => sms.setPhone(event.target.value)}
              placeholder={m.security_bind_phone_placeholder()}
              value={sms.phone}
            />
          </div>
          <div className="mb-3.5">
            <div className="text-ink-secondary mb-1.5 text-xs">{m.security_pw_code_label()}</div>
            <div className="flex items-stretch gap-3">
              <input
                aria-label={m.security_code_label()}
                className={`${inputClass} flex-1 tracking-[4px]`}
                inputMode="numeric"
                maxLength={6}
                onChange={(event) => sms.setCode(event.target.value)}
                placeholder={m.security_pw_code_placeholder()}
                value={sms.code}
              />
              <button
                className="border-line text-ink-secondary bg-surface-input rounded-control w-[112px] shrink-0 border-[1.5px] py-[13px] text-[13px] font-medium active:opacity-60 disabled:opacity-60"
                disabled={sms.countdown > 0 || sms.sending}
                onClick={() => void sms.send()}
                type="button"
              >
                {sms.countdown > 0
                  ? m.security_code_resend_in({ seconds: sms.countdown })
                  : m.security_code_send()}
              </button>
            </div>
            <div className="text-ink-tertiary mt-1.5 text-right text-[11px]">
              {m.security_forgot_code_window()}
            </div>
          </div>
          <div className="mb-3.5">
            <div className="text-ink-secondary mb-1.5 text-xs">{m.security_pw_new_label()}</div>
            <input
              aria-label={m.security_pw_new_label()}
              autoComplete="new-password"
              className={inputClass}
              onChange={(event) => setNewPassword(event.target.value)}
              placeholder={m.security_pw_new_placeholder()}
              type="password"
              value={newPassword}
            />
          </div>
          <div>
            <div className="text-ink-secondary mb-1.5 text-xs">{m.security_pw_confirm_label()}</div>
            <input
              aria-label={m.security_pw_confirm_label()}
              autoComplete="new-password"
              className={inputClass}
              onChange={(event) => setConfirmPassword(event.target.value)}
              placeholder={m.security_pw_confirm_placeholder()}
              type="password"
              value={confirmPassword}
            />
          </div>
          <button
            className="bg-cta-primary shadow-cta-primary rounded-control mt-4.5 w-full px-4 py-3.5 text-[15px] font-bold text-white disabled:opacity-50"
            disabled={pending}
            onClick={() => void onSubmit()}
            type="button"
          >
            {pending ? m.security_pending() : m.security_forgot_submit()}
          </button>
          <div className="text-ink-tertiary mt-3 text-center text-[11px]">
            {m.security_forgot_note()}
          </div>
        </div>
        {error !== null && <p className="text-danger-text text-[11.5px]">{error}</p>}
        {reason !== undefined && <p className="text-danger-text text-[11.5px]">{failureCopy}</p>}
      </div>
    </AmbientLayout>
  );
}
