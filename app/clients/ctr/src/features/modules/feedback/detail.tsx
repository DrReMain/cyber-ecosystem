import { useQuery as useConnectQuery, useMutation } from "@connectrpc/connect-query";
import {
  followUpFeedback,
  getMyFeedback,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/feedback-FeedbackService_connectquery";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate, useParams, useRouter } from "@tanstack/react-router";
import { ChevronLeft } from "lucide-react";
import { useState } from "react";
import { Button } from "react-aria-components";
import { toast } from "sonner";
import type { AmbientOrb } from "#/features/app/ambient/ambient-backdrop";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { localShortDateTimeOf } from "#/features/app/date";
import { StateScreen } from "#/features/app/states/state-screen";
import { m } from "#/paraglide/messages";
import { StatusBadge } from "./ui/status-badge";
import {
  categoryWordOf,
  SENDER_ADMIN,
  STATUS_CLOSED,
  wireFeedbackMessageOf,
  wireFeedbackOf,
} from "./wire";

const ORBS: readonly AmbientOrb[] = [];

const GLASS_CLASS =
  "border-line bg-surface-card shadow-card rounded-card border p-4 backdrop-blur-md";

// The feedback conversation (prototype feedback-detail?kind=fb): the
// ticket root card, the thread of turns, and the follow-up bar — a
// follow-up flips the ticket back to 处理中 (server-side flip, then the
// invalidation repaints the badge).
export function FeedbackDetailPage() {
  const navigate = useNavigate();
  const router = useRouter();
  const { feedbackId } = useParams({ from: "/feedback/$feedbackId" });
  const queryClient = useQueryClient();
  const [draft, setDraft] = useState("");
  const query = useConnectQuery(getMyFeedback, { feedbackId }, { throwOnError: false });
  const followUp = useMutation(followUpFeedback, {
    onSuccess: () => {
      // The prototype confirms the flip with its own toast line — the
      // badge repaint rides above the fold, the toast says what changed.
      setDraft("");
      toast.success(m.feedback_followup_ok());
      void queryClient.invalidateQueries();
    },
  });

  function backOrQueue() {
    if (router.history.canGoBack()) {
      router.history.back();
    } else {
      void navigate({ to: "/messages", search: { t: "fb" } });
    }
  }

  if (query.isPending) {
    return (
      <AmbientLayout orbs={ORBS}>
        <StateScreen variant="loading" />
      </AmbientLayout>
    );
  }

  const root = query.isSuccess && query.data !== undefined ? query.data.feedback : undefined;
  if (root === undefined) {
    return (
      <AmbientLayout orbs={ORBS}>
        <StateScreen
          action={
            <Button
              className="bg-cta-primary rounded-[12px] px-5 py-2.5 text-sm font-semibold text-white active:opacity-80"
              onPress={backOrQueue}
            >
              {m.common_back()}
            </Button>
          }
          title={m.feedback_error_title()}
          variant="empty"
        />
      </AmbientLayout>
    );
  }

  const ticket = wireFeedbackOf(root);
  const thread = (query.data?.messages ?? []).map(wireFeedbackMessageOf);
  const closed = ticket.status === STATUS_CLOSED;
  const createdMs = Date.parse(ticket.createdAt ?? "");
  const body = draft.trim();

  return (
    <AmbientLayout orbs={ORBS}>
      <div className="flex w-full flex-col gap-3.5 pb-6">
        <div className="flex items-center gap-3 pt-5">
          <button
            aria-label={m.common_back()}
            className="bg-surface-input text-ink-secondary flex h-10 w-10 shrink-0 items-center justify-center rounded-full active:opacity-60"
            onClick={backOrQueue}
            type="button"
          >
            <ChevronLeft aria-hidden size={20} />
          </button>
          <h1 className="text-ink text-[19px] font-bold">{m.feedback_detail_head()}</h1>
        </div>

        <div className={GLASS_CLASS}>
          <StatusBadge status={ticket.status} />
          <div className="text-ink-tertiary mt-1.5 text-[11px]">
            {m.feedback_detail_meta({
              time: Number.isNaN(createdMs) ? "" : localShortDateTimeOf(createdMs),
              type: categoryWordOf(ticket.category),
            })}
          </div>
          <div className="text-ink-secondary mt-3.5 text-[12.5px] leading-[1.8] whitespace-pre-line">
            {ticket.content}
          </div>
        </div>

        {thread.map((turn) => {
          const admin = turn.sender === SENDER_ADMIN;
          const ms = Date.parse(turn.createdAt ?? "");
          return (
            <div
              className={`${GLASS_CLASS} ${admin ? "border-[rgba(37,99,235,0.35)]" : ""}`}
              key={turn.id ?? turn.content}
            >
              <div className="text-ink-tertiary text-[10.5px]">
                {admin ? m.feedback_sender_admin() : m.feedback_sender_user()}
                {Number.isNaN(ms) ? "" : ` · ${localShortDateTimeOf(ms)}`}
              </div>
              <div className="text-ink mt-1.5 text-[12.5px] leading-[1.8] whitespace-pre-line">
                {turn.content}
              </div>
            </div>
          );
        })}

        {closed ? (
          <div className="text-ink-tertiary text-center text-[10.5px] leading-[1.7]">
            {m.feedback_closed_note()}
          </div>
        ) : (
          <>
            <div className="flex items-center gap-2">
              <input
                className="bg-surface-input text-ink placeholder:text-ink-tertiary min-w-0 flex-1 rounded-[12px] px-3.5 py-2.5 text-[13px] outline-none"
                maxLength={500}
                onChange={(e) => setDraft(e.target.value)}
                placeholder={m.feedback_followup_placeholder()}
                type="text"
                value={draft}
              />
              <Button
                className="bg-cta-primary shrink-0 rounded-[10px] px-4 py-2.5 text-[13px] font-bold text-white active:opacity-80 disabled:opacity-50"
                isDisabled={followUp.isPending || body === ""}
                onPress={() => followUp.mutate({ feedbackId, content: body })}
              >
                {m.feedback_followup_send()}
              </Button>
            </div>
            <div className="text-ink-tertiary text-center text-[10.5px] leading-[1.7]">
              {m.feedback_followup_note()}
            </div>
          </>
        )}
      </div>
    </AmbientLayout>
  );
}
