import type { User } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/user_pb";

export interface UserDraft {
  email: string;
  password: string;
  deptId?: string;
  roles: string[];
  avatar?: string;
  avatarChanged?: boolean;
}

export type OperationCatalog = Map<string, { comment: string; methods: Map<string, string> }>;

export function buildOperationCatalog(
  list: { fullName?: string; comment?: string; methods?: { name?: string; comment?: string }[] }[],
): OperationCatalog {
  const catalog: OperationCatalog = new Map();
  for (const svc of list) {
    if (!svc.fullName) continue;
    const methods = new Map<string, string>();
    for (const mt of svc.methods ?? []) {
      if (mt.name) methods.set(mt.name, mt.comment ?? "");
    }
    catalog.set(svc.fullName, { comment: svc.comment ?? "", methods });
  }
  return catalog;
}

export interface DeptNode {
  id: string;
  name: string;
  children: DeptNode[];
}

// biome-ignore lint/style/useConsistentTypeDefinitions: feeds DataTable<T extends Record<string, unknown>>
export type UserRow = {
  user: User;
  deptName: string | null;
  roleViews: { code: string; name: string; enabled: boolean }[];
};
