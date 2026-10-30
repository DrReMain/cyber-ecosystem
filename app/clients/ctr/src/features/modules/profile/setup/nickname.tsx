import { useQuery } from "@connectrpc/connect-query";
import { getProfile } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile-ProfileService_connectquery";
import { useNavigate } from "@tanstack/react-router";
import { Check, MessageCircle, TriangleAlert } from "lucide-react";
import { useEffect, useState } from "react";
import { Button } from "react-aria-components";
import { toast } from "sonner";
import { clearWechatProfile, readWechatProfile } from "#/features/app/wechat-profile";
import { PROFILE_STALE_MS } from "#/features/auth/profile-cache";
import { beginBindWechatFn } from "#/features/modules/security/security.fn";
import { m } from "#/paraglide/messages";
import { useProfileUpdate } from "../use-profile-update";
import { StepFrame } from "./step-frame";

const MAX_LEN = 12;
const MIN_LEN = 2;

// Emoji family the client gates visibly: pictographics, skin-tone
// modifiers, ZWJ sequences, and the variation selector. The server's
// contract validation stays the authority; this is the UX arm. Alternation,
// not one class — astral ranges in a class never match (biome's
// noMisleadingCharacterClass pins this).
const EMOJI_RE = /\p{Extended_Pictographic}|[\u{1F3FB}-\u{1F3FF}]|\u{200D}|\u{FE0F}/u;

// Same family, global, for stripping instead of testing — a wechat nickname
// arrives pre-decorated and the ruling is truncate-and-strip, never reject.
const EMOJI_STRIP_RE = /\p{Extended_Pictographic}|[\u{1F3FB}-\u{1F3FF}]|\u{200D}|\u{FE0F}/gu;

// Mirrors the frozen contract (UpdateProfileRequest.nickname 2..12) — code
// points, so 12 CJK characters pass exactly like 12 latin letters.
function weightLen(value: string): number {
  return Array.from(value).length;
}

// The wechat raw nickname becomes a submittable one: emoji out, edges
// trimmed, then cut to the contract's code-point cap. An all-emoji nickname
// sanitises to "" — the caller reads that as "no prefill".
function sanitizeWechatNickname(raw: string): string {
  const stripped = raw.replace(EMOJI_STRIP_RE, "").trim();
  return Array.from(stripped).slice(0, MAX_LEN).join("").trim();
}

function validate(value: string): string | null {
  if (value !== value.trim()) {
    return m.profile_nickname_err_trim();
  }
  if (EMOJI_RE.test(value)) {
    return m.profile_nickname_err_emoji();
  }
  const len = weightLen(value);
  if (len > MAX_LEN) {
    return m.profile_nickname_err_max();
  }
  if (len < MIN_LEN) {
    return m.profile_nickname_err_min();
  }
  return null;
}

