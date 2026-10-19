import { createFileRoute, Outlet } from "@tanstack/react-router";
import { noindexHead } from "#/domains/seo";
import { sessionGuard } from "#/features/auth/guard";

// Layout parent for the /profile tree (overview at the index, onboarding
// steps as dot-named children): children render through the outlet — the
// wechat-callback lesson generalizes to every nested child here. The tree
// carries the personal-flow noindex stance once at the root, and the member
// session guard once at the root too: unauthenticated hops to login before
// any profile query can fire (the admin dashboard-route law).
export const Route = createFileRoute("/profile")({
  beforeLoad: sessionGuard,
  head: () => ({ meta: [...noindexHead().meta] }),
  component: () => <Outlet />,
});
