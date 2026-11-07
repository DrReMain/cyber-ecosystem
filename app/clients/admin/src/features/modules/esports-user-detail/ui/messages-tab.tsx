import { useQuery } from "@connectrpc/connect-query";
import { listUserMessages } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/message_admin-MessageAdminService_connectquery";
import type { Message } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/message_pb";
import { Card, Table, Tag, Typography } from "antd";
import { fullTime } from "#/features/modules/esports-members/ui/dates";
import { m } from "#/paraglide/messages";

// The letter's type vocabulary keyed on the transport layer's wire names —
// the row fields carry those strings, not the enum members (the
// assessments tab's own precedent); an unknown word shows itself.
const TYPE_LABELS: Record<string, () => string> = {
  MESSAGE_TYPE_ANNOUNCEMENT: m.esports_user_detail_msg_type_announcement,
  MESSAGE_TYPE_REPORT: m.esports_user_detail_msg_type_report,
  MESSAGE_TYPE_MEMBERSHIP: m.esports_user_detail_msg_type_membership,
  MESSAGE_TYPE_FEEDBACK_REPLY: m.esports_user_detail_msg_type_feedback,
};

const typeLabel = (t: string) => TYPE_LABELS[t]?.() ?? t;

// The operator's read-only window into one member's letterbox: the same
// rows the member face lists, read state included, nothing mutated here.
export function MessagesTab({ userId }: Readonly<{ userId: string }>) {
  // A letterbox grows by whole-roster deliveries, so the newest page is
  // what an operator reconciles by; 100 rides the paginate ceiling the
  // family's bounded-fetch comment already names.
  const lettersQuery = useQuery(listUserMessages, {
    userId,
    page: { pageNo: 1, pageSize: 100 },
  });
  const letters = lettersQuery.data?.messages ?? [];
  const unread = lettersQuery.data?.unreadCount ?? 0;
  const total = lettersQuery.data?.page?.total ?? letters.length;

  return (
    <Card
      extra={
        <Typography.Text className="text-xs" type="secondary">
          {m.esports_user_detail_msg_note({ unread, total })}
        </Typography.Text>
      }
      size="small"
      title={m.esports_user_detail_msg_title()}
    >
      <Table<Message>
        columns={[
          {
            title: m.esports_user_detail_msg_col_time(),
            render: (_, r) => <span className="tabular-nums">{fullTime(r.createdAt)}</span>,
          },
          {
            title: m.esports_user_detail_msg_col_type(),
            render: (_, r) => <Tag className="m-0">{typeLabel(String(r.type))}</Tag>,
          },
          {
            title: m.esports_user_detail_msg_col_subject(),
            render: (_, r) => <span className="font-medium">{r.title || "-"}</span>,
          },
          {
            title: m.esports_user_detail_msg_col_content(),
            ellipsis: true,
            render: (_, r) => (
              <Typography.Text className="text-xs" ellipsis={{ tooltip: r.content || "-" }}>
                {r.content || "-"}
              </Typography.Text>
            ),
          },
          {
            title: m.esports_user_detail_msg_col_read(),
            render: (_, r) =>
              r.isRead ? (
                <Tag className="m-0">{m.esports_user_detail_msg_read()}</Tag>
              ) : (
                <Tag className="m-0" color="processing">
                  {m.esports_user_detail_msg_unread()}
                </Tag>
              ),
          },
        ]}
        dataSource={letters}
        loading={lettersQuery.isFetching}
        locale={{ emptyText: m.esports_user_detail_msg_empty() }}
        pagination={false}
        rowKey={(r) => r.id ?? String(r.createdAt ?? "") + r.title}
        size="small"
      />
    </Card>
  );
}
