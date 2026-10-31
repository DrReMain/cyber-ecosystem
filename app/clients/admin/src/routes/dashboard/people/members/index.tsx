import { EntitleAdminService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin_pb";
import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { EsportsMembersPage, parseMembersSearch } from "#/features/modules/esports-members";
import { m } from "#/paraglide/messages";

const operations = [op(EntitleAdminService, "listMemberships")];

export const Route = createFileRoute("/dashboard/people/members/")({
  staticData: {
    menu: { hide: true },
    operations,
  },
  validateSearch: parseMembersSearch,
  head: () => ({ meta: [{ title: pageTitle(m.esports_members_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: EsportsMembersPage,
});
