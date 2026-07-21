import { useId } from "react";
import { type RouterProgressConfig, resolveRouterProgressConfig } from "./config";
import { useRouterProgress } from "./use-router-progress";

export type RouterProgressProps = Partial<RouterProgressConfig>;

/** Displays loading progress for route transitions managed by TanStack Router. */
export function RouterProgress(props: RouterProgressProps = {}) {
  const config = resolveRouterProgressConfig(props);
  const { containerRef, barRef } = useRouterProgress(config);
  const animationName = `router-progress-cycle-${useId().replace(/[^a-zA-Z0-9_-]/g, "")}`;

  const positionStyle = config.position === "top" ? { top: 0 } : { bottom: 0 };

  return (
    <div
      aria-hidden="true"
      ref={containerRef}
      style={{
        display: "none",
        position: "fixed",
        left: 0,
        right: 0,
        height: config.height,
        zIndex: config.zIndex,
        pointerEvents: "none",
        ...positionStyle,
      }}
    >
      <style>{`
        @keyframes ${animationName} { to { background-position: 300% 0; } }
        @media (prefers-reduced-motion: reduce) {
          [data-router-progress-animation="${animationName}"] {
            animation: none !important;
            transition: none !important;
          }
        }
      `}</style>
      <div
        data-router-progress-animation={animationName}
        ref={barRef}
        style={{
          width: 0,
          height: "100%",
          background: config.color,
          backgroundSize: "300% 100%",
          animation:
            config.cycleDuration > 0
              ? `${animationName} ${config.cycleDuration}ms linear infinite`
              : undefined,
        }}
      />
    </div>
  );
}
