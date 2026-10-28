import { EntitleAdminService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin_pb";
import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { ReportPoliciesPage } from "#/features/modules/esports-report-policies";
import { m } from "#/paraglide/messages";

const operations = [
  op(EntitleAdminService, "listReportAccessPolicies"),
  op(EntitleAdminService, "setReportAccessPolicy"),
];

export const Route = createFileRoute("/dashboard/esports/report-policies/")({
  staticData: { title: "esports_report_policies_title", menu: { order: 20 }, operations },
  head: () => ({ meta: [{ title: pageTitle(m.esports_report_policies_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: ReportPoliciesPage,
});
