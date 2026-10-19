import { act, renderHook } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { z } from "zod";
import { pageNoField, pageSizeField } from "./schema";
import type { SearchStore } from "./types";
import { useServerPagination } from "./use-server-pagination";

const schema = z.object({
  pageNo: pageNoField(),
  pageSize: pageSizeField(20),
});

function makeStore(initial: Record<string, unknown>): {
  store: SearchStore<Record<string, unknown>>;
  patches: Array<Partial<Record<string, unknown>>>;
} {
  const patches: Array<Partial<Record<string, unknown>>> = [];
  const store: SearchStore<Record<string, unknown>> = {
    state: initial,
    patch: (partial) => {
      Object.assign(initial, partial);
      patches.push(partial);
    },
  };
  return { store, patches };
}

describe("useServerPagination", () => {
  it("falls back to search state while no response echo exists", () => {
    const { store } = makeStore({ pageNo: 3, pageSize: 50 });
    const { result } = renderHook(() => useServerPagination(store, schema, undefined));
    expect(result.current.current).toBe(3);
    expect(result.current.pageSize).toBe(50);
    expect(result.current.total).toBe(0);
  });

  it("prefers the response echo over search state", () => {
    const { store } = makeStore({ pageNo: 9, pageSize: 20 });
    const { result } = renderHook(() =>
      useServerPagination(store, schema, { pageNo: 2, pageSize: 20, total: 41 }),
    );
    expect(result.current.current).toBe(2);
    expect(result.current.total).toBe(41);
  });

  it("page changes patch the search store", () => {
    const { store, patches } = makeStore({ pageNo: 1, pageSize: 20 });
    const { result } = renderHook(() => useServerPagination(store, schema, undefined));
    act(() => {
      result.current.onChange?.(4, 100);
    });
    expect(patches[0]).toEqual({ pageNo: 4, pageSize: 100 });
  });

  it("carries the size options and total copy", () => {
    const { store } = makeStore({});
    const { result } = renderHook(() =>
      useServerPagination(
        store,
        schema,
        { pageNo: 1, pageSize: 20, total: 7 },
        {
          pageSizeOptions: [20, 50, 100],
          showTotal: (n) => `total ${n}`,
        },
      ),
    );
    expect(result.current.pageSizeOptions).toEqual([20, 50, 100]);
    expect(result.current.showTotal?.(7, [1, 7])).toBe("total 7");
  });

  it("falls back to the schema defaults when state lacks the page keys", () => {
    const { store } = makeStore({ keyword: "x" });
    const { result } = renderHook(() => useServerPagination(store, schema, undefined));
    expect(result.current.current).toBe(1);
    expect(result.current.pageSize).toBe(20);
  });

  it("falls back to 1/10 when the schema has no page fields", () => {
    const noPage = z.object({ keyword: z.string().optional() });
    const { store } = makeStore({});
    const { result } = renderHook(() => useServerPagination(store, noPage, undefined));
    expect(result.current.current).toBe(1);
    expect(result.current.pageSize).toBe(10);
  });
});
