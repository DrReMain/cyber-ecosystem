import type { LucideIcon } from "lucide-react";
import type { m } from "#/paraglide/messages";

export type MessageKey = keyof typeof m;

export type NoParamMessageKey = {
  [K in MessageKey]: (typeof m)[K] extends (inputs?: never, options?: never) => unknown ? K : never;
}[MessageKey];

export interface RouteMenuMeta {
  icon?: LucideIcon;
  type?: "group";
  hide?: boolean;
  dividerBefore?: boolean;
  order?: number;
}

declare module "@tanstack/react-router" {
  interface StaticDataRouteOption {
    title?: NoParamMessageKey;
    menu?: RouteMenuMeta;
    operations?: readonly string[];
  }
}
