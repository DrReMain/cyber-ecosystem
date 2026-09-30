import { ScopeKind } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/role_pb";
import { Button, Typography } from "antd";
import { X } from "lucide-react";
import { useMemo, useState } from "react";
import { m } from "#/paraglide/messages";
import {
  catalogDatascopeAny,
  GLOBAL_WILDCARD,
  type GrantableView,
  type GrantMap,
  opOf,
  type PolicyOptionView,
  summarizeGrantMap,
  wildcardOpOf,
} from "../grants";
import { PolicySubRow } from "./policy-sub-row";
import { PolicyToggle } from "./policy-toggle";
import { ScopeSelect } from "./scope-select";
import { ServiceBlock } from "./service-block";

const serviceTouched = (svc: GrantableView, grants: GrantMap): boolean =>
  grants[wildcardOpOf(svc.fullName)] !== undefined ||
  svc.methods.some((mt) => grants[opOf(svc.fullName, mt.name)] !== undefined);

interface GrantTreeProps {
  catalog: GrantableView[];
  grants: GrantMap;
  policies: PolicyOptionView[];
  onChange: (next: GrantMap) => void;
}

export function GrantTree({ catalog, grants, policies, onChange }: Readonly<GrantTreeProps>) {
  const setScope = (op: string, scope: ScopeKind | undefined) => {
    const next = { ...grants };
    if (scope === undefined) delete next[op];
    else next[op] = { scope, policies: next[op]?.policies ?? [] };
    onChange(next);
  };
  const setPolicies = (op: string, next: string[]) => {
    const entry = grants[op];
    if (entry === undefined) return; // only granted rows can attach policies
    onChange({ ...grants, [op]: { ...entry, policies: next } });
  };
  const setService = (svc: GrantableView, on: boolean) => {
    const wildcard = wildcardOpOf(svc.fullName);
    const next = { ...grants };
    delete next[wildcard];
    for (const op of svc.methods.map((mt) => opOf(svc.fullName, mt.name))) {
      if (!on) delete next[op];
      else if (next[op] === undefined) next[op] = { scope: ScopeKind.UNSPECIFIED, policies: [] };
    }
    if (on) setClosedOverride((prev) => ({ ...prev, [svc.fullName]: false }));
    onChange(next);
  };
  const setWildcard = (svc: GrantableView) => {
    const next = { ...grants };
    for (const op of svc.methods.map((mt) => opOf(svc.fullName, mt.name))) delete next[op];
    next[wildcardOpOf(svc.fullName)] = { scope: ScopeKind.UNSPECIFIED, policies: [] };
    onChange(next);
  };
  const summary = summarizeGrantMap(catalog, grants);
  const globalOn = grants[GLOBAL_WILDCARD] !== undefined;
  const [closedOverride, setClosedOverride] = useState<Record<string, boolean>>({});
  const [globalPolicyOpen, setGlobalPolicyOpen] = useState(false);
  const closed = useMemo(
    () => ({
      ...Object.fromEntries(
        catalog.map((svc, i) => [svc.fullName, !(i === 0 || serviceTouched(svc, grants))]),
      ),
      ...closedOverride,
    }),
    [catalog, grants, closedOverride],
  );

  return (
    <div className="flex flex-col gap-3">
      {globalOn && (
        <div className="flex flex-col gap-1 rounded-lg border border-black/8 px-3 py-2 dark:border-white/8">
          <div className="flex min-h-9 items-center gap-2">
            <span className="text-[13px]">{m.system_roles_perm_global()}</span>
            {catalogDatascopeAny(catalog) && (
              <ScopeSelect
                onChange={(value) => setScope(GLOBAL_WILDCARD, value)}
                value={grants[GLOBAL_WILDCARD]?.scope}
              />
            )}
            <PolicyToggle
              count={grants[GLOBAL_WILDCARD]?.policies.length ?? 0}
              onClick={() => setGlobalPolicyOpen((prev) => !prev)}
            />
            <Button
              aria-label="clear"
              icon={<X size={14} />}
              onClick={() => setScope(GLOBAL_WILDCARD, undefined)}
              size="small"
              type="text"
            />
          </div>
          {(globalPolicyOpen || (grants[GLOBAL_WILDCARD]?.policies.length ?? 0) > 0) && (
            <PolicySubRow
              grants={grants}
              onSetPolicies={setPolicies}
              op={GLOBAL_WILDCARD}
              policies={policies}
            />
          )}
        </div>
      )}
      {catalog.map((svc) => (
        <ServiceBlock
          globalOn={globalOn}
          grants={grants}
          key={svc.fullName}
          onSetPolicies={setPolicies}
          onSetScope={setScope}
          onSetService={setService}
          onSetWildcard={setWildcard}
          onToggle={() =>
            setClosedOverride((prev) => ({ ...prev, [svc.fullName]: !closed[svc.fullName] }))
          }
          open={!closed[svc.fullName]}
          policies={policies}
          svc={svc}
        />
      ))}
      <div className="rounded-md bg-black/3 px-3 py-2 dark:bg-white/5">
        <Typography.Text className="text-sm" type="secondary">
          {summary.global
            ? m.system_roles_perm_summary_global({ m: summary.services })
            : summary.wildcards > 0
              ? m.system_roles_perm_summary_wild({
                  ops: summary.ops,
                  m: summary.services,
                  w: summary.wildcards,
                })
              : m.system_roles_perm_summary({ ops: summary.ops, m: summary.services })}
          {summary.narrowed > 0 &&
            ` · ${m.system_roles_perm_summary_narrowed({ k: summary.narrowed })}`}
          {summary.attached > 0 &&
            ` · ${m.system_roles_perm_summary_attached({ k: summary.attached })}`}
        </Typography.Text>
      </div>
    </div>
  );
}
