import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { AgreementPage } from "#/features/modules/agreement/page";
import { m } from "#/paraglide/messages";

// The public legal face, privacy-policy document: no guard (pre-login
// reachable), indexed (fed by SITEMAP_ROUTES in config.ts).
export const Route = createFileRoute("/agreement/privacy")({
  component: () => <AgreementPage doc="privacy" />,
  head: () => ({
    meta: [
      { name: "description", content: m.agreement_privacy_description() },
      { title: pageTitle(m.agreement_privacy_title()) },
    ],
  }),
});
