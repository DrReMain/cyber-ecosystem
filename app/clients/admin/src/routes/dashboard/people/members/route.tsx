import { EntitleAdminService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin_pb";
import { createFileRoute, Outlet } from "@tanstack/react-router";
import { Crown } from "lucide-react";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";

const operations = [op(EntitleAdminService, "listMemberships")];

export const Route = createFileRoute("/dashboard/people/members")({
  staticData: {
    title: "esports_members_title",
    menu: { icon: Crown, order: 20 },
    operations,
  },
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: Outlet,
});
