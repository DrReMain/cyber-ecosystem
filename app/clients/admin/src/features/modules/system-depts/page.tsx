import { useMutation, useQuery } from "@connectrpc/connect-query";
import type { Dept } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/dept_pb";
import {
  createDept,
  deleteDept,
  listDeptMembers,
  listDepts,
  removeDeptMember,
  updateDept,
} from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/dept-DeptService_connectquery";
import { Filter } from "@cyber-ecosystem/shared-antd/filter";
import { useFilter } from "@cyber-ecosystem/shared-antd/use-search";
import { useSearch } from "@tanstack/react-router";
import { Card } from "antd";
import { useMemo, useState } from "react";
import { useAreaSearchStore } from "#/libs";
import { m } from "#/paraglide/messages";
import { deptsSearchSchema, parseDeptsSearch } from "./search";
import type { DeptDraft } from "./types";
import { buildDeptRows, buildParentTree } from "./types";
import { DeptDrawer } from "./ui/dept-drawer";
import type { DeptMemberView } from "./ui/dept-members";
import { DeptsTable } from "./ui/depts-table";

type EditingBody = { mode: "create"; parent: Dept | null } | { mode: "edit"; dept: Dept };
type Editing = { seq: number; open: boolean } & EditingBody;

function deptsInput(values: { keyword?: string }) {
  return {
    orderBy: ["createdAt:asc"],
    page: { all: true },
    name: values.keyword?.trim() || undefined,
  };
}

const AREA = "/dashboard/system/depts";

export function DeptsPage() {
  const search = useSearch({ from: AREA });
  const store = useAreaSearchStore(AREA, search, parseDeptsSearch);
  const { values, onFilter, onReset } = useFilter(store, deptsSearchSchema);
  const [editing, setEditing] = useState<Editing | null>(null);
  const openDrawer = (next: EditingBody) => {
    setEditing((prev) => ({ seq: (prev?.seq ?? 0) + 1, open: true, ...next }));
  };
  const beginClose = () => setEditing((prev) => (prev ? { ...prev, open: false } : prev));
  const editId = editing?.mode === "edit" ? editing.dept.id : undefined;

  const deptsQuery = useQuery(
    listDepts,
    useMemo(() => deptsInput(values), [values]),
  );
  const membersQuery = useQuery(
    listDeptMembers,
    { deptId: editId ?? "", page: { all: true } },
    { enabled: editId !== undefined },
  );
  const list = useMemo(() => deptsQuery.data?.list ?? [], [deptsQuery.data]);
  const rows = useMemo(() => buildDeptRows(list), [list]);
  const parentTree = useMemo(
    () => buildParentTree(list, editing?.mode === "edit" ? editing.dept.id : undefined),
    [list, editing],
  );
  const members = useMemo<DeptMemberView[]>(
    () =>
      (membersQuery.data?.list ?? []).map((mem) => ({
        userId: mem.userId ?? "",
        email: mem.email ?? "",
        enabled: mem.enabled ?? false,
      })),
    [membersQuery.data],
  );

  const refetchList = () => deptsQuery.refetch();
  const closeDrawer = () => {
    void refetchList();
    beginClose();
  };
  const createMut = useMutation(createDept, { onSuccess: closeDrawer });
  const updateMut = useMutation(updateDept, { onSuccess: closeDrawer });
  const deleteMut = useMutation(deleteDept, { onSuccess: refetchList });
  const removeMemberMut = useMutation(removeDeptMember, {
    onSuccess: () => {
      void membersQuery.refetch();
      refetchList();
    },
  });
  const saving = createMut.isPending || updateMut.isPending;
  const deletingId = deleteMut.isPending ? (deleteMut.variables?.id ?? null) : null;
  const pendingMemberId = removeMemberMut.isPending
    ? (removeMemberMut.variables?.userId ?? null)
    : null;

  const handleRemoveMember = (member: DeptMemberView) => {
    if (editId === undefined) return;
    removeMemberMut.mutate({ deptId: editId, userId: member.userId });
  };

  const handleSave = (draft: DeptDraft) => {
    if (editing === null) return;
    if (editing.mode === "create") {
      createMut.mutate({ name: draft.name, parentId: draft.parentId, remark: draft.remark });
      return;
    }
    updateMut.mutate({
      id: editing.dept.id ?? "",
      name: draft.name,
      parentId: draft.parentId,
      remark: draft.remark,
      fieldsMask: ["name", "parent_id", "remark"],
    });
  };

  return (
    <div className="flex flex-col gap-4 p-4">
      <Card>
        <Filter
          columns={2}
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
              label: m.system_depts_filter_name(),
              name: "keyword",
              placeholder: m.system_depts_filter_name_placeholder(),
            },
          ]}
        />
      </Card>
      <DeptsTable
        loading={deptsQuery.isFetching}
        onAddChild={(dept) => openDrawer({ mode: "create", parent: dept })}
        onCreate={() => openDrawer({ mode: "create", parent: null })}
        onDelete={(dept) => deleteMut.mutate({ id: dept.id ?? "" })}
        onEdit={(dept) => openDrawer({ mode: "edit", dept })}
        onRefresh={refetchList}
        pendingId={deletingId}
        rows={rows}
      />
      {editing !== null && (
        <DeptDrawer
          afterOpenChange={(open) => {
            if (!open) setEditing(null);
          }}
          key={editing.seq}
          members={editing.mode === "edit" ? members : []}
          mode={
            editing.mode === "create" && editing.parent !== null ? "create-child" : editing.mode
          }
          onCancel={beginClose}
          onRemoveMember={handleRemoveMember}
          onSave={handleSave}
          open={editing.open}
          parentTree={parentTree}
          pending={saving}
          pendingMemberId={pendingMemberId}
          source={editing.mode === "edit" ? editing.dept : editing.parent}
        />
      )}
    </div>
  );
}