export function SetupNicknamePage() {
  const navigate = useNavigate();
  const profileQuery = useQuery(getProfile, {}, { staleTime: PROFILE_STALE_MS });
  const updateMutation = useProfileUpdate();
  const [value, setValue] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [ready, setReady] = useState(false);
  // The wechat person once the OAuth moment parked it: nickname prefill
  // source, avatar carrier, and the pill's off-switch — showing the pill to
  // an identity that already carried a profile would offer a hop it doesn't
  // need. Null until the browser side reads session storage.
  const [wechat, setWechat] = useState<{ nickname: string; avatarUrl: string } | null>(null);
  const [wechatPrefill, setWechatPrefill] = useState<string | null>(null);
  const [wechatNote, setWechatNote] = useState<string | null>(null);
  const [wechatEnv, setWechatEnv] = useState(false);

  useEffect(() => {
    setWechatEnv(/MicroMessenger/i.test(navigator.userAgent));
    const parked = readWechatProfile();
    if (parked === null) {
      return;
    }
    setWechat(parked);
    const sanitized = sanitizeWechatNickname(parked.nickname);
    if (sanitized === "") {
      setWechatNote(m.profile_setup_wechat_empty());
      return;
    }
    if (sanitized !== parked.nickname.trim()) {
      setWechatNote(m.profile_setup_wechat_trim({ name: sanitized }));
    }
    setWechatPrefill(sanitized);
  }, []);

  // Prefill in priority order: a stored nickname (the edit path re-enters
  // with the user's own choice) outranks the wechat suggestion; never
  // overwrite what the user is typing.
  useEffect(() => {
    if (ready) {
      return;
    }
    const stored = profileQuery.data?.profile?.nickname;
    if (stored != null) {
      setValue(stored);
      setReady(true);
      return;
    }
    if (wechatPrefill != null && wechatPrefill !== "") {
      setValue(wechatPrefill);
      setReady(true);
    }
  }, [profileQuery.data?.profile?.nickname, ready, wechatPrefill]);

  const len = weightLen(value);
  const showPill = wechatEnv && wechat === null;

  async function onWechatPill() {
    const begin = await beginBindWechatFn({
      data: { redirectUri: `${window.location.origin}/profile/setup/wechat-callback` },
    });
    if (begin.status === "ok") {
      window.location.href = begin.authorizeUrl;
      return;
    }
    toast.error(m.profile_setup_wechat_bind_failed());
  }

  async function onNext() {
    const problem = validate(value);
    setError(problem);
    if (problem !== null) {
      return;
    }
    // The avatar lands with the nickname in the same write — its storage
    // slot is this profile row, and the OAuth moment was the only chance to
    // learn it. Never overwrite an avatar the profile already holds.
    const patch: { nickname: string; avatarUrl?: string } = { nickname: value };
    if (wechat?.avatarUrl && profileQuery.data?.profile?.avatarUrl == null) {
      patch.avatarUrl = wechat.avatarUrl;
    }
    try {
      await updateMutation.mutateAsync(patch);
    } catch {
      // The mutation cache already toasts the reason; swallowing the
      // rejection keeps it out of the window capture (Sentry noise).
      return;
    }
    // The park's cargo just landed in the profile row — retiring it here
    // keeps a completed wizard from re-serving the identity to a later
    // account in this tab.
    clearWechatProfile();
    await navigate({ to: "/profile/setup/age" });
  }

  return (
    <StepFrame onBack={() => void navigate({ to: "/profile" })} step={1}>
      <div className="border-line bg-surface-card shadow-card rounded-card border p-4 backdrop-blur-md">
        <div className="text-ink mb-1 text-sm font-extrabold">
          {m.profile_setup_nickname_title()}
        </div>
        <p className="text-ink-tertiary mb-3 text-[11.5px] leading-relaxed">
          {m.profile_setup_nickname_desc_lead()}
          <span className="font-bold">{m.profile_setup_nickname_desc_must()}</span>
          {m.profile_setup_nickname_desc_tail()}
        </p>
        <input
          aria-label={m.profile_nickname_label()}
          className="border-line bg-surface-input text-ink placeholder:text-ink-tertiary focus:border-focus rounded-control w-full border px-3.5 py-[13px] text-[15px] outline-none"
          // Double the contract cap in characters so the counter can show
          // the over-limit state instead of silently truncating input.
          maxLength={24}
          name="nickname"
          onChange={(event) => {
            setValue(event.target.value);
            setError(null);
          }}
          placeholder={m.profile_setup_nickname_placeholder()}
          type="text"
          value={value}
        />
        {showPill ? (
          <div className="mt-[11px] text-center">
            <button
              className="text-status-ok inline-flex items-center gap-1.5 rounded-full border-[1.5px] border-[rgba(74,222,128,0.55)] bg-[rgba(34,197,94,0.10)] px-[17px] py-2 text-[12.5px] font-bold"
              onClick={() => void onWechatPill()}
              type="button"
            >
              <MessageCircle aria-hidden size={14} />
              {m.profile_setup_wechat_pill()}
            </button>
          </div>
        ) : null}
        {wechatPrefill != null && wechatPrefill !== "" ? (
          <div className="text-status-ok mt-[11px] flex items-center justify-center gap-1 text-[11px]">
            <Check aria-hidden size={13} />
            {m.profile_setup_wechat_from()}
          </div>
        ) : null}
        {wechatNote !== null ? (
          <div className="text-status-pending mt-1.5 flex items-center justify-center gap-1 text-[10.5px]">
            <TriangleAlert aria-hidden size={12} />
            {wechatNote}
          </div>
        ) : null}
        <div className="mt-[7px] flex min-h-[15px] items-start justify-between">
          {error !== null && (
            <span className="text-danger-text text-[11.5px]" role="alert">
              {error}
            </span>
          )}
          <span
            className={`text-ink-tertiary text-[11px] tabular-nums ${
              len > MAX_LEN
                ? "text-danger-text"
                : len >= MIN_LEN
                  ? "text-status-ok"
                  : "text-ink-tertiary"
            }`}
          >
            {len} / {MAX_LEN}
          </span>
        </div>
      </div>

      <div className="border-line bg-surface-content text-ink-muted rounded-control border px-3.5 py-3 text-[11.5px] leading-[1.7] whitespace-pre-line">
        <span className="font-bold text-[#e5e7eb]">{m.profile_setup_nickname_rules_title()}</span>
        <br />
        {m.profile_setup_nickname_rules()}
      </div>

      <Button
        className="bg-cta-primary shadow-cta-primary rounded-control px-4 py-[14px] text-[15px] font-bold text-white disabled:opacity-50"
        isDisabled={updateMutation.isPending}
        onPress={() => void onNext()}
      >
        {updateMutation.isPending ? m.common_loading() : m.profile_setup_next()}
      </Button>
    </StepFrame>
  );
}
