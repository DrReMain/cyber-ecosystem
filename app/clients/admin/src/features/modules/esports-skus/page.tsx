import { useMutation, useQuery } from "@connectrpc/connect-query";
import type { AdminSku } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/sku_admin_pb";
import {
  createSku,
  listCatalogSkus,
  setSkuShelf,
  updateSkuPrice,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/sku_admin-SkuAdminService_connectquery";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { Button, Card, Tabs, Typography } from "antd";
import { toast } from "sonner";
import { Plus } from "lucide-react";
import { useState } from "react";
import { m } from "#/paraglide/messages";
import { RepriceModal } from "./ui/reprice-modal";
import { type SkuDraft, SkuFormModal } from "./ui/sku-form-modal";
import { SkusTable } from "./ui/skus-table";

const AREA = "/dashboard/pricing/skus/";

export function EsportsSkusPage() {
  const search = useSearch({ from: AREA });
  const navigate = useNavigate();
  const [formOpen, setFormOpen] = useState(false);
  const [repriceRow, setRepriceRow] = useState<AdminSku | null>(null);

  const patchSearch = (patch: Record<string, unknown>) => {
    void navigate({ to: "/dashboard/pricing/skus", search: { ...search, ...patch } });
  };

  const listQuery = useQuery(listCatalogSkus, {});
  const refetch = () => listQuery.refetch();

  const createMut = useMutation(createSku, {
    onSuccess: () => {
      setFormOpen(false);
      void toast.success(m.esports_skus_created());
      void refetch();
    },
  });
  const repriceMut = useMutation(updateSkuPrice, {
    onSuccess: (resp) => {
      setRepriceRow(null);
      void toast.success(m.esports_skus_repriced({ name: resp.sku?.sku?.name ?? "" }));
      void refetch();
    },
  });
  const shelfMut = useMutation(setSkuShelf, {
    onSuccess: (resp) => {
      void toast.success(
        resp.sku?.onShelf
          ? m.esports_skus_on_ok({ name: resp.sku?.sku?.name ?? "" })
          : m.esports_skus_off_ok({ name: resp.sku?.sku?.name ?? "" }),
      );
      void refetch();
    },
  });

  const all = listQuery.data?.skus ?? [];
  const rows = all.filter((s) =>
    search.shelf === "on" ? s.onShelf : search.shelf === "off" ? !s.onShelf : true,
  );

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-start justify-between">
        <div>
          <Typography.Title className="!mb-0" level={4}>
            {m.esports_skus_title()}
          </Typography.Title>
          <Typography.Text className="text-xs" type="secondary">
            {m.esports_skus_desc()}
          </Typography.Text>
        </div>
        <Button
          icon={<Plus size={14} />}
          onClick={() => setFormOpen(true)}
          size="small"
          type="primary"
        >
          {m.esports_skus_new()}
        </Button>
      </div>

      <Tabs
        activeKey={search.shelf ?? "all"}
        items={[
          { key: "all", label: m.esports_skus_filter_all() },
          { key: "on", label: m.esports_skus_filter_on() },
          { key: "off", label: m.esports_skus_filter_off() },
        ]}
        onChange={(key) =>
          patchSearch({ shelf: key === "all" ? undefined : (key as "on" | "off") })
        }
      />

      <Card>
        <SkusTable
          loading={listQuery.isLoading}
          onRefresh={refetch}
          onReprice={setRepriceRow}
          onToggleShelf={(row) =>
            shelfMut.mutate({ skuCode: row.sku?.skuCode, onShelf: !row.onShelf })
          }
          rows={rows}
        />
        <Typography.Text className="mt-2 block text-xs" type="secondary">
          {m.esports_skus_count({ n: rows.length, total: all.length })}
        </Typography.Text>
      </Card>

      <SkuFormModal
        onCancel={() => setFormOpen(false)}
        onSubmit={(draft: SkuDraft) =>
          createMut.mutate({
            name: draft.name,
            kind: draft.kind,
            priceFen: draft.priceFen,
            ...(draft.periodDays ? { periodDays: draft.periodDays } : {}),
            ...(draft.description ? { description: draft.description } : {}),
          })
        }
        open={formOpen}
        pending={createMut.isPending}
      />
      <RepriceModal
        onCancel={() => setRepriceRow(null)}
        onSubmit={(priceFen) =>
          repriceRow?.sku?.skuCode &&
          repriceMut.mutate({ skuCode: repriceRow.sku?.skuCode, priceFen })
        }
        pending={repriceMut.isPending}
        sku={repriceRow}
      />
    </div>
  );
}
