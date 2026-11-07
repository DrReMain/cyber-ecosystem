import { useMutation, useQuery } from "@connectrpc/connect-query";
import {
  getFeedback,
  replyFeedback,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/feedback_admin-FeedbackAdminService_connectquery";
import { Link, useNavigate, useParams } from "@tanstack/react-router";
import { Button, Card, Descriptions, Input, Tag, Typography } from "antd";
import { ArrowLeft, RefreshCw, Send } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { categoryLabel, statusTags } from "#/features/modules/esports-feedback-tickets/vocab";
import { fullTime } from "#/features/modules/esports-members/ui/dates";
import { m } from "#/paraglide/messages";

export function FeedbackDetailPage() {
  const { feedbackId } = useParams({ from: "/dashboard/feedback/tickets/$feedbackId" });
  const navigate = useNavigate();
  const detailQuery = useQuery(getFeedback, { feedbackId });
  const [draft, setDraft] = useState("");

  const replyMut = useMutation(replyFeedback, {
    onSuccess: () => {
      setDraft("");
      void toast.success(m.esports_feedback_detail_reply_ok());
      void detailQuery.refetch();
    },
  });

  const d = detailQuery.data;
  const ticket = d?.ticket ?? null;
  const tag = statusTags[String(ticket?.status ?? "")];
  const closed = String(ticket?.status) === "FEEDBACK_STATUS_CLOSED";
  const trimmed = draft.trim();
  const asker = ticket?.nickname ?? ticket?.userId ?? "-";

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <Button
          icon={<ArrowLeft size={14} />}
          onClick={() =>
            void navigate({
              to: "/dashboard/feedback/tickets",
              search: { pageNo: 1, pageSize: 20 },
            })
          }
          size="small"
          type="text"
        >
          {m.esports_feedback_detail_back()}
        </Button>
        <Button
          icon={<RefreshCw size={14} />}
          loading={detailQuery.isFetching}
          onClick={() => void detailQuery.refetch()}
          size="small"
          type="text"
        >
          {m.common_toolbar_refresh()}
        </Button>
      </div>

      <Card loading={detailQuery.isLoading}>
        <div className="mb-4 flex items-center gap-3">
          <Typography.Title className="!mb-0" level={4}>
            {m.esports_feedback_detail_title()}
          </Typography.Title>
          <span className="font-mono text-sm">#{feedbackId}</span>
          {tag ? (
            <Tag className="m-0" color={tag.color}>
              {tag.label()}
            </Tag>
          ) : null}
        </div>

        <Descriptions bordered column={2} size="small">
          <Descriptions.Item label={m.esports_feedback_detail_user()}>
            {ticket?.userId ? (
              <Link
                className="text-primary"
                params={{ id: ticket.userId }}
                to="/dashboard/people/users/$id"
              >
                {asker}
              </Link>
            ) : (
              "-"
            )}
          </Descriptions.Item>
          <Descriptions.Item label={m.esports_feedback_detail_category()}>
            {categoryLabel(String(ticket?.category ?? ""))}
          </Descriptions.Item>
          <Descriptions.Item label={m.esports_feedback_detail_time()}>
            {fullTime(ticket?.createdAt)}
          </Descriptions.Item>
          {ticket?.contact ? (
            // The member's own way back — the one channel they chose to
            // be reached on, not a system field the ticket never carried.
            <Descriptions.Item label={m.esports_feedback_detail_contact()} span={2}>
              {ticket.contact}
            </Descriptions.Item>
          ) : null}
          <Descriptions.Item label={m.esports_feedback_detail_content()} span={2}>
            {ticket?.content ?? "-"}
          </Descriptions.Item>
        </Descriptions>

        <div className="mt-4 flex flex-col gap-3">
          <Typography.Text strong>{m.esports_feedback_detail_thread()}</Typography.Text>
          {(d?.messages ?? []).map((msg) => {
            const admin = String(msg.sender) === "FEEDBACK_MESSAGE_SENDER_ADMIN";
            // The member's bubble signs with their own nickname (data,
            // not copy); the operator's signs with the seat's fixed label.
            const sender = admin ? m.esports_feedback_detail_sender_admin() : asker;
            return (
              <div
                className={
                  admin
                    ? "border-primary/40 bg-primary/5 rounded-lg border px-3 py-2"
                    : "rounded-lg border px-3 py-2"
                }
                key={msg.id ?? ""}
              >
                <div className="mb-1 flex items-baseline gap-2">
                  <Typography.Text className="text-xs" type="secondary">
                    {sender}
                  </Typography.Text>
                  <Typography.Text className="text-xs" type="secondary">
                    {fullTime(msg.createdAt)}
                  </Typography.Text>
                </div>
                <Typography.Text className="whitespace-pre-wrap">{msg.content}</Typography.Text>
              </div>
            );
          })}

          {closed ? null : (
            // Closed is terminal — the box leaves with the conversation;
            // a new question is a new ticket, not a reopened one.
            <div className="flex items-start gap-2">
              <Input.TextArea
                autoSize={{ maxRows: 8, minRows: 3 }}
                className="flex-1"
                maxLength={2000}
                onChange={(e) => setDraft(e.target.value)}
                placeholder={m.esports_feedback_detail_reply_placeholder()}
                showCount
                value={draft}
              />
              <Button
                disabled={!trimmed}
                icon={<Send aria-hidden size={14} />}
                loading={replyMut.isPending}
                onClick={() => trimmed && void replyMut.mutate({ feedbackId, content: trimmed })}
                type="primary"
              >
                {m.esports_feedback_detail_reply_send()}
              </Button>
            </div>
          )}
        </div>
      </Card>
    </div>
  );
}
