import { OrderAdminService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/order_admin_pb";
import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { EsportsOrderDetailPage } from "#/features/modules/esports-order-detail";
import { m } from "#/paraglide/messages";

const operations = [op(OrderAdminService, "getPaymentOrder")];

export const Route = createFileRoute("/dashboard/pricing/orders/$id")({
  staticData: {
    menu: { hide: true },
    operations,
  },
  head: () => ({ meta: [{ title: pageTitle(m.esports_order_detail_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: EsportsOrderDetailPage,
});
