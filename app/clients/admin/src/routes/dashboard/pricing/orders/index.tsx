import { OrderAdminService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/order_admin_pb";
import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { EsportsOrdersPage, parseOrdersSearch } from "#/features/modules/esports-orders";
import { m } from "#/paraglide/messages";

const operations = [op(OrderAdminService, "listPaymentOrders")];

export const Route = createFileRoute("/dashboard/pricing/orders/")({
  staticData: {
    menu: { hide: true },
    operations,
  },
  validateSearch: parseOrdersSearch,
  head: () => ({ meta: [{ title: pageTitle(m.esports_orders_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: EsportsOrdersPage,
});
