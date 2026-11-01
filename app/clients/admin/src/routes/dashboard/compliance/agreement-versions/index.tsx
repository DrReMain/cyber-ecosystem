import { AgreementAdminService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/agreement_admin_pb";
import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { AgreementVersionsPage } from "#/features/modules/esports-agreement-versions";
import { m } from "#/paraglide/messages";

// The read set is the page's one query; the create/promote/delete
// buttons surface denials through the error domain, not the page gate.
const operations = [op(AgreementAdminService, "listAgreementVersions")];

export const Route = createFileRoute("/dashboard/compliance/agreement-versions/")({
  staticData: {
    title: "esports_agreement_versions_title",
    menu: { order: 10 },
    operations,
  },
  head: () => ({ meta: [{ title: pageTitle(m.esports_agreement_versions_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: AgreementVersionsPage,
});
