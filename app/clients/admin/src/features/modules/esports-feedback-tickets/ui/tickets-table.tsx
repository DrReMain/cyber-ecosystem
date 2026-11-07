import type { FeedbackTicket } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/feedback_admin_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import { Link } from "@tanstack/react-router";
import type { TableProps } from "antd";
import { Tag, Tooltip, Typography } from "antd";
import { Ban, FileSearch, Reply } from "lucide-react";
import { ActionIconButton } from "#/components/action-icon-button";
import { fullTime } from "#/features/modules/esports-members/ui/dates";
import { m } from "#/paraglide/messages";
import { categoryLabel, statusTags } from "../vocab";

interface TicketsTableProps {
  rows: FeedbackTicket[];
  loading?: boolean;
  pagination?: TableProps<FeedbackTicket>["pagination"];
  onClose: (row: FeedbackTicket) => void;
  onRefresh: () => void;
  onReply: (row: FeedbackTicket) => void;
}

export function TicketsTable({
  rows,
  loading,
  pagination,
  onClose,
  onRefresh,
  onReply,
}: Readonly<TicketsTableProps>) {
  const { fieldAction } = useColumns();

  const columns: TableProps<FeedbackTicket>["columns"] = [
    {
      title: () => m.esports_feedback_tickets_col_user(),
      dataIndex: "user",
      width: 150,
      render: (_, row) =>
        row.userId ? (
          <Link
            className="text-primary"
            params={{ id: row.userId }}
            to="/dashboard/people/users/$id"
          >
            {row.nickname ?? row.userId}
          </Link>
        ) : (
          // No local profile fold — the raw id, never an error.
          <Typography.Text type="secondary">-</Typography.Text>
        ),
    },
    {
      title: () => m.esports_feedback_tickets_col_category(),
      dataIndex: "category",
      width: 110,
      render: (_, row) => categoryLabel(String(row.category ?? "")),
    },
    {
      title: () => m.esports_feedback_tickets_col_content(),
      dataIndex: "content",
      render: (_, row) => (
        // The ask itself, clamped to two lines — the full text lives one
        // click away in the conversation; the queue only needs the gist.
        <Tooltip overlayStyle={{ maxWidth: 480 }} title={row.content}>
          <Typography.Paragraph className="!mb-0" ellipsis={{ rows: 2 }} style={{ maxWidth: 520 }}>
            {row.content ?? "-"}
          </Typography.Paragraph>
        </Tooltip>
      ),
    },
    {
      title: () => m.esports_feedback_tickets_col_status(),
      dataIndex: "status",
      width: 90,
      render: (_, row) => {
        const tag = statusTags[String(row.status ?? "")];
        return tag ? (
          <Tag className="m-0" color={tag.color}>
            {tag.label()}
          </Tag>
        ) : (
          String(row.status ?? "")
        );
      },
    },
    {
      title: () => m.esports_feedback_tickets_col_time(),
      dataIndex: "time",
      width: 170,
      render: (_, row) => fullTime(row.createdAt),
    },
    {
      ...fieldAction<FeedbackTicket>(
        (_, row) => (
          <div className="flex gap-1">
            <ActionIconButton
              href={`/dashboard/feedback/tickets/${row.id}`}
              icon={FileSearch}
              label={m.esports_feedback_tickets_act_detail()}
            />
            {/* Only a pending row offers the quick reply — the backend
                takes admin turns on replied tickets too, but that move
                belongs to the conversation page where the thread is in
                view; the queue's shortcut is for the fresh ask. */}
            {String(row.status) === "FEEDBACK_STATUS_PENDING" ? (
              <ActionIconButton
                color="primary"
                icon={Reply}
                label={m.esports_feedback_tickets_act_reply()}
                onClick={() => onReply(row)}
              />
            ) : null}
            {String(row.status) !== "FEEDBACK_STATUS_CLOSED" ? (
              <ActionIconButton
                color="danger"
                confirm={{
                  description: m.esports_feedback_tickets_close_confirm_desc(),
                  onConfirm: () => onClose(row),
                  title: m.esports_feedback_tickets_close_confirm_title(),
                }}
                icon={Ban}
                label={m.esports_feedback_tickets_act_close()}
              />
            ) : null}
          </div>
        ),
        { title: m.esports_feedback_tickets_col_actions() },
      ),
      width: 110,
    },
  ];

  return (
    <DataTable<FeedbackTicket>
      columns={columns}
      dataSource={rows}
      emptyDescription={m.esports_feedback_tickets_empty()}
      loading={loading}
      pagination={pagination}
      rowKey={(r) => r.id ?? ""}
      toolbar={{
        labels: {
          density: m.common_toolbar_density(),
          densityLarge: m.common_toolbar_density_large(),
          densityMiddle: m.common_toolbar_density_middle(),
          densitySmall: m.common_toolbar_density_small(),
          refresh: m.common_toolbar_refresh(),
        },
        loading,
        onRefresh,
      }}
    />
  );
}
