import type { GrantMap, PolicyOptionView } from "../grants";
import { PolicySelect } from "./policy-select";

interface PolicySubRowProps {
  op: string;
  grants: GrantMap;
  policies: PolicyOptionView[];
  onSetPolicies: (op: string, policies: string[]) => void;
}

// A grant's policies edit in a sub-row under the operation row: the row
// itself stays light (one trailing toggle), and the picker gets full width
// instead of squeezing the description column.
export function PolicySubRow({ op, grants, policies, onSetPolicies }: Readonly<PolicySubRowProps>) {
  const attached = grants[op]?.policies ?? [];
  return (
    <div className="flex flex-col gap-1 px-3 ps-9 pb-2">
      <PolicySelect
        onChange={(next) => onSetPolicies(op, next)}
        options={policies}
        value={attached}
      />
    </div>
  );
}
