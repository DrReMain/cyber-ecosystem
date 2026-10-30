import { createFileRoute, Outlet } from "@tanstack/react-router";
import { noindexHead } from "#/domains/seo";

// Layout parent for the wechat-bind tree (bind index, the OAuth callbacks
// as dot-named children): children render through the outlet — a leaf here
// would swallow the callbacks like /profile/security once did.
export const Route = createFileRoute("/profile/security/bind-wechat")({
  head: () => ({ meta: [...noindexHead().meta] }),
  component: () => <Outlet />,
});
