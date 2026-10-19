import { useNavigate } from "@tanstack/react-router";
import { Button, Popconfirm, Tooltip } from "antd";
import type { LucideIcon } from "lucide-react";
import type { MouseEvent, ReactNode } from "react";
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
  const navigate = useNavigate();
  // Overlays open toward the table content — the inline-start side of the
  // fixed action column.
  const side = getTextDirection() === "rtl" ? "right" : "left";
  // Internal hrefs route through the router — a raw anchor would reload the
  // SPA; external URLs keep the native anchor. Modifier clicks stay native.
  const internal = href?.startsWith("/");
  const handleClick = (event?: MouseEvent) => {
    if (
      internal &&
      event !== undefined &&
      event.button === 0 &&
      !event.metaKey &&
      !event.ctrlKey &&
      !event.shiftKey &&
      !event.altKey
    ) {
      event.preventDefault();
      void navigate({ href });
      return;
    }
    onClick?.();
  };
  const button = (
    <Tooltip arrow={false} title={label}>
      <span className="inline-flex">
        <Button
          aria-label={label}
          color={color}
          disabled={disabled}
          href={href}
          icon={<Icon aria-hidden size={14} />}
          onClick={handleClick}
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
