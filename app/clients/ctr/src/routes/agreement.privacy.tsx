import { createQueryOptions } from "@connectrpc/connect-query";
import { getAgreementCurrent } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/agreement-AgreementService_connectquery";
import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { AgreementPage } from "#/features/modules/agreement/page";
import { m } from "#/paraglide/messages";
import { connectTransport } from "#/services/connect/transport";

export const Route = createFileRoute("/agreement/privacy")({
  head: () => ({
    meta: [
      { name: "description", content: m.agreement_privacy_description() },
      { title: pageTitle(m.agreement_privacy_title()) },
    ],
  }),
  // The version line's truth rides the loader: the server's current
  // version answers before first paint; a fetch failure only costs the
  // line, never the page.
  loader: async ({ context }) => {
    await context.queryClient
      .query({ ...createQueryOptions(getAgreementCurrent, {}, { transport: connectTransport }) })
      .catch(() => undefined);
  },
  component: () => <AgreementPage doc="privacy" />,
});
