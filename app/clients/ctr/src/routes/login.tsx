import { createFileRoute, Outlet } from "@tanstack/react-router";

// Layout parent for the /login tree: nested children (the wechat callback)
// render through the outlet. Without this shell the child matches (head
// applies) but never mounts — the callback effect silently never runs.
export const Route = createFileRoute("/login")({
  // Same-origin next only: protocol-relative "//host" is rejected at the
  // schema boundary too (routeAfterLogin re-checks on consumption).
  // expired rides the query string as "true" (the watcher and guard link
  // here with it) and strips after the notice fires.
  validateSearch: (search: Record<string, unknown>): { expired?: boolean; next?: string } => ({
    ...(typeof search.next === "string" &&
    search.next.startsWith("/") &&
    !search.next.startsWith("//")
      ? { next: search.next }
      : {}),
    ...(search.expired === true || search.expired === "true" ? { expired: true } : {}),
  }),
  component: () => <Outlet />,
});
