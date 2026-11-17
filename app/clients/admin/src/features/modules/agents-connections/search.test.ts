import fc from "fast-check";
import { describe, expect, it } from "vitest";

import { connectionsSearchSchema, parseConnectionsSearch } from "./search";

// The paginated face's contract: explicit values parse to themselves, absent
// page keys fill the defaults, the router's string digits coerce to numbers,
// and the mirror assertion keeps the handwritten arbitrary honest.
const legal = fc.record({
  pageNo: fc.integer({ min: 1 }),
  pageSize: fc.integer(),
});

describe("connectionsSearch", () => {
  it("the arbitrary mirrors the schema's declared keys", () => {
    expect(Object.keys(connectionsSearchSchema.shape).sort()).toEqual(["pageNo", "pageSize"]);
  });

  it("parses any explicit legal value to itself", () => {
    fc.assert(
      fc.property(legal, (value) => {
        expect(parseConnectionsSearch(value)).toEqual(value);
      }),
    );
  });

  it("fills page defaults on absence", () => {
    expect(parseConnectionsSearch({})).toEqual({ pageNo: 1, pageSize: 20 });
  });

  it("coerces router string digits to numbers", () => {
    expect(parseConnectionsSearch({ pageNo: "3", pageSize: "50" })).toEqual({
      pageNo: 3,
      pageSize: 50,
    });
  });
});
