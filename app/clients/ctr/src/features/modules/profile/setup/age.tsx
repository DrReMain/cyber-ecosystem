import { useMutation, useQuery } from "@connectrpc/connect-query";
import {
  getProfile,
  updateProfile,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile-ProfileService_connectquery";
import { useNavigate } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { Button } from "react-aria-components";
import { m } from "#/paraglide/messages";
import { bandLabel, bandOfAge } from "../bands";
import { StepFrame } from "./step-frame";

const MONTHS = Array.from({ length: 12 }, (_, i) => i + 1);
// 6–60 years old: the range must reach below 13 or the guardian gate has no
// reachable input (the prototype widened it for exactly this reason).
function years(): number[] {
  const now = new Date().getFullYear();
  return Array.from({ length: 55 }, (_, i) => now - 60 + i);
}

function ageOf(year: number, month: number): number {
  const now = new Date();
  let age = now.getFullYear() - year;
  if (now.getMonth() + 1 < month) {
    age--;
  }
  return age;
}

export function SetupAgePage() {
  const navigate = useNavigate();
  const profileQuery = useQuery(getProfile, {});
  const updateMutation = useMutation(updateProfile);
  const [year, setYear] = useState("");
  const [month, setMonth] = useState("");
  const [guardian, setGuardian] = useState(false);
  const [ready, setReady] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const p = profileQuery.data?.profile;
    if (!ready && p?.birthYear != null && p?.birthMonth != null) {
      setYear(String(p.birthYear));
      setMonth(String(p.birthMonth));
      setReady(true);
    }
  }, [profileQuery.data?.profile, ready]);

  const complete = year !== "" && month !== "";
  const age = complete ? ageOf(Number(year), Number(month)) : null;
  const derivedBand = age !== null ? bandOfAge(age) : null;
  const needGuardian = age !== null && age < 13;
  const minorNote = age !== null && age >= 13 && age < 18;

  async function onNext() {
    if (!complete) {
      setError(m.profile_setup_birth_incomplete());
      return;
    }
    if (needGuardian && !guardian) {
      setError(m.profile_minor_guard_required());
      return;
    }
    setError(null);
    try {
      await updateMutation.mutateAsync({
        ...(needGuardian ? { guardianAccepted: true } : {}),
        birthMonth: Number(month),
        birthYear: Number(year),
      });
    } catch {
      // The mutation cache already toasts the reason; swallowing the
      // rejection keeps it out of the window capture (Sentry noise).
      return;
    }
    await navigate({ to: "/profile/setup/games" });
  }

  const selectClass =
    "w-full appearance-none rounded-sm border border-line bg-surface-input px-3 py-3 text-ink outline-none focus:border-focus";

  return (
    <StepFrame onBack={() => void navigate({ to: "/profile/setup/nickname" })} step={2}>
      <div className="rounded-lg border border-line bg-surface-card p-4">
        <div className="mb-1 font-extrabold text-ink text-sm">{m.profile_birth_label()}</div>
        <p className="mb-3 text-[11.5px] text-ink-tertiary leading-relaxed">
          {m.profile_setup_birth_desc()}
        </p>
        <div className="flex gap-3">
          <select
            aria-label={m.profile_birth_year()}
            className={selectClass}
            name="birthYear"
            onChange={(event) => {
              setYear(event.target.value);
              setError(null);
            }}
            value={year}
          >
            <option value="">{m.profile_birth_year()}</option>
            {years().map((y) => (
              <option key={y} value={y}>
                {y}
              </option>
            ))}
          </select>
          <select
            aria-label={m.profile_birth_month()}
            className={selectClass}
            name="birthMonth"
            onChange={(event) => {
              setMonth(event.target.value);
              setError(null);
            }}
            value={month}
          >
            <option value="">{m.profile_birth_month()}</option>
            {MONTHS.map((mo) => (
              <option key={mo} value={mo}>
                {mo}
              </option>
            ))}
          </select>
        </div>
        {complete && derivedBand !== null && (
          <p className="mt-2 text-[10.5px] text-accent-info">
            {m.profile_setup_band_derived({ band: bandLabel(derivedBand) ?? "" })}
          </p>
        )}
      </div>

      {needGuardian || minorNote ? (
        <div className="rounded-lg border border-[rgba(245,158,11,0.5)] bg-surface-deep p-3.5 text-[11.5px] text-status-pending leading-relaxed">
          {needGuardian ? m.profile_minor_u13() : m.profile_minor_u18()}
        </div>
      ) : null}

      {needGuardian ? (
        <label className="flex flex-col gap-1.5 rounded-lg border border-line bg-surface-deep p-3.5">
          <span className="flex items-start gap-2 text-ink-secondary">
            <input
              checked={guardian}
              className="mt-0.5 h-4 w-4 shrink-0 accent-focus"
              onChange={(event) => {
                setGuardian(event.target.checked);
                setError(null);
              }}
              type="checkbox"
            />
            <span className="text-sm leading-snug">{m.profile_guardian()}</span>
          </span>
          <span className="ml-6 text-[10.5px] text-ink-tertiary leading-snug">
            {m.profile_guardian_note()}
          </span>
        </label>
      ) : null}

      {error !== null && (
        <p className="text-danger-text text-sm" role="alert">
          {error}
        </p>
      )}

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
