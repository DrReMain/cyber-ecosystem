import { useQuery } from "@connectrpc/connect-query";
import { getMyAgentConfig } from "@cyber-ecosystem/gen-connect-ts/cyber/agent/v1/agent_config-AgentConfigService_connectquery";
import { useNavigate } from "@tanstack/react-router";
import { useCallback } from "react";
import { ChatWorkspace } from "./chat-workspace";
import { EmptyState } from "./ui/empty-state";

export function AgentsChatPage() {
  const navigate = useNavigate();
  const cfgQuery = useQuery(getMyAgentConfig, {});
  const baseUrl = cfgQuery.data?.agentConfig?.baseUrl ?? "";

  const manage = useCallback(() => void navigate({ to: "/dashboard/profile/agents" }), [navigate]);

  if (cfgQuery.isPending) return null;
  if (baseUrl === "") {
    return <EmptyState onConfigure={manage} />;
  }
  return <ChatWorkspace baseUrl={baseUrl} />;
}
