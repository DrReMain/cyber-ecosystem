import { ResourceService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/resource_pb";
import { RoleService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/role_pb";
import { UserService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/user_pb";
import { createFileRoute } from "@tanstack/react-router";
import { Stethoscope } from "lucide-react";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { DiagPage, diagSearchSchema } from "#/features/modules/system-diag";
import { m } from "#/paraglide/messages";

const operations = [
  op(UserService, "listUsers"),
  op(ResourceService, "listResource"),
  op(RoleService, "listRolesByPrincipal"),
  op(RoleService, "previewGrants"),
];

export const Route = createFileRoute("/dashboard/system/diag")({
  staticData: {
    title: "system_diag_title",
    menu: { icon: Stethoscope, order: 30 },
    operations,
  },
  validateSearch: diagSearchSchema.parse,
  head: () => ({ meta: [{ title: pageTitle(m.system_diag_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: DiagPage,
});
