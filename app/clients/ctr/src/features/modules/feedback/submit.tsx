import { useMutation } from "@connectrpc/connect-query";
import { FeedbackCategory } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/feedback_pb";
import { submitFeedback } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/feedback-FeedbackService_connectquery";
import { useNavigate, useRouter } from "@tanstack/react-router";
import { ChevronLeft } from "lucide-react";
import { useState } from "react";
import { Button } from "react-aria-components";
import { toast } from "sonner";
import type { AmbientOrb } from "#/features/app/ambient/ambient-backdrop";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { m } from "#/paraglide/messages";
import {
  CATEGORY_ACCOUNT,
  CATEGORY_ASSESSMENT,
  CATEGORY_FEATURE_BUG,
  CATEGORY_OTHER,
  CATEGORY_PAYMENT,
  CATEGORY_REPORT,
  categoryWordOf,
} from "./wire";

const ORBS: readonly AmbientOrb[] = [];

const GLASS_CLASS =
  "border-line bg-surface-card shadow-card rounded-card border p-4 backdrop-blur-md";

// The prototype's own minimum: ten characters is the shortest
// description that can carry a repro (the contract's min_len agrees).
const CONTENT_MIN = 10;
const CONTENT_MAX = 500;
// The prototype caps the contact line at 40 characters — well inside the
// contract's 128, so the tighter input stays the honest shape.
const CONTACT_MAX = 40;

const CATEGORY_PICKS = [
  { value: FeedbackCategory.FEATURE_BUG, label: () => categoryWordOf(CATEGORY_FEATURE_BUG) },
  { value: FeedbackCategory.ASSESSMENT, label: () => categoryWordOf(CATEGORY_ASSESSMENT) },
  { value: FeedbackCategory.REPORT, label: () => categoryWordOf(CATEGORY_REPORT) },
  { value: FeedbackCategory.PAYMENT, label: () => categoryWordOf(CATEGORY_PAYMENT) },
  { value: FeedbackCategory.ACCOUNT, label: () => categoryWordOf(CATEGORY_ACCOUNT) },
  { value: FeedbackCategory.OTHER, label: () => categoryWordOf(CATEGORY_OTHER) },
] as const;

export function FeedbackSubmitPage() {
  const navigate = useNavigate();
  const router = useRouter();
  const [picked, setPicked] = useState<FeedbackCategory | null>(null);
  const [content, setContent] = useState("");
  const [contact, setContact] = useState("");
  const [err, setErr] = useState<string | null>(null);
  const submit = useMutation(submitFeedback, {
    onSuccess: () => {
      // The prototype lands the member back on their queue so the new
      // ticket's 处理中 badge is the confirmation.
      toast.success(m.feedback_submit_ok());
      void navigate({ to: "/messages", search: { t: "fb" } });
    },
  });

  function backOrCenter() {
    if (router.history.canGoBack()) {
      router.history.back();
    } else {
      void navigate({ to: "/messages", search: { t: "fb" } });
    }
  }

  function onSubmit() {
    const body = content.trim();
    if (picked === null) {
      setErr(m.feedback_err_category());
      return;
    }
    if (body.length < CONTENT_MIN) {
      setErr(m.feedback_err_content());
      return;
    }
    setErr(null);
    submit.mutate({
      category: picked,
      content: body,
      contact: contact.trim() === "" ? undefined : contact.trim(),
    });
  }

  return (
    <AmbientLayout orbs={ORBS}>
      <div className="flex w-full flex-col gap-3.5 pb-6">
        <div className="flex items-center gap-3 pt-5">
          <button
            aria-label={m.common_back()}
            className="bg-surface-input text-ink-secondary flex h-10 w-10 shrink-0 items-center justify-center rounded-full active:opacity-60"
            onClick={backOrCenter}
            type="button"
          >
            <ChevronLeft aria-hidden size={20} />
          </button>
          <h1 className="text-ink text-[19px] font-bold">{m.feedback_head()}</h1>
        </div>

        <div className={GLASS_CLASS}>
          <div className="text-ink-tertiary mb-2.5 text-[11.5px]">
            {m.feedback_category_label()}
          </div>
          {/* The prototype's .hscroll chip row — one horizontal rail, the
           * rail scrolls instead of wrapping. */}
          <div className="flex [scrollbar-width:none] gap-2.5 overflow-x-auto pb-1">
            {CATEGORY_PICKS.map((pick) => (
              <button
                className={`shrink-0 rounded-[10px] px-[13px] py-2 text-[12px] whitespace-nowrap ${
                  picked === pick.value
                    ? "bg-cta-primary font-semibold text-white"
                    : "border-line bg-surface-input/50 text-ink-tertiary border"
                }`}
                key={pick.value}
                onClick={() => setPicked(pick.value)}
                type="button"
              >
                {pick.label()}
              </button>
            ))}
          </div>
        </div>

        <div className={GLASS_CLASS}>
          <div className="text-ink-tertiary mb-2.5 text-[11.5px]">{m.feedback_content_label()}</div>
          <textarea
            className="bg-surface-input text-ink placeholder:text-ink-tertiary min-h-[110px] w-full resize-none rounded-[10px] border-0 p-3 text-[13px] leading-[1.7] outline-none"
            maxLength={CONTENT_MAX}
            onChange={(e) => setContent(e.target.value)}
            placeholder={m.feedback_content_placeholder()}
            value={content}
          />
          <div
            className={`mt-1.5 text-right text-[10.5px] ${
              content.length >= CONTENT_MAX ? "text-[#F87171]" : "text-ink-tertiary"
            }`}
          >
            {m.feedback_content_count({ used: String(content.length), max: String(CONTENT_MAX) })}
          </div>
        </div>

        <div className={GLASS_CLASS}>
          <div className="text-ink-tertiary mb-2.5 text-[11.5px]">{m.feedback_contact_label()}</div>
          <input
            className="bg-surface-input text-ink placeholder:text-ink-tertiary w-full rounded-[10px] border-0 px-3 py-2.5 text-[13px] outline-none"
            maxLength={CONTACT_MAX}
            onChange={(e) => setContact(e.target.value)}
            placeholder={m.feedback_contact_placeholder()}
            type="text"
            value={contact}
          />
          <div className="text-ink-tertiary mt-2 text-[10.5px] leading-[1.6]">
            {m.feedback_contact_note()}
          </div>
        </div>

        {err !== null ? <div className="text-[11px] text-[#F87171]">{err}</div> : null}

        <div className="text-ink-tertiary text-center text-[10.5px] leading-[1.7]">
          {m.feedback_lifecycle_note()}
        </div>

        <Button
          className="bg-cta-primary w-full rounded-[12px] py-[13px] text-[15px] font-bold text-white active:opacity-80 disabled:opacity-50"
          isDisabled={submit.isPending}
          onPress={onSubmit}
        >
          {m.feedback_submit()}
        </Button>
      </div>
    </AmbientLayout>
  );
}
