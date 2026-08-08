import type { ErrorBoundaryProps } from "react-error-boundary";
import { ErrorBoundary as RBErrorBoundary } from "react-error-boundary";
import { ErrorFallback } from "./ErrorFallback";
import { errorHandler } from "./error-handler";

export function ErrorBoundary({
  onError,
  ...props
}: Readonly<Omit<ErrorBoundaryProps, "FallbackComponent" | "fallback" | "fallbackRender">>) {
  return (
    <RBErrorBoundary
      {...props}
      FallbackComponent={ErrorFallback}
      onError={(error, info) => {
        errorHandler.handle(error, "render", { feedback: false });
        onError?.(error, info);
      }}
    />
  );
}
