import fc from "fast-check";
import { describe, expect, it } from "vitest";

import { agentsChatSearchSchema } from "./search";

// Same contract as the parser-backed faces, anchored on the schema directly
// (this face exports no parser — its consumer parses). The mirror assertion
// and the identity property together pin the shape to one source.
const legal = fc.record({
  session: fc.option(fc.string({ minLength: 1 }), { nil: undefined }),
});

describe("agentsChatSearch", () => {
  it("the arbitrary mirrors the schema's declared keys", () => {
    expect(Object.keys(agentsChatSearchSchema.shape)).toEqual(["session"]);
  });

  it("admits any non-empty session value", () => {
    fc.assert(
      fc.property(legal, (value) => {
        expect(agentsChatSearchSchema.parse(value)).toEqual(value);
      }),
    );
  });

  it("rejects an empty session string", () => {
    expect(() => agentsChatSearchSchema.parse({ session: "" })).toThrow();
  });
});
