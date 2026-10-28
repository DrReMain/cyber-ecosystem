import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { profileFlowGuard } from "#/features/auth/guard";
import { ReportPage } from "#/features/modules/report/page";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/report/$recordId")({
  beforeLoad: profileFlowGuard,
  head: () => ({ meta: [...noindexHead().meta, { title: pageTitle(m.report_title()) }] }),
  component: ReportPage,
});
