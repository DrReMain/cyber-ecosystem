import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { LoginPage } from "#/features/modules/login/page";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/login/")({
  head: () => ({
    meta: [...noindexHead().meta, { title: pageTitle(m.login_title()) }],
  }),
  component: LoginPage,
});
