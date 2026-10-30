import { useQuery } from "@connectrpc/connect-query";
import { getSecurityStatus } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/auth-AuthService_connectquery";
import { System } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/error_reason_pb";
import { useQueryClient } from "@tanstack/react-query";
import { Link, useNavigate, useSearch } from "@tanstack/react-router";
import { useState } from "react";
import { toast } from "sonner";
import { systemReason } from "#/domains/error";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { inputClass } from "#/features/modules/login/ui/sms-fields";
import { FlowHeader } from "#/features/modules/match/ui/flow-header";
import { m } from "#/paraglide/messages";
import { changePasswordFn, resetPasswordFn } from "./security.fn";
import { useSmsCode } from "./use-sms-code";

// The rejection story per reason, shared by both modes.
function failureCopy(value: string | undefined): string {
  switch (systemReason(value)) {
    case System.PASSWORD_MISMATCH:
      return m.error_SYSTEM_PASSWORD_MISMATCH();
    case System.SMS_CODE_INVALID:
      return m.error_SYSTEM_SMS_CODE_INVALID();
    case System.SMS_RATE_LIMITED:
      return m.error_SYSTEM_SMS_RATE_LIMITED();
    default:
      return m.pw_error_generic();
  }
}

// The two local password verdicts: the length floor and the echo match.
function pwErrorOf(newPassword: string, confirmPassword: string): string | null {
  if (newPassword.length < 6) {
    return m.login_password_short();
  }
  return newPassword !== confirmPassword ? m.pw_error_mismatch() : null;
}

// Branch-free view helpers: each boolean chain lives in its own function so
// the page's own story stays linear.
function noPhoneYet(isSet: boolean, phone: unknown, ready: boolean): boolean {
  return isSet && phone === undefined && ready;
}

function submitDisabled(pending: boolean, isSet: boolean, phone: unknown): boolean {
  return pending || (isSet && phone === undefined);
}

function submitLabel(pending: boolean, isSet: boolean): string {
  if (pending) {
    return m.login_pending();
  }
  return isSet ? m.pw_set_submit() : m.pw_change_submit();
}

// block4's password pair on one route: mode=set (wechat-born, no password —
// the proof is an SMS to the bound phone) and mode=change (the old password
// is the proof, the existing ChangePassword face).
export function PasswordPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { mode } = useSearch({ from: "/profile/security/password" });
  const securityQuery = useQuery(getSecurityStatus, {}, { throwOnError: false });
  // toJson elides empty repeated fields — credentials is absent when unbound.
  const phone = (securityQuery.data?.credentials ?? []).find((c) => c.kind === "sms");
  const sms = useSmsCode();
  const [oldPassword, setOldPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [reason, setReason] = useState<string | undefined>(undefined);
  const [pending, setPending] = useState(false);

  const isSet = mode !== "change";

  async function onSubmit() {
    setError(null);
    setReason(undefined);
    const invalid = pwErrorOf(newPassword, confirmPassword);
    if (invalid !== null) {
      setError(invalid);
      return;
    }
    setPending(true);
    const res = isSet
      ? await resetPasswordFn({
          data: { code: sms.code, newPassword, phone: sms.phone, state: sms.state },
        }).catch(() => null)
      : await changePasswordFn({ data: { newPassword, oldPassword } }).catch(() => null);
    setPending(false);
    if (res === null) {
      setError(m.pw_error_generic());
      return;
    }
    if (res.status === "rejected") {
      setReason(res.reason);
      return;
    }
    toast.success(isSet ? m.pw_set_done() : m.pw_change_done());
    await queryClient.invalidateQueries();
    await navigate({ to: "/profile/security", viewTransition: false });
  }

  return (
    <AmbientLayout>
      <FlowHeader
        onBack={() => void navigate({ to: "/profile/security" })}
        title={isSet ? m.pw_set_title() : m.pw_change_title()}
      />
      <div className="flex w-full flex-col gap-3.5 pt-3">
        <div className="text-ink-tertiary text-center text-xs leading-[1.6]">
          {isSet ? m.pw_set_pitch() : m.pw_change_pitch()}
        </div>

        {noPhoneYet(isSet, phone, securityQuery.isSuccess) ? <NoPhoneCard /> : null}

        <div className="border-line-accent-blue bg-identity rounded-card border px-4 py-5">
          {isSet ? <SetProofFields mask={phone?.maskedIdentifier ?? ""} sms={sms} /> : null}
          {!isSet && (
            <div className="mb-3.5">
              <div className="text-ink-secondary mb-1.5 text-xs">{m.pw_old_label()}</div>
              <input
                aria-label={m.pw_old_label()}
                autoComplete="current-password"
                className={inputClass}
                onChange={(event) => setOldPassword(event.target.value)}
                placeholder={m.pw_old_placeholder()}
                type="password"
                value={oldPassword}
              />
            </div>
          )}
          <div className="mb-3.5">
            <div className="text-ink-secondary mb-1.5 text-xs">{m.pw_new_label()}</div>
            <input
              aria-label={m.pw_new_label()}
              autoComplete="new-password"
              className={inputClass}
              onChange={(event) => setNewPassword(event.target.value)}
              placeholder={m.pw_new_placeholder()}
              type="password"
              value={newPassword}
            />
          </div>
          <div>
            <div className="text-ink-secondary mb-1.5 text-xs">{m.pw_confirm_label()}</div>
            <input
              aria-label={m.pw_confirm_label()}
              autoComplete="new-password"
              className={inputClass}
              onChange={(event) => setConfirmPassword(event.target.value)}
              placeholder={m.pw_confirm_placeholder()}
              type="password"
              value={confirmPassword}
            />
          </div>
          <button
            className="bg-cta-primary shadow-cta-primary rounded-control mt-4.5 w-full px-4 py-3.5 text-[15px] font-bold text-white disabled:opacity-50"
            disabled={submitDisabled(pending, isSet, phone)}
            onClick={() => void onSubmit()}
            type="button"
          >
            {submitLabel(pending, isSet)}
          </button>
          <div className="text-ink-tertiary mt-3 text-center text-[11px] leading-[1.7]">
            {isSet ? m.pw_set_note() : m.pw_change_note()}{" "}
            {!isSet ? (
              <Link className="text-link" to="/login/forgot">
                {m.pw_change_forgot()}
              </Link>
            ) : null}
          </div>
        </div>

        {error !== null && <p className="text-danger-text text-[11.5px]">{error}</p>}
        {reason !== undefined && (
          <p className="text-danger-text text-[11.5px]">{failureCopy(reason)}</p>
        )}
      </div>
    </AmbientLayout>
  );
}

