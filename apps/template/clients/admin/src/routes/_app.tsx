import { createFileRoute, Link, Outlet } from "@tanstack/react-router"
import { Layout } from "antd"
import { LocaleSwitcher } from "#/domains/i18n"
import { ThemeToggle } from "#/domains/theme"

export const Route = createFileRoute("/_app")({
  component: AppLayout,
})

function AppLayout() {
  return (
    <Layout className="min-h-screen">
      <header className="sticky top-0 z-50 flex items-center justify-between gap-4 border-antd-border-secondary border-b bg-antd-base/80 px-6 py-3 backdrop-blur-md">
        <div className="flex items-center gap-4">
          <Link to="/">
            <img alt="Logo" className="h-8 w-8" src="/logo.webp" />
          </Link>
          <nav className="flex items-center gap-1">
            <Link
              activeOptions={{ exact: true }}
              activeProps={{ className: "font-semibold" }}
              className="rounded px-2 py-1 text-sm transition-colors hover:bg-antd-fill"
              to="/"
            >
              Home
            </Link>
            <Link
              activeProps={{ className: "font-semibold" }}
              className="rounded px-2 py-1 text-sm transition-colors hover:bg-antd-fill"
              search={{
                pageNo: 1,
                pageSize: 10,
                sort: "sort:desc",
                status: "draft",
              }}
              to="/playground/connect"
            >
              Connect
            </Link>
            <Link
              activeProps={{ className: "font-semibold" }}
              className="rounded px-2 py-1 text-sm transition-colors hover:bg-antd-fill"
              search={{
                pageNo: 1,
                pageSize: 10,
                sort: undefined,
                status: "draft",
              }}
              to="/playground/http"
            >
              HTTP
            </Link>
            <Link
              activeProps={{ className: "font-semibold" }}
              className="rounded px-2 py-1 text-sm transition-colors hover:bg-antd-fill"
              to="/playground/file"
            >
              File
            </Link>
          </nav>
        </div>
        <div className="flex items-center gap-2">
          <LocaleSwitcher />
          <ThemeToggle />
        </div>
      </header>
      <div style={{ viewTransitionName: "route-content" }}>
        <Outlet />
      </div>
    </Layout>
  )
}
