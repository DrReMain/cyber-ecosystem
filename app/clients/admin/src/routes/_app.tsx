import { RouteTransition } from "@cyber-ecosystem/shared-theme/route-transition";
import { createFileRoute, Link, Outlet } from "@tanstack/react-router";

export const Route = createFileRoute("/_app")({
  component: AppLayout,
});

const NAV_ITEMS = [
  { to: "/", label: "Home", exact: true },
  { to: "/theme", label: "Theme", exact: false },
  { to: "/stores", label: "Stores", exact: false },
  { to: "/antd", label: "AntD", exact: false },
  { to: "/playground-connect", label: "Connect", exact: false },
  { to: "/playground-http", label: "HTTP", exact: false },
] as const;

function AppLayout() {
  return (
    <div className="flex min-h-screen">
      <aside className="flex w-48 shrink-0 flex-col border-r p-4">
        <nav className="flex flex-col gap-1">
          {NAV_ITEMS.map((item) => (
            <Link
              activeOptions={{ exact: item.exact }}
              activeProps={{ className: "font-semibold" }}
              className="rounded px-2 py-1 text-sm hover:bg-black/5 dark:hover:bg-white/10"
              key={item.to}
              to={item.to}
            >
              {item.label}
            </Link>
          ))}
        </nav>
      </aside>
      <RouteTransition as="main" className="flex-1 p-6">
        <Outlet />
      </RouteTransition>
    </div>
  );
}
