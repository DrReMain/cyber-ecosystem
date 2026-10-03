import type { AgentConfigView } from "@cyber-ecosystem/gen-connect-ts/cyber/agent/v1/agent_config_admin_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import type { TablePaginationConfig, TableProps } from "antd";
import { Tag } from "antd";
import { m } from "#/paraglide/messages";

interface ConnectionsTableProps {
  rows: AgentConfigView[];
  loading: boolean;
  pagination: TablePaginationConfig;
  emptyDescription: string;
  onRefresh: () => void;
}

export function ConnectionsTable({
  rows,
  loading,
  pagination,
  emptyDescription,
  onRefresh,
}: Readonly<ConnectionsTableProps>) {
  const { fieldTimestamp } = useColumns();

  const columns: TableProps<AgentConfigView>["columns"] = [
    {
      title: () => m.agents_connections_col_user(),
      dataIndex: "userEmail",
      width: 240,
      render: (_, row) =>
        row.userEmail ? (
          <span>{row.userEmail}</span>
        ) : (
          <span className="font-mono text-[12px] text-ink-tertiary">{row.userId}</span>
        ),
    },
    {
      title: () => m.agents_connections_col_base_url(),
      dataIndex: "baseUrl",
      render: (_, row) => <span className="truncate font-mono text-[12px]">{row.baseUrl}</span>,
    },
    {
      title: () => m.agents_connections_col_api_key(),
      dataIndex: "apiKeySet",
      width: 110,
      render: (_, row) =>
        row.apiKeySet === true ? (
          <Tag className="m-0" color="success">
            {m.agents_connections_key_set()}
          </Tag>
        ) : (
          <Tag className="m-0">{m.agents_connections_key_unset()}</Tag>
        ),
    },
    {
      ...fieldTimestamp<AgentConfigView>("createdAt", {
        title: m.agents_connections_col_created_at(),
        format: { time: "second" },
      }),
      width: 180,
    },
  ];

  return (
    <DataTable<AgentConfigView>
      columnSettings={{
        storageKey: "cyber.columns.agents-connections",
        labels: {
          title: m.common_column_settings(),
          reset: m.common_column_settings_reset(),
        },
      }}
      columns={columns}
      dataSource={rows}
      emptyDescription={emptyDescription}
      loading={loading}
      pagination={pagination}
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
