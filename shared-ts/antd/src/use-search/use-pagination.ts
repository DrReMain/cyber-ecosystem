import { useCallback } from "react"
import { getPartition } from "./schema"
import type { PaginationControls, SearchStore } from "./types"

export function usePagination<T extends Record<string, unknown>>(
  store: SearchStore<T>,
  schema: { shape: Record<string, unknown> },
): PaginationControls {
  const partition = getPartition(schema)
  const s = store.state as Record<string, unknown>

  const pageNo = partition.pageNoKey
    ? ((s[partition.pageNoKey] as number) ?? 1)
    : 1
  const pageSize = partition.pageSizeKey
    ? ((s[partition.pageSizeKey] as number) ?? 10)
    : 10

  const onPageChange = useCallback(
    (pageNo: number, pageSize: number) => {
      const patch: Record<string, unknown> = {}
      if (partition.pageNoKey) patch[partition.pageNoKey] = pageNo
      if (partition.pageSizeKey) patch[partition.pageSizeKey] = pageSize
      store.patch(patch as Partial<T>)
    },
    [store.patch, partition],
  )

  return { pageNo, pageSize, onPageChange }
}
