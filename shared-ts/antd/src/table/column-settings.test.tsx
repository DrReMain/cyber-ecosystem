import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { setColumnHidden, useColumnVisibility } from "./column-settings";

const KEY = "test.columns";

describe("useColumnVisibility", () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  it("shows every column when nothing is persisted", () => {
    const { result } = renderHook(() => useColumnVisibility(KEY, ["a", "b"]));
    expect(result.current).toEqual(["a", "b"]);
  });

  it("hides persisted ids only", () => {
    window.localStorage.setItem(KEY, JSON.stringify(["b"]));
    const { result } = renderHook(() => useColumnVisibility(KEY, ["a", "b"]));
    expect(result.current).toEqual(["a"]);
  });

  it("new columns default to visible even with older persisted state", () => {
    window.localStorage.setItem(KEY, JSON.stringify(["b"]));
    const { result } = renderHook(() => useColumnVisibility(KEY, ["a", "b", "new"]));
    expect(result.current).toEqual(["a", "new"]);
  });

  it("treats corrupt payloads as nothing hidden", () => {
    window.localStorage.setItem(KEY, "{not json");
    const { result } = renderHook(() => useColumnVisibility(KEY, ["a"]));
    expect(result.current).toEqual(["a"]);
  });

  it("setColumnHidden updates live subscribers and storage", () => {
    const first = renderHook(() => useColumnVisibility(KEY, ["a", "b"]));
    act(() => setColumnHidden(KEY, ["a"]));
    expect(first.result.current).toEqual(["b"]);
    expect(window.localStorage.getItem(KEY)).toBe(JSON.stringify(["a"]));
  });

  it("undefined storage key keeps everything visible (settings disabled)", () => {
    const { result } = renderHook(() => useColumnVisibility(undefined, ["a"]));
    expect(result.current).toEqual(["a"]);
  });
});
