import { SkuAdminService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/sku_admin_pb";
import { createFileRoute, Outlet } from "@tanstack/react-router";
import { Tags } from "lucide-react";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";

const operations = [op(SkuAdminService, "listCatalogSkus")];

export const Route = createFileRoute("/dashboard/pricing/skus")({
  staticData: {
    title: "esports_skus_title",
    menu: { icon: Tags, order: 10 },
    operations,
  },
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: Outlet,
});
