import fc from "fast-check";
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

// The writable domain is every JSON value except objects carrying a
// `__proto__` own key: superjson refuses those at serialization time (its
// prototype-pollution guard throws), so the jar cannot carry them and the
// identity is promised on the rest.
const hasProtoKey = (v: unknown): boolean => {
  if (Array.isArray(v)) return v.some(hasProtoKey);
  if (v && typeof v === "object")
    return Object.entries(v).some(([k, x]) => k === "__proto__" || hasProtoKey(x));
  return false;
};

// Property face (fc): the round-trip is an identity for every JSON value the
// jar can carry — the example above pins the shape once, this pins it for
// the whole domain; and every failure input answers the fallback, never a
// throw, never a partial parse.
describe("cookie properties", () => {
  it("write→read is the identity for any JSON value", () => {
    fc.assert(
      fc.property(
        fc.jsonValue().filter((v) => !hasProtoKey(v)),
        (value) => {
          writeCookie("prop", value);
          expect(readCookie(rawOf("prop"), undefined, null)).toEqual(value);
        },
      ),
    );
  });

  it("never throws on any raw input", () => {
    fc.assert(
      fc.property(fc.string(), (raw) => {
        expect(() => readCookie(raw, undefined, { safe: true })).not.toThrow();
      }),
    );
  });
});
