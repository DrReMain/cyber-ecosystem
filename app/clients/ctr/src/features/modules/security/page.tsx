import { useQuery } from "@connectrpc/connect-query";
import { getProfile } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile-ProfileService_connectquery";
import type { GetSecurityStatusResponse } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/auth_pb";
import { getSecurityStatus } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/auth-AuthService_connectquery";
import { Link, useNavigate } from "@tanstack/react-router";
import { ChevronRight, KeyRound, MessageCircle, Smartphone } from "lucide-react";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { PROFILE_STALE_MS } from "#/features/auth/profile-cache";
import { FlowHeader } from "#/features/modules/match/ui/flow-header";
import { m } from "#/paraglide/messages";

const CARD_CLASS = "border-line-accent-blue bg-identity rounded-card border px-4 py-[15px]";

// The bound chip, the prototype's st-ok.
const OK_CHIP =
  "bg-[rgba(34,197,94,0.15)] text-[#6EE7B7] rounded-full px-2.5 py-1 text-[10px] font-semibold";

// Prototype 33: the binding surface as status rows — each row shows what is
// bound (masked server-side) and routes to the bind page only when it can
// act. The unbind note is the corpus's service-desk ruling, not the
// prototype's detail-page wording (those pages do not exist by ruling).
// toJson elides empty repeated fields — a member with nothing bound has no
// credentials key at all. Normalize before reading.
function boundCredentials(status: GetSecurityStatusResponse | undefined) {
  return status?.credentials ?? [];
}

export function SecurityPage() {
  const navigate = useNavigate();
  const securityQuery = useQuery(getSecurityStatus, {}, { throwOnError: false });
  const profileQuery = useQuery(
    getProfile,
    {},
    { staleTime: PROFILE_STALE_MS, throwOnError: false },
  );
  const status = securityQuery.data;
  const credentials = boundCredentials(status);
  const phone = credentials.find((c) => c.kind === "sms");
  const wechat = credentials.some((c) => c.kind === "wechat_mp");
  const profile = profileQuery.data?.profile;

  return (
    <AmbientLayout>
      <FlowHeader onBack={() => void navigate({ to: "/profile" })} title={m.security_title()} />
      <div className="flex w-full flex-col gap-3 pt-3">
        <div className="text-ink-secondary border-line-accent-blue bg-identity rounded-card border px-3.5 py-3 text-[11.5px] leading-[1.8]">
          {m.security_bind_note()}
        </div>

        <div className={CARD_CLASS}>
          <div className="flex items-center gap-[11px]">
            <span className="rounded-control flex h-[34px] w-[34px] items-center justify-center bg-[rgba(34,197,94,0.16)] text-[#4ADE80]">
              <MessageCircle aria-hidden size={17} />
            </span>
            <div className="min-w-0 flex-1">
              <div className="text-ink text-[13px] font-bold">{m.security_row_wechat()}</div>
              <div className="text-ink-tertiary mt-[3px] text-[10.5px] leading-[1.6]">
                {wechat ? m.security_wechat_bound_desc() : m.security_wechat_unbound_desc()}
              </div>
            </div>
            {wechat ? (
              <span className={OK_CHIP}>{m.security_status_bound()}</span>
            ) : (
              <Link
                className="tap-zone text-link flex items-center gap-1 py-1 font-bold active:opacity-60"
                search={{ from: "me" }}
                to="/profile/security/bind-wechat"
              >
                {m.security_action_bind()}
                <ChevronRight aria-hidden size={14} />
              </Link>
            )}
          </div>
        </div>

        <div className={CARD_CLASS}>
          <div className="flex items-center gap-[11px]">
            <span className="rounded-control flex h-[34px] w-[34px] items-center justify-center bg-[rgba(37,99,235,0.16)] text-[#93C5FD]">
              <Smartphone aria-hidden size={17} />
            </span>
            <div className="min-w-0 flex-1">
              <div className="text-ink text-[13px] font-bold">{m.security_row_phone()}</div>
              <div className="text-ink-tertiary mt-[3px] text-[10.5px] leading-[1.6]">
                {phone
                  ? m.security_phone_bound_desc({ mask: phone.maskedIdentifier })
                  : m.security_phone_unbound_desc()}
              </div>
            </div>
            {phone ? (
              <span className={OK_CHIP}>{m.security_status_bound()}</span>
            ) : (
              <Link
                className="tap-zone text-link flex items-center gap-1 py-1 font-bold active:opacity-60"
                search={{ from: "me" }}
                to="/profile/security/bind-phone"
              >
                {m.security_action_bind()}
                <ChevronRight aria-hidden size={14} />
              </Link>
            )}
          </div>
        </div>

        <div className={CARD_CLASS}>
          <div className="flex items-center gap-[11px]">
            <span className="text-ink-secondary rounded-control flex h-[34px] w-[34px] items-center justify-center bg-[rgba(107,114,128,0.18)]">
              <KeyRound aria-hidden size={17} />
            </span>
            <div className="min-w-0 flex-1">
              <div className="text-ink text-[13px] font-bold">{m.security_row_password()}</div>
              <div className="text-ink-tertiary mt-[3px] text-[10.5px] leading-[1.6]">
                {status?.hasPassword
                  ? m.security_password_set_desc()
                  : m.security_password_unset_desc()}
              </div>
            </div>
            <Link
              className="tap-zone text-link flex items-center gap-1 py-1 font-bold active:opacity-60"
              search={{ mode: status?.hasPassword ? "change" : "set" }}
              to="/profile/security/password"
            >
              {status?.hasPassword ? m.security_action_change() : m.security_action_set()}
              <ChevronRight aria-hidden size={14} />
            </Link>
          </div>
        </div>

        <div className="text-ink-tertiary border-line-accent-blue bg-identity rounded-card border px-3.5 py-3 text-[11px] leading-[1.8]">
          {m.security_unbind_note()}
        </div>

        <div className={CARD_CLASS}>
          <div className="text-ink mb-1 text-[13px] font-extrabold">
            {m.security_account_info()}
          </div>
          <InfoRow
            label={m.security_info_nickname()}
            value={profile?.nickname ?? m.security_info_unset()}
          />
          <InfoRow
            label={m.security_info_sex()}
            value={
              profile?.sex === 1
                ? m.security_sex_male()
                : profile?.sex === 2
                  ? m.security_sex_female()
                  : m.security_info_unfilled()
            }
          />
          <InfoRow
            label={m.security_info_birth()}
            value={
              profile?.birthYear != null && profile?.birthMonth != null
                ? `${profile.birthYear}-${String(profile.birthMonth).padStart(2, "0")}`
                : m.security_info_unfilled()
            }
          />
          <InfoRow
            label={m.security_info_games()}
            value={
              profile?.games && profile.games.length > 0
                ? profile.games
                    .map((g) => g.customName || g.gameId || "")
                    .filter(Boolean)
                    .join("、")
                : m.security_info_none()
            }
          />
        </div>

        <button
          className="border-line text-ink-secondary bg-surface-input rounded-control border-[1.5px] px-4 py-3.5 text-[15px] font-medium active:opacity-60"
          onClick={() => void navigate({ to: "/profile" })}
          type="button"
        >
          {m.security_back_profile()}
        </button>
      </div>
    </AmbientLayout>
  );
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between py-[7px] text-xs">
      <span className="text-ink-secondary">{label}</span>
      <span className="text-ink max-w-[200px] text-right">{value}</span>
    </div>
  );
}
