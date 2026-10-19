import { createFileRoute, Outlet } from "@tanstack/react-router";

// Layout parent for the /login tree: nested children (the wechat callback)
// render through the outlet. Without this shell the child matches (head
// applies) but never mounts — the callback effect silently never runs.
export const Route = createFileRoute("/login")({
  // Same-origin next only: protocol-relative "//host" is rejected at the
  // schema boundary too (routeAfterLogin re-checks on consumption).
  validateSearch: (search: Record<string, unknown>): { next?: string } => ({
    ...(typeof search.next === "string" &&
    search.next.startsWith("/") &&
    !search.next.startsWith("//")
      ? { next: search.next }
      : {}),
  }),
  component: () => <Outlet />,
});
