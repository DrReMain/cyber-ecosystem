import type { Method } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/resource_pb";
import { ScopeKind } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/role_pb";
import { Checkbox, Tooltip, Typography } from "antd";
import { m } from "#/paraglide/messages";
import { type GrantMap, opOf, type PolicyOptionView } from "../grants";
import { PolicySubRow } from "./policy-sub-row";
import { PolicyToggle } from "./policy-toggle";
import { ScopeSelect } from "./scope-select";

interface MethodRowProps {
  service: string;
  method: Method;
  grants: GrantMap;
  policies: PolicyOptionView[];
  coveredBy: boolean;
  expanded: boolean;
  onToggleExpand: () => void;
  onSetScope: (op: string, scope: ScopeKind | undefined) => void;
  onSetPolicies: (op: string, policies: string[]) => void;
}

export function MethodRow({
  service,
  method,
  grants,
  policies,
  coveredBy,
  expanded,
  onToggleExpand,
  onSetScope,
  onSetPolicies,
}: Readonly<MethodRowProps>) {
  const op = opOf(service, method.name);
  const on = coveredBy || grants[op] !== undefined;
  const granted = on && !coveredBy;
  const box = (
    <Checkbox
      checked={on}
      disabled={coveredBy}
      onChange={(e) => onSetScope(op, e.target.checked ? ScopeKind.UNSPECIFIED : undefined)}
    />
  );
  return (
    <div className="flex flex-col">
      <div className="flex min-h-9 items-center gap-1 px-3 py-1.5">
        {coveredBy ? (
          <Tooltip title={m.system_roles_perm_covered()}>
            <span>{box}</span>
          </Tooltip>
        ) : (
          box
        )}

        <span className="text-[13px]">{method.comment}</span>

        <Typography.Text className="flex-1 font-mono text-[14px]" type="secondary">
          {method.name}
        </Typography.Text>

        {granted && method.datascope ? (
          <ScopeSelect onChange={(value) => onSetScope(op, value)} value={grants[op]?.scope} />
        ) : null}

        <div className="ms-auto">
          {granted ? (
            <PolicyToggle count={grants[op]?.policies.length ?? 0} onClick={onToggleExpand} />
          ) : null}
        </div>
      </div>
      {granted && expanded ? (
        <PolicySubRow grants={grants} onSetPolicies={onSetPolicies} op={op} policies={policies} />
      ) : null}
    </div>
  );
}
