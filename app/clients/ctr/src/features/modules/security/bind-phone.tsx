import { useQueryClient } from "@tanstack/react-query";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { useState } from "react";
import { toast } from "sonner";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { mainlandMobile } from "#/features/modules/login/sms";
import { badInputClass, inputClass } from "#/features/modules/login/ui/sms-fields";
import { FlowHeader } from "#/features/modules/match/ui/flow-header";
import { m } from "#/paraglide/messages";
import { setMergeFlow } from "./merge-flow";
import type { MergeOfferDTO } from "./security.fn";
import { previewMergePhoneFn } from "./security.fn";
import { useBindSms } from "./use-bind-sms";

const sendBtnClass =
  "border-focus text-link bg-transparent rounded-control w-[116px] shrink-0 border px-3 py-[13px] text-[13px] font-bold active:opacity-60 disabled:opacity-60";

type Preview = { offer: MergeOfferDTO } | null;

function entryLabelOf(from: string | undefined): string {
  if (from === "login") {
    return m.security_bind_entry_login();
  }
  return from === "guide" ? m.security_bind_entry_guide() : m.security_bind_entry_me();
}

// An offer that names a direction (either side empty) is the mergeable
// conflict; BLOCKED still routes onward — its confirm-page pane is the
// manual path the corpus prescribes for two weighted accounts.
function actionableOffer(preview: Preview): MergeOfferDTO | null {
  if (preview === null) {
    return null;
  }
  const s = preview.offer.status;
  return s === "KEEP_CALLER" || s === "KEEP_OTHER" || s === "BLOCKED" ? preview.offer : null;
}

function conflictCtaLabel(offer: MergeOfferDTO): string {
  return offer.status === "BLOCKED" ? m.security_bind_phone_blocked_cta() : m.security_bind_phone_merge_cta();
}

// Bind a phone to a wechat-born account. The preview runs
// before any SMS is spent — an occupied phone answers with the conflict
// card (no send, per the prototype's "no code needed" law), and an
// empty-owner conflict upgrades into the merge offer instead.
export function BindPhonePage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { from } = useSearch({ from: "/profile/security/bind-phone" });
  const bind = useBindSms();
  const [preview, setPreview] = useState<Preview>(null);
  const [checking, setChecking] = useState(false);
  const [previewFailed, setPreviewFailed] = useState(false);
  const conflict = actionableOffer(preview);
  const doneTo = from === "me" ? "/profile/security" : "/";

  // The occupancy check ahead of the send button: FREE alone may spend an
  // SMS; everything else renders the card and keeps the window closed.
  async function onSend() {
    if (!mainlandMobile.test(bind.phone)) {
      return;
    }
    setChecking(true);
    setPreviewFailed(false);
    try {
      const res = await previewMergePhoneFn({ data: { phone: bind.phone } });
      if (res.status === "rejected") {
        setPreviewFailed(true);
        return;
      }
      if (res.offer.status === "FREE") {
        setPreview(null);
        await bind.send();
        return;
      }
      if (res.offer.status === "SELF") {
        toast.info(m.security_bind_phone_self());
        setPreview(null);
        return;
      }
      setPreview({ offer: res.offer });
    } catch {
      setPreviewFailed(true);
    } finally {
      setChecking(false);
    }
  }

  async function onSubmit() {
    const verdict = await bind.submit();
    if (verdict === "ok") {
      toast.success(m.security_bind_phone_done());
      await queryClient.invalidateQueries();
      await navigate({ to: doneTo, viewTransition: false });
    }
    // A rejected/failed bind keeps the page up — the reason line tells the
    // story; a 6005 race re-runs the preview so the card replaces the form.
  }

  return (
    <AmbientLayout>
      <FlowHeader
        onBack={() => void navigate({ to: "/profile/security" })}
        title={m.security_bind_phone_title()}
      />
      <div className="flex w-full flex-col gap-3.5 pt-3">
        <div className="border-line-accent-blue bg-identity rounded-card border px-4 py-5">
          <div className="text-ink text-base font-extrabold">{m.security_bind_phone_title()}</div>
          <div className="text-ink-secondary mt-2 text-[12.5px] leading-[1.8]">
            {m.security_bind_phone_pitch()}
          </div>
        </div>

        <input
          aria-label={m.security_phone_label()}
          className={bind.phoneErr ? badInputClass : inputClass}
          inputMode="numeric"
          maxLength={11}
          onBlur={bind.blurPhone}
          onChange={(event) => bind.setPhone(event.target.value)}
          placeholder={m.security_bind_phone_placeholder()}
          value={bind.phone}
        />
        <div className="flex items-stretch gap-3">
          <input
            aria-label={m.security_code_label()}
            className={`${bind.codeErr ? badInputClass : inputClass} flex-1 tracking-[4px]`}
            inputMode="numeric"
            maxLength={6}
            onChange={(event) => bind.setCode(event.target.value)}
            placeholder={m.security_bind_phone_code_placeholder()}
            value={bind.code}
          />
          <button
            className={sendBtnClass}
            disabled={
              bind.countdown > 0 || bind.sending || checking || preview !== null || previewFailed
            }
            onClick={() => void onSend()}
            type="button"
          >
            {bind.countdown > 0
              ? m.security_code_resend_in({ seconds: bind.countdown })
              : checking
                ? m.security_bind_phone_checking()
                : m.security_code_send()}
          </button>
        </div>

        {bind.reason === "SYSTEM_SMS_RATE_LIMITED" ? (
          <p className="text-danger-text text-[11.5px]">{m.error_SYSTEM_SMS_RATE_LIMITED()}</p>
        ) : null}
        {bind.codeErr ? (
          <p className="text-danger-text text-[11.5px]">{m.security_bind_phone_code_err()}</p>
        ) : null}
        {previewFailed ? (
          <p className="text-danger-text text-[11.5px]">{m.security_bind_phone_preview_failed()}</p>
        ) : null}

        {preview !== null ? (
          <div className="border-line-accent-blue bg-identity rounded-card border px-3.5 py-3">
            <p className="text-danger-text text-[12px] leading-[1.8]">
              <b>{m.security_bind_phone_conflict_head()}</b>
              <br />
              {m.security_bind_phone_conflict_body()}
            </p>
            {conflict !== null ? (
              <button
                className="bg-cta-primary shadow-cta-primary rounded-control mt-3 w-full px-4 py-3 text-[14px] font-bold text-white active:opacity-60"
                onClick={() => {
                  setMergeFlow({ offer: conflict, phone: bind.phone });
                  void navigate({ to: "/profile/security/merge", viewTransition: false });
                }}
                type="button"
              >
                {conflictCtaLabel(conflict)}
              </button>
            ) : null}
          </div>
        ) : null}

        <div className="text-ink-tertiary text-[11.5px] leading-[1.8]">
          {m.security_bind_phone_entry_line({ entry: entryLabelOf(from) })}
        </div>
        <button
          className="bg-cta-primary shadow-cta-primary rounded-control px-4 py-3.5 text-[15px] font-bold text-white disabled:opacity-50"
          disabled={bind.pending}
          onClick={() => void onSubmit()}
          type="button"
        >
          {m.security_bind_phone_submit()}
        </button>
        <button
          className="border-line text-ink-secondary bg-surface-input rounded-control border-[1.5px] px-4 py-3.5 text-[15px] font-medium active:opacity-60"
          onClick={() => {
            toast.info(m.security_bind_phone_skipped());
            void navigate({ to: doneTo, viewTransition: false });
          }}
          type="button"
        >
          {m.security_bind_phone_skip()}
        </button>
      </div>
    </AmbientLayout>
  );
}
