import type { Role } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/role_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import type { TableProps } from "antd";
import { Avatar, Button, Popconfirm, Space, Tag, Typography } from "antd";
import { m } from "#/paraglide/messages";
import { getTextDirection } from "#/paraglide/runtime";
import { formatGrantsSummary, summarizeGrants } from "../grants";

const AVATAR_COLORS = ["#1677ff", "#722ed1", "#13c2c2", "#fa8c16", "#eb2f96", "#52c41a"];

function avatarColor(code: string): string {
  let sum = 0;
  for (const ch of code) sum += ch.charCodeAt(0);
  return AVATAR_COLORS[sum % AVATAR_COLORS.length] ?? "#1677ff";
}

interface RolesTableProps {
  roles: Role[];
  loading?: boolean;
  pendingId: string | null;
  onToggle: (role: Role, next: boolean) => void;
  onDelete: (role: Role) => void;
  onEdit: (role: Role) => void;
  onRefresh: () => void;
  onCreate: () => void;
}

export function RolesTable({
  roles,
  loading,
  pendingId,
  onToggle,
  onDelete,
  onEdit,
  onRefresh,
  onCreate,
}: Readonly<RolesTableProps>) {
  const { fieldTimestamp, fieldAction, fixed } = useColumns();
  const columns: TableProps<Role>["columns"] = [
    {
      title: () => m.system_roles_col_role(),
      dataIndex: "name",
      width: 200,
      fixed: fixed("left"),
      render: (_, role) => (
        <Space>
          <Avatar style={{ backgroundColor: avatarColor(role.code ?? ""), flexShrink: 0 }}>
            {role.name?.charAt(0)}
          </Avatar>
          <div className="flex flex-col">
            <span>{role.name}</span>
            <Typography.Text className="font-mono" type="secondary">
              {role.code}
            </Typography.Text>
          </div>
        </Space>
      ),
    },
    {
      title: () => m.system_roles_col_remark(),
      dataIndex: "remark",
      render: (_, role) =>
        role.remark ? (
          <Typography.Text className="max-w-xs" ellipsis={{ tooltip: role.remark }}>
            {role.remark}
          </Typography.Text>
        ) : (
          <Typography.Text type="secondary">-</Typography.Text>
        ),
    },
    {
      title: () => m.system_roles_col_grants(),
      render: (_, role) =>
        formatGrantsSummary(
          summarizeGrants(
            role.globalGrant ?? false,
            Number(role.grantCount ?? 0),
            Number(role.wildcardCount ?? 0),
            Number(role.narrowedCount ?? 0),
            Number(role.policyCount ?? 0),
          ),
        ),
    },
    {
      align: getTextDirection() === "rtl" ? "left" : "right",
      title: () => m.system_roles_col_bound(),
      width: 170,
      render: (_, role) => <span className="font-mono">{Number(role.boundCount ?? 0)}</span>,
    },
    {
      align: "center",
      title: () => m.system_roles_col_status(),
      width: 110,
      render: (_, role) =>
        role.enabled ? (
          <Tag color="success">{m.system_roles_status_on()}</Tag>
        ) : (
          <Tag>{m.system_roles_status_off()}</Tag>
        ),
    },
    fieldTimestamp<Role>("createdAt", { title: m.system_roles_col_created_at() }),
    fieldTimestamp<Role>("updatedAt", { title: m.system_roles_col_updated_at() }),
    fieldAction<Role>(
      (_, role) => (
        <Space size={0}>
          <Button color="primary" onClick={() => onEdit(role)} size="small" variant="text">
            {m.system_roles_act_edit()}
          </Button>
          {role.enabled ? (
            <Popconfirm
              cancelButtonProps={{ variant: "filled", color: "default" }}
              okButtonProps={{ variant: "filled", color: "danger" }}
              onConfirm={() => onToggle(role, false)}
              placement={getTextDirection() === "rtl" ? "right" : "left"}
              title={m.system_roles_disable_confirm()}
            >
              <Button
                color="danger"
                disabled={pendingId === (role.id ?? "")}
                size="small"
                variant="text"
              >
                {m.system_roles_act_disable()}
              </Button>
            </Popconfirm>
          ) : (
            <Button
              color="primary"
              disabled={pendingId === (role.id ?? "")}
              onClick={() => onToggle(role, true)}
              size="small"
              variant="text"
            >
              {m.system_roles_act_enable()}
            </Button>
          )}
          <Popconfirm
            cancelButtonProps={{ variant: "filled", color: "default" }}
            okButtonProps={{ variant: "filled", color: "danger" }}
            onConfirm={() => onDelete(role)}
            placement={getTextDirection() === "rtl" ? "right" : "left"}
            title={m.system_roles_delete_confirm({ name: role.name ?? "" })}
          >
            <Button color="danger" disabled={pendingId === role.id} size="small" variant="text">
              {m.system_roles_act_delete()}
            </Button>
          </Popconfirm>
        </Space>
      ),
      { title: m.system_roles_col_actions() },
    ),
  ];
  return (
    <DataTable<Role>
      columnSettings={{
        storageKey: "cyber.columns.system-roles",
        locked: ["_ACTION"],
        labels: {
          title: m.common_column_settings(),
          reset: m.common_column_settings_reset(),
        },
      }}
      columns={columns}
      dataSource={roles}
      emptyAction={
        <Button color="primary" onClick={onCreate} variant="filled">
          {m.system_roles_create()}
        </Button>
      }
      emptyDescription={m.system_roles_empty()}
      loading={loading}
      rowClassName={(role) => (role.enabled ? "" : "opacity-60")}
      rowKey="id"
      toolbar={{
        extra: (
          <Button color="primary" onClick={onCreate} variant="filled">
            {m.system_roles_create()}
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
