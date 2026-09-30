import { Button } from "antd";
import { m } from "#/paraglide/messages";

interface PolicyToggleProps {
  count: number;
  onClick: () => void;
}

export function PolicyToggle({ count, onClick }: Readonly<PolicyToggleProps>) {
  return (
    <Button onClick={onClick} size="small" type="text">
      {count > 0 ? m.system_roles_policy_count({ n: count }) : m.system_roles_policy_placeholder()}
    </Button>
  );
}
