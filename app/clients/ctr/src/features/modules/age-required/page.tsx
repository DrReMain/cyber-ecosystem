import { useQuery } from "@connectrpc/connect-query";
import { getProfile } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile-ProfileService_connectquery";
import { useNavigate } from "@tanstack/react-router";
import { Cake } from "lucide-react";
import { useEffect, useState } from "react";
import { Button } from "react-aria-components";
import { toast } from "sonner";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { ageOf, MONTHS, years } from "#/features/app/birth";
import { PROFILE_STALE_MS } from "#/features/auth/profile-cache";
import { useProfileUpdate } from "#/features/modules/profile/use-profile-update";
import { m } from "#/paraglide/messages";

const selectClass =
  "w-full appearance-none rounded-control border border-line bg-surface-input px-2.5 py-[13px] text-[15px] text-ink outline-none focus:border-focus";

/**
 * The pre-payment birth backfill (prototype age-required): the full form
 * (save via UpdateProfile, honest exits), reachable by direct URL.
 */
export function AgeRequiredPage() {
  const navigate = useNavigate();
  const profileQuery = useQuery(getProfile, {}, { staleTime: PROFILE_STALE_MS });
  const updateMutation = useProfileUpdate();
  const [year, setYear] = useState("");
  const [month, setMonth] = useState("");
  const [ready, setReady] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Under the current completeness gate every profile carries birth, so this
  // backfill re-edits for direct visits.
  useEffect(() => {
    const p = profileQuery.data?.profile;
    if (!ready && p?.birthYear != null && p?.birthMonth != null) {
      setYear(String(p.birthYear));
      setMonth(String(p.birthMonth));
      setReady(true);
    }
  }, [profileQuery.data?.profile, ready]);

  async function onSave() {
    if (year === "" || month === "") {
      setError(m.age_required_err_incomplete());
      return;
    }
    // The pay hard gate (< 18 cannot buy) blocks here.
    if (ageOf(Number(year), Number(month)) < 18) {
      setError(null);
      toast.error(m.age_required_minor_toast());
      return;
    }
    setError(null);
    try {
      await updateMutation.mutateAsync({
        birthMonth: Number(month),
        birthYear: Number(year),
      });
    } catch {
      // The mutation cache already toasts the reason; swallowing the
      // rejection keeps it out of the window capture (Sentry noise).
      return;
    }
    toast.success(m.age_required_saved_toast());
    await navigate({ to: "/" });
  }

  return (
    <AmbientLayout
      orbs={[{ className: "top-[-60px] right-[-60px] size-[min(57vw,220px)]", tone: "orange" }]}
    >
      <div className="flex w-full flex-col gap-3 pb-6">
        <h1 className="text-ink text-center text-[19px] font-bold">{m.age_required_title()}</h1>

        <div className="border-line bg-surface-card rounded-card border px-4 py-[18px]">
          <div className="mb-3 flex items-center gap-2.5">
            <div className="bg-cta-practice flex size-11 shrink-0 items-center justify-center rounded-[12px]">
              <Cake aria-hidden className="text-white" size={20} />
            </div>
            <div>
              <div className="text-ink text-[14px] font-extrabold">
                {m.age_required_card_title()}
              </div>
              <div className="text-ink-tertiary mt-0.5 text-[11px]">
                {m.age_required_card_sub()}
              </div>
            </div>
          </div>
          <p className="text-ink-tertiary mb-3 text-[11.5px] leading-[1.75]">
            {m.age_required_desc_lead()}
            <span className="font-bold">{m.age_required_desc_band()}</span>
            {m.age_required_desc_mid()}
            <span className="font-bold">{m.age_required_desc_exact()}</span>
            {m.age_required_desc_tail()}
          </p>
          <div className="flex gap-2.5">
            <select
              aria-label={m.age_required_birth_year()}
              className={selectClass}
              name="birthYear"
              onChange={(event) => {
                setYear(event.target.value);
                setError(null);
              }}
              value={year}
            >
              <option value="">{m.age_required_birth_year()}</option>
              {years().map((y) => (
                <option key={y} value={y}>
                  {m.profile_birth_year_opt({ y })}
                </option>
              ))}
            </select>
            <select
              aria-label={m.age_required_birth_month()}
              className={selectClass}
              name="birthMonth"
              onChange={(event) => {
                setMonth(event.target.value);
                setError(null);
              }}
              value={month}
            >
              <option value="">{m.age_required_birth_month()}</option>
              {MONTHS.map((mo) => (
                <option key={mo} value={mo}>
                  {m.profile_birth_month_opt({ month: mo })}
                </option>
              ))}
            </select>
          </div>
          <p className="text-danger-text min-h-[15px] text-[11.5px] leading-[1.4]" role="alert">
            {error ?? ""}
          </p>
        </div>

        <div className="border-line bg-surface-content text-ink-tertiary rounded-control border px-[13px] py-3 text-[11.5px] leading-[1.85]">
          <span className="font-bold text-[#e5e7eb]">{m.age_required_warn_must()}</span>
          {m.age_required_warn_must_tail()}
          <br />
          {m.age_required_warn_exit_lead()}
          <span className="font-bold text-[#e5e7eb]">{m.age_required_warn_exit_bold()}</span>
          {m.age_required_warn_exit_tail()}
        </div>

        <Button
          className="bg-cta-practice shadow-cta-practice rounded-control w-full px-4 py-[14px] text-[15px] font-bold text-white active:opacity-80 disabled:opacity-50"
          isDisabled={updateMutation.isPending}
          onPress={() => void onSave()}
        >
          {updateMutation.isPending ? m.common_loading() : m.age_required_btn_save()}
        </Button>
        <Button
          className="bg-cta-primary shadow-cta-primary rounded-control w-full px-4 py-[14px] text-[15px] font-bold text-white active:opacity-80"
          onPress={() => void navigate({ to: "/" })}
        >
          {m.age_required_btn_free()}
        </Button>
        <Button
          className="bg-surface-input rounded-control w-full border-[1.5px] border-[#4b5563] px-4 py-[14px] text-[15px] font-medium text-[#d1d5db] active:opacity-80"
          onPress={() => void navigate({ to: "/" })}
        >
          {m.age_required_btn_home()}
        </Button>
      </div>
    </AmbientLayout>
  );
}
