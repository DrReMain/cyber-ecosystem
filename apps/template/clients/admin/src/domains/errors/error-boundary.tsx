export { useErrorBoundary } from "react-error-boundary"

import type { ComponentProps } from "react"
import { ErrorBoundary as BaseErrorBoundary } from "react-error-boundary"
import { reportError } from "./report"

export function ErrorBoundary({
  onError,
  ...rest
}: ComponentProps<typeof BaseErrorBoundary>) {
  return (
    <BaseErrorBoundary
      onError={(error, info) => {
        reportError(error instanceof Error ? error : new Error(String(error)), {
          source: "component",
        })
        onError?.(error, info)
      }}
      {...rest}
    />
  )
}
