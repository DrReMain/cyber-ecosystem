import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { profileFlowGuard } from "#/features/auth/guard";
import { CareerPage } from "#/features/modules/report/career-page";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/report/career")({
  beforeLoad: profileFlowGuard,
  head: () => ({
    meta: [...noindexHead().meta, { title: pageTitle(m.report_career_title()) }],
  }),
  component: CareerPage,
});
