import { createFileRoute } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { profileFlowGuard } from "#/features/auth/guard";
import { SessionDonePage } from "#/features/modules/report/session-done";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/report/session/$sessionId")({
  beforeLoad: profileFlowGuard,
  head: () => ({ meta: [...noindexHead().meta, { title: pageTitle(m.session_done_title()) }] }),
  component: SessionDonePage,
});
