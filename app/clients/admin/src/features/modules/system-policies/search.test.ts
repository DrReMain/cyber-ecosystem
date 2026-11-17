import fc from "fast-check";
import { describe, expect, it } from "vitest";

import { parsePoliciesSearch, policiesSearchSchema } from "./search";

const legal = fc.record({
  keyword: fc.option(fc.string(), { nil: undefined }),
  kind: fc.option(fc.constantFrom("time_window", "calendar"), { nil: undefined }),
});

describe("policiesSearch", () => {
  it("the arbitrary mirrors the schema's declared keys", () => {
    expect(Object.keys(policiesSearchSchema.shape).sort()).toEqual(["keyword", "kind"]);
  });

  it("parses any schema-legal value to itself", () => {
    fc.assert(
      fc.property(legal, (value) => {
        expect(parsePoliciesSearch(value)).toEqual(value);
      }),
    );
  });

  it("rejects values outside the kind enum", () => {
    expect(() => parsePoliciesSearch({ kind: "weird" })).toThrow();
  });
});
