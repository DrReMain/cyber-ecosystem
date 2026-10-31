import { SkuAdminService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/sku_admin_pb";
import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { EsportsSkusPage, parseSkusSearch } from "#/features/modules/esports-skus";
import { m } from "#/paraglide/messages";

const operations = [op(SkuAdminService, "listCatalogSkus")];

export const Route = createFileRoute("/dashboard/pricing/skus/")({
  staticData: {
    menu: { hide: true },
    operations,
  },
  validateSearch: parseSkusSearch,
  head: () => ({ meta: [{ title: pageTitle(m.esports_skus_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: EsportsSkusPage,
});
