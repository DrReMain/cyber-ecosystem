import { useQuery as useConnectQuery, useMutation } from "@connectrpc/connect-query";
import {
  listMessages,
  markMessageRead,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/message-MessageService_connectquery";
import { useQueryClient } from "@tanstack/react-query";
import { useNavigate, useParams, useRouter } from "@tanstack/react-router";
import { ChevronLeft } from "lucide-react";
import { useEffect, useRef } from "react";
import { Button } from "react-aria-components";
import type { AmbientOrb } from "#/features/app/ambient/ambient-backdrop";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { localDateOf } from "#/features/app/date";
import { StateScreen } from "#/features/app/states/state-screen";
import { m } from "#/paraglide/messages";
import { TYPE_FEEDBACK_REPLY, TYPE_MEMBERSHIP, TYPE_REPORT, wireMessageOf } from "./wire";

const DETAIL_ORBS: readonly AmbientOrb[] = [];

const GLASS_CLASS =
  "border-line bg-surface-card shadow-card rounded-card border p-4 backdrop-blur-md";

// The notice detail (prototype feedback-detail?kind=notice): the
// letter's full body, the meta line naming its kind, and the one deep
// link the kind carries — a report letter points at the report, a
// feedback reply at the conversation.
export function MessageDetailPage() {
  const navigate = useNavigate();
  const router = useRouter();
  const { messageId } = useParams({ from: "/messages/$messageId" });
  const queryClient = useQueryClient();
  // No single-letter read in the contract — the letterbox list is a
  // letter's only home, so the detail finds its row there; a deep link
  // loads the same query the center already warms.
  const listQuery = useConnectQuery(
    listMessages,
    { page: { pageSize: 50 } },
    { throwOnError: false },
  );
  const markRead = useMutation(markMessageRead, {
    onSuccess: () => void queryClient.invalidateQueries(),
  });

  const raw =
    listQuery.isSuccess && listQuery.data !== undefined
      ? listQuery.data.messages.find((letter) => letter.id === messageId)
      : undefined;
  const found = raw !== undefined ? wireMessageOf(raw) : undefined;

  // Reading is the visit: opening the letter flips its row server-side,
  // and the blanket invalidation walks the badge and the list dot back
  // to truth. Idempotent by contract, so StrictMode's double pass and a
  // racing re-read are both no-ops.
  const markedRef = useRef(false);
  useEffect(() => {
    if (
      found !== undefined &&
      found.isRead === false &&
      found.id !== undefined &&
      !markedRef.current
    ) {
      markedRef.current = true;
      markRead.mutate({ messageId: found.id });
    }
  });

  function backOrCenter() {
    if (router.history.canGoBack()) {
      router.history.back();
    } else {
      void navigate({ to: "/messages", search: { t: "msg" } });
    }
  }

  if (listQuery.isPending) {
    return (
      <AmbientLayout orbs={DETAIL_ORBS}>
        <StateScreen variant="loading" />
      </AmbientLayout>
    );
  }
  if (found === undefined) {
    return (
      <AmbientLayout orbs={DETAIL_ORBS}>
        <StateScreen
          action={
            <Button
              className="bg-cta-primary rounded-[12px] px-5 py-2.5 text-sm font-semibold text-white active:opacity-80"
              onPress={backOrCenter}
            >
              {m.common_back()}
            </Button>
          }
          title={listQuery.isSuccess ? m.messages_not_found() : m.messages_error_title()}
          variant="empty"
        />
      </AmbientLayout>
    );
  }

  const ms = Date.parse(found.createdAt ?? "");
  const reportCta = found.type === TYPE_REPORT && found.refId !== undefined;
  const feedbackCta = found.type === TYPE_FEEDBACK_REPLY && found.refId !== undefined;

  return (
    <AmbientLayout orbs={DETAIL_ORBS}>
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
          <h1 className="text-ink text-[19px] font-bold">{m.messages_notice_head()}</h1>
        </div>

        <div className={GLASS_CLASS}>
          <div className="text-ink text-[15px] font-bold">{found.title}</div>
          <div className="text-ink-tertiary mt-1.5 text-[11px]">
            {typeWordOf(found.type)}
            {Number.isNaN(ms) ? "" : ` · ${localDateOf(ms)}`}
          </div>
          <div className="text-ink-secondary mt-3.5 text-[12.5px] leading-[1.8] whitespace-pre-line">
            {found.content}
          </div>
        </div>

        {reportCta ? (
          <Button
            className="bg-cta-primary w-full rounded-[12px] py-3.25 text-[15px] font-bold text-white active:opacity-80"
            onPress={() =>
              void navigate({
                to: "/report/$recordId",
                params: { recordId: found.refId as string },
              })
            }
          >
            {m.messages_cta_report()}
          </Button>
        ) : null}
        {feedbackCta ? (
          <Button
            className="bg-cta-primary w-full rounded-[12px] py-3.25 text-[15px] font-bold text-white active:opacity-80"
            onPress={() =>
              void navigate({
                to: "/feedback/$feedbackId",
                params: { feedbackId: found.refId as string },
              })
            }
          >
            {m.messages_cta_feedback()}
          </Button>
        ) : null}
      </div>
    </AmbientLayout>
  );
}

function typeWordOf(type: string | undefined): string {
  switch (type) {
    case TYPE_REPORT:
      return m.messages_type_report();
    case TYPE_MEMBERSHIP:
      return m.messages_type_membership();
    case TYPE_FEEDBACK_REPLY:
      return m.messages_type_feedback_reply();
    default:
      return m.messages_type_announcement();
  }
}
