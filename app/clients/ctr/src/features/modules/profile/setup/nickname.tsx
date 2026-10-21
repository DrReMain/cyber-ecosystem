import { useQuery } from "@connectrpc/connect-query";
import { getProfile } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile-ProfileService_connectquery";
import { useNavigate } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { Button } from "react-aria-components";
import { m } from "#/paraglide/messages";
import { useProfileUpdate } from "../use-profile-update";
import { StepFrame } from "./step-frame";

const MAX_LEN = 12;
const MIN_LEN = 2;

// Mirrors the frozen contract (UpdateProfileRequest.nickname 2..12) — code
// points, so 12 CJK characters pass exactly like 12 latin letters.
function weightLen(value: string): number {
  return Array.from(value).length;
}

function validate(value: string): string | null {
  if (value !== value.trim()) {
    return m.profile_nickname_err_trim();
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
  const profileQuery = useQuery(getProfile, {});
  const updateMutation = useProfileUpdate();
  const [value, setValue] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [ready, setReady] = useState(false);

  // Prefill once the profile lands (edit path re-enters this step with the
  // stored nickname); never overwrite what the user is typing.
  useEffect(() => {
    if (!ready && profileQuery.data?.profile?.nickname != null) {
      setValue(profileQuery.data.profile.nickname);
      setReady(true);
    }
  }, [profileQuery.data?.profile?.nickname, ready]);

  const len = weightLen(value);

  async function onNext() {
    const problem = validate(value);
    setError(problem);
    if (problem !== null) {
      return;
    }
    try {
      await updateMutation.mutateAsync({ nickname: value });
    } catch {
      // The mutation cache already toasts the reason; swallowing the
      // rejection keeps it out of the window capture (Sentry noise).
      return;
    }
    await navigate({ to: "/profile/setup/age" });
  }

  return (
    <StepFrame onBack={() => void navigate({ to: "/profile" })} step={1}>
      <div className="rounded-lg border border-line bg-surface-card p-4">
        <div className="mb-1 font-extrabold text-ink text-sm">
          {m.profile_setup_nickname_title()}
        </div>
        <p className="mb-3 text-[11.5px] text-ink-tertiary leading-relaxed">
          {m.profile_setup_nickname_desc()}
        </p>
        <input
          aria-label={m.profile_nickname_label()}
          className="w-full rounded-sm border border-line bg-surface-input px-3 py-3 text-ink outline-none focus:border-focus"
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
        <div className="mt-2 flex items-center justify-between">
          {error !== null && (
            <span className="text-danger-text text-sm" role="alert">
              {error}
            </span>
          )}
          <span
            className={`text-xs ${
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

      <div className="whitespace-pre-line rounded-lg border border-line bg-surface-deep p-3.5 text-[11.5px] text-ink-tertiary leading-relaxed">
        <span className="font-bold text-ink-secondary">
          {m.profile_setup_nickname_rules_title()}
        </span>
        <br />
        {m.profile_setup_nickname_rules()}
      </div>

      <Button
        className="rounded-sm bg-cta-primary px-4 py-3 font-semibold text-white disabled:opacity-50"
        isDisabled={updateMutation.isPending}
        onPress={() => void onNext()}
      >
        {updateMutation.isPending ? m.common_loading() : m.profile_setup_next()}
      </Button>
    </StepFrame>
  );
}
