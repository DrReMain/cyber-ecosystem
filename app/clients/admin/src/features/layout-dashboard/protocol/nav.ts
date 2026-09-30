import type { StaticDataRouteOption } from "@tanstack/react-router";
import { useRouteContext, useRouter } from "@tanstack/react-router";
import { useMemo } from "react";
import { m } from "#/paraglide/messages";
import { AREA_PATH } from "../area";
import { isOperationAllowed } from "../auth/permissions";
import type { NoParamMessageKey, RouteMenuMeta } from "./route-meta";

export interface NavRoute {
  fullPath: string;
  options: { staticData?: StaticDataRouteOption };
  children?: NavRoute[];
}

export interface NavNode {
  path: string;
  title: NoParamMessageKey;
  meta: RouteMenuMeta | undefined;
  children: NavNode[];
}

export const isHome = (path: string) => path === AREA_PATH;

export const text = (key: NoParamMessageKey) => m[key]();

export function navTo(router: ReturnType<typeof useRouter>, target: string) {
  if (target.includes("?")) router.navigate({ href: target });
  else router.navigate({ to: target as never });
}

const byOrder = (a: NavNode, b: NavNode) =>
  (a.meta?.order ?? Number.MAX_SAFE_INTEGER) - (b.meta?.order ?? Number.MAX_SAFE_INTEGER);

function collect(route: NavRoute, patterns: readonly string[]): NavNode | null {
  const { title, menu: meta, operations } = route.options.staticData ?? {};
  if (meta?.hide) return null;
  if (operations && !operations.every((operation) => isOperationAllowed(patterns, operation))) {
    return null;
  }
  if (!title) return null;

  const rawChildren = route.children ?? [];
  const children = rawChildren
    .map((child) => collect(child, patterns))
    .filter((node): node is NavNode => node !== null)
    .sort(byOrder);
  if (rawChildren.length > 0 && children.length === 0) return null;

  return { path: route.fullPath.replace(/\/$/, "") || "/", title, meta, children };
}

export function buildNavNodes(routeTree: NavRoute, patterns: readonly string[]): NavNode[] {
  const dashboard = (routeTree.children ?? []).find((r) => r.fullPath === AREA_PATH);
  if (!dashboard) return [];
  return (dashboard.children ?? [])
    .map((route) => collect(route, patterns))
    .filter((node): node is NavNode => node !== null)
    .sort(byOrder);
}

export function indexByPath(nodes: NavNode[]): Map<string, NavNode> {
  const map = new Map<string, NavNode>();
  for (const node of nodes) {
    map.set(node.path, node);
    for (const [k, v] of indexByPath(node.children)) map.set(k, v);
  }
  return map;
}

export function useNavNodes(): NavNode[] {
  const router = useRouter();
  const { permissions } = useRouteContext({ from: AREA_PATH });
  return useMemo(
    () => buildNavNodes(router.routeTree as unknown as NavRoute, permissions),
    [router, permissions],
  );
}

export function useNavIndex(): Map<string, NavNode> {
  const nodes = useNavNodes();
  return useMemo(() => indexByPath(nodes), [nodes]);
}

export function firstLeafOf(nodes: NavNode[]): NavNode | null {
  for (const node of nodes) {
    if (node.children.length === 0) return node;
    const leaf = firstLeafOf(node.children);
    if (leaf) return leaf;
  }
  return null;
}
