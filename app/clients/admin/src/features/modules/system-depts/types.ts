import type { Dept } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/dept_pb";

export interface DeptNode {
  id: string;
  name: string;
  children: DeptNode[];
}

export interface DeptDraft {
  name: string;
  parentId?: string;
  remark?: string;
}

// biome-ignore lint/style/useConsistentTypeDefinitions: feeds DataTable<T extends Record<string, unknown>>
export type DeptRow = {
  dept: Dept;
  children?: DeptRow[];
};

export function buildDeptTree(
  depts: { id?: string; name?: string; parentId?: string }[],
): DeptNode[] {
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

// Leaf rows keep children undefined so antd renders no expander for them.
export function buildDeptRows(list: Dept[]): DeptRow[] {
  const nodes = new Map<string, DeptRow>();
  for (const d of list) {
    if (d.id) nodes.set(d.id, { dept: d });
  }
  const roots: DeptRow[] = [];
  for (const d of list) {
    const node = d.id ? nodes.get(d.id) : undefined;
    const parent = d.parentId ? nodes.get(d.parentId) : undefined;
    if (node === undefined) continue;
    if (parent !== undefined && parent !== node) {
      parent.children = [...(parent.children ?? []), node];
    } else {
      roots.push(node);
    }
  }
  return roots;
}

// Parent-picker tree for edit mode: keeps a node only when the excluded id is
// absent from its ancestor chain (walk-up per node, cycle-guarded). This is UX
// prevention only - the backend self-parent/cycle guards remain the truth.
export function buildParentTree(list: Dept[], excludeId?: string): DeptNode[] {
  if (!excludeId) return buildDeptTree(list);
  const parentOf = new Map<string, string | undefined>(list.map((d) => [d.id ?? "", d.parentId]));
  const allowed = (d: Dept): boolean => {
    let cur = d.id;
    const seen = new Set<string>();
    while (cur) {
      if (cur === excludeId) return false;
      if (seen.has(cur)) return true;
      seen.add(cur);
      cur = parentOf.get(cur);
    }
    return true;
  };
  return buildDeptTree(list.filter(allowed));
}

export function collectRowKeys(rows: DeptRow[]): string[] {
  const keys: string[] = [];
  for (const row of rows) {
    if (row.dept.id) keys.push(row.dept.id);
    if (row.children) keys.push(...collectRowKeys(row.children));
  }
  return keys;
}
