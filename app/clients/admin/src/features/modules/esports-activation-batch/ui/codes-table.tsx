import type { ActivationCode } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import { Link } from "@tanstack/react-router";
import type { TablePaginationConfig, TableProps } from "antd";
import { Tag, Typography } from "antd";
import { fullTime } from "#/features/modules/esports-members/ui/dates";
import { m } from "#/paraglide/messages";

interface CodesTableProps {
  rows: ActivationCode[];
  loading?: boolean;
  pagination: TablePaginationConfig;
  onRefresh: () => void;
}

export function CodesTable({ rows, loading, pagination, onRefresh }: Readonly<CodesTableProps>) {
  const { fieldTimestamp } = useColumns();

  const columns: TableProps<ActivationCode>["columns"] = [
    {
      title: () => m.esports_activation_batch_col_code(),
      dataIndex: "code",
      width: 210,
      render: (_, row) => (
        <Typography.Text className="font-mono" copyable={{ text: row.code ?? "" }}>
          {row.code}
        </Typography.Text>
      ),
    },
    {
      title: () => m.esports_activation_batch_col_state(),
      dataIndex: "usedCount",
      width: 100,
      render: (_, row) =>
        Number(row.usedCount ?? 0) > 0 ? (
          <Tag className="m-0">{m.esports_activation_batch_state_redeemed()}</Tag>
        ) : (
          <Tag className="m-0" color="success">
            {m.esports_activation_batch_state_unused()}
          </Tag>
        ),
    },
    {
      title: () => m.esports_activation_batch_col_user(),
      dataIndex: "nickname",
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
          <Typography.Text type="secondary">-</Typography.Text>
        ),
    },
    {
      // The trail is the code row's own fact — absent on an unredeemed
      // code renders the dash, never a zero-date.
      title: () => m.esports_activation_batch_col_redeemed(),
      dataIndex: "redeemedAt",
      width: 180,
      render: (_, row) =>
        row.redeemedAt ? (
          <span className="tabular-nums">{fullTime(row.redeemedAt)}</span>
        ) : (
          <Typography.Text type="secondary">-</Typography.Text>
        ),
    },
    {
      ...fieldTimestamp<ActivationCode>("expiresAt", {
        title: m.esports_activation_batch_col_expires(),
        format: { time: "second" },
      }),
      width: 180,
    },
  ];

  return (
    <DataTable<ActivationCode>
      columns={columns}
      dataSource={rows}
      emptyDescription={m.esports_activation_batch_empty()}
      loading={loading}
      pagination={pagination}
      rowKey={(r) => r.code ?? ""}
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
