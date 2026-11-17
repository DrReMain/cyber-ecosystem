import fc from "fast-check";
import { describe, expect, it } from "vitest";

import { parseUsersSearch, usersSearchSchema } from "./search";

// The search face's contract, property-first: every generated legal value
// parses to itself, an enum face rejects outside values, and the handwritten
// arbitrary never drifts from the schema — a new schema key without a test
// key turns the mirror assertion red, an arbitrary wider than the schema
// turns the identity property red (parse throws). Two sources, both anchored.
const legal = fc.record({
  keyword: fc.option(fc.string(), { nil: undefined }),
  status: fc.option(fc.constantFrom("on", "off"), { nil: undefined }),
  deptId: fc.option(fc.string(), { nil: undefined }),
});

describe("usersSearch", () => {
  it("the arbitrary mirrors the schema's declared keys", () => {
    expect(Object.keys(usersSearchSchema.shape).sort()).toEqual(["deptId", "keyword", "status"]);
  });

  it("parses any schema-legal value to itself", () => {
    fc.assert(
      fc.property(legal, (value) => {
        expect(parseUsersSearch(value)).toEqual(value);
      }),
    );
  });

  it("rejects values outside the status enum", () => {
    expect(() => parseUsersSearch({ status: "weird" })).toThrow();
  });
});
