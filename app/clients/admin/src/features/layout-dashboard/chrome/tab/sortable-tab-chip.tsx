import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import type { CSSProperties } from "react";
import { TabChip, type TabChipProps } from "./tab-chip";

export function SortableTabChip(props: Omit<Readonly<TabChipProps>, "drag">) {
  const { attributes, isDragging, listeners, setNodeRef, transform, transition } = useSortable({
    id: props.tab.key,
  });
  const style: CSSProperties = {
    transform: CSS.Transform.toString(transform),
    transition,
  };
  return (
    <TabChip {...props} drag={{ attributes, dragging: isDragging, listeners, setNodeRef, style }} />
  );
}
