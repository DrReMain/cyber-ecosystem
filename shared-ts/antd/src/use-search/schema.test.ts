import { describe, expect, it } from "vitest";
import { z } from "zod";
import { getPartition, pageNoField, pageSizeField, sortField } from "./schema";

describe("search schema fields", () => {
  it("pageNoField defaults and coerces to a positive int", () => {
    const parsed = z.object({ pageNo: pageNoField() }).parse({});
    expect(parsed.pageNo).toBe(1);
    expect(() => z.object({ pageNo: pageNoField() }).parse({ pageNo: 0 })).toThrow();
  });

  it("pageSizeField carries its default", () => {
    expect(z.object({ pageSize: pageSizeField(20) }).parse({}).pageSize).toBe(20);
  });

  it("sortField is optional without default and set with one", () => {
    expect(z.object({ sort: sortField() }).parse({}).sort).toBeUndefined();
    expect(z.object({ sort: sortField("name:asc") }).parse({}).sort).toBe("name:asc");
  });
});

describe("getPartition", () => {
  it("identifies the reserved keys and excludes them from filter values", () => {
    const schema = z.object({
      pageNo: pageNoField(),
      pageSize: pageSizeField(),
      sort: sortField(),
      keyword: z.string().optional(),
    });
    const p = getPartition(schema);
    expect(p.pageNoKey).toBe("pageNo");
    expect(p.pageSizeKey).toBe("pageSize");
    expect(p.sortKey).toBe("sort");
    expect(p.nonFilterKeys.has("keyword")).toBe(false);
  });

  it("returns the empty partition for shapeless schemas and caches", () => {
    const shapeless = {} as { shape: Record<string, unknown> };
    const p1 = getPartition(shapeless);
    const p2 = getPartition(shapeless);
    expect(p1.pageNoKey).toBeUndefined();
    expect(p1).toBe(p2); // WeakMap cache: same object back
  });

  it("carries the schema defaults in the partition", () => {
    const schema = z.object({
      pageNo: pageNoField(),
      pageSize: pageSizeField(20),
      sort: sortField("name:asc"),
      keyword: z.string().optional(),
    });
    const p = getPartition(schema);
    expect(p.defaults).toEqual({ pageNo: 1, pageSize: 20, sort: "name:asc" });
  });

  it("leaves defaults empty when the schema has no safeParse", () => {
    const shapeOnly = {
      shape: { pageNo: pageNoField() },
    } as { shape: Record<string, unknown> };
    expect(getPartition(shapeOnly).defaults).toEqual({});
  });
});
