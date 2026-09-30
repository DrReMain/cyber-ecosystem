import { createFileRoute } from "@tanstack/react-router";
import { LayoutDashboard } from "lucide-react";
import { pageTitle } from "#/config";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/dashboard/")({
  staticData: { title: "layout_dashboard_workbench", menu: { icon: LayoutDashboard, order: 0 } },
  head: () => ({ meta: [{ title: pageTitle(m.layout_dashboard_workbench()) }] }),
  component: DashboardHome,
});

function DashboardHome() {
  return <div className="flex flex-col gap-4">{/* TODO */}</div>;
}
