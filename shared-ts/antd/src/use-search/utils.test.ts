import { describe, expect, it } from "vitest";
import { buildSearchPatch } from "./utils";

describe("buildSearchPatch", () => {
  it("merges the patch over prev", () => {
    expect(buildSearchPatch({ a: 1, b: 2 }, { b: 3 })).toEqual({ a: 1, b: 3 });
  });

  it("deletes keys whose patch value is undefined", () => {
    expect(buildSearchPatch({ a: 1, b: 2 }, { b: undefined })).toEqual({ a: 1 });
  });

  it("keeps null values (explicit cleared state) in the URL", () => {
    const prev: Record<string, unknown> = { a: 1 };
    expect(buildSearchPatch(prev, { b: null })).toEqual({ a: 1, b: null });
  });

  it("does not mutate prev", () => {
    const prev = { a: 1 };
    buildSearchPatch(prev, { a: 2 });
    expect(prev).toEqual({ a: 1 });
  });
});
