import { timestampFromMs } from "@bufbuild/protobuf/wkt"

interface PaginationInput {
  pageNo?: number
  pageSize?: number
  createdAtA?: number
  createdAtZ?: number
  updatedAtA?: number
  updatedAtZ?: number
}

export function buildConnectPage(search: PaginationInput) {
  return {
    pageNo: search.pageNo ?? 1,
    pageSize: search.pageSize ?? 10,
    ...(search.createdAtA != null
      ? { createdAtA: timestampFromMs(search.createdAtA) }
      : {}),
    ...(search.createdAtZ != null
      ? { createdAtZ: timestampFromMs(search.createdAtZ) }
      : {}),
    ...(search.updatedAtA != null
      ? { updatedAtA: timestampFromMs(search.updatedAtA) }
      : {}),
    ...(search.updatedAtZ != null
      ? { updatedAtZ: timestampFromMs(search.updatedAtZ) }
      : {}),
  }
}

export function buildHTTPPage(search: PaginationInput) {
  return {
    "page.pageNo": search.pageNo ?? 1,
    "page.pageSize": search.pageSize ?? 10,
    ...(search.createdAtA != null
      ? { "page.createdAtA": new Date(search.createdAtA).toISOString() }
      : {}),
    ...(search.createdAtZ != null
      ? { "page.createdAtZ": new Date(search.createdAtZ).toISOString() }
      : {}),
    ...(search.updatedAtA != null
      ? { "page.updatedAtA": new Date(search.updatedAtA).toISOString() }
      : {}),
    ...(search.updatedAtZ != null
      ? { "page.updatedAtZ": new Date(search.updatedAtZ).toISOString() }
      : {}),
  }
}

export function buildOrderBy(sort?: string): string[] {
  if (!sort) return []
  return sort.split(",").filter((s) => s.includes(":"))
}

export function defaultKV<T>(key: string, value: T): { [key]?: T } {
  return value ? { [key]: value } : {}
}
