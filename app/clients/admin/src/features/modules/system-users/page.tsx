import { useMutation, useQuery } from "@connectrpc/connect-query";
import { listDepts } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/dept-DeptService_connectquery";
import { listResource } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/resource-ResourceService_connectquery";
import {
  listRoles,
  previewGrants,
} from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/role-RoleService_connectquery";
import type { User } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/user_pb";
import {
  createUser,
  deleteUser,
  listUsers,
  updateUser,
  updateUserStatus,
} from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/user-UserService_connectquery";
import { Filter } from "@cyber-ecosystem/shared-antd/filter";
import { useFilter } from "@cyber-ecosystem/shared-antd/use-search";
import { keepPreviousData } from "@tanstack/react-query";
import { useSearch } from "@tanstack/react-router";
import { Card, TreeSelect } from "antd";
import { useMemo, useState } from "react";
import { useFileUrls } from "#/features/app/file/use-file-url";
import { useProxyUpload } from "#/features/app/file/use-proxy-upload";
import { useAreaSearchStore } from "#/libs";
import { m } from "#/paraglide/messages";
import { parseUsersSearch, usersSearchSchema } from "./search";
import type { DeptNode, OperationCatalog, UserDraft, UserRow } from "./types";
import { buildOperationCatalog } from "./types";
import { UserDrawer } from "./ui/user-drawer";
import { UsersTable } from "./ui/users-table";

type EditingBody = { mode: "create" } | { mode: "edit"; user: User };
type Editing = { seq: number; open: boolean } & EditingBody;

function usersInput(values: { keyword?: string; status?: string; deptId?: string }) {
  return {
    orderBy: ["createdAt:asc"],
    page: { all: true },
    email: values.keyword?.trim() || undefined,
    enabled: values.status === "on" ? true : values.status === "off" ? false : undefined,
    deptId: values.deptId || undefined,
  };
}

function buildDeptTree(depts: { id?: string; name?: string; parentId?: string }[]): DeptNode[] {
  const nodes = new Map<string, DeptNode>();
  for (const d of depts) {
    if (d.id) nodes.set(d.id, { id: d.id, name: d.name ?? "", children: [] });
  }
  const roots: DeptNode[] = [];
  for (const d of depts) {
    const node = d.id ? nodes.get(d.id) : undefined;
    const parent = d.parentId ? nodes.get(d.parentId) : undefined;
    if (node === undefined) continue;
    if (parent !== undefined && parent !== node) parent.children.push(node);
    else roots.push(node);
  }
  return roots;
}

