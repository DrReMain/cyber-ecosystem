import { describe, expect, it, vi } from "vitest";
import {
  createErrorDomain,
  type ErrorView,
  isEnvelope,
  KINDS,
  makeEnvelope,
  resolveKind,
  type SourceAdapter,
} from "./index";

const named = (name: string, message = "boom"): Error => {
  const error = new Error(message);
  error.name = name;
  return error;
};

describe("word table", () => {
  it("resolves known words and falls back to unknown", () => {
    expect(resolveKind("business")).toBe("business");
    expect(resolveKind("nope")).toBe("unknown");
  });

  it("marks canceled as the only silent word", () => {
    const silent = Object.keys(KINDS).filter((kind) => KINDS[kind as keyof typeof KINDS].silent);
    expect(silent).toEqual(["canceled"]);
  });
});

describe("makeEnvelope", () => {
  it("defaults the source to local and brands the envelope", () => {
    const envelope = makeEnvelope("server", "down");
    expect(envelope.source).toBe("local");
    expect(isEnvelope(envelope)).toBe(true);
    expect(isEnvelope(new Error("nope"))).toBe(false);
  });

  it("keeps one diagnostic frame from an Error cause and from a string cause", () => {
    const inner = new Error("inner");
    const fromError = makeEnvelope("server", "outer", { cause: inner });
    expect(fromError.cause?.name).toBe("Error");
    expect(fromError.cause?.message).toBe("inner");
    const fromString = makeEnvelope("server", "outer", { cause: "raw text" });
    expect(fromString.cause).toEqual({ message: "raw text" });
    expect(makeEnvelope("server", "outer").cause).toBeUndefined();
  });

  it("defends a corrupted kind string", () => {
    expect(makeEnvelope("weird" as "server", "x").kind).toBe("unknown");
  });
});

describe("normalize", () => {
  const fakeSource: SourceAdapter = {
    source: "fake",
    recognize: (errorLike) =>
      errorLike === "fake-token"
        ? makeEnvelope("network", "fake says network", { source: "fake" })
        : undefined,
  };
  const domain = createErrorDomain({ sources: [fakeSource] });

  it("tries sources in order, then falls back to local shapes", () => {
    expect(domain.normalize("fake-token").source).toBe("fake");
    expect(domain.normalize(named("AbortError")).kind).toBe("canceled");
    expect(domain.normalize(named("TimeoutError")).kind).toBe("network");
    expect(domain.normalize(new Error("plain")).kind).toBe("unknown");
    expect(domain.normalize("thrown string").message).toBe("thrown string");
  });

  it("passes envelopes through, repairing a wire-corrupted kind", () => {
    const envelope = makeEnvelope("auth", "expired");
    expect(domain.normalize(envelope)).toMatchObject({ brand: "error-envelope", kind: "auth" });
    const corrupted = { ...envelope, kind: "hacked" };
    expect(domain.normalize(corrupted).kind).toBe("unknown");
  });

  it("never throws on hostile shapes", () => {
    const hostile = {
      get brand(): string {
        throw new Error("getter boom");
      },
    };
    const stringy = {
      toString(): string {
        throw new Error("toString boom");
      },
    };
    expect(() => domain.capture(hostile)).not.toThrow();
    expect(domain.normalize(hostile)).toMatchObject({ kind: "unknown", message: "unclassifiable" });
    expect(domain.normalize(stringy).kind).toBe("unknown");
  });
});

describe("capture policy ladder", () => {
  it("dedups by object identity across sinks", () => {
    const feedback = vi.fn();
    const reporter = vi.fn();
    const domain = createErrorDomain({ feedback, reporters: [reporter] });
    const error = new Error("once");
    domain.capture(error);
    domain.capture(error);
    expect(feedback).toHaveBeenCalledTimes(1);
    expect(reporter).toHaveBeenCalledTimes(1);
  });

  it("honors feedback:false and report:false overrides", () => {
    const feedback = vi.fn();
    const reporter = vi.fn();
    const domain = createErrorDomain({ feedback, reporters: [reporter] });
    domain.capture(new Error("a"), { feedback: false });
    domain.capture(new Error("b"), { report: false });
    expect(feedback).toHaveBeenCalledTimes(1);
    expect(reporter).toHaveBeenCalledTimes(1);
  });

  it("never feeds silent words back but still reports them", () => {
    const feedback = vi.fn();
    const reporter = vi.fn();
    const domain = createErrorDomain({ feedback, reporters: [reporter] });
    domain.capture(named("AbortError"), { feedback: true });
    expect(feedback).not.toHaveBeenCalled();
    expect(reporter).toHaveBeenCalledTimes(1);
  });

  it("fans out to every reporter and swallows sink throws", () => {
    const feedback = vi.fn(() => {
      throw new Error("sink blowup");
    });
    const exploding = vi.fn(() => {
      throw new Error("reporter blowup");
    });
    const sane = vi.fn();
    const domain = createErrorDomain({ feedback, reporters: [exploding, sane] });
    expect(() => domain.capture(new Error("ok"))).not.toThrow();
    expect(sane).toHaveBeenCalledTimes(1);
  });

  it("is safe with no sinks at all", () => {
    const domain = createErrorDomain();
    expect(domain.capture(new Error("bare")).kind).toBe("unknown");
  });
});

describe("view", () => {
  const copy = {
    title: (kind: string) => `T:${kind}`,
    message: (envelope: { message: string }) => `M:${envelope.message}`,
  };
  const domain = createErrorDomain({ copy });

  const expectView = (errorLike: unknown): ErrorView => domain.view(errorLike);

  it("derives http from the word table and honors adapter overrides", () => {
    expect(expectView(new Error("x")).http).toBe(500);
    expect(expectView(makeEnvelope("server", "x", { http: 503 })).http).toBe(503);
  });

  it("resolves copy only through the injected resolver", () => {
    const view = expectView(makeEnvelope("business", "raw"));
    expect(view).toMatchObject({ title: "T:business", message: "M:raw", silent: false });
  });

  it("exposes the silent bit renderers branch on", () => {
    expect(expectView(named("AbortError")).silent).toBe(true);
  });

  it("passes raw messages through when no copy is injected", () => {
    expect(createErrorDomain().view(new Error("plain")).title).toBe("unknown");
  });
});
