import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { profileFlowGuard } from "#/features/auth/guard";
import { TrendPage } from "#/features/modules/report/trend-page";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/report/trend")({
  beforeLoad: profileFlowGuard,
  head: () => ({
    meta: [...noindexHead().meta, { title: pageTitle(m.report_trend_title()) }],
  }),
  component: TrendPage,
});
