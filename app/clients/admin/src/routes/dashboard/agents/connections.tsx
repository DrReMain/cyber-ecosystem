import { AgentConfigAdminService } from "@cyber-ecosystem/gen-connect-ts/cyber/agent/v1/agent_config_admin_pb";
import { createFileRoute } from "@tanstack/react-router";
import { Cable } from "lucide-react";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { ConnectionsPage, parseConnectionsSearch } from "#/features/modules/agents-connections";
import { m } from "#/paraglide/messages";

const operations = [op(AgentConfigAdminService, "listAgentConfigs")];

export const Route = createFileRoute("/dashboard/agents/connections")({
  staticData: {
    title: "agents_connections_title",
    menu: { icon: Cable, order: 10 },
    operations,
  },
  validateSearch: parseConnectionsSearch,
  head: () => ({ meta: [{ title: pageTitle(m.agents_connections_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: ConnectionsPage,
});
