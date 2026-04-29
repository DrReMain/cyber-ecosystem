import { useCallback } from "react"
import { getPartition } from "./schema"
import type { SearchStore, SortControls } from "./types"

export function useSort<T extends Record<string, unknown>>(
  store: SearchStore<T>,
  schema: { shape: Record<string, unknown> },
): SortControls {
  const partition = getPartition(schema)
  const s = store.state as Record<string, unknown>

  const sort = partition.sortKey
    ? (s[partition.sortKey] as string | undefined)
    : undefined

  const onSortToggle = useCallback(
    (field: string) => {
      const state = store.state as Record<string, unknown>
      const current = partition.sortKey
        ? (state[partition.sortKey] as string) || ""
        : ""
      const parts = current ? current.split(",").filter(Boolean) : []

      const idx = parts.findIndex((p) => p.startsWith(`${field}:`))
      if (idx >= 0) {
        const [f, dir] = parts[idx].split(":")
        parts[idx] = `${f}:${dir === "asc" ? "desc" : "asc"}`
      } else {
        parts.push(`${field}:asc`)
      }

      const patch: Record<string, unknown> = {}
      if (partition.sortKey) patch[partition.sortKey] = parts.join(",")
      if (partition.pageNoKey) patch[partition.pageNoKey] = 1
      store.patch(patch as Partial<T>)
    },
    [store.state, store.patch, partition],
  )

  const onSortClear = useCallback(() => {
    const patch: Record<string, unknown> = {}
    if (partition.sortKey) patch[partition.sortKey] = undefined
    if (partition.pageNoKey) patch[partition.pageNoKey] = 1
    store.patch(patch as Partial<T>)
  }, [store.patch, partition])

  return { sort, onSortToggle, onSortClear }
}
