import { ScopeKind } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/role_pb";
import { Select } from "antd";
import { m } from "#/paraglide/messages";

const scopeOptions = () => [
  { label: m.system_roles_scope_unspecified(), value: ScopeKind.UNSPECIFIED },
  { label: m.system_roles_scope_self(), value: ScopeKind.SELF },
  { label: m.system_roles_scope_dept_tree(), value: ScopeKind.DEPT_TREE },
];

const SCOPE_SELECT_CLS = "m-auto min-w-[180px]";

interface ScopeSelectProps {
  value: ScopeKind | undefined;
  onChange: (value: ScopeKind) => void;
}

export function ScopeSelect({ value, onChange }: Readonly<ScopeSelectProps>) {
  return (
    <Select
      className={SCOPE_SELECT_CLS}
      onChange={onChange}
      options={scopeOptions()}
      size="small"
      value={value}
      virtual={false}
    />
  );
}
