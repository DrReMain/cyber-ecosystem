import { useQuery } from "@connectrpc/connect-query";
import { listAgentConfigs } from "@cyber-ecosystem/gen-connect-ts/cyber/agent/v1/agent_config_admin-AgentConfigAdminService_connectquery";
import { useServerPagination } from "@cyber-ecosystem/shared-antd/use-search";
import { useSearch } from "@tanstack/react-router";
import { useAreaSearchStore } from "#/libs";
import { m } from "#/paraglide/messages";
import { connectionsSearchSchema, parseConnectionsSearch } from "./search";
import { ConnectionsTable } from "./ui/connections-table";

const AREA = "/dashboard/agents/connections";

export function ConnectionsPage() {
  const search = useSearch({ from: AREA });
  const store = useAreaSearchStore(AREA, search, parseConnectionsSearch);
  const listQuery = useQuery(listAgentConfigs, {
    page: { pageNo: search.pageNo, pageSize: search.pageSize },
  });
  const pagination = useServerPagination(store, connectionsSearchSchema, listQuery.data?.page, {
    pageSizeOptions: [20, 50, 100],
    showTotal: (t) => m.agents_connections_total({ n: t }),
  });

  return (
    <div className="p-4">
      <ConnectionsTable
        emptyDescription={m.agents_connections_empty()}
        loading={listQuery.isFetching}
        onRefresh={() => listQuery.refetch()}
        pagination={pagination}
        rows={listQuery.data?.list ?? []}
      />
    </div>
  );
}
