import { AgreementAdminService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/agreement_admin_pb";
import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import {
  AgreementConsentsPage,
  parseConsentsSearch,
} from "#/features/modules/esports-agreement-consents";
import { m } from "#/paraglide/messages";

// The version ledger feeds the version dropdown beside the consent
// rows — both reads belong to the page's read set.
const operations = [
  op(AgreementAdminService, "listAgreementConsents"),
  op(AgreementAdminService, "listAgreementVersions"),
];

export const Route = createFileRoute("/dashboard/compliance/agreement-consents/")({
  staticData: {
    title: "esports_agreement_consents_title",
    menu: { order: 20 },
    operations,
  },
  validateSearch: parseConsentsSearch,
  head: () => ({ meta: [{ title: pageTitle(m.esports_agreement_consents_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: AgreementConsentsPage,
});
