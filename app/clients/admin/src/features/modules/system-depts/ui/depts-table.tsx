import type { Dept } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/dept_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import type { TableProps } from "antd";
import { Button, Space, Typography } from "antd";
import { GitBranchPlus, Pencil, Trash } from "lucide-react";
import { useMemo, useState } from "react";
import { ActionIconButton } from "#/components/action-icon-button";
import { m } from "#/paraglide/messages";
import { collectRowKeys, type DeptRow } from "../types";

interface DeptsTableProps {
  rows: DeptRow[];
  loading?: boolean;
  pendingId: string | null;
  onEdit: (dept: Dept) => void;
  onAddChild: (dept: Dept) => void;
  onDelete: (dept: Dept) => void;
  onRefresh: () => void;
  onCreate: () => void;
}

export function DeptsTable({
  rows,
  loading,
  pendingId,
  onEdit,
  onAddChild,
  onDelete,
  onRefresh,
  onCreate,
}: Readonly<DeptsTableProps>) {
  const { fieldTimestamp, fieldAction } = useColumns();
  const allKeys = useMemo(() => collectRowKeys(rows), [rows]);
  const [touched, setTouched] = useState<string[] | null>(null);
  const columns: TableProps<DeptRow>["columns"] = [
    {
      title: () => m.system_depts_col_name(),
      dataIndex: ["dept", "name"],
      width: 320,
    },
    {
      title: () => m.system_depts_col_user_count(),
      dataIndex: ["dept", "userCount"],
      width: 110,
      align: "center",
      render: (_, row) => {
        const count = Number(row.dept.userCount ?? 0);
        return <span className="font-mono">{count}</span>;
      },
    },
    {
      title: () => m.system_depts_col_remark(),
      dataIndex: ["dept", "remark"],
      render: (_, row) =>
        row.dept.remark ? (
          <Typography.Text className="max-w-xs" ellipsis={{ tooltip: row.dept.remark }}>
            {row.dept.remark}
          </Typography.Text>
        ) : (
          <Typography.Text type="secondary">-</Typography.Text>
        ),
    },
    fieldTimestamp<DeptRow>(["dept", "createdAt"], { title: m.system_depts_col_created_at() }),
    fieldTimestamp<DeptRow>(["dept", "updatedAt"], { title: m.system_depts_col_updated_at() }),
    fieldAction<DeptRow>(
      (_, row) => (
        <Space size={0}>
          <ActionIconButton
            icon={Pencil}
            label={m.system_depts_act_edit()}
            onClick={() => onEdit(row.dept)}
          />
          <ActionIconButton
            icon={GitBranchPlus}
            label={m.system_depts_act_add_child()}
            onClick={() => onAddChild(row.dept)}
          />
          <ActionIconButton
            confirm={{
              description: m.system_depts_delete_confirm_desc(),
              onConfirm: () => onDelete(row.dept),
              title: m.system_depts_delete_confirm(),
            }}
            disabled={pendingId === row.dept.id}
            icon={Trash}
            label={m.system_depts_act_delete()}
          />
        </Space>
      ),
      { title: m.system_depts_col_actions() },
    ),
  ];
  return (
    <DataTable<DeptRow>
      columnSettings={{
        storageKey: "cyber.columns.system-depts",
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
          {m.system_depts_create()}
        </Button>
      }
      emptyDescription={m.system_depts_empty()}
      expandable={{
        expandedRowKeys: touched ?? allKeys,
        onExpandedRowsChange: (keys) => setTouched([...keys].map(String)),
      }}
      loading={loading}
      rowKey={(row) => row.dept.id ?? ""}
      toolbar={{
        extra: (
          <Button color="primary" onClick={onCreate} variant="filled">
            {m.system_depts_create()}
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
