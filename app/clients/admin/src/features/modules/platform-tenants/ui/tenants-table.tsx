import type { Tenant } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/tenant_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import type { TableProps } from "antd";
import { Avatar, Space, Tag, Typography } from "antd";
import { Power, PowerOff } from "lucide-react";
import { ActionIconButton } from "#/components/action-icon-button";
import { avatarColor } from "#/libs";
import { m } from "#/paraglide/messages";

interface TenantsTableProps {
  tenants: Tenant[];
  loading?: boolean;
  pendingId: string | null;
  onToggle: (tenant: Tenant, next: boolean) => void;
  onRefresh: () => void;
}

export function TenantsTable({
  tenants,
  loading,
  pendingId,
  onToggle,
  onRefresh,
}: Readonly<TenantsTableProps>) {
  const { fieldTimestamp, fieldAction, fixed } = useColumns();
  const columns: TableProps<Tenant>["columns"] = [
    {
      title: () => m.platform_tenants_col_tenant(),
      dataIndex: "name",
      width: 240,
      fixed: fixed("left"),
      render: (_, tenant) => (
        <Space>
          <Avatar style={{ backgroundColor: avatarColor(tenant.code ?? ""), flexShrink: 0 }}>
            {tenant.name?.charAt(0)}
          </Avatar>
          <div className="flex flex-col">
            <span>{tenant.name}</span>
            <Typography.Text className="font-mono" type="secondary">
              {tenant.code}
            </Typography.Text>
          </div>
        </Space>
      ),
    },
    {
      align: "center",
      title: () => m.platform_tenants_col_status(),
      width: 110,
      render: (_, tenant) =>
        tenant.enabled ? (
          <Tag color="success">{m.platform_tenants_status_on()}</Tag>
        ) : (
          <Tag>{m.platform_tenants_status_off()}</Tag>
        ),
    },
    fieldTimestamp<Tenant>("createdAt", { title: m.platform_tenants_col_created_at() }),
    fieldTimestamp<Tenant>("updatedAt", { title: m.platform_tenants_col_updated_at() }),
    fieldAction<Tenant>(
      (_, tenant) => (
        <Space size={0}>
          {tenant.enabled ? (
            <ActionIconButton
              confirm={{
                onConfirm: () => onToggle(tenant, false),
                title: m.platform_tenants_disable_confirm(),
              }}
              disabled={pendingId === (tenant.id ?? "")}
              icon={PowerOff}
              label={m.platform_tenants_act_disable()}
            />
          ) : (
            <ActionIconButton
              disabled={pendingId === (tenant.id ?? "")}
              icon={Power}
              label={m.platform_tenants_act_enable()}
              onClick={() => onToggle(tenant, true)}
            />
          )}
        </Space>
      ),
      { title: m.platform_tenants_col_actions() },
    ),
  ];
  return (
    <DataTable<Tenant>
      columnSettings={{
        storageKey: "cyber.columns.platform-tenants",
        locked: ["_ACTION"],
        labels: {
          title: m.common_column_settings(),
          reset: m.common_column_settings_reset(),
        },
      }}
      columns={columns}
      dataSource={tenants}
      emptyDescription={m.platform_tenants_empty()}
      loading={loading}
      rowClassName={(tenant) => (tenant.enabled ? "" : "opacity-60")}
      rowKey="id"
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
