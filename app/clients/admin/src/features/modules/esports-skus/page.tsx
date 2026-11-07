import { useMutation, useQuery } from "@connectrpc/connect-query";
import type { AdminSku } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/sku_admin_pb";
import {
  createSku,
  listCatalogSkus,
  setSkuShelf,
  updateSkuPrice,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/sku_admin-SkuAdminService_connectquery";
import { Filter } from "@cyber-ecosystem/shared-antd/filter";
import { useFilter } from "@cyber-ecosystem/shared-antd/use-search";
import { useSearch } from "@tanstack/react-router";
import { Card, Typography } from "antd";
import { useState } from "react";
import { toast } from "sonner";
import { useAreaSearchStore } from "#/libs";
import { m } from "#/paraglide/messages";
import { parseSkusSearch, skusSearchSchema } from "./search";
import { RepriceModal } from "./ui/reprice-modal";
import { type SkuDraft, SkuFormModal } from "./ui/sku-form-modal";
import { SkusTable } from "./ui/skus-table";

const AREA = "/dashboard/pricing/skus/";

export function EsportsSkusPage() {
  const search = useSearch({ from: AREA });
  const store = useAreaSearchStore(AREA, search, parseSkusSearch);
  const { values, onFilter, onReset } = useFilter(store, skusSearchSchema);
  const [formOpen, setFormOpen] = useState(false);
  const [repriceRow, setRepriceRow] = useState<AdminSku | null>(null);

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

  // Shelf is a client-side fold over the unpaginated catalog — the whole
  // list is already in memory, so the filter only narrows what renders.
  const all = listQuery.data?.skus ?? [];
  const rows = all.filter((s) =>
    values.shelf === "on" ? s.onShelf : values.shelf === "off" ? !s.onShelf : true,
  );

  return (
    <div className="flex flex-col gap-4 p-4">
      <div>
        <Typography.Title className="!mb-0" level={4}>
          {m.esports_skus_title()}
        </Typography.Title>
        <Typography.Text className="text-xs" type="secondary">
          {m.esports_skus_desc()}
        </Typography.Text>
      </div>

      <Card>
        <Filter
          columns={4}
          initialValues={values}
          labels={{
            search: m.common_filter_search(),
            reset: m.common_filter_reset(),
            fold: m.common_filter_fold(),
            expand: m.common_filter_expand(),
          }}
          onFilter={onFilter}
          onReset={onReset}
          options={[
            {
              label: m.esports_skus_filter_shelf(),
              name: "shelf",
              type: "select",
              placeholder: m.esports_skus_filter_all(),
              options: [
                { label: m.esports_skus_filter_on(), value: "on" },
                { label: m.esports_skus_filter_off(), value: "off" },
              ],
            },
          ]}
        />
      </Card>

      <SkusTable
        loading={listQuery.isLoading}
        onCreate={() => setFormOpen(true)}
        onRefresh={refetch}
        onReprice={setRepriceRow}
        onToggleShelf={(row) =>
          shelfMut.mutate({ skuCode: row.sku?.skuCode, onShelf: !row.onShelf })
        }
        rows={rows}
      />
      <Typography.Text className="text-xs" type="secondary">
        {m.esports_skus_count({ n: rows.length, total: all.length })}
      </Typography.Text>

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
          repriceMut.mutate({ skuCode: repriceRow?.sku?.skuCode, priceFen })
        }
        pending={repriceMut.isPending}
        sku={repriceRow}
      />
    </div>
  );
}
