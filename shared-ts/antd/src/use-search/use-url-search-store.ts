import { useCallback, useRef } from "react"
import type { SearchStore } from "./types"

export function useUrlSearchStore<T extends Record<string, unknown>>(
  search: T,
  options: { onNavigate: (patch: Partial<T>) => void },
): SearchStore<T> {
  const navigateRef = useRef(options.onNavigate)
  navigateRef.current = options.onNavigate

  const patch = useCallback(
    (partial: Partial<T>) => navigateRef.current(partial),
    [],
  )

  return { state: search, patch }
}
