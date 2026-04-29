import { z } from "zod"

const SEARCH_META = z.registry<{
  searchRole: "pageNo" | "pageSize" | "sort"
}>()

export const pageNoField = (def = 1) =>
  z
    .number()
    .int()
    .min(def)
    .default(def)
    .register(SEARCH_META, { searchRole: "pageNo" })

export const pageSizeField = (def = 10) =>
  z
    .number()
    .int()
    .default(def)
    .register(SEARCH_META, { searchRole: "pageSize" })

export const sortField = (def?: string) => {
  const s = def ? z.string().default(def) : z.string().optional()
  return s.register(SEARCH_META, { searchRole: "sort" })
}

export interface SearchPartition {
  pageNoKey: string | undefined
  pageSizeKey: string | undefined
  sortKey: string | undefined
  nonFilterKeys: Set<string>
}

const cache = new WeakMap<object, SearchPartition>()

const EMPTY: SearchPartition = {
  pageNoKey: undefined,
  pageSizeKey: undefined,
  sortKey: undefined,
  nonFilterKeys: new Set(),
}

export function getPartition(schema: {
  shape: Record<string, unknown>
}): SearchPartition {
  const cached = cache.get(schema)
  if (cached) return cached

  const shape = schema.shape
  if (!shape) {
    cache.set(schema, EMPTY)
    return EMPTY
  }

  let pageNoKey: string | undefined
  let pageSizeKey: string | undefined
  let sortKey: string | undefined

  for (const [key, fieldSchema] of Object.entries(shape)) {
    const meta = SEARCH_META.get(
      fieldSchema as Parameters<typeof SEARCH_META.get>[0],
    )
    if (!meta) continue
    if (meta.searchRole === "pageNo") pageNoKey = key
    else if (meta.searchRole === "pageSize") pageSizeKey = key
    else if (meta.searchRole === "sort") sortKey = key
  }

  const nonFilterKeys = new Set<string>()
  if (pageNoKey) nonFilterKeys.add(pageNoKey)
  if (pageSizeKey) nonFilterKeys.add(pageSizeKey)
  if (sortKey) nonFilterKeys.add(sortKey)

  const partition = { pageNoKey, pageSizeKey, sortKey, nonFilterKeys }
  cache.set(schema, partition)
  return partition
}
