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
import { useServerPagination } from "@cyber-ecosystem/shared-antd/use-search";
import { useNavigate, useSearch } from "@tanstack/react-router";
import { Button, Card, Select, Tabs, Typography } from "antd";
import { Plus } from "lucide-react";
import { useState } from "react";
import { useAreaSearchStore } from "#/libs";
import { m } from "#/paraglide/messages";
import { membersSearchSchema, parseMembersSearch } from "./search";
import { MembersTable } from "./ui/members-table";
import { GrantModal, MEMBER_MONTH_SKU } from "./ui/grant-modal";
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
  const navigate = useNavigate();

  const listQuery = useQuery(listMemberships, {
    status: TAB_TO_STATUS[search.tab ?? "all"],
    source: search.source ? SOURCE_TO_WIRE[search.source] : undefined,
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

  const patchSearch = (patch: Record<string, unknown>) => {
    void navigate({
      to: "/dashboard/people/members",
      search: { ...search, ...patch },
    });
  };

  const counts = listQuery.data;
  const tabOf = (key: string) => ({
    key,
    label:
      key === "all"
        ? m.esports_members_tab_all()
        : key === "active"
          ? m.esports_members_tab_active()
          : key === "expiring"
            ? m.esports_members_tab_expiring()
            : m.esports_members_tab_expired(),
  });
  const badgeOf = (key: string): number | undefined =>
    key === "all"
      ? (counts?.totalCount ?? undefined)
      : key === "active"
        ? (counts?.activeCount ?? undefined)
        : key === "expiring"
          ? (counts?.expiringSoonCount ?? undefined)
          : (counts?.expiredCount ?? undefined);

  return (
    <div className="flex flex-col gap-4 p-4">
      <Card>
        <div className="flex flex-wrap items-center justify-between gap-3">
          <Tabs
            activeKey={search.tab ?? "all"}
            items={["all", "active", "expiring", "expired"].map((key) => {
              const n = badgeOf(key);
              const tab = tabOf(key);
              return { ...tab, label: n != null ? `${tab.label} (${n})` : tab.label };
            })}
            onChange={(key) =>
              patchSearch({ tab: key === "all" ? undefined : key, pageNo: undefined })
            }
          />
          <div className="flex items-center gap-2">
            <Select
              allowClear
              className="w-40"
              onChange={(v) => patchSearch({ source: v ?? undefined, pageNo: undefined })}
              options={[
                { label: m.esports_members_source_order(), value: "order" },
                { label: m.esports_members_source_manual(), value: "manual" },
                { label: m.esports_members_source_legacy(), value: "legacy_code" },
              ]}
              placeholder={m.esports_members_source_filter()}
              value={search.source ?? null}
            />
            <Button
              icon={<Plus size={14} />}
              onClick={() => setGrantRow({ open: true, row: null })}
              type="primary"
            >
              {m.esports_members_act_grant()}
            </Button>
          </div>
        </div>
      </Card>
      <MembersTable
        loading={listQuery.isFetching}
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
