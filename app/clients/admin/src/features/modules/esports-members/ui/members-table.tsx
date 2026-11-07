import type { MembershipSummary } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import type { TableProps } from "antd";
import { Button, Tag, Typography } from "antd";
import { Ban, FileSearch, RefreshCw } from "lucide-react";
import { ActionIconButton } from "#/components/action-icon-button";
import { m } from "#/paraglide/messages";
import { fullDate } from "./dates";

export const statusTags: Record<string, { color: string; label: () => string }> = {
  MEMBERSHIP_STATUS_ACTIVE: { color: "success", label: () => m.esports_members_status_active() },
  MEMBERSHIP_STATUS_EXPIRING_SOON: {
    color: "warning",
    label: () => m.esports_members_status_expiring(),
  },
  MEMBERSHIP_STATUS_EXPIRED: { color: "default", label: () => m.esports_members_status_expired() },
  MEMBERSHIP_STATUS_TERMINATED: {
    color: "error",
    label: () => m.esports_members_status_terminated(),
  },
};

const sourceLabels: Record<string, () => string> = {
  MEMBERSHIP_SOURCE_ORDER: () => m.esports_members_source_order(),
  MEMBERSHIP_SOURCE_MANUAL: () => m.esports_members_source_manual(),
  MEMBERSHIP_SOURCE_LEGACY_CODE: () => m.esports_members_source_legacy(),
};

export const sourceText = (source: string): string => sourceLabels[source]?.() ?? source;

export const isLive = (row: MembershipSummary): boolean =>
  String(row.status) === "MEMBERSHIP_STATUS_ACTIVE" ||
  String(row.status) === "MEMBERSHIP_STATUS_EXPIRING_SOON";

interface MembersTableProps {
  rows: MembershipSummary[];
  loading?: boolean;
  onCreate: () => void;
  onRefresh: () => void;
  onRenew: (row: MembershipSummary) => void;
  onTerminate: (row: MembershipSummary) => void;
  pagination?: TableProps<MembershipSummary>["pagination"];
}

export function MembersTable({
  rows,
  loading,
  onCreate,
  onRefresh,
  onRenew,
  onTerminate,
  pagination,
}: Readonly<MembersTableProps>) {
  const { fieldAction, fieldTimestamp } = useColumns();

  const columns: TableProps<MembershipSummary>["columns"] = [
    {
      title: () => m.esports_members_col_member(),
      dataIndex: "nickname",
      render: (_, row) => (
        <div className="flex flex-col">
          <span>{row.nickname || row.userId}</span>
          {row.nickname ? (
            <Typography.Text className="text-xs" type="secondary">
              {row.userId}
            </Typography.Text>
          ) : null}
        </div>
      ),
    },
    {
      title: () => m.esports_members_col_phone(),
      dataIndex: "phoneMasked",
      width: 140,
      render: (_, row) => row.phoneMasked ?? <Typography.Text type="secondary">-</Typography.Text>,
    },
    {
      ...fieldTimestamp<MembershipSummary>("grantedAt", {
        title: m.esports_members_col_granted(),
        format: { time: "second" },
      }),
      width: 170,
    },
    {
      title: () => m.esports_members_col_expires(),
      dataIndex: "expiresAt",
      width: 190,
      render: (_, row) => (
        <div className="flex flex-col">
          <span>{fullDate(row.expiresAt)}</span>
          {row.daysLeft != null && (
            <Typography.Text className="text-xs" type="secondary">
              {m.esports_members_days_left({ n: row.daysLeft })}
            </Typography.Text>
          )}
        </div>
      ),
    },
    {
      title: () => m.esports_members_col_unlocks(),
      dataIndex: "unlockCount",
      width: 140,
      render: (_, row) => m.esports_members_unlocks_count({ n: row.unlockCount ?? 0 }),
    },
    {
      title: () => m.esports_members_col_source(),
      dataIndex: "source",
      width: 110,
      render: (_, row) => sourceLabels[String(row.source ?? "")]?.() ?? String(row.source),
    },
    {
      title: () => m.esports_members_col_status(),
      dataIndex: "status",
      width: 110,
      render: (_, row) => {
        const tag = statusTags[String(row.status ?? "")];
        return tag ? (
          <Tag className="m-0" color={tag.color}>
            {tag.label()}
          </Tag>
        ) : (
          String(row.status)
        );
      },
    },
    {
      ...fieldAction<MembershipSummary>(
        (_, row) => (
          <div className="flex gap-1">
            <ActionIconButton
              href={`/dashboard/people/members/${row.userId}`}
              icon={FileSearch}
              label={m.esports_members_act_detail()}
            />
            {isLive(row) && (
              <>
                <ActionIconButton
                  icon={RefreshCw}
                  label={m.esports_members_act_renew()}
                  onClick={() => onRenew(row)}
                />
                <ActionIconButton
                  icon={Ban}
                  label={m.esports_members_act_terminate()}
                  onClick={() => onTerminate(row)}
                />
              </>
            )}
            {!isLive(row) && (
              <Typography.Text type="secondary">{m.esports_members_act_ended()}</Typography.Text>
            )}
          </div>
        ),
        { title: m.esports_members_col_actions() },
      ),
      width: 170,
    },
  ];

  return (
    <DataTable<MembershipSummary>
      columns={columns}
      dataSource={rows}
      emptyDescription={m.esports_members_empty()}
      loading={loading}
      pagination={pagination}
      rowKey="userId"
      toolbar={{
        extra: (
          <Button color="primary" onClick={onCreate} variant="filled">
            {m.esports_members_act_grant()}
          </Button>
        ),
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
