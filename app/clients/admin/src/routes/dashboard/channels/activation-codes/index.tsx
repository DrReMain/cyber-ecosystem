import { EntitleAdminService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin_pb";
import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { ActivationCodesPage, parseCodesSearch } from "#/features/modules/esports-activation-codes";
import { m } from "#/paraglide/messages";

const operations = [
  op(EntitleAdminService, "listActivationCodes"),
  op(EntitleAdminService, "generateActivationCodes"),
];

export const Route = createFileRoute("/dashboard/channels/activation-codes/")({
  staticData: { title: "esports_activation_codes_title", menu: { order: 10 }, operations },
  validateSearch: parseCodesSearch,
  head: () => ({ meta: [{ title: pageTitle(m.esports_activation_codes_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: ActivationCodesPage,
});
