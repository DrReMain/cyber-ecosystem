import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { profileFlowGuard } from "#/features/auth/guard";
import { CognitivePage } from "#/features/modules/report/cognitive-page";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/report/cognitive")({
  beforeLoad: profileFlowGuard,
  head: () => ({
    meta: [...noindexHead().meta, { title: pageTitle(m.report_cognitive_title()) }],
  }),
  component: CognitivePage,
});
