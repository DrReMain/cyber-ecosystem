import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { useTable } from "./use-table";

describe("useTable", () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  it("defaults to middle density", () => {
    const { result } = renderHook(() => useTable());
    expect(result.current.tableSize).toBe("middle");
  });

  it("shares one global state across hook instances", () => {
    const a = renderHook(() => useTable());
    const b = renderHook(() => useTable());
    act(() => a.result.current.setTableSize("small"));
    expect(a.result.current.tableSize).toBe("small");
    expect(b.result.current.tableSize).toBe("small");
  });

  it("persists the choice to localStorage", () => {
    const { result } = renderHook(() => useTable());
    act(() => result.current.setTableSize("large"));
    expect(window.localStorage.getItem("cyber.table-size")).toBe("large");
  });

  it("keeps a stable setter identity across renders", () => {
    const { result, rerender } = renderHook(() => useTable());
    const first = result.current.setTableSize;
    rerender();
    expect(result.current.setTableSize).toBe(first);
  });
});
