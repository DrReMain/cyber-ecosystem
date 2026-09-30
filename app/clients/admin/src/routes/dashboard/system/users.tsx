import { DeptService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/dept_pb";
import { ResourceService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/resource_pb";
import { RoleService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/role_pb";
import { UserService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/user_pb";
import { createFileRoute } from "@tanstack/react-router";
import { Users } from "lucide-react";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { parseUsersSearch, UsersPage } from "#/features/modules/system-users";
import { m } from "#/paraglide/messages";

const operations = [
  op(UserService, "listUsers"),
  op(RoleService, "listRoles"),
  op(DeptService, "listDepts"),
  op(ResourceService, "listResource"),
  op(RoleService, "previewGrants"),
];

export const Route = createFileRoute("/dashboard/system/users")({
  staticData: {
    title: "system_users_title",
    menu: { icon: Users, order: 10 },
    operations,
  },
  validateSearch: parseUsersSearch,
  head: () => ({ meta: [{ title: pageTitle(m.system_users_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: UsersPage,
});
