import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { SecurityPage } from "#/features/modules/security/page";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/profile/security/")({
  head: () => ({
    meta: [{ title: pageTitle(m.security_title()) }],
  }),
  component: SecurityPage,
});
