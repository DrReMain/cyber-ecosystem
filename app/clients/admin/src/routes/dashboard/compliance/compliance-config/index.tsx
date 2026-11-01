import { ComplianceAdminService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/compliance_admin_pb";
import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { ComplianceConfigPage } from "#/features/modules/esports-compliance-config";
import { m } from "#/paraglide/messages";

// The read set is the page's one query; the three setters surface denials
// through the error domain, not the page gate.
const operations = [op(ComplianceAdminService, "getComplianceConfig")];

export const Route = createFileRoute("/dashboard/compliance/compliance-config/")({
  staticData: {
    title: "esports_compliance_config_title",
    menu: { order: 30 },
    operations,
  },
  head: () => ({ meta: [{ title: pageTitle(m.esports_compliance_config_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: ComplianceConfigPage,
});
