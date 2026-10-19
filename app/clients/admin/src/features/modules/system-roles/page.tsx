import { useMutation, useQuery } from "@connectrpc/connect-query";
import { listPolicies } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/policy-PolicyService_connectquery";
import { listResource } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/resource-ResourceService_connectquery";
import type { Role } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/role_pb";
import {
  createRole,
  deleteRole,
  getRole,
  listRoleMembers,
  listRoles,
  removeRoleMember,
  updateRole,
  updateRoleStatus,
} from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/role-RoleService_connectquery";
import { listUsers } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/user-UserService_connectquery";
import { Filter } from "@cyber-ecosystem/shared-antd/filter";
import { useFilter } from "@cyber-ecosystem/shared-antd/use-search";
import { useSearch } from "@tanstack/react-router";
import { Card } from "antd";
import { useMemo, useState } from "react";
import { useAreaSearchStore } from "#/libs";
import { m } from "#/paraglide/messages";
import { grantableServices, type RoleDraft } from "./grants";
import { parseRolesSearch, rolesSearchSchema } from "./search";
import { RoleDrawer } from "./ui/role-drawer";
import type { MemberView } from "./ui/role-members";
import { RolesTable } from "./ui/roles-table";

const ROLES_INPUT = { orderBy: ["createdAt:asc"], page: { all: true } };

type EditingBody = { mode: "create" } | { mode: "edit"; role: Role };
type Editing = { seq: number; open: boolean } & EditingBody;

function matchKeyword(role: Role, keyword: string): boolean {
  if (!keyword) return true;
  const inName = role.name?.toLowerCase().includes(keyword) ?? false;
  const inCode = role.code?.toLowerCase().includes(keyword) ?? false;
  return inName || inCode;
}

function matchStatus(role: Role, status: string | undefined): boolean {
  if (status === "on") return role.enabled ?? false;
  if (status === "off") return !(role.enabled ?? false);
  return true;
}

function applySearch(roles: Role[], values: { keyword?: string; status?: string }): Role[] {
  const keyword = values.keyword?.trim().toLowerCase();
  return roles.filter(
    (role) => matchKeyword(role, keyword ?? "") && matchStatus(role, values.status),
  );
}

function useRolesActions({
  editId,
  refetchList,
  refetchMembers,
  refetchRole,
  closeDrawer,
}: Readonly<{
  editId: string | undefined;
  refetchList: () => Promise<unknown>;
  refetchMembers: () => Promise<unknown>;
  refetchRole: () => Promise<unknown>;
  closeDrawer: () => void;
}>) {
  const toggleMut = useMutation(updateRoleStatus, { onSuccess: refetchList });
  const deleteMut = useMutation(deleteRole, { onSuccess: refetchList });
  const unbindMut = useMutation(removeRoleMember, {
    onSuccess: () => {
      refetchList();
      refetchMembers();
    },
  });
  const createMut = useMutation(createRole, {
    onSuccess: () => {
      refetchList();
      closeDrawer();
    },
  });
  const updateMut = useMutation(updateRole, {
    onSuccess: async () => {
      await Promise.allSettled([refetchList(), refetchRole()]);
      closeDrawer();
    },
  });

  const pendingId = toggleMut.isPending
    ? (toggleMut.variables?.id ?? null)
    : deleteMut.isPending
      ? (deleteMut.variables?.id ?? null)
      : null;
  const saving = createMut.isPending || updateMut.isPending;
  const pendingMemberId = unbindMut.isPending ? (unbindMut.variables?.principalId ?? null) : null;

  return {
    pendingId,
    saving,
    pendingMemberId,
    handleToggle: (role: Role, next: boolean) =>
      toggleMut.mutate({ id: role.id ?? "", enabled: next }),
    handleDelete: (role: Role) => deleteMut.mutate({ id: role.id ?? "" }),
    handleUnbind: (target: MemberView) => {
      if (editId === undefined) return;
      unbindMut.mutate({
        roleId: editId,
        principalType: target.principalType,
        principalId: target.principalId,
      });
    },
    handleSave: (draft: RoleDraft) => {
      if (editId === undefined) {
        createMut.mutate({
          code: draft.code,
          name: draft.name,
          remark: draft.remark,
          grants: draft.grants,
        });
        return;
      }
      updateMut.mutate({
        id: editId,
        name: draft.name,
        remark: draft.remark,
        grants: draft.grants,
        fieldsMask: ["name", "remark", "grants"],
      });
    },
  };
}

