import type { AdminSku } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/sku_admin_pb";
import { DataTable, useColumns } from "@cyber-ecosystem/shared-antd/table";
import type { TableProps } from "antd";
import { Tag, Typography } from "antd";
import { CircleCheckBig, CircleOff, PenLine } from "lucide-react";
import { ActionIconButton } from "#/components/action-icon-button";
import { m } from "#/paraglide/messages";
import { fenToYuan } from "./fen";

const kindTags: Record<string, { color: string; label: () => string }> = {
  SKU_KIND_TRIAL: { color: "cyan", label: () => m.esports_skus_kind_trial() },
  SKU_KIND_MEMBERSHIP: { color: "purple", label: () => m.esports_skus_kind_membership() },
};

interface SkusTableProps {
  rows: AdminSku[];
  loading?: boolean;
  onRefresh: () => void;
  onReprice: (row: AdminSku) => void;
  onToggleShelf: (row: AdminSku) => void;
}

export function SkusTable({
  rows,
  loading,
  onRefresh,
  onReprice,
  onToggleShelf,
}: Readonly<SkusTableProps>) {
  const { fieldAction } = useColumns();

  const columns: TableProps<AdminSku>["columns"] = [
    {
      title: () => m.esports_skus_col_sku(),
      dataIndex: "name",
      render: (_, row) => (
        <div className="flex flex-col">
          <span>{row.sku?.name}</span>
          <Typography.Text className="text-xs" type="secondary">
            {row.sku?.skuCode}
          </Typography.Text>
        </div>
      ),
    },
    {
      title: () => m.esports_skus_col_kind(),
      dataIndex: "kind",
      width: 90,
      render: (_, row) => {
        const tag = kindTags[String(row.sku?.kind ?? "")];
        return tag ? (
          <Tag className="m-0" color={tag.color}>
            {tag.label()}
          </Tag>
        ) : (
          String(row.sku?.kind ?? "")
        );
      },
    },
    {
      title: () => m.esports_skus_col_price(),
      dataIndex: "price",
      width: 110,
      render: (_, row) => <span className="font-semibold">{fenToYuan(row.sku?.priceFen)}</span>,
    },
    {
      title: () => m.esports_skus_col_list_price(),
      dataIndex: "listPrice",
      width: 110,
      render: (_, row) =>
        row.sku?.listPriceFen != null && row.sku?.listPriceFen !== row.sku?.priceFen ? (
          fenToYuan(row.sku?.listPriceFen)
        ) : (
          <Typography.Text type="secondary">-</Typography.Text>
        ),
    },
    {
      title: () => m.esports_skus_col_shelf(),
      dataIndex: "onShelf",
      width: 90,
      render: (_, row) =>
        row.onShelf ? (
          <Tag className="m-0" color="success">
            {m.esports_skus_shelf_on()}
          </Tag>
        ) : (
          <Tag className="m-0">{m.esports_skus_shelf_off()}</Tag>
        ),
    },
    {
      title: () => m.esports_skus_col_desc(),
      dataIndex: "description",
      render: (_, row) => row.sku?.description ?? "",
    },
    {
      ...fieldAction<AdminSku>(
        (_, row) => (
          <div className="flex gap-1">
            <ActionIconButton
              icon={PenLine}
              label={m.esports_skus_act_reprice()}
              onClick={() => onReprice(row)}
            />
            <ActionIconButton
              color={row.onShelf ? "danger" : "primary"}
              confirm={{
                description: row.onShelf ? m.esports_skus_off_audit() : m.esports_skus_on_audit(),
                onConfirm: () => onToggleShelf(row),
                title: row.onShelf
                  ? m.esports_skus_off_confirm({ name: row.sku?.name ?? "" })
                  : m.esports_skus_on_confirm({ name: row.sku?.name ?? "" }),
              }}
              icon={row.onShelf ? CircleOff : CircleCheckBig}
              label={row.onShelf ? m.esports_skus_act_off() : m.esports_skus_act_on()}
            />
          </div>
        ),
        { title: m.esports_skus_col_actions() },
      ),
      width: 130,
    },
  ];

  return (
    <DataTable<AdminSku>
      columns={columns}
      dataSource={rows}
      emptyDescription={m.esports_skus_empty()}
      loading={loading}
      rowKey={(r) => r.sku?.skuCode ?? ""}
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
