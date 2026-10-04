import { createFileRoute } from "@tanstack/react-router";
import { LayoutDashboard } from "lucide-react";
import { pageTitle } from "#/config";
import { AgentsChatPage, agentsChatSearchSchema } from "#/features/modules/agents-chat";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/dashboard/")({
  staticData: {
    title: "layout_dashboard_workbench",
    menu: { icon: LayoutDashboard, order: 0 },
    fullHeight: true,
  },
  validateSearch: agentsChatSearchSchema.parse,
  head: () => ({ meta: [{ title: pageTitle(m.layout_dashboard_workbench()) }] }),
  component: AgentsChatPage,
});
