import { EntitleAdminService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin_pb";
import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { EsportsMemberDetailPage } from "#/features/modules/esports-members";
import { m } from "#/paraglide/messages";

const operations = [op(EntitleAdminService, "getMembership")];

export const Route = createFileRoute("/dashboard/people/members/$id")({
  staticData: {
    menu: { hide: true },
    operations,
  },
  head: () => ({ meta: [{ title: pageTitle(m.esports_members_detail_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: EsportsMemberDetailPage,
});
