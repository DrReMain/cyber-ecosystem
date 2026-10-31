import { useMutation, useQuery } from "@connectrpc/connect-query";
import { listMembers } from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/member-MemberService_connectquery";
import { updateUserStatus } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/user-UserService_connectquery";
import { Filter } from "@cyber-ecosystem/shared-antd/filter";
import { useFilter, useServerPagination } from "@cyber-ecosystem/shared-antd/use-search";
import { useSearch } from "@tanstack/react-router";
import { Card } from "antd";
import { useAreaSearchStore } from "#/libs";
import { m } from "#/paraglide/messages";
import { parseUsersSearch, usersSearchSchema } from "./search";
import { UsersTable } from "./ui/users-table";

const AREA = "/dashboard/people/users/";

export function EsportsUsersPage() {
  const search = useSearch({ from: AREA });
  const store = useAreaSearchStore(AREA, search, parseUsersSearch);
  const { values, onFilter, onReset } = useFilter(store, usersSearchSchema);
  const listQuery = useQuery(listMembers, {
    keyword: search.keyword?.trim() || undefined,
    enabled: search.status === "on" ? true : search.status === "off" ? false : undefined,
    page: { pageNo: search.pageNo, pageSize: search.pageSize },
  });
  const pagination = useServerPagination(store, usersSearchSchema, listQuery.data?.page, {
    pageSizeOptions: [20, 50, 100],
    showTotal: (t) => m.esports_users_total({ n: t }),
  });
  const toggleMut = useMutation(updateUserStatus, { onSuccess: () => listQuery.refetch() });

  return (
    <div className="flex flex-col gap-4 p-4">
      <Card>
        <Filter
          columns={4}
          initialValues={values}
          labels={{
            search: m.common_filter_search(),
            reset: m.common_filter_reset(),
            fold: m.common_filter_fold(),
            expand: m.common_filter_expand(),
          }}
          onFilter={onFilter}
          onReset={onReset}
          options={[
            {
              label: m.esports_users_filter_keyword(),
              name: "keyword",
              placeholder: m.esports_users_filter_keyword_placeholder(),
            },
            {
              label: m.esports_users_filter_status(),
              name: "status",
              type: "select",
              placeholder: m.esports_users_filter_any(),
              options: [
                { label: m.esports_users_status_on(), value: "on" },
                { label: m.esports_users_status_off(), value: "off" },
              ],
            },
          ]}
        />
      </Card>
      <UsersTable
        loading={listQuery.isFetching}
        onRefresh={() => listQuery.refetch()}
        onToggle={(id, next) => toggleMut.mutate({ id, enabled: next })}
        pagination={pagination}
        pendingId={toggleMut.isPending ? (toggleMut.variables?.id ?? null) : null}
        rows={listQuery.data?.members ?? []}
      />
    </div>
  );
}