// The set mode's proof half: the bound-phone notice and the code rail —
// the SMS to the account's own phone is the whole gate.
function SetProofFields({ mask, sms }: { mask: string; sms: ReturnType<typeof useSmsCode> }) {
  return (
    <>
      <div className="mb-3.5 w-fit rounded-full bg-[rgba(34,197,94,0.15)] px-2.5 py-1 text-[10.5px] font-semibold text-[#6EE7B7]">
        {m.pw_set_send_to({ mask })}
      </div>
      <div className="mb-3.5">
        <div className="text-ink-secondary mb-1.5 text-xs">{m.forgot_phone_label()}</div>
        <input
          aria-label={m.login_phone_label()}
          className={inputClass}
          inputMode="numeric"
          maxLength={11}
          onChange={(event) => sms.setPhone(event.target.value)}
          placeholder={m.pw_set_phone_placeholder({ mask })}
          value={sms.phone}
        />
      </div>
      <div className="mb-3.5">
        <div className="text-ink-secondary mb-1.5 text-xs">{m.pw_code_label()}</div>
        <div className="flex items-stretch gap-3">
          <input
            aria-label={m.login_code_label()}
            className={`${inputClass} flex-1 tracking-[4px]`}
            inputMode="numeric"
            maxLength={6}
            onChange={(event) => sms.setCode(event.target.value)}
            placeholder={m.pw_code_placeholder()}
            value={sms.code}
          />
          <button
            className="border-line text-ink-secondary bg-surface-input rounded-control w-[112px] shrink-0 border-[1.5px] py-[13px] text-[13px] font-medium active:opacity-60 disabled:opacity-60"
            disabled={sms.countdown > 0 || sms.sending}
            onClick={() => void sms.send()}
            type="button"
          >
            {sms.countdown > 0
              ? m.login_code_resend_in({ seconds: sms.countdown })
              : m.login_code_send()}
          </button>
        </div>
      </div>
    </>
  );
}

// The set mode's dead end: no bound phone, no proof channel — bind first.
function NoPhoneCard() {
  return (
    <div className="rounded-card border border-[rgba(220,38,38,0.55)] bg-[rgba(127,29,29,0.22)] px-3.5 py-3 text-[12px] leading-[1.8]">
      <p className="text-danger-text">{m.pw_set_no_phone()}</p>
      <Link
        className="tap-zone text-link font-bold"
        search={{ from: "me" }}
        to="/profile/security/bind-phone"
      >
        {m.pw_set_go_bind()}
      </Link>
    </div>
  );
}
