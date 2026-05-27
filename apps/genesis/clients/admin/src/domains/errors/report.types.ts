export type ErrorSource = "route" | "component" | "manual"

export interface ErrorReportContext {
  source: ErrorSource
}
