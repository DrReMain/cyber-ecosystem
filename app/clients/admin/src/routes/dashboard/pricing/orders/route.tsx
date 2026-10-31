import { OrderAdminService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/order_admin_pb";
import { createFileRoute, Outlet } from "@tanstack/react-router";
import { ReceiptText } from "lucide-react";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";

const operations = [op(OrderAdminService, "listPaymentOrders")];

export const Route = createFileRoute("/dashboard/pricing/orders")({
  staticData: {
    title: "esports_orders_title",
    menu: { icon: ReceiptText, order: 20 },
    operations,
  },
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: Outlet,
});
