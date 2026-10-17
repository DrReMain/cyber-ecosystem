import { Button, Popconfirm, Tooltip } from "antd";
import type { LucideIcon } from "lucide-react";
import type { ReactNode } from "react";
import { getTextDirection } from "#/paraglide/runtime";

interface ActionIconConfirm {
  description?: ReactNode;
  onConfirm: () => void;
  title: ReactNode;
}

interface ActionIconButtonProps {
  color?: "primary" | "default" | "danger";
  confirm?: ActionIconConfirm;
  disabled?: boolean;
  href?: string;
  icon: LucideIcon;
  label: string;
  onClick?: () => void;
  rel?: string;
  target?: string;
}

export function ActionIconButton({
  color = "default",
  confirm,
  disabled,
  href,
  icon: Icon,
  label,
  onClick,
  rel,
  target,
}: Readonly<ActionIconButtonProps>) {
  // Overlays open toward the table content — the inline-start side of the
  // fixed action column.
  const side = getTextDirection() === "rtl" ? "right" : "left";
  const button = (
    <Tooltip arrow={false} title={label}>
      <span className="inline-flex">
        <Button
          aria-label={label}
          color={color}
          disabled={disabled}
          href={href}
          icon={<Icon aria-hidden size={14} />}
          onClick={onClick}
          rel={rel}
          size="small"
          target={target}
          variant="text"
        />
      </span>
    </Tooltip>
  );
  if (confirm === undefined) return button;
  return (
    <Popconfirm
      cancelButtonProps={{ variant: "filled", color: "default" }}
      description={confirm.description}
      okButtonProps={{ variant: "filled", color: "danger" }}
      onConfirm={confirm.onConfirm}
      placement={side}
      title={confirm.title}
    >
      <span className="inline-flex">{button}</span>
    </Popconfirm>
  );
}