function useUsersActions({
  editId,
  refetchList,
  closeDrawer,
  onResetDone,
}: Readonly<{
  editId: string | undefined;
  refetchList: () => Promise<unknown>;
  closeDrawer: () => void;
  onResetDone: (password: string) => void;
}>) {
  const toggleMut = useMutation(updateUserStatus, { onSuccess: refetchList });
  const deleteMut = useMutation(deleteUser, { onSuccess: refetchList });
  const createMut = useMutation(createUser, {
    onSuccess: () => {
      refetchList();
      closeDrawer();
    },
  });
  const updateMut = useMutation(updateUser, {
    onSuccess: () => {
      refetchList();
      closeDrawer();
    },
  });
  const resetMut = useMutation(updateUser, {
    onSuccess: (_data, variables) => onResetDone(variables?.password ?? ""),
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
    resetPending: resetMut.isPending,
    handleToggle: (user: User, next: boolean) =>
      toggleMut.mutate({ id: user.id ?? "", enabled: next }),
    handleDelete: (user: User) => deleteMut.mutate({ id: user.id ?? "" }),
    handleResetPassword: (user: User, password: string) =>
      resetMut.mutate({ id: user.id ?? "", password, fieldsMask: ["password"] }),
    handleSave: (draft: UserDraft) => {
      if (editId === undefined) {
        createMut.mutate({
          email: draft.email,
          password: draft.password,
          deptId: draft.deptId,
          roles: draft.roles,
          avatar: draft.avatar,
        });
        return;
      }
      const fieldsMask = ["dept_id", "roles"];
      if (draft.avatarChanged) fieldsMask.push("avatar");
      updateMut.mutate({
        id: editId,
        deptId: draft.deptId,
        roles: draft.roles,
        avatar: draft.avatarChanged ? draft.avatar : undefined,
        fieldsMask,
      });
    },
  };
}

const AREA = "/dashboard/system/users";

export function UsersPage() {
  const search = useSearch({ from: AREA });
  const store = useAreaSearchStore(AREA, search, parseUsersSearch);
  const { values, onFilter, onReset } = useFilter(store, usersSearchSchema);
  const [editing, setEditing] = useState<Editing | null>(null);
  const avatarUpload = useProxyUpload();
  const [selected, setSelected] = useState<string[]>([]);
  const [resetResult, setResetResult] = useState<string | null>(null);
  const openDrawer = (next: EditingBody) => {
    setEditing((prev) => ({ seq: (prev?.seq ?? 0) + 1, open: true, ...next }));
    setSelected(next.mode === "edit" ? (next.user.roles ?? []) : []);
    setResetResult(null);
  };
  const beginClose = () => setEditing((prev) => (prev ? { ...prev, open: false } : prev));
  const editId = editing?.mode === "edit" ? editing.user.id : undefined;

  const usersQuery = useQuery(
    listUsers,
    useMemo(() => usersInput(values), [values]),
  );
  const rolesQuery = useQuery(listRoles, { orderBy: ["createdAt:asc"], page: { all: true } });
  const deptsQuery = useQuery(listDepts, { page: { all: true } });
  const catalogQuery = useQuery(listResource, {});
  const catalog = useMemo<OperationCatalog>(
    () => buildOperationCatalog(catalogQuery.data?.list ?? []),
    [catalogQuery.data],
  );
  const previewQuery = useQuery(
    previewGrants,
    { roleCodes: selected },
    { enabled: editing !== null, placeholderData: keepPreviousData },
  );

  const rows = useMemo<UserRow[]>(() => {
    const roles = rolesQuery.data?.list ?? [];
    const depts = deptsQuery.data?.list ?? [];
    const deptById = new Map(depts.map((d) => [d.id ?? "", d.name ?? ""]));
    return (usersQuery.data?.list ?? []).map((user) => ({
      user,
      deptName: user.deptId ? (deptById.get(user.deptId) ?? null) : null,
      roleViews: roles
        .filter((r) => user.roles?.includes(r.code ?? "") ?? false)
        .map((r) => ({ code: r.code ?? "", name: r.name ?? "", enabled: r.enabled ?? false })),
    }));
  }, [usersQuery.data, rolesQuery.data, deptsQuery.data]);
  const avatarUrls = useFileUrls(rows.map((r) => r.user.avatar ?? ""));
  const deptTree = useMemo(() => buildDeptTree(deptsQuery.data?.list ?? []), [deptsQuery.data]);

  const actions = useUsersActions({
    editId,
    refetchList: () => usersQuery.refetch(),
    closeDrawer: beginClose,
    onResetDone: setResetResult,
  });

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
              label: m.system_users_filter_keyword(),
              name: "keyword",
              placeholder: m.system_users_filter_keyword_placeholder(),
            },
            {
              label: m.system_users_filter_status(),
              name: "status",
              type: "select",
              placeholder: m.system_users_filter_any(),
              options: [
                { label: m.system_users_status_on(), value: "on" },
                { label: m.system_users_status_off(), value: "off" },
              ],
            },
            {
              label: m.system_users_filter_dept(),
              name: "deptId",
              element: (
                <TreeSelect
                  allowClear
                  fieldNames={{ label: "name", value: "id", children: "children" }}
                  placeholder={m.system_users_filter_dept_any()}
                  treeData={deptTree}
                  treeDefaultExpandAll
                />
              ),
            },
          ]}
        />
      </Card>
      <UsersTable
        avatarUrls={avatarUrls}
        loading={usersQuery.isFetching}
        onCreate={() => openDrawer({ mode: "create" })}
        onDelete={actions.handleDelete}
        onEdit={(user) => openDrawer({ mode: "edit", user })}
        onRefresh={() => usersQuery.refetch()}
        onToggle={actions.handleToggle}
        pendingId={actions.pendingId}
        rows={rows}
      />
      {editing !== null && (
        <UserDrawer
          afterOpenChange={(open) => {
            if (!open) setEditing(null);
          }}
          avatarUploading={avatarUpload.isPending}
          catalog={catalog}
          deptTree={deptTree}
          key={editing.seq}
          mode={editing.mode}
          onCancel={beginClose}
          onResetPassword={actions.handleResetPassword}
          onSave={actions.handleSave}
          onSelectionChange={setSelected}
          onUploadAvatar={avatarUpload.upload}
          open={editing.open}
          pending={actions.saving}
          previewLoading={previewQuery.isFetching}
          previewOperations={previewQuery.data?.operations ?? []}
          resetPending={actions.resetPending}
          resetResult={resetResult}
          roles={rolesQuery.data?.list ?? []}
          selected={selected}
          user={editing.mode === "edit" ? editing.user : null}
        />
      )}
    </div>
  );
}