const AREA = "/dashboard/system/roles";

export function RolesPage() {
  const search = useSearch({ from: AREA });
  const store = useAreaSearchStore(AREA, search, parseRolesSearch);
  const { values, onFilter, onReset } = useFilter(store, rolesSearchSchema);
  const [editing, setEditing] = useState<Editing | null>(null);
  const openDrawer = (next: EditingBody) =>
    setEditing((prev) => ({ seq: (prev?.seq ?? 0) + 1, open: true, ...next }));
  const beginClose = () => setEditing((prev) => (prev ? { ...prev, open: false } : prev));
  const editId = editing?.mode === "edit" ? editing.role.id : undefined;

  const rolesQuery = useQuery(listRoles, ROLES_INPUT);
  const catalogQuery = useQuery(listResource, {});
  const policiesQuery = useQuery(listPolicies, { orderBy: ["name:asc"], page: { all: true } });
  const roleQuery = useQuery(getRole, { id: editId ?? "" }, { enabled: editId !== undefined });
  const membersQuery = useQuery(
    listRoleMembers,
    { roleId: editId ?? "", page: { all: true } },
    { enabled: editId !== undefined },
  );
  const usersQuery = useQuery(listUsers, { page: { all: true } });

  const roles = useMemo(
    () => applySearch(rolesQuery.data?.list ?? [], values),
    [rolesQuery.data, values],
  );
  const catalog = useMemo(
    () => grantableServices(catalogQuery.data?.list ?? []),
    [catalogQuery.data],
  );
  const policyOptions = useMemo(
    () =>
      (policiesQuery.data?.list ?? []).map((p) => ({
        id: p.id ?? "",
        name: p.name ?? "",
        kind: p.kind ?? "",
        enabled: p.enabled ?? false,
      })),
    [policiesQuery.data],
  );
  const members = useMemo(() => {
    const emailById = new Map(
      (usersQuery.data?.list ?? []).map((u) => [u.id ?? "", u.email ?? ""] as const),
    );
    return (membersQuery.data?.list ?? []).map((mem) => ({
      principalType: mem.principalType ?? "user",
      principalId: mem.principalId ?? "",
      createdAt: mem.createdAt,
      displayName: emailById.get(mem.principalId ?? "") || (mem.principalId ?? ""),
    }));
  }, [membersQuery.data, usersQuery.data]);

  const actions = useRolesActions({
    editId,
    refetchList: () => rolesQuery.refetch(),
    refetchMembers: () => membersQuery.refetch(),
    refetchRole: () => roleQuery.refetch(),
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
              label: m.system_roles_filter_keyword(),
              name: "keyword",
              placeholder: m.system_roles_filter_keyword_placeholder(),
            },
            {
              label: m.system_roles_filter_status(),
              name: "status",
              type: "select",
              placeholder: m.system_roles_filter_any(),
              options: [
                { label: m.system_roles_status_on(), value: "on" },
                { label: m.system_roles_status_off(), value: "off" },
              ],
            },
          ]}
        />
      </Card>
      <RolesTable
        loading={rolesQuery.isFetching}
        onCreate={() => openDrawer({ mode: "create" })}
        onDelete={actions.handleDelete}
        onEdit={(role) => openDrawer({ mode: "edit", role })}
        onRefresh={() => rolesQuery.refetch()}
        onToggle={actions.handleToggle}
        pendingId={actions.pendingId}
        roles={roles}
      />
      {editing !== null && (
        <RoleDrawer
          afterOpenChange={(open) => {
            if (!open) setEditing(null);
          }}
          catalog={catalog}
          initialGrants={editing.mode === "edit" ? (roleQuery.data?.grants ?? []) : []}
          key={editing.seq}
          loading={editing.mode === "edit" && roleQuery.data === undefined}
          members={editing.mode === "edit" ? members : []}
          mode={editing.mode}
          onCancel={beginClose}
          onSave={actions.handleSave}
          onUnbind={actions.handleUnbind}
          open={editing.open}
          pending={actions.saving}
          pendingMemberId={actions.pendingMemberId}
          policies={policyOptions}
          role={editing.mode === "edit" ? editing.role : null}
        />
      )}
    </div>
  );
}
