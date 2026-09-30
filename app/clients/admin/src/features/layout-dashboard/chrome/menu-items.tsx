import type { MenuProps } from "antd";
import { type NavNode, text } from "../protocol/nav";

type MenuItem = NonNullable<MenuProps["items"]>[number];

const iconOf = (node: NavNode) => {
  const Icon = node.meta?.icon;
  return Icon ? <Icon aria-hidden className="size-4" /> : undefined;
};

export function toMenuItems(nodes: NavNode[], collapsed: boolean): MenuItem[] {
  const items: MenuItem[] = [];
  for (const node of nodes) {
    if (node.meta?.dividerBefore && items.length > 0) {
      items.push({ type: "divider" });
    }
    if (node.children.length === 0) {
      items.push({ key: node.path, icon: iconOf(node), label: text(node.title) });
      continue;
    }
    const children = toMenuItems(node.children, collapsed);
    if (node.meta?.type === "group" && !collapsed) {
      items.push({ type: "group", label: text(node.title), children });
    } else {
      items.push({ key: node.path, icon: iconOf(node), label: text(node.title), children });
    }
  }
  return items;
}
