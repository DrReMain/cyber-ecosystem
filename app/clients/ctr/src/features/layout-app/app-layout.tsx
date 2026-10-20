import { Outlet } from "@tanstack/react-router";
import { useSessionWatcher } from "#/features/auth/use-session-watcher";

/**
 * ctr's app shell — the root-mounted counterpart of admin's
 * DashboardLayout: one component wrapping the whole route tree (the root
 * route's component slot, so no per-route wrapping), owning the
 * cross-cutting app concerns, starting with the live session watcher.
 * Owns no chrome and no tabbar: the tab surface spans the public home
 * and the profile overview, so TabBar stays with the tab-owning pages.
 */
export function AppLayout() {
  useSessionWatcher();
  return <Outlet />;
}
