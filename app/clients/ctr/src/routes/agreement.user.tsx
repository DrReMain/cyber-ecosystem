import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { AgreementPage } from "#/features/modules/agreement/page";
import { m } from "#/paraglide/messages";

// The public legal face, user-agreement document: no guard (pre-login
// reachable), indexed (fed by SITEMAP_ROUTES in config.ts).
export const Route = createFileRoute("/agreement/user")({
  component: () => <AgreementPage doc="user" />,
  head: () => ({
    meta: [
      { name: "description", content: m.agreement_user_description() },
      { title: pageTitle(m.agreement_user_title()) },
    ],
  }),
});
