import { useQuery as useConnectQuery } from "@connectrpc/connect-query";
import { listMyFeedbacks } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/feedback-FeedbackService_connectquery";
import { listMessages } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/message-MessageService_connectquery";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { ChevronRight, Crown, FileText, Megaphone, Reply, User } from "lucide-react";
import { Button } from "react-aria-components";
import type { AmbientOrb } from "#/features/app/ambient/ambient-backdrop";
import { AmbientLayout } from "#/features/app/ambient/ambient-layout";
import { localShortDateOf } from "#/features/app/date";
import { StateScreen } from "#/features/app/states/state-screen";
import { TabBar } from "#/features/app/tabbar";
import { StatusBadge } from "#/features/modules/feedback/ui/status-badge";
import {
  STATUS_CLOSED,
  STATUS_PENDING,
  STATUS_REPLIED,
  type WireFeedback,
  wireFeedbackOf,
} from "#/features/modules/feedback/wire";
import { m } from "#/paraglide/messages";
import {
  TYPE_FEEDBACK_REPLY,
  TYPE_MEMBERSHIP,
  TYPE_REPORT,
  type WireMessage,
  wireMessageOf,
} from "./wire";

// The prototype's message surfaces carry no orb art — a plain canvas.
const MSG_ORBS: readonly AmbientOrb[] = [];

const GLASS_CLASS =
  "border-line bg-surface-card shadow-card rounded-card border p-4 backdrop-blur-md";

function messageIconOf(type: string | undefined) {
  switch (type) {
    case TYPE_REPORT:
      return FileText;
    case TYPE_MEMBERSHIP:
      return Crown;
    case TYPE_FEEDBACK_REPLY:
      return Reply;
    default:
      // Announcement and any unrecognized kind alike read as the
      // loudspeaker — the fallback face, not a wrong domain's icon.
      return Megaphone;
  }
}

// The list meta words differ from the badge words only for a replied
// ticket — the queue says who acted (管理员已回复), the badge says the
// state (已回复).
function feedbackMetaOf(status: string | undefined): string | null {
  switch (status) {
    case STATUS_REPLIED:
      return m.feedback_list_meta_replied();
    case STATUS_CLOSED:
      return m.feedback_list_meta_closed();
    case STATUS_PENDING:
      return m.feedback_list_meta_pending();
    default:
      return null;
  }
}

function MessageRow({ row }: Readonly<{ row: WireMessage }>) {
  const navigate = useNavigate();
  const Icon = messageIconOf(row.type);
  const ms = Date.parse(row.createdAt ?? "");
  return (
    <button
      className={`${GLASS_CLASS} w-full text-left active:opacity-70`}
      onClick={() => {
        if (row.id !== undefined) {
          void navigate({ to: "/messages/$messageId", params: { messageId: row.id } });
        }
      }}
      type="button"
    >
      <div className="flex items-start gap-2.75">
        <Icon aria-hidden className="text-ink-secondary mt-0.5 shrink-0" size={17} />
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-1.75">
            <span className="text-ink truncate text-[13px] font-bold">{row.title}</span>
            {row.isRead === false ? (
              <span aria-hidden className="size-1.5 shrink-0 rounded-full bg-[#EF4444]" />
            ) : null}
          </div>
          <div className="text-ink-tertiary mt-1.25 line-clamp-2 text-[11.5px] leading-[1.7]">
            {row.content}
          </div>
        </div>
        {Number.isNaN(ms) ? null : (
          <span className="text-ink-tertiary shrink-0 text-[10.5px]">{localShortDateOf(ms)}</span>
        )}
      </div>
    </button>
  );
}

function FeedbackRow({ row }: Readonly<{ row: WireFeedback }>) {
  const navigate = useNavigate();
  const meta = feedbackMetaOf(row.status);
  const ms = Date.parse(row.updatedAt ?? row.createdAt ?? "");
  return (
    <button
      className={`${GLASS_CLASS} w-full text-left active:opacity-70`}
      onClick={() => {
        if (row.id !== undefined) {
          void navigate({ to: "/feedback/$feedbackId", params: { feedbackId: row.id } });
        }
      }}
      type="button"
    >
      <div className="flex items-center gap-2">
        <div className="min-w-0 flex-1">
          <div className="text-ink truncate text-[13px] font-bold">{row.content}</div>
          <div className="text-ink-tertiary mt-1 truncate text-[11px]">
            {meta}
            {meta !== null && !Number.isNaN(ms) ? " · " : ""}
            {Number.isNaN(ms) ? "" : localShortDateOf(ms)}
          </div>
        </div>
        <StatusBadge status={row.status} />
        <ChevronRight aria-hidden className="text-ink-tertiary shrink-0" size={14} />
      </div>
    </button>
  );
}

