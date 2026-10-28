import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { profileFlowGuard } from "#/features/auth/guard";
import { CenterPage } from "#/features/modules/report/center-page";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/report/center")({
  beforeLoad: profileFlowGuard,
  head: () => ({
    meta: [...noindexHead().meta, { title: pageTitle(m.report_center_title()) }],
  }),
  component: CenterPage,
});
