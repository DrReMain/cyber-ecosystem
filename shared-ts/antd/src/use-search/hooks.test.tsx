import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { z } from "zod";
import { pageNoField, pageSizeField, sortField } from "./schema";
import type { SearchStore } from "./types";
import { useFilter } from "./use-filter";
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
