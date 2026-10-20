import { createFileRoute, Outlet } from "@tanstack/react-router";
import { noindexHead } from "#/domains/seo";
import { profileFlowGuard } from "#/features/auth/guard";

// Layout parent for the /match flow tree: the personal-flow noindex stance
// once at the root, and the flow-entry guard — authed plus a complete
// profile (the same completeness law the post-login router enforces).
export const Route = createFileRoute("/match")({
  beforeLoad: profileFlowGuard,
  head: () => ({ meta: [...noindexHead().meta] }),
  component: () => <Outlet />,
});
