import { describe, expect, it } from "vitest";
import { z } from "zod";
import { makeSearchParser } from "./parse";
import { pageNoField, pageSizeField, sortField } from "./schema";

const schema = z.looseObject({
  keyword: z.string().optional(),
  createdAtA: z.number().optional(),
  createdAtZ: z.number().optional(),
  lens: z.enum(["all", "denied"]).default("all"),
  pageNo: pageNoField(),
  pageSize: pageSizeField(20),
  sort: sortField("createdAt:desc"),
});

const parseSearch = makeSearchParser(schema, ["createdAtA", "createdAtZ"]);

describe("makeSearchParser", () => {
  it("coerces string page params and extra number keys from URL input", () => {
    const parsed = parseSearch({ pageNo: "3", pageSize: "50", createdAtA: "1727000000000" });
    expect(parsed.pageNo).toBe(3);
    expect(parsed.pageSize).toBe(50);
    expect(parsed.createdAtA).toBe(1727000000000);
  });

  it("applies schema defaults to missing keys and leaves strings untouched", () => {
    const parsed = parseSearch({ keyword: "abc" });
    expect(parsed.lens).toBe("all");
    expect(parsed.sort).toBe("createdAt:desc");
    expect(parsed.pageNo).toBe(1);
    expect(parsed.keyword).toBe("abc");
  });

  it("keeps empty strings un-coerced so validation rejects them loudly", () => {
    expect(() => parseSearch({ createdAtA: "" })).toThrow();
  });
});
