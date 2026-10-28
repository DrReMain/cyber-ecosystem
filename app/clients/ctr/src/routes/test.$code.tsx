import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { profileFlowGuard } from "#/features/auth/guard";
import { TestPage } from "#/features/modules/test/page";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/test/$code")({
  beforeLoad: profileFlowGuard,
  // review=1 lands straight on the stored verdict — the done step's
  // read-back entry (the prototype's ro=1), no local run.
  validateSearch: (search: Record<string, unknown>): { review?: boolean } => ({
    review: search.review === true || search.review === "1",
  }),
  head: () => ({ meta: [...noindexHead().meta, { title: pageTitle(m.test_title()) }] }),
  component: TestPage,
});
