import fc from "fast-check";
import { describe, expect, it } from "vitest";

import { auditSearchSchema, parseAuditSearch } from "./search";

// The richest search face: numeric wire fields the router delivers as
// strings (status, createdAtA/Z, page keys — all coerced), a lens enum with
// a default, and free-text filters. The identity property runs over
// explicit values; the defaults and coercion rows pin the transform.
const legal = fc.record({
  actor: fc.option(fc.string(), { nil: undefined }),
  operation: fc.option(fc.string(), { nil: undefined }),
  status: fc.option(fc.integer(), { nil: undefined }),
  denyReason: fc.option(fc.string(), { nil: undefined }),
  lens: fc.constantFrom("all", "denied"),
  createdAtA: fc.option(fc.integer(), { nil: undefined }),
  createdAtZ: fc.option(fc.integer(), { nil: undefined }),
  pageNo: fc.integer({ min: 1 }),
  pageSize: fc.integer(),
  sort: fc.string(),
});

describe("auditSearch", () => {
  it("the arbitrary mirrors the schema's declared keys", () => {
    expect(Object.keys(auditSearchSchema.shape).sort()).toEqual([
      "actor",
      "createdAtA",
      "createdAtZ",
      "denyReason",
      "lens",
      "operation",
      "pageNo",
      "pageSize",
      "sort",
      "status",
    ]);
  });

  it("parses any explicit legal value to itself", () => {
    fc.assert(
      fc.property(legal, (value) => {
        expect(parseAuditSearch(value)).toEqual(value);
      }),
    );
  });

  it("fills defaults on absence", () => {
    expect(parseAuditSearch({})).toEqual({
      lens: "all",
      pageNo: 1,
      pageSize: 20,
      sort: "createdAt:desc",
    });
  });

  it("coerces every router string digit field to numbers", () => {
    expect(
      parseAuditSearch({
        status: "403",
        createdAtA: "100",
        createdAtZ: "200",
        pageNo: "2",
        pageSize: "50",
      }),
    ).toEqual({
      lens: "all",
      status: 403,
      createdAtA: 100,
      createdAtZ: 200,
      pageNo: 2,
      pageSize: 50,
      sort: "createdAt:desc",
    });
  });

  it("rejects values outside the lens enum", () => {
    expect(() => parseAuditSearch({ lens: "weird" })).toThrow();
  });
});
