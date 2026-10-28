import type { ActivationCode } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import type { TablePaginationConfig, TableProps } from "antd";
import { Progress, Tag, Typography } from "antd";
import { m } from "#/paraglide/messages";

interface CodesTableProps {
  rows: ActivationCode[];
  loading?: boolean;
  pagination: TablePaginationConfig;
  onRefresh: () => void;
}

const typeLabels: Record<string, () => string> = {
  REPORT_TYPE_COGNITIVE: () => m.esports_report_type_cognitive(),
  REPORT_TYPE_CAREER: () => m.esports_report_type_career(),
  REPORT_TYPE_SCHOOL: () => m.esports_report_type_school(),
};

export function CodesTable({ rows, loading, pagination, onRefresh }: Readonly<CodesTableProps>) {
  const { fieldTimestamp } = useColumns();

  const columns: TableProps<ActivationCode>["columns"] = [
    {
      title: () => m.esports_activation_codes_col_code(),
      dataIndex: "code",
      render: (_, row) => (
        <Typography.Text className="font-mono" copyable={{ text: row.code ?? "" }}>
          {row.code}
        </Typography.Text>
      ),
    },
    {
      title: () => m.esports_activation_codes_col_type(),
      dataIndex: "reportType",
      width: 130,
      render: (_, row) => (
        <Tag className="m-0">{typeLabels[row.reportType ?? ""]?.() ?? row.reportType}</Tag>
      ),
    },
    {
      // The prototype's 生成/核销/核销进度 triple, one column: the count
      // pair plus the bar, so a half-spent batch reads at a glance.
      title: () => m.esports_activation_codes_col_progress(),
      dataIndex: "usedCount",
      width: 180,
      render: (_, row) => {
        const used = Number(row.usedCount ?? 0);
        const total = Number(row.quantity ?? 0);
        return (
          <div className="flex w-36 flex-col gap-1">
            <span className="text-xs">
              {used} / {total}
            </span>
            <Progress
              percent={total > 0 ? Math.round((used / total) * 100) : 0}
              showInfo={false}
              size="small"
              status={used >= total && total > 0 ? "exception" : "active"}
            />
          </div>
        );
      },
    },
    {
      ...fieldTimestamp<ActivationCode>("expiresAt", {
        title: m.esports_activation_codes_col_expires(),
        format: { time: "second" },
      }),
      width: 180,
    },
    {
      ...fieldTimestamp<ActivationCode>("createdAt", {
        title: m.esports_activation_codes_col_created(),
        format: { time: "second" },
      }),
      width: 180,
    },
  ];

  return (
    <DataTable<ActivationCode>
      columns={columns}
      dataSource={rows}
      emptyDescription={m.esports_activation_codes_empty()}
      loading={loading}
      pagination={pagination}
      rowKey="code"
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
