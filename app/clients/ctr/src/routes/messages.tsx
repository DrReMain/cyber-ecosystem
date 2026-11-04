import { createFileRoute, Outlet } from "@tanstack/react-router";
import { pageTitle } from "#/config";
import { noindexHead } from "#/domains/seo";
import { sessionGuard } from "#/features/auth/guard";
import { m } from "#/paraglide/messages";

// Layout parent for the /messages tree: the letterbox index and the
// letter detail both sit behind the session guard once, here; each child
// owns its own head and URL shape.
export const Route = createFileRoute("/messages")({
  beforeLoad: sessionGuard,
  head: () => ({ meta: [...noindexHead().meta, { title: pageTitle(m.messages_head()) }] }),
  component: () => <Outlet />,
});
