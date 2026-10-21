import { System } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/error_reason_pb";
import { describe, expect, it } from "vitest";
import { systemReason } from "./system-reason";

describe("systemReason", () => {
  it("decodes a prefixed wire reason onto the typed enum", () => {
    expect(systemReason("SYSTEM_LOGIN_SEAT_REQUIRED")).toBe(System.LOGIN_SEAT_REQUIRED);
    expect(systemReason("SYSTEM_USER_DISABLED")).toBe(System.USER_DISABLED);
  });

  it("returns undefined for foreign prefixes and unknown names", () => {
    expect(systemReason("GENERAL_ERROR_UNSPECIFIED")).toBeUndefined();
    expect(systemReason("SYSTEM_NO_SUCH_REASON")).toBeUndefined();
    expect(systemReason(undefined)).toBeUndefined();
  });
});
