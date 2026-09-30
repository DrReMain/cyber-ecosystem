import { PolicyService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/policy_pb";
import { ResourceService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/resource_pb";
import { RoleService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/role_pb";
import { UserService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/user_pb";
import { createFileRoute } from "@tanstack/react-router";
import { ShieldCheck } from "lucide-react";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { parseRolesSearch, RolesPage } from "#/features/modules/system-roles";
import { m } from "#/paraglide/messages";

const operations = [
  op(RoleService, "listRoles"),
  op(RoleService, "getRole"),
  op(RoleService, "listRoleMembers"),
  op(PolicyService, "listPolicies"),
  op(UserService, "listUsers"),
  op(ResourceService, "listResource"),
];

export const Route = createFileRoute("/dashboard/system/roles")({
  staticData: {
    title: "system_roles_title",
    menu: { icon: ShieldCheck, order: 15 },
    operations,
  },
  validateSearch: parseRolesSearch,
  head: () => ({ meta: [{ title: pageTitle(m.system_roles_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: RolesPage,
});
