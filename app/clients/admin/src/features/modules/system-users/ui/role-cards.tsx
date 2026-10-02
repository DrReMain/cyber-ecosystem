import type { Role } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/role_pb";
import { Checkbox, Tag, Typography } from "antd";
import { m } from "#/paraglide/messages";
import { formatGrantsSummary, summarizeGrants } from "../../system-roles/grants";

function grantsSummary(role: Role): string {
  return formatGrantsSummary(
    summarizeGrants(
      role.globalGrant ?? false,
      Number(role.grantCount ?? 0),
      Number(role.wildcardCount ?? 0),
      Number(role.narrowedCount ?? 0),
      Number(role.policyCount ?? 0),
    ),
  );
}

interface RoleCardsProps {
  roles: Role[];
  selected: string[];
  onChange: (next: string[]) => void;
}

export function RoleCards({ roles, selected, onChange }: Readonly<RoleCardsProps>) {
  if (roles.length === 0) {
    return <Typography.Text type="secondary">{m.system_users_roles_empty()}</Typography.Text>;
  }
  const toggle = (code: string, on: boolean) => {
    onChange(on ? [...selected, code] : selected.filter((c) => c !== code));
  };
  return (
    <div className="grid grid-cols-2 gap-2">
      {roles.map((role) => (
        <div
          className={`flex flex-col gap-1 rounded-lg border border-line-soft px-3 py-2 ${
            role.enabled ? "" : "opacity-50"
          }`}
          key={role.code}
        >
          <Checkbox
            checked={selected.includes(role.code ?? "")}
            onChange={(e) => toggle(role.code ?? "", e.target.checked)}
          >
            <span className="font-medium text-[13px]">{role.name}</span>
          </Checkbox>
          <div className="flex items-center gap-2 ps-6">
            <Typography.Text className="font-mono text-[12px]" type="secondary">
              {role.code}
            </Typography.Text>
            {!role.enabled && <Tag className="ms-auto">{m.system_users_role_disabled()}</Tag>}
          </div>
          <div className="ps-6 text-[12px] text-ink-tertiary">{grantsSummary(role)}</div>
        </div>
      ))}
    </div>
  );
}
