import type { ReactNode } from "react"

type BaseOption = {
  label: string
  show?: boolean
}

export type ElementFilterOption = BaseOption & {
  name: string
  element: ReactNode
}

export type InputFilterOption = BaseOption & {
  name: string
  type?: "input"
  placeholder?: string
}

export type SelectFilterOption = BaseOption & {
  name: string
  type: "select"
  options: { label: string; value: string | number }[]
  placeholder?: string
  mode?: "multiple" | "tags"
}

export type DateFilterOption = BaseOption & {
  name: string
  type: "date"
  picker?: "date" | "week" | "month" | "quarter" | "year"
  showTime?: boolean
  placeholder?: string
}

export type NumberFilterOption = BaseOption & {
  name: string
  type: "number"
  placeholder?: string
  min?: number
  max?: number
}

export type RangeNumberFilterOption = BaseOption & {
  name: [string, string]
  type: "range-number"
  placeholder?: [string, string]
}

export type RangeDateFilterOption = BaseOption & {
  name: [string, string]
  type: "range-date" | "range-datetime"
  placeholder?: [string, string]
}

export type RangeFilterOption =
  | RangeNumberFilterOption
  | RangeDateFilterOption

export type BuiltinFilterOption =
  | InputFilterOption
  | SelectFilterOption
  | DateFilterOption
  | NumberFilterOption
  | RangeNumberFilterOption
  | RangeDateFilterOption

export type FilterOption = ElementFilterOption | BuiltinFilterOption

export interface FilterLabels {
  search?: string
  reset?: string
  fold?: string
  expand?: string
}

export interface FilterProps<
  T extends Record<string, unknown> = Record<string, unknown>,
> {
  options: FilterOption[]
  onFilter?: (values: T) => void
  onReset?: () => void
  disabled?: boolean
  initialValues?: T
  columns?: 2 | 3 | 4 | 6 | 8 | 12 | 24
  size?: "small" | "middle" | "large"
  labels?: FilterLabels
  buttonVariant?: "icon-text" | "icon" | "text"
}
