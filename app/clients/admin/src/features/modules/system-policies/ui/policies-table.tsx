import type { Policy } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/policy_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import type { TableProps } from "antd";
import { Button, Popconfirm, Space, Tag } from "antd";
import { m } from "#/paraglide/messages";
import { getTextDirection } from "#/paraglide/runtime";
import { KindTag } from "./kind-tag";
import { PolicyParamsSummary } from "./params-summary";

interface PoliciesTableProps {
  policies: Policy[];
  loading?: boolean;
  pendingId: string | null;
  onToggle: (policy: Policy, next: boolean) => void;
  onDelete: (policy: Policy) => void;
  onEdit: (policy: Policy) => void;
  onRefresh: () => void;
  onCreate: () => void;
}

export function PoliciesTable({
  policies,
  loading,
  pendingId,
  onToggle,
  onDelete,
  onEdit,
  onRefresh,
  onCreate,
}: Readonly<PoliciesTableProps>) {
  const { fieldTimestamp, fieldAction, fixed } = useColumns();
  const columns: TableProps<Policy>["columns"] = [
    {
      title: () => m.system_policies_col_name(),
      dataIndex: "name",
      width: 200,
      fixed: fixed("left"),
      render: (_, policy) => <span>{policy.name}</span>,
    },
    {
      title: () => m.system_policies_col_kind(),
      width: 110,
      render: (_, policy) => <KindTag policy={policy} />,
    },
    {
      title: () => m.system_policies_col_params(),
      render: (_, policy) => <PolicyParamsSummary policy={policy} />,
    },
    {
      align: getTextDirection() === "rtl" ? "left" : "right",
      title: () => m.system_policies_col_bound(),
      width: 100,
      render: (_, policy) => <span className="font-mono">{Number(policy.boundCount ?? 0)}</span>,
    },
    {
      align: "center",
      title: () => m.system_policies_col_status(),
      width: 110,
      render: (_, policy) =>
        policy.enabled ? (
          <Tag color="success">{m.system_policies_status_on()}</Tag>
        ) : (
          <Tag>{m.system_policies_status_off()}</Tag>
        ),
    },
    fieldTimestamp<Policy>("createdAt", { title: m.system_policies_col_created_at() }),
    fieldTimestamp<Policy>("updatedAt", { title: m.system_policies_col_updated_at() }),
    fieldAction<Policy>(
      (_, policy) => (
        <Space size={0}>
          <Button color="primary" onClick={() => onEdit(policy)} size="small" variant="text">
            {m.system_policies_act_edit()}
          </Button>
          {policy.enabled ? (
            <Popconfirm
              cancelButtonProps={{ variant: "filled", color: "default" }}
              okButtonProps={{ variant: "filled", color: "danger" }}
              onConfirm={() => onToggle(policy, false)}
              placement={getTextDirection() === "rtl" ? "right" : "left"}
              title={m.system_policies_disable_confirm()}
            >
              <Button
                color="danger"
                disabled={pendingId === (policy.id ?? "")}
                size="small"
                variant="text"
              >
                {m.system_policies_act_disable()}
              </Button>
            </Popconfirm>
          ) : (
            <Button
              color="primary"
              disabled={pendingId === (policy.id ?? "")}
              onClick={() => onToggle(policy, true)}
              size="small"
              variant="text"
            >
              {m.system_policies_act_enable()}
            </Button>
          )}
          <Popconfirm
            cancelButtonProps={{ variant: "filled", color: "default" }}
            okButtonProps={{ variant: "filled", color: "danger" }}
            onConfirm={() => onDelete(policy)}
            placement={getTextDirection() === "rtl" ? "right" : "left"}
            title={
              <span className="max-w-64">
                {Number(policy.boundCount ?? 0) > 0
                  ? m.system_policies_delete_confirm_bound({
                      name: policy.name ?? "",
                      n: Number(policy.boundCount ?? 0),
                    })
                  : m.system_policies_delete_confirm({ name: policy.name ?? "" })}
              </span>
            }
          >
            <Button color="danger" disabled={pendingId === policy.id} size="small" variant="text">
              {m.system_policies_act_delete()}
            </Button>
          </Popconfirm>
        </Space>
      ),
      { title: m.system_policies_col_actions() },
    ),
  ];
  return (
    <DataTable<Policy>
      columnSettings={{
        storageKey: "cyber.columns.system-policies",
        locked: ["_ACTION"],
        labels: {
          title: m.common_column_settings(),
          reset: m.common_column_settings_reset(),
        },
      }}
      columns={columns}
      dataSource={policies}
      emptyAction={
        <Button color="primary" onClick={onCreate} variant="filled">
          {m.system_policies_create()}
        </Button>
      }
      emptyDescription={m.system_policies_empty()}
      loading={loading}
      rowClassName={(policy) => (policy.enabled ? "" : "opacity-60")}
      rowKey="id"
      toolbar={{
        extra: (
          <Button color="primary" onClick={onCreate} variant="filled">
            {m.system_policies_create()}
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
