import { useMutation, useQuery } from "@connectrpc/connect-query";
import type { Policy } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/policy_pb";
import {
  createPolicy,
  deletePolicy,
  getPolicy,
  listPolicies,
  updatePolicy,
} from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/policy-PolicyService_connectquery";
import { Filter } from "@cyber-ecosystem/shared-antd/filter";
import { useFilter } from "@cyber-ecosystem/shared-antd/use-search";
import { useSearch } from "@tanstack/react-router";
import { Card } from "antd";
import { useMemo, useState } from "react";
import { useAreaSearchStore } from "#/libs";
import { m } from "#/paraglide/messages";
import { type PolicyFormDraft, paramsDraftOf, paramsInputOf } from "./params";
import { parsePoliciesSearch, policiesSearchSchema } from "./search";
import { PoliciesTable } from "./ui/policies-table";
import { PolicyDrawer } from "./ui/policy-drawer";

type EditingBody = { mode: "create" } | { mode: "edit"; policy: Policy };
type Editing = { seq: number; open: boolean } & EditingBody;

function policiesInput(values: { keyword?: string; kind?: string }) {
  return {
    orderBy: ["name:asc"],
    page: { all: true },
    name: values.keyword?.trim() || undefined,
    kind: values.kind || undefined,
  };
}

function usePoliciesActions({
  editId,
  refetchList,
  closeDrawer,
}: Readonly<{
  editId: string | undefined;
  refetchList: () => Promise<unknown>;
  closeDrawer: () => void;
}>) {
  const toggleMut = useMutation(updatePolicy, { onSuccess: refetchList });
  const deleteMut = useMutation(deletePolicy, { onSuccess: refetchList });
  const createMut = useMutation(createPolicy, {
    onSuccess: () => {
      refetchList();
      closeDrawer();
    },
  });
  const updateMut = useMutation(updatePolicy, {
    onSuccess: () => {
      refetchList();
      closeDrawer();
    },
  });

  const pendingId = toggleMut.isPending
    ? (toggleMut.variables?.id ?? null)
    : deleteMut.isPending
      ? (deleteMut.variables?.id ?? null)
      : null;
  const saving = createMut.isPending || updateMut.isPending;

  return {
    pendingId,
    saving,
    handleToggle: (policy: Policy, next: boolean) =>
      toggleMut.mutate({ id: policy.id ?? "", enabled: next, fieldsMask: ["enabled"] }),
    handleDelete: (policy: Policy) => deleteMut.mutate({ id: policy.id ?? "" }),
    handleSave: (draft: PolicyFormDraft) => {
      if (editId === undefined) {
        createMut.mutate({ name: draft.name, params: paramsInputOf(paramsDraftOf(draft)) });
        return;
      }
      updateMut.mutate({
        id: editId,
        name: draft.name,
        enabled: draft.enabled,
        params: paramsInputOf(paramsDraftOf(draft)),
        fieldsMask: ["name", "params", "enabled"],
      });
    },
  };
}

const AREA = "/dashboard/system/policies";

export function PoliciesPage() {
  const search = useSearch({ from: AREA });
  const store = useAreaSearchStore(AREA, search, parsePoliciesSearch);
  const { values, onFilter, onReset } = useFilter(store, policiesSearchSchema);
  const [editing, setEditing] = useState<Editing | null>(null);
  const openDrawer = (next: EditingBody) =>
    setEditing((prev) => ({ seq: (prev?.seq ?? 0) + 1, open: true, ...next }));
  const beginClose = () => setEditing((prev) => (prev ? { ...prev, open: false } : prev));
  const editId = editing?.mode === "edit" ? editing.policy.id : undefined;

  const policiesQuery = useQuery(
    listPolicies,
    useMemo(() => policiesInput(values), [values]),
  );
  const policyQuery = useQuery(getPolicy, { id: editId ?? "" }, { enabled: editId !== undefined });
  const policies = policiesQuery.data?.list ?? [];

  const actions = usePoliciesActions({
    editId,
    refetchList: () => policiesQuery.refetch(),
    closeDrawer: beginClose,
  });

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
              label: m.system_policies_filter_keyword(),
              name: "keyword",
              placeholder: m.system_policies_filter_keyword_placeholder(),
            },
            {
              label: m.system_policies_filter_kind(),
              name: "kind",
              type: "select",
              placeholder: m.system_policies_filter_kind_any(),
              options: [
                { label: m.system_policies_kind_time_window(), value: "time_window" },
                { label: m.system_policies_kind_calendar(), value: "calendar" },
              ],
            },
          ]}
        />
      </Card>
      <PoliciesTable
        loading={policiesQuery.isFetching}
        onCreate={() => openDrawer({ mode: "create" })}
        onDelete={actions.handleDelete}
        onEdit={(policy) => openDrawer({ mode: "edit", policy })}
        onRefresh={() => policiesQuery.refetch()}
        onToggle={actions.handleToggle}
        pendingId={actions.pendingId}
        policies={policies}
      />
      {editing !== null && (
        <PolicyDrawer
          afterOpenChange={(open) => {
            if (!open) setEditing(null);
          }}
          key={editing.seq}
          loading={editing.mode === "edit" && policyQuery.data === undefined}
          mode={editing.mode}
          onCancel={beginClose}
          onSave={actions.handleSave}
          open={editing.open}
          pending={actions.saving}
          policy={editing.mode === "edit" ? (policyQuery.data?.policy ?? null) : null}
        />
      )}
    </div>
  );
}
