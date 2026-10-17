import { TenantService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/tenant_pb";
import { createFileRoute } from "@tanstack/react-router";
import { Layers } from "lucide-react";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { parseTenantsSearch, TenantsPage } from "#/features/modules/platform-tenants";
import { m } from "#/paraglide/messages";

const operations = [op(TenantService, "listTenants")];

export const Route = createFileRoute("/dashboard/platform/tenants")({
  staticData: {
    title: "platform_tenants_title",
    menu: { icon: Layers, order: 10 },
    operations,
  },
  validateSearch: parseTenantsSearch,
  head: () => ({ meta: [{ title: pageTitle(m.platform_tenants_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: TenantsPage,
});
