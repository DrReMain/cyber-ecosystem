import { useQuery } from "@connectrpc/connect-query";
import { ProfileSex } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile_pb";
import { getProfile } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile-ProfileService_connectquery";
import { useNavigate } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { Button } from "react-aria-components";
import { ageOf, MONTHS, years } from "#/features/app/birth";
import { PROFILE_STALE_MS } from "#/features/auth/profile-cache";
import { m } from "#/paraglide/messages";
import { bandLabel, bandOfAge } from "../bands";
import { useProfileUpdate } from "../use-profile-update";
import { StepFrame } from "./step-frame";

// The read-side wire names (proto-JSON) for the sex enum — the write side
// takes the enum itself, so the mapping lives only where reads land.
const SEX_FROM_WIRE: Record<string, ProfileSex> = {
  PROFILE_SEX_MALE: ProfileSex.MALE,
  PROFILE_SEX_FEMALE: ProfileSex.FEMALE,
  PROFILE_SEX_UNDETERMINED: ProfileSex.UNDETERMINED,
};

export function SetupAgePage() {
  const navigate = useNavigate();
  const profileQuery = useQuery(getProfile, {}, { staleTime: PROFILE_STALE_MS });
  const updateMutation = useProfileUpdate();
  const [year, setYear] = useState("");
  const [month, setMonth] = useState("");
  const [sex, setSex] = useState<ProfileSex | null>(null);
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

  // The sex card (D-AN) prefills from the profile read — whenever the
  // server starts deriving sex from wechat, this prefill lights up.
  // Queries deliver proto-JSON shapes (CONVENTIONS §1): the enum arrives
  // as its wire-name string, mapped back to the enum before state.
  useEffect(() => {
    if (sex !== null) {
      return;
    }
    const wire = profileQuery.data?.profile?.sex as unknown as string | null | undefined;
    const mapped = wire == null ? undefined : SEX_FROM_WIRE[wire];
    if (mapped !== undefined) {
      setSex(mapped);
    }
  }, [profileQuery.data?.profile?.sex, sex]);

  const sexOptions = [
    { value: ProfileSex.MALE, label: m.profile_sex_male() },
    { value: ProfileSex.FEMALE, label: m.profile_sex_female() },
    { value: ProfileSex.UNDETERMINED, label: m.profile_sex_undetermined() },
  ];

  const complete = year !== "" && month !== "";
  const age = complete ? ageOf(Number(year), Number(month)) : null;
  const derivedBand = age !== null ? bandOfAge(age) : null;
  const needGuardian = age !== null && age < 13;
  const minorNote = age !== null && age >= 13 && age < 18;

  async function onNext() {
    if (sex === null) {
      setError(m.profile_setup_sex_required());
      return;
    }
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
        sex,
      });
    } catch {
      // The mutation cache already toasts the reason; swallowing the
      // rejection keeps it out of the window capture (Sentry noise).
      return;
    }
    await navigate({ to: "/profile/setup/games" });
  }

  const selectClass =
    "w-full appearance-none rounded-control border border-line bg-surface-input px-2.5 py-[13px] text-[15px] text-ink outline-none focus:border-focus";

  return (
    <StepFrame onBack={() => void navigate({ to: "/profile/setup/nickname" })} step={2}>
      <div className="border-line bg-surface-card shadow-card rounded-card border p-4 backdrop-blur-md">
        <div className="text-ink mb-1 text-sm font-extrabold">
          {m.profile_sex_label()}
          <span className="text-danger-text"> *</span>
        </div>
        <p className="text-ink-muted mb-2.5 text-[11.5px]">{m.profile_sex_hint()}</p>
        {/* A native radio per option, visually hijacked by the segmented
            skin: the sr-only input carries the semantics (checked,
            arrow-key grouping), the sibling span carries the paint, and
            peer-checked keeps the two in lockstep with no state mirror. */}
        <div
          aria-label={m.profile_sex_label()}
          className="bg-surface-input border-line-soft rounded-control flex gap-1 border p-1"
          role="radiogroup"
        >
          {sexOptions.map((option) => (
            <label className="flex-1 cursor-pointer" key={option.value}>
              <input
                checked={sex === option.value}
                className="peer sr-only"
                name="profile-sex"
                onChange={() => {
                  setSex(option.value);
                  setError(null);
                }}
                type="radio"
                value={option.value}
              />
              <span className="text-ink-muted peer-checked:bg-focus peer-focus-visible:ring-focus block rounded-[9px] py-2.75 text-center text-[13.5px] peer-checked:font-bold peer-checked:text-white peer-checked:shadow-[0_4px_12px_rgba(59,130,246,0.4)] peer-focus-visible:ring-2">
                {option.label}
              </span>
            </label>
          ))}
        </div>
      </div>

      <div className="border-line bg-surface-card shadow-card rounded-card border p-4 backdrop-blur-md">
        <div className="text-ink mb-1 text-sm font-extrabold">{m.profile_birth_label()}</div>
        <p className="text-ink-tertiary mb-3 text-[11.5px] leading-relaxed whitespace-pre-line">
          {m.profile_setup_birth_desc_lead()}
          <span className="font-bold">{m.profile_setup_birth_desc_auto()}</span>
          {m.profile_setup_birth_desc_mid()}
          <span className="font-bold">{m.profile_setup_birth_desc_precise()}</span>
          {m.profile_setup_birth_desc_tail()}
        </p>
        <div className="flex gap-2.5">
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
                {m.profile_birth_year_opt({ y })}
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
            <option value="">{m.profile_birth_month_ph()}</option>
            {MONTHS.map((mo) => (
              <option key={mo} value={mo}>
                {m.profile_birth_month_opt({ month: mo })}
              </option>
            ))}
          </select>
        </div>
        {complete && derivedBand !== null && (
          <p className="text-accent-info mt-2.25 text-[10.5px]">
            {m.profile_setup_band_derived({ band: bandLabel(derivedBand) ?? "" })}
          </p>
        )}
      </div>

      {needGuardian || minorNote ? (
        <div className="bg-surface-content text-status-pending rounded-control border border-[rgba(245,158,11,0.5)] px-3.25 py-2.75 text-[11.5px] leading-[1.7]">
          {needGuardian ? (
            <>
              {m.profile_minor_u13_lead()}
              <span className="font-bold">{m.profile_minor_u13_bold()}</span>
              {m.profile_minor_u13_tail()}
            </>
          ) : (
            <>
              {m.profile_minor_u18_lead()}
              <span className="font-bold">{m.profile_minor_u18_bold()}</span>
              {m.profile_minor_u18_tail()}
            </>
          )}
        </div>
      ) : null}

      {needGuardian ? (
        <div>
          <label className="text-ink-muted flex cursor-pointer items-start gap-2.25 text-[12px] leading-[1.6]">
            <input
              checked={guardian}
              className="mt-px h-4 w-4 shrink-0 accent-[#2563eb]"
              onChange={(event) => {
                setGuardian(event.target.checked);
                setError(null);
              }}
              type="checkbox"
            />
            <span>
              {m.profile_guardian_lead()}
              <span className="font-bold">{m.profile_guardian_bold()}</span>
              {m.profile_guardian_tail()}
            </span>
          </label>
          <p className="text-ink-tertiary ml-6.25 text-[10.5px] leading-snug">
            {m.profile_guardian_note()}
          </p>
        </div>
      ) : null}

      {error !== null && (
        <p className="text-danger-text text-[11.5px]" role="alert">
          {error}
        </p>
      )}

      <Button
        className="bg-cta-primary shadow-cta-primary rounded-control px-4 py-3.5 text-[15px] font-bold text-white disabled:opacity-50"
        isDisabled={updateMutation.isPending}
        onPress={() => void onNext()}
      >
        {updateMutation.isPending ? m.common_loading() : m.profile_setup_next()}
      </Button>
    </StepFrame>
  );
}
