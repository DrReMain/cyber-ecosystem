import { DeptService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/dept_pb";
import { createFileRoute } from "@tanstack/react-router";
import { Building2 } from "lucide-react";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { DeptsPage, parseDeptsSearch } from "#/features/modules/system-depts";
import { m } from "#/paraglide/messages";

const operations = [op(DeptService, "listDepts"), op(DeptService, "listDeptMembers")];

export const Route = createFileRoute("/dashboard/system/depts")({
  staticData: {
    title: "system_depts_title",
    menu: { icon: Building2, order: 20 },
    operations,
  },
  validateSearch: parseDeptsSearch,
  head: () => ({ meta: [{ title: pageTitle(m.system_depts_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: DeptsPage,
});
