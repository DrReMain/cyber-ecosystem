import { Select, Tag } from "antd";
import { m } from "#/paraglide/messages";
import type { PolicyOptionView } from "../grants";

// Wire contract cap: RoleGrant.policy_ids is unique, max 16 items.
const MAX_POLICIES = 16;

function optionLabel(policy: PolicyOptionView) {
  return (
    <span className={policy.enabled ? "" : "opacity-55"}>
      <Tag className="me-1" color={policy.kind === "calendar" ? "gold" : "blue"}>
        {policy.kind === "calendar"
          ? m.system_policies_kind_calendar()
          : m.system_policies_kind_time_window()}
      </Tag>
      {policy.name}
      {!policy.enabled && ` · ${m.system_roles_policy_disabled()}`}
    </span>
  );
}

interface PolicySelectProps {
  options: PolicyOptionView[];
  value: string[];
  onChange: (next: string[]) => void;
}

export function PolicySelect({ options, value, onChange }: Readonly<PolicySelectProps>) {
  const byId = new Map(options.map((p) => [p.id, p]));
  return (
    <Select
      allowClear
      className="w-full"
      maxTagCount="responsive"
      mode="multiple"
      onChange={(next) => onChange([...new Set(next)].slice(0, MAX_POLICIES))}
      options={options.map((p) => ({
        value: p.id,
        label: optionLabel(p),
      }))}
      placeholder={m.system_roles_policy_placeholder()}
      size="small"
      tagRender={(props) => {
        const policy = byId.get(props.value);
        return (
          <Tag
            className="me-1"
            closable
            color={policy?.enabled ? undefined : "default"}
            onClose={(e) => props.onClose(e)}
          >
            {policy?.name ?? props.value}
            {policy && !policy.enabled ? ` · ${m.system_roles_policy_disabled()}` : ""}
          </Tag>
        );
      }}
      value={value}
      virtual={false}
    />
  );
}
