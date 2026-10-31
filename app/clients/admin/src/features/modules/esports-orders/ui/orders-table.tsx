import type { AdminOrder } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/order_admin_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import type { TableProps } from "antd";
import { Tag, Typography } from "antd";
import { Link } from "@tanstack/react-router";
import { Ban, FileSearch } from "lucide-react";
import { ActionIconButton } from "#/components/action-icon-button";
import { fullDate } from "#/features/modules/esports-members/ui/dates";
import { fenToYuan } from "#/features/modules/esports-skus/ui/fen";
import { m } from "#/paraglide/messages";

export const stateTags: Record<string, { color: string; label: () => string }> = {
  ORDER_STATE_PENDING: { color: "warning", label: () => m.esports_orders_state_pending() },
  ORDER_STATE_PAID: { color: "success", label: () => m.esports_orders_state_paid() },
  ORDER_STATE_CLOSED: { color: "default", label: () => m.esports_orders_state_closed() },
};

const scopeLabels: Record<string, () => string> = {
  ORDER_SCOPE_TRIAL: () => m.esports_orders_scope_trial(),
  ORDER_SCOPE_PERIOD: () => m.esports_orders_scope_period(),
};

interface OrdersTableProps {
  rows: AdminOrder[];
  loading?: boolean;
  pagination?: TableProps<AdminOrder>["pagination"];
  onClose: (row: AdminOrder) => void;
  onRefresh: () => void;
}

export function OrdersTable({ rows, loading, pagination, onClose, onRefresh }: Readonly<OrdersTableProps>) {
  const { fieldAction } = useColumns();

  const columns: TableProps<AdminOrder>["columns"] = [
    {
      title: () => m.esports_orders_col_no(),
      dataIndex: "orderNo",
      width: 190,
      render: (_, row) => <span className="font-mono text-xs">{row.order?.orderNo}</span>,
    },
    {
      title: () => m.esports_orders_col_user(),
      dataIndex: "user",
      width: 170,
      render: (_, row) =>
        row.nickname || row.phoneMasked ? (
          <div className="flex flex-col">
            <span>
              {row.userId ? (
                <Link
                  className="text-primary"
                  params={{ id: row.userId }}
                  to="/dashboard/people/users/$id"
                >
                  {row.nickname ?? "-"}
                </Link>
              ) : (
                (row.nickname ?? "-")
              )}
            </span>
            {row.phoneMasked ? (
              <Typography.Text className="text-xs" type="secondary">
                {row.phoneMasked}
              </Typography.Text>
            ) : null}
          </div>
        ) : (
          // Identity planes answered nothing — the lazy shape, not an error.
          <Typography.Text type="secondary">-</Typography.Text>
        ),
    },
    {
      title: () => m.esports_orders_col_sku(),
      dataIndex: "sku",
      render: (_, row) => row.skuName ?? row.order?.skuCode ?? "-",
    },
    {
      title: () => m.esports_orders_col_amount(),
      dataIndex: "amount",
      width: 100,
      render: (_, row) => <span className="font-semibold">{fenToYuan(row.order?.amountFen)}</span>,
    },
    {
      title: () => m.esports_orders_col_scope(),
      dataIndex: "scope",
      width: 80,
      render: (_, row) => scopeLabels[String(row.order?.scope ?? "")]?.() ?? "-",
    },
    {
      title: () => m.esports_orders_col_channel(),
      dataIndex: "channel",
      width: 100,
      render: (_, row) => row.order?.channel ?? "-",
    },
    {
      title: () => m.esports_orders_col_state(),
      dataIndex: "state",
      width: 90,
      render: (_, row) => {
        const tag = stateTags[String(row.order?.state ?? "")];
        return tag ? (
          <Tag className="m-0" color={tag.color}>
            {tag.label()}
          </Tag>
        ) : (
          String(row.order?.state ?? "")
        );
      },
    },
    {
      title: () => m.esports_orders_col_time(),
      dataIndex: "time",
      width: 170,
      render: (_, row) => fullDate(row.order?.createdAt),
    },
    {
      ...fieldAction<AdminOrder>(
        (_, row) => (
          <div className="flex gap-1">
            <ActionIconButton
              href={`/dashboard/pricing/orders/${row.order?.orderNo}`}
              icon={FileSearch}
              label={m.esports_orders_act_detail()}
            />
            {/* The wire speaks the enum's full name; the generated constant is
                the number — compare on the string, the M2-1 precedent shape. */}
            {String(row.order?.state) === "ORDER_STATE_PENDING" ? (
              <ActionIconButton
                color="danger"
                confirm={{
                  description: m.esports_orders_close_audit(),
                  onConfirm: () => onClose(row),
                  title: m.esports_orders_close_confirm({
                    no: row.order?.orderNo ?? "",
                  }),
                }}
                icon={Ban}
                label={m.esports_orders_act_close()}
              />
            ) : null}
          </div>
        ),
        { title: m.esports_orders_col_actions() },
      ),
      width: 110,
    },
  ];

  return (
    <DataTable<AdminOrder>
      columns={columns}
      dataSource={rows}
      emptyDescription={m.esports_orders_empty()}
      loading={loading}
      pagination={pagination}
      rowKey={(r) => r.order?.orderNo ?? ""}
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
