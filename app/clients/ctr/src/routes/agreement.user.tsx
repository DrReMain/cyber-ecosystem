import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { AgreementPage } from "#/features/modules/agreement/page";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/agreement/user")({
  head: () => ({
    meta: [
      { name: "description", content: m.agreement_user_description() },
      { title: pageTitle(m.agreement_user_title()) },
    ],
  }),
  component: () => <AgreementPage doc="user" />,
});
