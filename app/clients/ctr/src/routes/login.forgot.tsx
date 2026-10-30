import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { ForgotPasswordPage } from "#/features/modules/security/forgot";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/login/forgot")({
  head: () => ({
    meta: [...noindexHead().meta, { title: pageTitle(m.security_forgot_title()) }],
  }),
  component: ForgotPasswordPage,
});
