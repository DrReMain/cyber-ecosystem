import { MemberService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/member_pb";
import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { EsportsUsersPage, parseUsersSearch } from "#/features/modules/esports-users";
import { m } from "#/paraglide/messages";

const operations = [op(MemberService, "listMembers")];

export const Route = createFileRoute("/dashboard/esports/users/")({
  staticData: {
    menu: { hide: true },
    operations,
  },
  validateSearch: parseUsersSearch,
  head: () => ({ meta: [{ title: pageTitle(m.esports_users_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: EsportsUsersPage,
});
