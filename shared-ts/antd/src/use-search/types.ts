// ── Store abstraction ────────────────────────────────────────────
export interface SearchStore<T extends Record<string, unknown>> {
  state: T
  patch: (partial: Partial<T>) => void
}

// ── Hook return types ────────────────────────────────────────────
export interface FilterControls<T extends Record<string, unknown>> {
  values: Partial<T>
  onFilter: (values: Partial<T>) => void
  onReset: () => void
}

export interface PaginationControls {
  pageNo: number
  pageSize: number
  onPageChange: (pageNo: number, pageSize: number) => void
}

export interface SortControls {
  sort: string | undefined
  onSortToggle: (field: string) => void
  onSortClear: () => void
}
