import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { sessionGuard } from "#/features/auth/guard";
import { AgeRequiredPage } from "#/features/modules/age-required/page";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/age-required")({
  beforeLoad: sessionGuard,
  head: () => ({ meta: [...noindexHead().meta, { title: pageTitle(m.age_required_title()) }] }),
  component: AgeRequiredPage,
});
