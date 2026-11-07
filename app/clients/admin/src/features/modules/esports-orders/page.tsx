import { useMutation, useQuery } from "@connectrpc/connect-query";
import {
  closePaymentOrder,
  listPaymentOrders,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/order_admin-OrderAdminService_connectquery";
import { OrderScope, OrderState } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/order_pb";
import { Filter } from "@cyber-ecosystem/shared-antd/filter";
import { useFilter, useServerPagination } from "@cyber-ecosystem/shared-antd/use-search";
import { useSearch } from "@tanstack/react-router";
import { Card, Typography } from "antd";
import { toast } from "sonner";
import { useAreaSearchStore } from "#/libs";
import { m } from "#/paraglide/messages";
import { ordersSearchSchema, parseOrdersSearch } from "./search";
import { OrdersTable } from "./ui/orders-table";

const AREA = "/dashboard/pricing/orders/";

// "all" stays a legal URL value (legacy shares) but never re-enters the
// filter — a cleared select is the all case.
const STATE_TO_WIRE: Record<string, OrderState | undefined> = {
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
  const { values, onFilter, onReset } = useFilter(store, ordersSearchSchema);

  const listQuery = useQuery(listPaymentOrders, {
    state: values.state ? STATE_TO_WIRE[values.state] : undefined,
    scope: values.scope ? SCOPE_TO_WIRE[values.scope] : undefined,
    keyword: values.keyword?.trim() || undefined,
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
    <div className="flex flex-col gap-4 p-4">
      <div>
        <Typography.Title className="!mb-0" level={4}>
          {m.esports_orders_title()}
        </Typography.Title>
        <Typography.Text className="text-xs" type="secondary">
          {m.esports_orders_desc()}
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
              label: m.esports_orders_filter_no(),
              name: "keyword",
              placeholder: m.esports_orders_keyword_ph(),
            },
            {
              label: m.esports_orders_filter_state(),
              name: "state",
              type: "select",
              placeholder: m.esports_orders_filter_all(),
              options: [
                { label: m.esports_orders_filter_paid(), value: "paid" },
                { label: m.esports_orders_filter_pending(), value: "pending" },
              ],
            },
            {
              label: m.esports_orders_filter_scope(),
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
      </Card>

      <OrdersTable
        loading={listQuery.isLoading}
        onClose={(row) => closeMut.mutate({ orderNo: row.order?.orderNo })}
        onRefresh={() => listQuery.refetch()}
        pagination={pagination}
        rows={listQuery.data?.orders ?? []}
      />
    </div>
  );
}
