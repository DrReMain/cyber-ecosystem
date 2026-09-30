import { create } from "@bufbuild/protobuf";
import type { Method, Service } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/resource_pb";
import {
  type RoleGrant,
  RoleGrantSchema,
  ScopeKind,
} from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/role_pb";
import { m } from "#/paraglide/messages";

export interface RoleDraft {
  name: string;
  code: string;
  remark: string;
  grants: RoleGrant[];
}

export interface GrantableView {
  name: string;
  fullName: string;
  comment: string;
  methods: Method[];
}

// One grant = operation + scope + attached constraint policies (AND, ≤16 ids
// on the wire). policies live in the map so every edit path carries them.
interface GrantEntry {
  scope: ScopeKind;
  policies: string[];
}

export type GrantMap = Record<string, GrantEntry>;

// Light view of a policy row for the attachment picker (wire fields only).
export interface PolicyOptionView {
  id: string;
  name: string;
  kind: string;
  enabled: boolean;
}

const WILDCARD_SUFFIX = "/*";

export const opOf = (service: string, method: string): string => `/${service}/${method}`;

export const wildcardOpOf = (service: string): string => `/${service}${WILDCARD_SUFFIX}`;

export function grantableServices(catalog: readonly Service[]): GrantableView[] {
  return catalog
    .map((svc) => ({
      name: svc.name,
      fullName: svc.fullName,
      comment: svc.comment,
      methods: (svc.methods ?? []).filter((mt) => mt.access === "ADMIN" && !mt.builtin),
    }))
    .filter((svc) => svc.methods.length > 0);
}

export const serviceDatascopeAny = (svc: GrantableView): boolean =>
  svc.methods.some((mt) => mt.datascope);

export const catalogDatascopeAny = (catalog: readonly GrantableView[]): boolean =>
  catalog.some(serviceDatascopeAny);

export function toScopeKind(value: number | string | undefined): ScopeKind {
  if (typeof value === "number") return value;
  // Wire names carry the full proto prefix ("SCOPE_KIND_SELF"); the TS enum
  // keys are the short names, and a numeric string would hit the reverse
  // mapping - so strip the prefix, then type-check the lookup.
  const key = value?.replace(/^SCOPE_KIND_/, "") ?? "";
  const hit = key === "" ? undefined : (ScopeKind as Record<string, number | string>)[key];
  return typeof hit === "number" ? hit : ScopeKind.UNSPECIFIED;
}

export function grantMapOf(grants: readonly RoleGrant[]): GrantMap {
  const map: GrantMap = {};
  for (const grant of grants) {
    const scope = toScopeKind(grant.scopeKind);
    map[grant.operation] = {
      scope: scope === ScopeKind.ALL ? ScopeKind.UNSPECIFIED : scope,
      policies: [...(grant.policyIds ?? [])],
    };
  }
  return map;
}

export function roleGrantsOf(map: GrantMap): RoleGrant[] {
  return Object.entries(map).map(([operation, entry]) =>
    create(RoleGrantSchema, {
      operation,
      scopeKind: entry.scope,
      policyIds: entry.policies,
    }),
  );
}

export const GLOBAL_WILDCARD = "/*";

const NARROWING: ReadonlySet<ScopeKind> = new Set([ScopeKind.SELF, ScopeKind.DEPT_TREE]);

export function summarizeGrantMap(
  catalog: readonly GrantableView[],
  grants: GrantMap,
): {
  global: boolean;
  ops: number;
  services: number;
  wildcards: number;
  narrowed: number;
  attached: number;
} {
  const narrowed = Object.values(grants).filter((e) => NARROWING.has(e.scope)).length;
  const attached = Object.values(grants).filter((e) => e.policies.length > 0).length;
  if (grants[GLOBAL_WILDCARD] !== undefined) {
    return { global: true, ops: 0, services: catalog.length, wildcards: 0, narrowed, attached };
  }
  let ops = 0;
  let services = 0;
  let wildcards = 0;
  for (const svc of catalog) {
    const wild = grants[wildcardOpOf(svc.fullName)] !== undefined;
    const checked = svc.methods.filter(
      (mt) => grants[opOf(svc.fullName, mt.name)] !== undefined,
    ).length;
    ops += checked;
    if (wild) {
      wildcards += 1;
      services += 1;
    } else if (checked > 0) {
      services += 1;
    }
  }
  return { global: false, ops, services, wildcards, narrowed, attached };
}

type GrantsSummary = {
  narrowed: number;
  attached: number;
} & (
  | { kind: "all" }
  | { kind: "count"; count: number }
  | { kind: "wildcard"; wildcards: number }
  | { kind: "mixed"; ops: number; wildcards: number }
);

export function summarizeGrants(
  globalGrant: boolean,
  grantCount: number,
  wildcardCount: number,
  narrowedCount = 0,
  policyCount = 0,
): GrantsSummary {
  const wildcards = Number(wildcardCount) || 0;
  const ops = (Number(grantCount) || 0) - wildcards;
  const extra = { narrowed: Number(narrowedCount) || 0, attached: Number(policyCount) || 0 };
  if (globalGrant) return { kind: "all", ...extra };
  if (wildcards > 0 && ops > 0) return { kind: "mixed", ops, wildcards, ...extra };
  if (wildcards > 0) return { kind: "wildcard", wildcards, ...extra };
  return { kind: "count", count: ops, ...extra };
}

export function summarizePatterns(operations: string[]): GrantsSummary {
  const global = operations.includes("/*");
  const wildcards = operations.filter((op) => op.endsWith("/*") && op !== "/*").length;
  return summarizeGrants(global, operations.length, wildcards);
}

// The one grant-summary sentence for compact surfaces (table cell, role card,
// preview line) - domain vocabulary shared with the drawer summary bar.
export function formatGrantsSummary(s: GrantsSummary): string {
  let out: string;
  if (s.kind === "all") out = m.system_roles_grants_all();
  else if (s.kind === "wildcard") out = m.system_roles_grants_wildcard({ m: s.wildcards });
  else if (s.kind === "mixed") out = m.system_roles_grants_mixed({ ops: s.ops, m: s.wildcards });
  else out = m.system_roles_grants_count({ count: s.count });
  if (s.narrowed > 0) out += ` · ${m.system_roles_perm_summary_narrowed({ k: s.narrowed })}`;
  if (s.attached > 0) out += ` · ${m.system_roles_perm_summary_attached({ k: s.attached })}`;
  return out;
}

// ScopeKind is enum vocabulary - one label set shared by the grant tree
// picker and the diag attribution tag.
export function scopeKindLabel(kind: ScopeKind): string {
  switch (kind) {
    case ScopeKind.ALL:
      return m.system_roles_scope_all();
    case ScopeKind.SELF:
      return m.system_roles_scope_self();
    case ScopeKind.DEPT_TREE:
      return m.system_roles_scope_dept_tree();
    default:
      return m.system_roles_scope_unspecified();
  }
}
