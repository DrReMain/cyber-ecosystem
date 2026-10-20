import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { profileFlowGuard } from "#/features/auth/guard";
import { TestPage } from "#/features/modules/test/page";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/test/$code")({
  beforeLoad: profileFlowGuard,
  head: () => ({ meta: [...noindexHead().meta, { title: pageTitle(m.test_title()) }] }),
  component: TestPage,
});
