import { timestampFromMs } from "@bufbuild/protobuf/wkt";
import { useQuery } from "@connectrpc/connect-query";
import type { AuditLog } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/audit_pb";
import { listAuditLogs } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/audit-AuditService_connectquery";
import { UserService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/user_pb";
import { Filter } from "@cyber-ecosystem/shared-antd/filter";
import { useFilter, useServerPagination } from "@cyber-ecosystem/shared-antd/use-search";
import { useRouteContext, useSearch } from "@tanstack/react-router";
import { Card, Input, Select } from "antd";
import { useMemo, useState } from "react";
import { useUserDirectory } from "#/features/app/use-user-directory";
import { isOperationAllowed, op } from "#/features/layout-dashboard/auth/permissions";
import { useAreaSearchStore } from "#/libs";
import { m } from "#/paraglide/messages";
import { auditSearchSchema, parseAuditSearch } from "./search";
import { AuditDrawer } from "./ui/audit-drawer";
import { AuditTable } from "./ui/audit-table";
import { denyLabel } from "./ui/event-tags";

const denyOptions = () => [
  { label: denyLabel("NO_GRANT"), value: "NO_GRANT" },
  { label: denyLabel("ABAC_CONSTRAINT"), value: "ABAC_CONSTRAINT" },
];

const clean = (v: string | undefined): string | undefined =>
  v !== undefined && v !== "" ? v : undefined;

const AREA = "/dashboard/system/audit";

export function AuditPage() {
  const search = useSearch({ from: AREA });
  const { permissions } = useRouteContext({ from: AREA });
  const store = useAreaSearchStore(AREA, search, parseAuditSearch);
  const { values, onFilter, onReset } = useFilter(store, auditSearchSchema);
  const [detail, setDetail] = useState<{ event: AuditLog; open: boolean } | null>(null);

  const canListUsers = isOperationAllowed(permissions, op(UserService, "listUsers"));
  const { byId: usersById, byEmail, known: actorKnown } = useUserDirectory(canListUsers);
  const actorName = (id: string) => usersById.get(id);

  const input = useMemo(() => {
    const actor = clean(values.actor);
    return {
      page: {
        pageNo: search.pageNo,
        pageSize: search.pageSize,
        createdAtA:
          values.createdAtA !== undefined ? timestampFromMs(values.createdAtA) : undefined,
        createdAtZ:
          values.createdAtZ !== undefined ? timestampFromMs(values.createdAtZ) : undefined,
      },
      orderBy: [search.sort ?? "createdAt:desc"],
      actor: actor === undefined ? undefined : (byEmail.get(actor) ?? actor),
      operation: clean(values.operation?.trim()),
      status: values.status,
      denyReason: clean(values.denyReason),
      denied: values.lens === "denied" ? true : undefined,
    };
  }, [values, search.pageNo, search.pageSize, search.sort, byEmail]);

  const auditQuery = useQuery(listAuditLogs, input);
  const pagination = useServerPagination(store, auditSearchSchema, auditQuery.data?.page, {
    pageSizeOptions: [20, 50, 100],
    showTotal: (t) => m.system_audit_total({ n: t }),
  });

  const filtersActive =
    clean(values.actor) !== undefined ||
    clean(values.operation) !== undefined ||
    values.status !== undefined ||
    clean(values.denyReason) !== undefined ||
    values.createdAtA !== undefined ||
    values.createdAtZ !== undefined ||
    values.lens === "denied";

  return (
    <div className="flex flex-col gap-4 p-4">
      <Card>
        <Filter
          columns={3}
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
              label: m.system_audit_filter_lens(),
              name: "lens",
              type: "select",
              options: [
                { label: m.system_audit_lens_all(), value: "all" },
                { label: m.system_audit_lens_denied(), value: "denied" },
              ],
            },
            {
              label: m.system_audit_filter_actor(),
              name: "actor",
              element: canListUsers ? (
                <Select
                  allowClear
                  options={Array.from(usersById, ([id, email]) => ({ label: email, value: id }))}
                  placeholder={m.system_audit_filter_actor_ph()}
                  showSearch
                  style={{ width: "100%" }}
                />
              ) : (
                <Input allowClear placeholder={m.system_audit_filter_actor_id_ph()} />
              ),
            },
            {
              label: m.system_audit_filter_operation(),
              name: "operation",
              placeholder: m.system_audit_filter_operation_ph(),
            },
            {
              label: m.system_audit_filter_status(),
              name: "status",
              type: "number",
              min: 100,
              max: 599,
              placeholder: "403",
            },
            {
              label: m.system_audit_filter_deny(),
              name: "denyReason",
              type: "select",
              placeholder: m.system_audit_filter_deny_any(),
              options: denyOptions(),
            },
            {
              label: m.system_audit_filter_range(),
              name: ["createdAtA", "createdAtZ"],
              type: "range-datetime",
              placeholder: [m.system_audit_filter_from(), m.system_audit_filter_to()],
            },
          ]}
        />
      </Card>
      <AuditTable
        actorKnown={actorKnown}
        actorName={actorName}
        emptyDescription={filtersActive ? m.system_audit_empty_filtered() : m.system_audit_empty()}
        events={auditQuery.data?.list ?? []}
        loading={auditQuery.isFetching}
        onDetail={(event) => setDetail({ event, open: true })}
        onRefresh={() => auditQuery.refetch()}
        pagination={pagination}
      />
      <AuditDrawer
        actorKnown={actorKnown}
        actorName={actorName}
        afterOpenChange={(open) => {
          if (!open) setDetail(null);
        }}
        event={detail?.event ?? null}
        onClose={() => setDetail((prev) => (prev ? { ...prev, open: false } : prev))}
        open={detail?.open ?? false}
      />
    </div>
  );
}
