import { Tag, Tooltip } from "antd";
import { Check, X } from "lucide-react";
import { m } from "#/paraglide/messages";

export interface ExplainPolicyView {
  kind: string;
  name: string;
  state: string; // passed | failed - the wire reserves indeterminate
}

export function PolicyChip({ policy }: Readonly<{ policy: ExplainPolicyView }>) {
  const failed = policy.state === "failed";
  return (
    <Tooltip
      title={`${m.system_diag_policy_and_hint()} · ${
        failed ? m.system_diag_policy_failed() : m.system_diag_policy_passed()
      }`}
    >
      <Tag
        className="me-0"
        color={failed ? "error" : policy.kind === "calendar" ? "gold" : "blue"}
        icon={failed ? <X size={11} /> : <Check size={11} />}
      >
        {policy.name}
      </Tag>
    </Tooltip>
  );
}
