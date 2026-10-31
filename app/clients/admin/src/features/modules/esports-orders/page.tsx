import { useMutation, useQuery } from "@connectrpc/connect-query";
import {
  closePaymentOrder,
  listPaymentOrders,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/order_admin-OrderAdminService_connectquery";
import { OrderScope, OrderState } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/order_pb";
import { Filter } from "@cyber-ecosystem/shared-antd/filter";
import { useFilter, useServerPagination } from "@cyber-ecosystem/shared-antd/use-search";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { Card, Tabs, Typography } from "antd";
import { toast } from "sonner";
import { useAreaSearchStore } from "#/libs";
import { m } from "#/paraglide/messages";
import { ordersSearchSchema, parseOrdersSearch } from "./search";
import { OrdersTable } from "./ui/orders-table";

const AREA = "/dashboard/pricing/orders/";

const TAB_TO_STATE: Record<string, OrderState | undefined> = {
  all: undefined,
  paid: OrderState.PAID,
  pending: OrderState.PENDING,
};

const SCOPE_TO_WIRE: Record<string, OrderScope | undefined> = {
  all: undefined,
  trial: OrderScope.TRIAL,
  period: OrderScope.PERIOD,
};

export function EsportsOrdersPage() {
  const search = useSearch({ from: AREA });
  const store = useAreaSearchStore(AREA, search, parseOrdersSearch);
  const navigate = useNavigate();
  const { values, onFilter, onReset } = useFilter(store, ordersSearchSchema);

  const patchSearch = (patch: Record<string, unknown>) => {
    void navigate({ to: "/dashboard/pricing/orders", search: { ...search, ...patch } });
  };

  const listQuery = useQuery(listPaymentOrders, {
    state: TAB_TO_STATE[search.state ?? "all"],
    scope: SCOPE_TO_WIRE[search.scope ?? "all"],
    keyword: search.keyword?.trim() || undefined,
    page: { pageNo: search.pageNo, pageSize: search.pageSize },
  });
  // Only a pending row offers the close action, so the mutation always
  // rides the operator's confirm — the backend arbitrates pending-only
  // anyway; the UI gate is reachability, not enforcement.
  const closeMut = useMutation(closePaymentOrder, {
    onSuccess: (resp) => {
      void toast.success(m.esports_orders_close_ok({ no: resp.order?.order?.orderNo ?? "" }));
      void listQuery.refetch();
    },
  });

  const pagination = useServerPagination(store, ordersSearchSchema, listQuery.data?.page, {
    pageSizeOptions: [20, 50, 100],
    showTotal: (t) => m.esports_orders_total({ n: t }),
  });

  return (
    <div className="flex flex-col gap-3">
      <div>
        <Typography.Title className="!mb-0" level={4}>
          {m.esports_orders_title()}
        </Typography.Title>
        <Typography.Text className="text-xs" type="secondary">
          {m.esports_orders_desc()}
        </Typography.Text>
      </div>

      <Tabs
        activeKey={search.state ?? "all"}
        items={[
          { key: "all", label: m.esports_orders_filter_all() },
          { key: "paid", label: m.esports_orders_filter_paid() },
          { key: "pending", label: m.esports_orders_filter_pending() },
        ]}
        onChange={(key) =>
          patchSearch({
            state: key === "all" ? undefined : (key as "paid" | "pending"),
            pageNo: undefined,
          })
        }
      />

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
              label: m.esports_orders_col_no(),
              name: "keyword",
              placeholder: m.esports_orders_keyword_ph(),
            },
            {
              label: m.esports_orders_col_scope(),
              name: "scope",
              type: "select",
              placeholder: m.esports_orders_scope_all(),
              options: [
                { label: m.esports_orders_scope_trial(), value: "trial" },
                { label: m.esports_orders_scope_period(), value: "period" },
              ],
            },
          ]}
        />
        <OrdersTable
          loading={listQuery.isLoading}
          onClose={(row) => closeMut.mutate({ orderNo: row.order?.orderNo })}
          onRefresh={() => listQuery.refetch()}
          pagination={pagination}
          rows={listQuery.data?.orders ?? []}
        />
      </Card>
    </div>
  );
}
