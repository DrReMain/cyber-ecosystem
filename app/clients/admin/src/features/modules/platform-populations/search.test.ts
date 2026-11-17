import fc from "fast-check";
import { describe, expect, it } from "vitest";

import { parsePopulationsSearch, populationsSearchSchema } from "./search";

const legal = fc.record({
  keyword: fc.option(fc.string(), { nil: undefined }),
});

describe("populationsSearch", () => {
  it("the arbitrary mirrors the schema's declared keys", () => {
    expect(Object.keys(populationsSearchSchema.shape)).toEqual(["keyword"]);
  });

  it("parses any schema-legal value to itself", () => {
    fc.assert(
      fc.property(legal, (value) => {
        expect(parsePopulationsSearch(value)).toEqual(value);
      }),
    );
  });

  it("parses absence to an empty object", () => {
    expect(parsePopulationsSearch({})).toEqual({});
  });
});
