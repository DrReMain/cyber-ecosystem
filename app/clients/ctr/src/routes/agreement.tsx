import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { AgreementPage } from "#/features/modules/agreement/page";
import type { AgreementDoc } from "#/features/modules/agreement/versions";
import { m } from "#/paraglide/messages";

// The public legal face: no guard (pre-login reachable), indexed (fed by
// SITEMAP_ROUTES in config.ts). p is loosely narrowed — any other value is
// stripped to the default user doc, so the public face always renders.
interface AgreementSearch {
  p?: AgreementDoc;
}

export const Route = createFileRoute("/agreement")({
  validateSearch: (search: Record<string, unknown>): AgreementSearch => ({
    ...(search.p === "user" || search.p === "privacy" ? { p: search.p } : {}),
  }),
  head: ({ match }) => ({
    meta: [
      {
        name: "description",
        content:
          match.search.p === "privacy"
            ? m.agreement_privacy_description()
            : m.agreement_user_description(),
      },
      {
        title: pageTitle(
          match.search.p === "privacy" ? m.agreement_privacy_title() : m.agreement_user_title(),
        ),
      },
    ],
  }),
  component: AgreementPage,
});
