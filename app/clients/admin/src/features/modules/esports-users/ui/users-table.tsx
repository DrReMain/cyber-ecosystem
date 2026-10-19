import type { Member } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/member_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import type { TablePaginationConfig, TableProps } from "antd";
import { Space, Tag, Typography } from "antd";
import { Eye, Power, PowerOff } from "lucide-react";
import { ActionIconButton } from "#/components/action-icon-button";
import { UserAvatar } from "#/features/modules/system-users/ui/user-avatar";
import { m } from "#/paraglide/messages";
import { bandLabel } from "../bands";

interface UsersTableProps {
  rows: Member[];
  loading?: boolean;
  pagination: TablePaginationConfig;
  pendingId: string | null;
  onRefresh: () => void;
  onToggle: (id: string, next: boolean) => void;
}

export function UsersTable({
  rows,
  loading,
  pagination,
  pendingId,
  onRefresh,
  onToggle,
}: Readonly<UsersTableProps>) {
  const { fieldTimestamp, fieldAction, fixed } = useColumns();

  const columns: TableProps<Member>["columns"] = [
    {
      title: () => m.esports_users_col_user(),
      dataIndex: "nickname",
      fixed: fixed("left"),
      render: (_, row) => (
        <Space>
          <UserAvatar account={row.account ?? row.userId ?? ""} url={undefined} />
          <div className="flex flex-col">
            <span>{row.nickname ?? m.esports_users_nickname_unset()}</span>
            <Typography.Text className="font-mono" type="secondary">
              {row.account}
            </Typography.Text>
          </div>
        </Space>
      ),
    },
    {
      title: () => m.esports_users_col_age_band(),
      dataIndex: "ageBand",
      width: 110,
      render: (_, row) =>
        row.ageBand ? (
          <Tag className="m-0">{bandLabel(row.ageBand)}</Tag>
        ) : (
          <Typography.Text type="secondary">-</Typography.Text>
        ),
    },
    {
      ...fieldTimestamp<Member>("registeredAt", {
        title: m.esports_users_col_registered_at(),
        format: { time: "second" },
      }),
      width: 180,
    },
    {
      title: () => m.esports_users_col_status(),
      dataIndex: "enabled",
      width: 100,
      render: (_, row) =>
        row.enabled ? (
          <Tag className="m-0" color="success">
            {m.esports_users_status_on()}
          </Tag>
        ) : (
          <Tag className="m-0" color="error">
            {m.esports_users_status_off()}
          </Tag>
        ),
    },
    {
      ...fieldAction<Member>(
        (_, row) => (
          <Space size={0}>
            <ActionIconButton
              href={`/dashboard/esports/users/${row.userId}`}
              icon={Eye}
              label={m.esports_users_act_detail()}
            />
            <ActionIconButton
              confirm={
                row.enabled
                  ? {
                      title: m.esports_users_ban_confirm_title(),
                      description: m.esports_users_ban_confirm_description(),
                      onConfirm: () => onToggle(row.userId ?? "", false),
                    }
                  : {
                      title: m.esports_users_unban_confirm_title(),
                      onConfirm: () => onToggle(row.userId ?? "", true),
                    }
              }
              disabled={pendingId === (row.userId ?? "")}
              icon={row.enabled ? PowerOff : Power}
              label={row.enabled ? m.esports_users_act_ban() : m.esports_users_act_unban()}
            />
          </Space>
        ),
        { title: m.esports_users_col_actions() },
      ),
    },
  ];

  return (
    <DataTable<Member>
      columnSettings={{
        storageKey: "cyber.columns.esports-users",
        labels: {
          title: m.common_column_settings(),
          reset: m.common_column_settings_reset(),
        },
      }}
      columns={columns}
      dataSource={rows}
      emptyDescription={m.esports_users_empty()}
      loading={loading}
      pagination={pagination}
      rowKey="userId"
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
