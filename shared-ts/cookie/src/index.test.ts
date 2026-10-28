import { describe, expect, it, vi } from "vitest";
import { z } from "zod";
import { readCookie, writeCookie } from "./index";

// The hydration gate's laws: a written shape survives the full cookie
// round-trip, every failure path answers the fallback, and the fallback is
// always a clone — a mutated default must never leak into the next read.

const rawOf = (key: string) => {
  const pair = document.cookie.split("; ").find((c) => c.startsWith(`${key}=`));
  return pair ? decodeURIComponent(pair.slice(key.length + 1)) : undefined;
};

describe("readCookie", () => {
  const schema = z.object({ n: z.number() });
  const fallback = { n: 0 };

  it("round-trips a written value through the real cookie jar", () => {
    writeCookie("rt", { n: 1 });
    expect(readCookie(rawOf("rt"), schema, fallback)).toEqual({ n: 1 });
  });

  it("answers the fallback on absence, garbage, and schema mismatch", () => {
    expect(readCookie(undefined, schema, fallback)).toEqual(fallback);
    expect(readCookie("garbage{", schema, fallback)).toEqual(fallback);
    writeCookie("mm", "a-string");
    expect(readCookie(rawOf("mm"), schema, fallback)).toEqual(fallback);
  });

  it("passes the parsed value through when no schema narrows it", () => {
    writeCookie("ns", { n: "not-a-number" });
    expect(readCookie(rawOf("ns"), undefined, fallback)).toEqual({ n: "not-a-number" });
  });

  it("clones the fallback so mutations never leak across reads", () => {
    const shared = { n: 0 };
    const got = readCookie(undefined, undefined, shared);
    got.n = 42;
    expect(readCookie(undefined, undefined, shared)).toEqual({ n: 0 });
  });
});

describe("writeCookie", () => {
  it("writes the hydration attributes", () => {
    const spy = vi.spyOn(document, "cookie", "set");
    writeCookie("attr", { a: 1 }, { maxAge: 60, path: "/x", sameSite: "strict" });
    const written = spy.mock.calls[0]?.[0] ?? "";
    expect(written).toContain("path=/x");
    expect(written).toContain("max-age=60");
    expect(written).toContain("samesite=strict");
    spy.mockRestore();
  });

  it("defaults to the year-long lax cookie", () => {
    const spy = vi.spyOn(document, "cookie", "set");
    writeCookie("def", 1);
    const written = spy.mock.calls[0]?.[0] ?? "";
    expect(written).toContain("path=/");
    expect(written).toContain("max-age=31536000");
    expect(written).toContain("samesite=lax");
    spy.mockRestore();
  });
});
