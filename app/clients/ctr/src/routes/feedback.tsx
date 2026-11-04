import { createFileRoute, Outlet } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { sessionGuard } from "#/features/auth/guard";
import { m } from "#/paraglide/messages";

// Layout parent for the /feedback tree: the submit form and the
// conversation detail both sit behind the session guard once, here.
export const Route = createFileRoute("/feedback")({
  beforeLoad: sessionGuard,
  head: () => ({ meta: [...noindexHead().meta, { title: pageTitle(m.feedback_head()) }] }),
  component: () => <Outlet />,
});
