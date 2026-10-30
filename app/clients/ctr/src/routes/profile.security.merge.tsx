import { createFileRoute, Outlet } from "@tanstack/react-router";
import { noindexHead } from "#/domains/seo";

// Layout parent for the merge tree (confirm index, success, the two OAuth
// callbacks as dot-named children): children render through the outlet —
// a leaf here would swallow /merge/success like /profile/security once did.
export const Route = createFileRoute("/profile/security/merge")({
  head: () => ({ meta: [...noindexHead().meta] }),
  component: () => <Outlet />,
});
