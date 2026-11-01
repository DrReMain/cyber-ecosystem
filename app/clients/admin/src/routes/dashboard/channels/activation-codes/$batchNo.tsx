import { EntitleAdminService } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin_pb";
import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import {
  ActivationBatchPage,
  parseBatchCodesSearch,
} from "#/features/modules/esports-activation-batch";
import { m } from "#/paraglide/messages";

// Reads only — the export button is a mutation and surfaces denials
// through the error domain.
const operations = [
  op(EntitleAdminService, "listActivationCodes"),
  op(EntitleAdminService, "listActivationBatches"),
];

export const Route = createFileRoute("/dashboard/channels/activation-codes/$batchNo")({
  staticData: { menu: { hide: true }, operations },
  validateSearch: parseBatchCodesSearch,
  head: () => ({ meta: [{ title: pageTitle(m.esports_activation_batch_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: () => <ActivationBatchPage batchNo={Route.useParams().batchNo} />,
});