interface MsgTabLeg {
  leg: "msg";
  pending: boolean;
  ok: boolean;
  messages: readonly WireMessage[];
  unread: number | undefined;
}
interface FbTabLeg {
  leg: "fb";
  pending: boolean;
  ok: boolean;
  feedbacks: readonly WireFeedback[];
}
type TabLeg = MsgTabLeg | FbTabLeg;

function TabPane({ leg }: Readonly<{ leg: TabLeg }>) {
  const navigate = useNavigate();
  if (leg.pending) {
    return <StateScreen variant="loading" />;
  }
  if (!leg.ok) {
    return (
      <StateScreen
        title={leg.leg === "msg" ? m.messages_error_title() : m.feedback_error_title()}
        variant="empty"
      />
    );
  }
  if (leg.leg === "msg") {
    return (
      <>
        {leg.unread !== undefined && leg.unread > 0 ? (
          <div className="bg-surface-input text-ink-tertiary rounded-card px-3.25 py-2.5 text-[11.5px]">
            <b className="text-ink">{leg.unread}</b> {m.messages_unread_unit()}
          </div>
        ) : null}
        {leg.messages.length === 0 ? (
          <div className={`${GLASS_CLASS} py-5.5 text-center`}>
            <div className="text-ink-tertiary text-[12px]">{m.messages_empty()}</div>
          </div>
        ) : (
          leg.messages.map((row) => <MessageRow key={row.id ?? row.title} row={row} />)
        )}
      </>
    );
  }
  return (
    <>
      {leg.feedbacks.length === 0 ? (
        <div className={`${GLASS_CLASS} py-5.5 text-center`}>
          <div className="text-ink-tertiary text-[12px]">{m.messages_feedback_empty()}</div>
        </div>
      ) : (
        leg.feedbacks.map((row) => <FeedbackRow key={row.id ?? row.content} row={row} />)
      )}
      <Button
        className="bg-cta-primary w-full rounded-[12px] py-3.25 text-[15px] font-bold text-white active:opacity-80"
        onPress={() => void navigate({ to: "/feedback" })}
      >
        {m.messages_feedback_new()}
      </Button>
    </>
  );
}

// The message center (prototype message-center): one page, two tabs —
// the letterbox and the member's own feedback queue. The tab is the
// URL's search (?t=), read here like the pay wall reads its own; both
// queries stay live so a tab switch shows warm data instead of a fresh
// pend.
export function MessagesPage() {
  const navigate = useNavigate();
  const { t: tab } = useSearch({ from: "/messages/" });
  const msgQuery = useConnectQuery(
    listMessages,
    { page: { pageSize: 50 } },
    { throwOnError: false },
  );
  const fbQuery = useConnectQuery(
    listMyFeedbacks,
    { page: { pageSize: 50 } },
    { throwOnError: false },
  );

  const leg: TabLeg =
    tab === "msg"
      ? {
          leg: "msg",
          pending: msgQuery.isPending,
          ok: msgQuery.isSuccess && msgQuery.data !== undefined,
          messages: (msgQuery.data?.messages ?? []).map(wireMessageOf),
          unread: msgQuery.data?.unreadCount,
        }
      : {
          leg: "fb",
          pending: fbQuery.isPending,
          ok: fbQuery.isSuccess && fbQuery.data !== undefined,
          feedbacks: (fbQuery.data?.feedbacks ?? []).map(wireFeedbackOf),
        };

  return (
    <AmbientLayout chrome={<TabBar active="messages" />} orbs={MSG_ORBS}>
      <div className="flex w-full flex-col gap-3.5 pb-6">
        <div className="flex items-center justify-between pt-5">
          <h1 className="text-ink text-[19px] font-bold">{m.messages_title()}</h1>
          <button
            aria-label={m.common_tab_me()}
            className="bg-surface-input text-ink-secondary flex h-10 w-10 items-center justify-center rounded-full active:opacity-60"
            onClick={() => void navigate({ to: "/profile" })}
            type="button"
          >
            <User aria-hidden size={20} />
          </button>
        </div>

        <div className="border-line-soft bg-surface-input/80 flex gap-1 rounded-[12px] border p-1">
          {(
            [
              ["msg", m.messages_tab_msg()],
              ["fb", m.messages_tab_feedback()],
            ] as const
          ).map(([key, label]) => (
            <button
              className={`flex-1 rounded-[9px] py-2.75 text-[13.5px] ${
                key === tab ? "bg-cta-primary font-bold text-white" : "text-ink-tertiary"
              }`}
              key={key}
              onClick={() => {
                if (key !== tab) {
                  void navigate({ to: "/messages", search: { t: key }, replace: true });
                }
              }}
              type="button"
            >
              {label}
            </button>
          ))}
        </div>

        <TabPane leg={leg} />
      </div>
    </AmbientLayout>
  );
}
