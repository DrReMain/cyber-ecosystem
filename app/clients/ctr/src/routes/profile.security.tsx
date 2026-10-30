import { createFileRoute, Outlet } from "@tanstack/react-router";
import { noindexHead } from "#/domains/seo";

// Layout parent for the security tree (the binding pages, the password
// pair, the merge flow as dot-named children): children render through the
// outlet — the /login tree's lesson, no child mounts without it.
export const Route = createFileRoute("/profile/security")({
  head: () => ({ meta: [...noindexHead().meta] }),
  component: () => <Outlet />,
});
