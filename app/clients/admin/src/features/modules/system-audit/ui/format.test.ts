import { describe, expect, it } from "vitest";

import { dash } from "./format";

describe("dash", () => {
  it("renders absence as a dash, presence verbatim", () => {
    expect(dash(undefined)).toBe("-");
    expect(dash("")).toBe("-");
    expect(dash("0")).toBe("0");
    expect(dash("alice")).toBe("alice");
  });
});
