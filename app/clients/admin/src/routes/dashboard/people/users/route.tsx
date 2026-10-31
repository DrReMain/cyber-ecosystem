import { MemberService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/member_pb";
import { createFileRoute, Outlet } from "@tanstack/react-router";
import { Users } from "lucide-react";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";

const operations = [op(MemberService, "listMembers")];

export const Route = createFileRoute("/dashboard/people/users")({
  staticData: {
    title: "esports_users_title",
    menu: { icon: Users, order: 10 },
    operations,
  },
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: Outlet,
});
