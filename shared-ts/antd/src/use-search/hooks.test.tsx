import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { z } from "zod";
import { pageNoField, pageSizeField, sortField } from "./schema";
import type { SearchStore } from "./types";
import { useFilter } from "./use-filter";
import { useSort } from "./use-sort";
import { useUrlSearchStore } from "./use-url-search-store";

const schema = z.object({
  pageNo: pageNoField(),
  pageSize: pageSizeField(),
  sort: sortField(),
  status: z.enum(["on", "off"]).optional(),
  keyword: z.string().optional(),
});

function makeStore(initial: Record<string, unknown>): {
  store: SearchStore<Record<string, unknown>>;
  patches: Array<Partial<Record<string, unknown>>>;
} {
  const patches: Array<Partial<Record<string, unknown>>> = [];
  const store: SearchStore<Record<string, unknown>> = {
    state: initial,
    // A URL-like store: patch merges, undefined deletes the key.
    patch: (partial) => {
      for (const [key, value] of Object.entries(partial)) {
        if (value === undefined) delete initial[key];
        else initial[key] = value;
      }
      patches.push(partial);
    },
  };
  return { store, patches };
}

describe("useUrlSearchStore", () => {
  it("routes patches through onNavigate with a stable patch fn", () => {
    const onNavigate = vi.fn();
    const { rerender, result } = renderHook(
      ({ search }) => useUrlSearchStore(search, { onNavigate }),
      {
        initialProps: { search: { keyword: "x" } as Record<string, unknown> },
      },
    );
    const first = result.current.patch;
    rerender({ search: { keyword: "y" } as Record<string, unknown> });
    expect(result.current.patch).toBe(first); // stable identity
    act(() => result.current.patch({ keyword: "z" }));
    expect(onNavigate).toHaveBeenCalledWith({ keyword: "z" });
    expect(result.current.state).toEqual({ keyword: "y" }); // state follows props, not patch
  });
});

describe("useFilter", () => {
  it("exposes non-reserved keys as values", () => {
    const { result } = renderHook(() =>
      useFilter(makeStore({ pageNo: 3, keyword: "a", status: "on" }).store, schema),
    );
    expect(result.current.values).toEqual({ keyword: "a", status: "on" });
  });

  it("normalizes empty strings, keeps clear state, and resets pageNo", () => {
    const { store, patches } = makeStore({ pageNo: 3, keyword: "a" });
    const { result } = renderHook(() => useFilter(store, schema));
    act(() => result.current.onFilter({ keyword: "", status: "off" }));
    expect(patches[0]).toEqual({ status: "off", keyword: undefined, pageNo: 1 });
  });

  it("reset removes filter keys and never touches the reserved ones", () => {
    const { store, patches } = makeStore({ pageNo: 2, keyword: "a", status: "on" });
    const { result } = renderHook(() => useFilter(store, schema));
    act(() => result.current.onReset());
    expect(patches[0]).toEqual({ keyword: undefined, status: undefined });
  });
});

describe("useSort", () => {
  it("toggles asc→desc per field, tracks several fields, and resets page", () => {
    const { store, patches } = makeStore({ pageNo: 4, sort: "name:asc" });
    const { result } = renderHook(() => useSort(store, schema));
    act(() => result.current.onSortToggle("name"));
    expect(patches[0]).toEqual({ sort: "name:desc", pageNo: 1 });
    act(() => result.current.onSortToggle("code"));
    expect(patches[1]).toEqual({ sort: "name:desc,code:asc", pageNo: 1 });
  });

  it("clear removes the sort key", () => {
    const { store, patches } = makeStore({ sort: "name:asc" });
    const { result } = renderHook(() => useSort(store, schema));
    act(() => result.current.onSortClear());
    expect(patches[0]).toEqual({ sort: undefined, pageNo: 1 });
  });
});
