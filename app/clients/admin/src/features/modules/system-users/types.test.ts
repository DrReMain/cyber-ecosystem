import { describe, expect, it } from "vitest";

import { buildOperationCatalog } from "./types";

describe("buildOperationCatalog", () => {
  it("indexes by fullName, skips nameless services, and defaults missing copy", () => {
    const catalog = buildOperationCatalog([
      {
        fullName: "system.System",
        comment: "IAM base",
        methods: [{ name: "GetUser", comment: "read" }, { name: "ListUsers" }, { name: undefined }],
      },
      { comment: "orphan service without a name" },
    ]);
    expect(catalog.size).toBe(1);
    expect(catalog.get("system.System")).toEqual({
      comment: "IAM base",
      methods: new Map([
        ["GetUser", "read"],
        ["ListUsers", ""],
      ]),
    });
  });

  it("answers an empty catalog for an empty face", () => {
    expect(buildOperationCatalog([]).size).toBe(0);
  });
});
