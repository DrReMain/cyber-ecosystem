import type { ElementType, PropsWithChildren } from "react";
import "./route-transition.css";

// keep in sync with route-transition.css ::view-transition-*(route-content)
export const ROUTE_VT = "route-content";

const SELECTOR = "[data-route-transition]";

function routeEl(): HTMLElement | null {
  if (typeof document === "undefined") return null;
  return document.querySelector<HTMLElement>(SELECTOR);
}

export function suppressRouteVT(): void {
  const el = routeEl();
  if (el) el.style.viewTransitionName = "none";
}

export function restoreRouteVT(): void {
  const el = routeEl();
  if (el) el.style.viewTransitionName = ROUTE_VT;
}

type RouteTransitionProps = PropsWithChildren<{
  as?: ElementType;
  className?: string;
}>;

export function RouteTransition({ as, className, children, ...rest }: RouteTransitionProps) {
  const Tag = (as ?? "div") as ElementType;
  return (
    <Tag
      className={className}
      data-route-transition=""
      style={{ viewTransitionName: ROUTE_VT }}
      {...rest}
    >
      {children}
    </Tag>
  );
}
