import type { TablePaginationConfig } from "antd";
import { useCallback } from "react";
import { getPartition } from "./schema";
import type { SearchSchema, SearchStore } from "./types";

export interface ServerPageEcho {
  pageNo?: number;
  pageSize?: number;
  total?: number;
}

export interface ServerPaginationOptions {
  pageSizeOptions?: number[];
  showTotal?: (total: number) => string;
}

function readNumber(
  state: Record<string, unknown>,
  defaults: Record<string, unknown>,
  key: string | undefined,
  fallback: number,
): number {
  if (!key) return fallback;
  return (state[key] as number | undefined) ?? (defaults[key] as number | undefined) ?? fallback;
}

export function useServerPagination<T extends Record<string, unknown>>(
  store: SearchStore<T>,
  schema: SearchSchema<T>,
  page: ServerPageEcho | undefined,
  options?: ServerPaginationOptions,
): TablePaginationConfig {
  const partition = getPartition(schema);
  const defaults = partition.defaults;
  const state = store.state as Record<string, unknown>;

  const onChange = useCallback(
    (pageNo: number, pageSize: number) => {
      const patch: Record<string, unknown> = {};
      if (partition.pageNoKey) patch[partition.pageNoKey] = pageNo;
      if (partition.pageSizeKey) patch[partition.pageSizeKey] = pageSize;
      store.patch(patch as Partial<T>);
    },
    [store.patch, partition],
  );

  return {
    current: page?.pageNo ?? readNumber(state, defaults, partition.pageNoKey, 1),
    onChange,
    pageSize: page?.pageSize ?? readNumber(state, defaults, partition.pageSizeKey, 10),
    pageSizeOptions: options?.pageSizeOptions ?? [10, 20, 50],
    showSizeChanger: true,
    showTotal: options?.showTotal,
    total: page?.total ?? 0,
  };
}
