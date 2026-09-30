import type { DraggableAttributes, DraggableSyntheticListeners } from "@dnd-kit/core";
import { Dropdown, type MenuProps } from "antd";
import clsx from "clsx";
import { Pin, X } from "lucide-react";
import type { CSSProperties } from "react";
import { m } from "#/paraglide/messages";
import { text } from "../../protocol/nav";
import type { TabRecord } from "./tab-session";

const CHIP =
  "group flex h-full flex-none cursor-pointer select-none items-center gap-1.5 border-line-soft pe-2 ps-3 text-sm whitespace-nowrap [&:not(:first-child)]:border-s";
const CHIP_ACTIVE = "bg-primary/10 font-medium text-primary dark:bg-primary/15";
const CHIP_IDLE = "text-ink-secondary hover:bg-fill-secondary";
const CLOSE_BTN =
  "flex size-4 flex-none items-center justify-center rounded-sm opacity-60 hover:bg-fill-secondary hover:opacity-100";

export interface TabChipDragHandle {
  attributes: DraggableAttributes;
  dragging: boolean;
  listeners: DraggableSyntheticListeners;
  setNodeRef: (node: HTMLElement | null) => void;
  style: CSSProperties;
}

export interface TabChipProps {
  drag?: TabChipDragHandle;
  isActive: boolean;
  menu: { items: NonNullable<MenuProps["items"]>; onClick: (info: { key: string }) => void };
  onActivate: () => void;
  onClose: () => void;
  siblingCount: number;
  tab: TabRecord;
}

export function TabChip({
  drag,
  isActive,
  menu,
  onActivate,
  onClose,
  siblingCount,
  tab,
}: Readonly<TabChipProps>) {
  const Icon = tab.icon;
  return (
    <Dropdown menu={menu} trigger={["contextMenu"]}>
      <div
        {...(drag?.listeners ?? {})}
        aria-describedby={drag?.attributes["aria-describedby"]}
        aria-roledescription={drag?.attributes["aria-roledescription"]}
        aria-selected={isActive}
        className={clsx(
          CHIP,
          isActive ? CHIP_ACTIVE : CHIP_IDLE,
          drag && "touch-none",
          drag?.dragging && "z-20 bg-container shadow-md",
        )}
        data-tab-key={tab.key}
        onClick={onActivate}
        onKeyDown={(e) => {
          if (e.key === "Enter" || e.key === " ") {
            e.preventDefault();
            onActivate();
          }
        }}
        ref={drag?.setNodeRef}
        role="tab"
        style={drag?.style}
        tabIndex={isActive ? 0 : -1}
      >
        {Icon ? <Icon aria-hidden className="size-3.5 flex-none" /> : null}
        <span className="max-w-40 truncate">{text(tab.title)}</span>
        {!tab.affix && siblingCount > 1 ? (
          <button
            aria-label={m.layout_dashboard_tab_close()}
            className={CLOSE_BTN}
            onClick={(e) => {
              e.stopPropagation();
              onClose();
            }}
            type="button"
          >
            <X aria-hidden className="size-3" />
          </button>
        ) : tab.affix && siblingCount > 1 ? (
          <Pin aria-hidden className="size-3 flex-none opacity-50" />
        ) : null}
      </div>
    </Dropdown>
  );
}
