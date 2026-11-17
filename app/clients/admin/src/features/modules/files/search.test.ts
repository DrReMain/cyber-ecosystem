import fc from "fast-check";
import { describe, expect, it } from "vitest";

import { filesSearchSchema, parseFilesSearch } from "./search";

// The files face: two enum faces (status, source), free-text filters, page
// defaults, and the sort default — each anchored once, the whole domain by
// the identity property.
const legal = fc.record({
  name: fc.option(fc.string(), { nil: undefined }),
  contentType: fc.option(fc.string(), { nil: undefined }),
  status: fc.option(fc.constantFrom("UPLOADING", "CONFIRMED", "PROCESSING", "FAILED"), {
    nil: undefined,
  }),
  source: fc.option(fc.constantFrom("CLIENT_UPLOAD", "SERVER_GENERATED"), { nil: undefined }),
  pageNo: fc.integer({ min: 1 }),
  pageSize: fc.integer(),
  sort: fc.string(),
});

describe("filesSearch", () => {
  it("the arbitrary mirrors the schema's declared keys", () => {
    expect(Object.keys(filesSearchSchema.shape).sort()).toEqual([
      "contentType",
      "name",
      "pageNo",
      "pageSize",
      "sort",
      "source",
      "status",
    ]);
  });

  it("parses any explicit legal value to itself", () => {
    fc.assert(
      fc.property(legal, (value) => {
        expect(parseFilesSearch(value)).toEqual(value);
      }),
    );
  });

  it("fills page and sort defaults on absence", () => {
    expect(parseFilesSearch({})).toEqual({ pageNo: 1, pageSize: 20, sort: "createdAt:desc" });
  });

  it("rejects values outside the status enum", () => {
    expect(() => parseFilesSearch({ status: "weird" })).toThrow();
  });
});
