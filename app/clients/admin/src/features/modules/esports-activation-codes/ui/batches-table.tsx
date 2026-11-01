import type { ActivationBatch } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import { Link } from "@tanstack/react-router";
import type { TablePaginationConfig, TableProps } from "antd";
import { Progress, Tag } from "antd";
import { CircleOff, CirclePlay, FileSearch } from "lucide-react";
import { ActionIconButton } from "#/components/action-icon-button";
import { m } from "#/paraglide/messages";
import { reportTypeOptions } from "../vocab";

// The prototype's restock bar: ≥95% redeemed flips the bar and adds the
// tag — a number the operator reads to mint the next batch in time.
const RESTOCK_AT = 95;

interface BatchesTableProps {
  rows: ActivationBatch[];
  loading?: boolean;
  pagination: TablePaginationConfig;
  onFlip: (row: ActivationBatch, disabled: boolean) => void;
  onRefresh: () => void;
}

export function BatchesTable({
  rows,
  loading,
  pagination,
  onFlip,
  onRefresh,
}: Readonly<BatchesTableProps>) {
  const { fieldAction, fieldTimestamp } = useColumns();

  const columns: TableProps<ActivationBatch>["columns"] = [
    {
      title: () => m.esports_activation_codes_col_batch(),
      dataIndex: "batchNo",
      width: 150,
      render: (_, row) => (
        <Link
          className="text-primary font-mono font-semibold"
          params={{ batchNo: row.batchNo ?? "" }}
          search={{ pageNo: 1, pageSize: 50 }}
          to="/dashboard/channels/activation-codes/$batchNo"
        >
          {row.batchNo}
        </Link>
      ),
    },
    {
      title: () => m.esports_activation_codes_col_type(),
      dataIndex: "reportType",
      width: 110,
      render: (_, row) => (
        <Tag className="m-0">
          {reportTypeOptions.find((o) => o.value === String(row.reportType))?.label() ??
            row.reportType}
        </Tag>
      ),
    },
    {
      // The prototype's 生成/核销 pair plus the burn-down bar, one column.
      title: () => m.esports_activation_codes_col_progress(),
      dataIndex: "usedCount",
      width: 210,
      render: (_, row) => {
        const used = Number(row.usedCount ?? 0);
        const total = Number(row.count ?? 0);
        const pct = total > 0 ? (used / total) * 100 : 0;
        return (
          <div className="flex w-44 flex-col gap-1">
            <span className="text-xs tabular-nums">
              {m.esports_activation_codes_gen_used({
                gen: total.toLocaleString("en-US"),
                used: used.toLocaleString("en-US"),
              })}
            </span>
            <div className="flex items-center gap-2">
              <Progress
                percent={Math.round(pct)}
                showInfo={false}
                size="small"
                strokeColor={pct >= RESTOCK_AT ? "#eab308" : undefined}
              />
              <span className="text-xs tabular-nums">{pct.toFixed(1)}%</span>
              {pct >= RESTOCK_AT ? (
                <Tag className="m-0" color="warning">
                  {m.esports_activation_codes_restock()}
                </Tag>
              ) : null}
            </div>
          </div>
        );
      },
    },
    {
      ...fieldTimestamp<ActivationBatch>("expiresAt", {
        title: m.esports_activation_codes_col_expires(),
        format: { time: "second" },
      }),
      width: 180,
    },
    {
      title: () => m.esports_activation_codes_col_state(),
      dataIndex: "disabledAt",
      width: 90,
      render: (_, row) =>
        row.disabledAt ? (
          <Tag className="m-0">{m.esports_activation_codes_state_disabled()}</Tag>
        ) : (
          <Tag className="m-0" color="success">
            {m.esports_activation_codes_state_live()}
          </Tag>
        ),
    },
    {
      ...fieldAction<ActivationBatch>(
        (_, row) => {
          const disabled = Boolean(row.disabledAt);
          return (
            <div className="flex gap-1">
              <ActionIconButton
                href={`/dashboard/channels/activation-codes/${row.batchNo}`}
                icon={FileSearch}
                label={m.esports_activation_codes_act_detail()}
              />
              <ActionIconButton
                color={disabled ? "default" : "danger"}
                confirm={{
                  description: disabled
                    ? m.esports_activation_codes_enable_audit()
                    : m.esports_activation_codes_disable_audit(),
                  onConfirm: () => onFlip(row, !disabled),
                  title: disabled
                    ? m.esports_activation_codes_enable_confirm({
                        no: row.batchNo ?? "",
                      })
                    : m.esports_activation_codes_disable_confirm({
                        no: row.batchNo ?? "",
                      }),
                }}
                icon={disabled ? CirclePlay : CircleOff}
                label={
                  disabled
                    ? m.esports_activation_codes_act_enable()
                    : m.esports_activation_codes_act_disable()
                }
              />
            </div>
          );
        },
        { title: m.esports_activation_codes_col_actions() },
      ),
      width: 110,
    },
  ];

  return (
    <DataTable<ActivationBatch>
      columns={columns}
      dataSource={rows}
      emptyDescription={m.esports_activation_codes_empty()}
      loading={loading}
      pagination={pagination}
      rowKey={(r) => r.batchNo ?? ""}
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
