import { useQuery } from "@connectrpc/connect-query";
import { getProfile } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/profile-ProfileService_connectquery";
import { useNavigate } from "@tanstack/react-router";
import { Cake } from "lucide-react";
import { useEffect, useState } from "react";
import { Button } from "react-aria-components";
import { toast } from "sonner";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { PROFILE_STALE_MS } from "#/features/auth/profile-cache";
import { useProfileUpdate } from "#/features/modules/profile/use-profile-update";
import { m } from "#/paraglide/messages";
import { ageOf, MONTHS, years } from "./birth";

const selectClass =
  "w-full appearance-none rounded-sm border border-line bg-surface-input px-3 py-3 text-ink outline-none focus:border-focus";

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
    <AmbientLayout>
      <div className="flex w-full flex-col gap-3 pb-6">
        <h1 className="text-center font-extrabold text-ink text-lg">{m.age_required_title()}</h1>

        <div className="rounded-lg border border-line bg-surface-card p-4">
          <div className="mb-3 flex items-center gap-2.5">
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-tile bg-cta-practice">
              <Cake aria-hidden className="text-white" size={20} />
            </div>
            <div>
              <div className="font-extrabold text-[14px] text-ink">
                {m.age_required_card_title()}
              </div>
              <div className="mt-0.5 text-[11px] text-ink-tertiary">
                {m.age_required_card_sub()}
              </div>
            </div>
          </div>
          <p className="mb-3 text-[11.5px] text-ink-tertiary leading-[1.75]">
            {m.age_required_desc()}
          </p>
          <div className="flex gap-3">
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
                  {y}
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
                  {mo}
                </option>
              ))}
            </select>
          </div>
          {error !== null && (
            <p className="mt-2 text-danger-text text-sm" role="alert">
              {error}
            </p>
          )}
        </div>

        <div className="rounded-lg border border-line bg-surface-deep p-3.5 text-[11.5px] text-ink-tertiary leading-[1.85]">
          <span className="font-bold text-ink-secondary">{m.age_required_warn_must()}</span>
          <br />
          {m.age_required_warn_exit()}
        </div>

        <Button
          className="w-full rounded-sm bg-cta-practice px-4 py-3 font-semibold text-white active:opacity-80 disabled:opacity-50"
          isDisabled={updateMutation.isPending}
          onPress={() => void onSave()}
        >
          {updateMutation.isPending ? m.common_loading() : m.age_required_btn_save()}
        </Button>
        <Button
          className="w-full rounded-sm bg-cta-primary px-4 py-3 font-semibold text-white active:opacity-80"
          onPress={() => void navigate({ to: "/" })}
        >
          {m.age_required_btn_free()}
        </Button>
        <Button
          className="w-full rounded-sm border border-line px-4 py-3 font-semibold text-ink-secondary active:opacity-80"
          onPress={() => void navigate({ to: "/" })}
        >
          {m.age_required_btn_home()}
        </Button>
      </div>
    </AmbientLayout>
  );
}
