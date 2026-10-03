import type { LucideIcon } from "lucide-react";
import type { m } from "#/paraglide/messages";

export type MessageKey = keyof typeof m;

export interface RouteMenuMeta {
  icon?: LucideIcon;
  type?: "group";
  hide?: boolean;
  dividerBefore?: boolean;
  order?: number;
}

export const compactParams = (
  params: Record<string, string | undefined> | undefined,
): Record<string, string> | undefined => {
  if (!params) return undefined;
  const out: Record<string, string> = {};
  for (const [k, v] of Object.entries(params)) {
    if (typeof v === "string") out[k] = v;
  }
  return Object.keys(out).length > 0 ? out : undefined;
};

declare module "@tanstack/react-router" {
  interface StaticDataRouteOption {
    title?: MessageKey;
    menu?: RouteMenuMeta;
    operations?: readonly string[];
  }
}
