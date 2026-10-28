import type { Test } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/test_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import type { TableProps } from "antd";
import { Card, Tag, Typography } from "antd";
import { BookOpen, SlidersHorizontal } from "lucide-react";
import { ActionIconButton } from "#/components/action-icon-button";
import { m } from "#/paraglide/messages";
import { abilityLabel, categoryLabel } from "../bands";

interface TestsTableProps {
  rows: Test[];
  loading?: boolean;
  onRefresh: () => void;
  onOpenConfig: (row: Test) => void;
  onOpenBank: (row: Test) => void;
}

export function TestsTable({
  rows,
  loading,
  onRefresh,
  onOpenConfig,
  onOpenBank,
}: Readonly<TestsTableProps>) {
  const { fieldAction } = useColumns();

  const columns: TableProps<Test>["columns"] = [
    {
      title: () => m.esports_tests_col_code(),
      dataIndex: "code",
      render: (_, row) => <Typography.Text className="font-mono">{row.code}</Typography.Text>,
    },
    {
      title: () => m.esports_tests_col_name(),
      dataIndex: "name",
    },
    {
      title: () => m.esports_tests_col_category(),
      dataIndex: "category",
      width: 110,
      render: (_, row) => <Tag className="m-0">{categoryLabel(row.category)}</Tag>,
    },
    {
      // Label plus the vocabulary word, the prototype's 瞄准精度 aiming pair.
      title: () => m.esports_tests_col_ability(),
      dataIndex: "ability",
      width: 160,
      render: (_, row) =>
        row.ability ? (
          <div className="flex flex-col">
            <span>{abilityLabel(row.ability)}</span>
            <Typography.Text type="secondary">{row.ability}</Typography.Text>
          </div>
        ) : (
          <Typography.Text type="secondary">-</Typography.Text>
        ),
    },
    {
      title: () => m.esports_tests_col_duration(),
      dataIndex: "durationSeconds",
      width: 90,
      render: (_, row) => `${row.durationSeconds ?? 0}s`,
    },
    {
      ...fieldAction<Test>(
        (_, row) => (
          <>
            <ActionIconButton
              icon={SlidersHorizontal}
              label={m.esports_tests_act_config()}
              onClick={() => onOpenConfig(row)}
            />
            <ActionIconButton
              icon={BookOpen}
              label={m.esports_tests_act_bank()}
              onClick={() => onOpenBank(row)}
            />
          </>
        ),
        { title: m.esports_tests_col_actions() },
      ),
      width: 100,
    },
  ];

  return (
    <div className="flex flex-col gap-4">
      <DataTable<Test>
        columns={columns}
        dataSource={rows}
        emptyDescription={m.esports_tests_empty()}
        loading={loading}
        pagination={false}
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
      <Card size="small">
        <Typography.Text className="text-xs" type="secondary">
          {m.esports_tests_hint()}
        </Typography.Text>
      </Card>
    </div>
  );
}
