import { useQuery } from "@connectrpc/connect-query";
import type { GetSecurityStatusResponse } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/auth_pb";
import { getSecurityStatus } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/auth-AuthService_connectquery";
import { Link } from "@tanstack/react-router";
import { ChevronRight, KeyRound, MessageCircle, Smartphone } from "lucide-react";
import { m } from "#/paraglide/messages";

const CARD_CLASS = "border-line-accent-blue bg-identity rounded-card border px-4 py-3.75";
const OK_CHIP = "bg-status-ok-bg text-[#6EE7B7] rounded-full px-2.5 py-1 text-[10px] font-semibold";
const LOCK_CHIP =
  "bg-[rgba(107,114,128,0.22)] text-ink-tertiary rounded-full px-2.5 py-1 text-[10px] font-semibold";

// Prototype profile-overview ②: the account-security block is a status
// view, not an entry list — each row shows what is bound (masked
// server-side) and acts only when it can: bind what is missing, manage the
// rest on the security page.
// The transport's toJson elides empty repeated fields: a member with nothing
// bound gets no credentials key at all — normalize before reading.
function boundCredentials(status: GetSecurityStatusResponse | undefined) {
  return status?.credentials ?? [];
}

export function SecurityCard() {
  const securityQuery = useQuery(getSecurityStatus, {}, { throwOnError: false });
  const status = securityQuery.data;
  const credentials = boundCredentials(status);
  const phone = credentials.find((c) => c.kind === "sms");
  const wechat = credentials.some((c) => c.kind === "wechat_mp");

  return (
    <div className={CARD_CLASS}>
      <div className="mb-2.75 flex items-baseline justify-between">
        <span className="text-ink text-[14.5px] font-extrabold">{m.profile_card_account()}</span>
        <Link
          className="tap-zone text-ink-tertiary text-[10.5px] active:opacity-60"
          to="/profile/security"
        >
          {m.security_manage()}
        </Link>
      </div>

      <div className="flex items-center justify-between py-2.25">
        <span className="text-ink flex items-center gap-2.25 text-[13px] font-bold">
          <span className="rounded-control flex h-7.5 w-7.5 shrink-0 items-center justify-center bg-[rgba(34,197,94,0.16)] text-[#4ADE80]">
            <MessageCircle aria-hidden size={14} />
          </span>
          {m.profile_row_wechat()}
        </span>
        <span className="flex items-center gap-1.5">
          <span className={wechat ? OK_CHIP : LOCK_CHIP}>
            {wechat ? m.security_status_bound() : m.security_status_unbound()}
          </span>
          <Link
            className="tap-zone text-link flex items-center py-1 font-bold active:opacity-60"
            search={wechat ? undefined : { from: "login" }}
            to={wechat ? "/profile/security" : "/profile/security/bind-wechat"}
          >
            {wechat ? m.security_action_manage() : m.security_action_bind()}
            <ChevronRight aria-hidden size={14} />
          </Link>
        </span>
      </div>

      <div className="flex items-center justify-between border-t border-[rgba(55,65,81,0.4)] py-2.25">
        <span className="text-ink flex items-center gap-2.25 text-[13px] font-bold">
          <span className="rounded-control text-accent-info flex h-7.5 w-7.5 shrink-0 items-center justify-center bg-[rgba(37,99,235,0.16)]">
            <Smartphone aria-hidden size={14} />
          </span>
          {m.profile_row_phone()}
        </span>
        <span className="flex items-center gap-1.5">
          {phone && (
            <span className="text-ink-secondary num text-[12px]">{phone.maskedIdentifier}</span>
          )}
          <span className={phone ? OK_CHIP : LOCK_CHIP}>
            {phone ? m.security_status_bound() : m.security_status_unbound()}
          </span>
          <Link
            className="tap-zone text-link flex items-center py-1 font-bold active:opacity-60"
            search={phone ? undefined : { from: "login" }}
            to={phone ? "/profile/security" : "/profile/security/bind-phone"}
          >
            {phone ? m.security_action_manage() : m.security_action_bind()}
            <ChevronRight aria-hidden size={14} />
          </Link>
        </span>
      </div>

      <div className="flex items-center justify-between border-t border-[rgba(55,65,81,0.4)] py-2.25">
        <span className="text-ink flex items-center gap-2.25 text-[13px] font-bold">
          <span className="text-ink-secondary rounded-control bg-status-lock-bg flex h-7.5 w-7.5 shrink-0 items-center justify-center">
            <KeyRound aria-hidden size={14} />
          </span>
          {m.profile_row_password()}
        </span>
        <span className="flex items-center gap-1.5">
          <span className={status?.hasPassword ? OK_CHIP : LOCK_CHIP}>
            {status?.hasPassword ? m.security_status_set() : m.security_status_unset()}
          </span>
          <Link
            className="tap-zone text-link flex items-center py-1 font-bold active:opacity-60"
            search={{ mode: status?.hasPassword ? "change" : "set" }}
            to="/profile/security/password"
          >
            {status?.hasPassword ? m.security_action_change() : m.security_action_set()}
            <ChevronRight aria-hidden size={14} />
          </Link>
        </span>
      </div>
    </div>
  );
}
