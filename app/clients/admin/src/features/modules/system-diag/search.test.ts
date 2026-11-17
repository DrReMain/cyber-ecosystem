import fc from "fast-check";
import { describe, expect, it } from "vitest";

import { diagSearchSchema } from "./search";

// Schema-only face (no parser export — the consumer parses): the shape
// anchored directly.
const legal = fc.record({
  userId: fc.option(fc.string(), { nil: undefined }),
});

describe("diagSearch", () => {
  it("the arbitrary mirrors the schema's declared keys", () => {
    expect(Object.keys(diagSearchSchema.shape)).toEqual(["userId"]);
  });

  it("admits any schema-legal value", () => {
    fc.assert(
      fc.property(legal, (value) => {
        expect(diagSearchSchema.parse(value)).toEqual(value);
      }),
    );
  });
});
