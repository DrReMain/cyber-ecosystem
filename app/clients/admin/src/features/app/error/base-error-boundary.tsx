import { useRouterState } from "@tanstack/react-router";
import type { PropsWithChildren } from "react";
import { ErrorBoundary } from "react-error-boundary";
import { capture } from "#/domains/error";
import { ErrorFallback } from "./error-fallback";

export function BaseErrorBoundary({ children }: Readonly<PropsWithChildren>) {
  const leafRouteId = useRouterState({ select: (s) => s.matches.at(-1)?.routeId });
  return (
    <ErrorBoundary
      fallbackRender={({ error, resetErrorBoundary }) => (
        <ErrorFallback error={error} onRetry={resetErrorBoundary} />
      )}
      onError={(error) => capture(error, { feedback: false })}
      resetKeys={[leafRouteId]}
    >
      {children}
    </ErrorBoundary>
  );
}
