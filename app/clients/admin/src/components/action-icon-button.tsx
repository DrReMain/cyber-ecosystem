import { useNavigate } from "@tanstack/react-router";
import { Button, Popconfirm, Tooltip } from "antd";
import type { LucideIcon } from "lucide-react";
import type { MouseEvent, ReactNode } from "react";
import { getTextDirection, localizeUrl } from "#/paraglide/runtime";

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
  // The native anchor carries the localized URL so copies and new-tab opens
  // keep the active locale — the router strips the prefix back off on its
  // input rewrite, so navigation itself still takes the raw route href.
  const internal = href !== undefined && href.startsWith("/");
  // localizeUrl may hand back a URL object; the anchor wants the string form.
  const anchorHref = internal ? localizeUrl(href).toString() : href;
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
          href={anchorHref}
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
