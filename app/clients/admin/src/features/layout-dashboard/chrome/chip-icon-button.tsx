import { Tooltip } from "antd";
import clsx from "clsx";
import type { MouseEventHandler, ReactNode } from "react";

interface ChipIconButtonProps {
  children: ReactNode;
  className?: string;
  disabled?: boolean;
  label: string;
  onClick: MouseEventHandler<HTMLButtonElement>;
  tooltip?: boolean;
}

export function ChipIconButton({
  children,
  className,
  disabled = false,
  label,
  tooltip = true,
  onClick,
}: Readonly<ChipIconButtonProps>) {
  const btn = (
    <button
      aria-label={label}
      className={clsx(
        "inline-flex cursor-pointer items-center justify-center rounded-full border border-line bg-fill-quaternary p-2 text-ink-secondary outline-none backdrop-blur-[6px] transition-colors hover:border-ink-tertiary hover:text-ink focus-visible:border-primary focus-visible:text-ink disabled:cursor-not-allowed disabled:opacity-40",
        className,
      )}
      disabled={disabled || undefined}
      onClick={onClick}
      type="button"
    >
      {children}
    </button>
  );

  if (tooltip && label) return <Tooltip title={label}>{btn}</Tooltip>;
  return btn;
}
