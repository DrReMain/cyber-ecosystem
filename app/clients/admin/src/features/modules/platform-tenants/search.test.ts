import fc from "fast-check";
import { describe, expect, it } from "vitest";

import { parseTenantsSearch, tenantsSearchSchema } from "./search";

const legal = fc.record({
  keyword: fc.option(fc.string(), { nil: undefined }),
  status: fc.option(fc.constantFrom("on", "off"), { nil: undefined }),
});

describe("tenantsSearch", () => {
  it("the arbitrary mirrors the schema's declared keys", () => {
    expect(Object.keys(tenantsSearchSchema.shape).sort()).toEqual(["keyword", "status"]);
  });

  it("parses any schema-legal value to itself", () => {
    fc.assert(
      fc.property(legal, (value) => {
        expect(parseTenantsSearch(value)).toEqual(value);
      }),
    );
  });

  it("rejects values outside the status enum", () => {
    expect(() => parseTenantsSearch({ status: "weird" })).toThrow();
  });
});
