import { Code, ConnectError } from "@connectrpc/connect";
import { ErrorInfoSchema } from "@cyber-ecosystem/gen-connect-ts/google/rpc/error_details_pb";
import { describe, expect, it } from "vitest";
import { connectSource } from "./connect-source";

const withReason = (code: Code, reason: string, message = "machine message"): ConnectError =>
  new ConnectError(message, code, undefined, [{ desc: ErrorInfoSchema, value: { reason } }]);

describe("connectSource recognition", () => {
  it("ignores non-ConnectError shapes", () => {
    expect(connectSource.recognize(new Error("plain"))).toBeUndefined();
    expect(connectSource.recognize("string")).toBeUndefined();
  });

  it("pins the load-bearing gRPC code mappings", () => {
    const cases: ReadonlyArray<[Code, string, number]> = [
      [Code.Unauthenticated, "auth", 401],
      [Code.PermissionDenied, "forbidden", 403],
      [Code.Canceled, "canceled", 499],
      [Code.InvalidArgument, "business", 400],
      [Code.Internal, "server", 500],
      [Code.Unavailable, "network", 503],
    ];
    for (const [code, kind, http] of cases) {
      const envelope = connectSource.recognize(new ConnectError("m", code));
      expect(envelope).toMatchObject({ kind, http, source: "connect" });
    }
  });

  it("extracts ErrorInfo reason and the wire code alongside the raw message", () => {
    const envelope = connectSource.recognize(
      withReason(Code.InvalidArgument, "SYSTEM_LOGIN_FAILED"),
    );
    expect(envelope).toMatchObject({
      kind: "business",
      reason: "SYSTEM_LOGIN_FAILED",
      code: "invalid_argument",
      message: "machine message",
    });
  });

  it("falls back to the reason string when rawMessage is empty", () => {
    const envelope = connectSource.recognize(
      withReason(Code.InvalidArgument, "SYSTEM_LOGIN_FAILED", ""),
    );
    expect(envelope?.message).toBe("SYSTEM_LOGIN_FAILED");
  });

  it("treats Unknown carried on a fetch TypeError as a network outcome, not a server fault", () => {
    const fetchFailure = new ConnectError(
      "failed to fetch",
      Code.Unknown,
      undefined,
      undefined,
      new TypeError("fetch failed"),
    );
    // The network special case carries no http override; the view layer's
    // word-table default (502/503 family) applies.
    expect(connectSource.recognize(fetchFailure)).toMatchObject({
      kind: "network",
      http: undefined,
    });
  });

  it("keeps plain Unknown as a server fault", () => {
    const envelope = connectSource.recognize(new ConnectError("boom", Code.Unknown));
    expect(envelope?.kind).toBe("server");
  });

  it("carries the raw stack and a cause frame", () => {
    const envelope = connectSource.recognize(
      new ConnectError("outer", Code.Internal, undefined, undefined, new TypeError("inner")),
    );
    expect(envelope?.cause?.message).toBe("inner");
  });
});
