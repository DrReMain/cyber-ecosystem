import type { ElementType, PropsWithChildren } from "react";
import "./route-transition.css";

// keep in sync with route-transition.css ::view-transition-*(route-content)
export const ROUTE_VT = "route-content";

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
