import { useMutation, useQuery } from "@connectrpc/connect-query";
import {
  MembershipSource,
  MembershipStatus,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin_pb";
import {
  grantMembership,
  listMemberships,
  terminateMembership,
} from "@cyber-ecosystem/gen-connect-ts/cyber/esports/v1/activation_admin-EntitleAdminService_connectquery";
import { Filter } from "@cyber-ecosystem/shared-antd/filter";
import { useFilter, useServerPagination } from "@cyber-ecosystem/shared-antd/use-search";
import { useSearch } from "@tanstack/react-router";
import { Card, Typography } from "antd";
import { useState } from "react";
import { useAreaSearchStore } from "#/libs";
import { m } from "#/paraglide/messages";
import { membersSearchSchema, parseMembersSearch } from "./search";
import { GrantModal, MEMBER_MONTH_SKU } from "./ui/grant-modal";
import { MembersTable } from "./ui/members-table";
import { TerminateModal } from "./ui/terminate-modal";

const AREA = "/dashboard/people/members/";

const TAB_TO_STATUS: Record<string, MembershipStatus | undefined> = {
  all: undefined,
  active: MembershipStatus.ACTIVE,
  expiring: MembershipStatus.EXPIRING_SOON,
  expired: MembershipStatus.EXPIRED,
};

const SOURCE_TO_WIRE: Record<string, MembershipSource | undefined> = {
  order: MembershipSource.ORDER,
  manual: MembershipSource.MANUAL,
  legacy_code: MembershipSource.LEGACY_CODE,
};

export function EsportsMembersPage() {
  const search = useSearch({ from: AREA });
  const store = useAreaSearchStore(AREA, search, parseMembersSearch);
  const { values, onFilter, onReset } = useFilter(store, membersSearchSchema);

  const listQuery = useQuery(listMemberships, {
    status: TAB_TO_STATUS[values.tab ?? "all"],
    source: values.source ? SOURCE_TO_WIRE[values.source] : undefined,
    page: { pageNo: search.pageNo, pageSize: search.pageSize },
  });
  const pagination = useServerPagination(store, membersSearchSchema, listQuery.data?.page, {
    pageSizeOptions: [20, 50, 100],
    showTotal: (t) => m.esports_users_total({ n: t }),
  });

  const [grantRow, setGrantRow] = useState<{
    open: boolean;
    row: Parameters<typeof GrantModal>[0]["row"];
  }>({ open: false, row: null });
  const [terminateRow, setTerminateRow] =
    useState<Parameters<typeof TerminateModal>[0]["row"]>(null);

  const refetch = () => listQuery.refetch();
  const grantMut = useMutation(grantMembership, {
    onSuccess: () => {
      setGrantRow({ open: false, row: null });
      void refetch();
    },
  });
  const terminateMut = useMutation(terminateMembership, {
    onSuccess: () => {
      setTerminateRow(null);
      void refetch();
    },
  });

  // The tabs' live counters are gone with the tabs; this line keeps every
  // count the backend still reports, below the table instead of on it.
  const counts = listQuery.data;

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
              label: m.esports_members_filter_status(),
              name: "tab",
              type: "select",
              placeholder: m.esports_members_tab_all(),
              options: [
                { label: m.esports_members_tab_active(), value: "active" },
                { label: m.esports_members_tab_expiring(), value: "expiring" },
                { label: m.esports_members_tab_expired(), value: "expired" },
              ],
            },
            {
              label: m.esports_members_source_filter(),
              name: "source",
              type: "select",
              placeholder: m.esports_members_tab_all(),
              options: [
                { label: m.esports_members_source_order(), value: "order" },
                { label: m.esports_members_source_manual(), value: "manual" },
                { label: m.esports_members_source_legacy(), value: "legacy_code" },
              ],
            },
          ]}
        />
      </Card>
      <MembersTable
        loading={listQuery.isFetching}
        onCreate={() => setGrantRow({ open: true, row: null })}
        onRefresh={refetch}
        onRenew={(row) => setGrantRow({ open: true, row })}
        onTerminate={(row) => setTerminateRow(row)}
        pagination={pagination}
        rows={listQuery.data?.memberships ?? []}
      />
      {counts ? (
        <Typography.Text type="secondary">
          {m.esports_members_count_line({
            shown: listQuery.data?.memberships?.length ?? 0,
            total: counts.totalCount ?? 0,
            active: counts.activeCount ?? 0,
            expiring: counts.expiringSoonCount ?? 0,
            expired: counts.expiredCount ?? 0,
          })}
        </Typography.Text>
      ) : null}
      <GrantModal
        onCancel={() => setGrantRow({ open: false, row: null })}
        onSubmit={(userId) => grantMut.mutate({ userId, skuCode: MEMBER_MONTH_SKU })}
        open={grantRow.open}
        pending={grantMut.isPending}
        row={grantRow.row}
      />
      <TerminateModal
        onCancel={() => setTerminateRow(null)}
        onSubmit={(userId) => terminateMut.mutate({ userId })}
        pending={terminateMut.isPending}
        row={terminateRow}
      />
    </div>
  );
}
