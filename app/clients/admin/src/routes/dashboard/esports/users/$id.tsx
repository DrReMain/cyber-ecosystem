import { MemberService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/member_pb";
import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { EsportsUserDetailPage } from "#/features/modules/esports-user-detail";
import { m } from "#/paraglide/messages";

const operations = [op(MemberService, "getMember")];

export const Route = createFileRoute("/dashboard/esports/users/$id")({
  staticData: {
    menu: { hide: true },
    operations,
  },
  head: () => ({ meta: [{ title: pageTitle(m.esports_user_detail_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: EsportsUserDetailPage,
});
