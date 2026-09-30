import type { User } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/user_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import type { TableProps } from "antd";
import { Button, Popconfirm, Space, Tag, Tooltip, Typography } from "antd";
import { m } from "#/paraglide/messages";
import { getTextDirection } from "#/paraglide/runtime";
import type { UserRow } from "../types";
import { localPart, UserAvatar } from "./user-avatar";

interface UsersTableProps {
  rows: UserRow[];
  loading?: boolean;
  pendingId: string | null;
  avatarUrls: Map<string, string>;
  onToggle: (user: User, next: boolean) => void;
  onDelete: (user: User) => void;
  onEdit: (user: User) => void;
  onRefresh: () => void;
  onCreate: () => void;
}

export function UsersTable({
  rows,
  loading,
  pendingId,
  avatarUrls,
  onToggle,
  onDelete,
  onEdit,
  onRefresh,
  onCreate,
}: Readonly<UsersTableProps>) {
  const { fieldTimestamp, fieldAction, fixed } = useColumns();
  const columns: TableProps<UserRow>["columns"] = [
    {
      title: () => m.system_users_col_user(),
      dataIndex: "user",
      width: 240,
      fixed: fixed("left"),
      render: (_, row) => (
        <Space>
          <UserAvatar
            email={row.user.email ?? ""}
            url={row.user.avatar ? avatarUrls.get(row.user.avatar) : undefined}
          />
          <div className="flex flex-col">
            <span>{localPart(row.user.email ?? "")}</span>
            <Typography.Text className="font-mono" type="secondary">
              {row.user.email}
            </Typography.Text>
          </div>
        </Space>
      ),
    },
    {
      title: () => m.system_users_col_dept(),
      width: 140,
      render: (_, row) => row.deptName ?? <Typography.Text type="secondary">-</Typography.Text>,
    },
    {
      title: () => m.system_users_col_roles(),
      render: (_, row) =>
        row.roleViews.length > 0 ? (
          <Space size={[4, 4]} wrap>
            {row.roleViews.map((rv) => (
              <Tag className={rv.enabled ? "" : "opacity-50"} key={rv.code}>
                {rv.name}
              </Tag>
            ))}
          </Space>
        ) : (
          <Tooltip title={m.system_users_unassigned_hint()}>
            <Tag>{m.system_users_unassigned()}</Tag>
          </Tooltip>
        ),
    },
    {
      align: "center",
      title: () => m.system_users_col_status(),
      width: 110,
      render: (_, row) =>
        row.user.enabled ? (
          <Tag color="success">{m.system_users_status_on()}</Tag>
        ) : (
          <Tag>{m.system_users_status_off()}</Tag>
        ),
    },
    fieldTimestamp<UserRow>(["user", "createdAt"], { title: m.system_users_col_created_at() }),
    fieldTimestamp<UserRow>(["user", "updatedAt"], { title: m.system_users_col_updated_at() }),
    fieldAction<UserRow>(
      (_, row) => (
        <Space size={0}>
          <Button color="primary" onClick={() => onEdit(row.user)} size="small" variant="text">
            {m.system_users_act_edit()}
          </Button>
          {row.user.enabled ? (
            <Popconfirm
              cancelButtonProps={{ variant: "filled", color: "default" }}
              okButtonProps={{ variant: "filled", color: "danger" }}
              onConfirm={() => onToggle(row.user, false)}
              placement={getTextDirection() === "rtl" ? "right" : "left"}
              title={m.system_users_disable_confirm()}
            >
              <Button
                color="danger"
                disabled={pendingId === (row.user.id ?? "")}
                size="small"
                variant="text"
              >
                {m.system_users_act_disable()}
              </Button>
            </Popconfirm>
          ) : (
            <Button
              color="primary"
              disabled={pendingId === (row.user.id ?? "")}
              onClick={() => onToggle(row.user, true)}
              size="small"
              variant="text"
            >
              {m.system_users_act_enable()}
            </Button>
          )}
          <Popconfirm
            cancelButtonProps={{ variant: "filled", color: "default" }}
            okButtonProps={{ variant: "filled", color: "danger" }}
            onConfirm={() => onDelete(row.user)}
            placement={getTextDirection() === "rtl" ? "right" : "left"}
            title={m.system_users_delete_confirm()}
          >
            <Button color="danger" disabled={pendingId === row.user.id} size="small" variant="text">
              {m.system_users_act_delete()}
            </Button>
          </Popconfirm>
        </Space>
      ),
      { title: m.system_users_col_actions() },
    ),
  ];
  return (
    <DataTable<UserRow>
      columnSettings={{
        storageKey: "cyber.columns.system-users",
        locked: ["_ACTION"],
        labels: {
          title: m.common_column_settings(),
          reset: m.common_column_settings_reset(),
        },
      }}
      columns={columns}
      dataSource={rows}
      emptyAction={
        <Button color="primary" onClick={onCreate} variant="filled">
          {m.system_users_create()}
        </Button>
      }
      emptyDescription={m.system_users_empty()}
      loading={loading}
      rowClassName={(row) => (row.user.enabled ? "" : "opacity-60")}
      rowKey={(row) => row.user.id ?? ""}
      toolbar={{
        extra: (
          <Button color="primary" onClick={onCreate} variant="filled">
            {m.system_users_create()}
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
