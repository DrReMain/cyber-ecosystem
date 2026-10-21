import type { DescMethod } from "@bufbuild/protobuf";
import { create } from "@bufbuild/protobuf";
import { TimestampSchema } from "@bufbuild/protobuf/wkt";
import type { Interceptor } from "@connectrpc/connect";
import {
  PageRequestSchema,
  PageResponseSchema,
} from "@cyber-ecosystem/gen-connect-ts/cyber/shared/common/v1/pagination_pb";
import { UserSchema } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/user_pb";
import { beforeEach, describe, expect, it, vi } from "vitest";

const { unaryImpl, streamImpl, wsImpl, serverCookie, branch } = vi.hoisted(() => ({
  unaryImpl: vi.fn(),
  streamImpl: vi.fn(),
  wsImpl: vi.fn(),
  serverCookie: vi.fn((): string | undefined => undefined),
  branch: { mode: "client" as "client" | "server" },
}));

// The react-start isomorphic helper is mocked with a switchable branch so
// the cookie forwarder's server leg is exercised for real: the lazy
// "@tanstack/react-start/server" import resolves to the mock below, whose
// getRequestHeader returns the controllable cookie fixture.
vi.mock("@tanstack/react-start", () => {
  const createIsomorphicFn = () => {
    const impls: Record<string, () => unknown> = {};
    const f = (() =>
      branch.mode === "server" ? impls.server?.() : impls.client?.()) as (() => unknown) & {
      server: (impl: () => unknown) => unknown;
      client: (impl: () => unknown) => unknown;
    };
    f.server = (impl: () => unknown) => {
      impls.server = impl;
      return f;
    };
    f.client = (impl: () => unknown) => {
      impls.client = impl;
      return f;
    };
    return f;
  };
  return { createIsomorphicFn };
});

vi.mock("@tanstack/react-start/server", () => ({
  getRequestHeader: (name: string) => (name === "cookie" ? serverCookie() : undefined),
}));

vi.mock("@connectrpc/connect-web", () => ({
  createConnectTransport: vi.fn(() => ({
    unary: (...args: unknown[]) => unaryImpl(...args),
    stream: (...args: unknown[]) => streamImpl(...args),
  })),
}));

vi.mock("./ws-stream", () => ({
  wsStreamCall: (...args: unknown[]) => wsImpl(...args),
}));

// Real generated schemas as fixtures: protobuf v2 has no public runtime
// type-construction API, and a dev-only gen dependency keeps the test
// fixtures honest about production shapes. UserSchema carries the WKT
// pair the flattening law must pin: StringValue fields and a Timestamp.
const Input = PageRequestSchema;
const Output = PageResponseSchema;

const service = { typeName: "test.v1.EchoService" };
const unaryMethod = {
  methodKind: "unary",
  name: "Echo",
  parent: service,
  input: Input,
  output: Output,
} as unknown as DescMethod & { methodKind: "unary" };

const streamMethod = {
  methodKind: "server_streaming",
  name: "EchoStream",
  parent: service,
  input: Input,
  output: Output,
} as unknown as DescMethod & { methodKind: "server_streaming" };

const bidiMethod = {
  methodKind: "bidi_streaming",
  name: "EchoBidi",
  parent: service,
  input: Input,
  output: Output,
} as unknown as DescMethod & { methodKind: "bidi_streaming" };

const wktMethod = {
  methodKind: "unary",
  name: "GetUser",
  parent: service,
  input: Input,
  output: UserSchema,
} as unknown as DescMethod & { methodKind: "unary" };

import { createConnectTransport } from "@connectrpc/connect-web";
import {
  createTransportFactory,
  SESSION_NAMESPACE_HEADER,
  sessionCookieNameFor,
} from "./transport";

const normalize = vi.fn((error: unknown) => new Error(`normalized:${String(error)}`));

const bindings = {
  getLocale: () => "zh-CN",
  getSiteUrl: async () => originHolder.origin,
  proxyPath: "/connect",
  getSessionNamespace: () => "member",
  normalize,
};

const originHolder = { origin: "http://origin" };

function makeFactory() {
  return createTransportFactory(bindings);
}

type HeaderReq = { header: Headers };

function lastInterceptors(): Interceptor[] {
  const calls = vi.mocked(createConnectTransport).mock.calls;
  const options = calls.at(-1)?.[0];
  return options?.interceptors ?? [];
}

async function runInterceptors(req: HeaderReq): Promise<HeaderReq> {
  const finalNext = async (r: HeaderReq) => r;
  const chain = lastInterceptors().reduceRight(
    (next: (r: HeaderReq) => Promise<HeaderReq>, interceptor: Interceptor) =>
      async (r: HeaderReq) => {
        await interceptor(next as never)(r as never);
        return r;
      },
    finalNext,
  );
  return await chain(req);
}

function okUnary() {
  unaryImpl.mockResolvedValue({ header: new Headers(), message: create(Output, {}) });
}

beforeEach(() => {
  vi.clearAllMocks();
  branch.mode = "client";
  originHolder.origin = "http://origin";
});

describe("createTransportFactory", () => {
  it("stamps the Accept-Language header from the locale binding", async () => {
    const transport = makeFactory().createTransport({ baseUrl: "http://api" });
    okUnary();
    await transport.unary(unaryMethod, undefined, undefined, new Headers(), {}, undefined);
    const req = { header: new Headers() };
    await runInterceptors(req);
    expect(req.header.get("Accept-Language")).toBe("zh-CN");
  });

  it("stamps the session-namespace header from the binding, or omits it when empty", async () => {
    const transport = makeFactory().createTransport({ baseUrl: "http://api" });
    okUnary();
    await transport.unary(unaryMethod, undefined, undefined, new Headers(), {}, undefined);
    const req = { header: new Headers() };
    await runInterceptors(req);
    expect(req.header.get(SESSION_NAMESPACE_HEADER)).toBe("member");
    const bare = createTransportFactory({
      getLocale: () => "zh-CN",
      getSiteUrl: () => "http://sync-origin",
      proxyPath: "/connect",
      getSessionNamespace: () => "",
      normalize,
    });
    bare.createTransport({ baseUrl: "http://api" });
    const bareReq = { header: new Headers() };
    await runInterceptors(bareReq);
    expect(bareReq.header.has(SESSION_NAMESPACE_HEADER)).toBe(false);
  });

  it("composes the session cookie name per namespace", () => {
    // The wire rule's per-language anchor — Go's SessionCookieNameFor must
    // compose the identical names; the live login chain is the cross-check.
    expect(sessionCookieNameFor("")).toBe("session_token");
    expect(sessionCookieNameFor("member")).toBe("session_token_member");
    expect(sessionCookieNameFor("user")).toBe("session_token_user");
  });

  it("forwards an explicit baseUrl to the underlying connect transport", () => {
    makeFactory().createTransport({ baseUrl: "http://api" });
    expect(vi.mocked(createConnectTransport).mock.calls.at(-1)?.[0].baseUrl).toBe("http://api");
  });

  it("appends caller interceptors after the built-ins", () => {
    const extra: Interceptor = (next) => next;
    makeFactory().createTransport({ baseUrl: "http://api", interceptors: [extra] });
    const interceptors = lastInterceptors();
    expect(interceptors.at(-1)).toBe(extra);
    expect(interceptors.length).toBeGreaterThanOrEqual(3);
  });

  it("runs the locale interceptor before caller interceptors", async () => {
    let seenLocale: string | null = null;
    const extra: Interceptor = (next) => async (req) => {
      seenLocale = req.header.get("Accept-Language");
      return await next(req);
    };
    const transport = makeFactory().createTransport({
      baseUrl: "http://api",
      interceptors: [extra],
    });
    okUnary();
    await transport.unary(unaryMethod, undefined, undefined, new Headers(), {}, undefined);
    await runInterceptors({ header: new Headers() });
    expect(seenLocale).toBe("zh-CN");
  });

  it("exposes a static-keyed default transport for connect-query", () => {
    const factory = makeFactory();
    expect(factory.connectTransport).toBeDefined();
  });

  it("routes calls through the static-keyed default transport", async () => {
    okUnary();
    const res = await makeFactory().connectTransport.unary(
      unaryMethod,
      undefined,
      undefined,
      new Headers(),
      {},
      undefined,
    );
    expect(res.message).toEqual({});
  });

  it("flattens the unary response message to its JSON shape", async () => {
    const transport = makeFactory().createTransport({ baseUrl: "http://api" });
    unaryImpl.mockResolvedValue({ header: new Headers(), message: create(Output, {}) });
    const res = await transport.unary(
      unaryMethod,
      undefined,
      undefined,
      new Headers(),
      {},
      undefined,
    );
    expect(res.message).toEqual({});
  });

  it("normalizes unary failures through the binding", async () => {
    const transport = makeFactory().createTransport({ baseUrl: "http://api" });
    unaryImpl.mockRejectedValue(new Error("boom"));
    await expect(
      transport.unary(unaryMethod, undefined, undefined, new Headers(), {}, undefined),
    ).rejects.toThrow("normalized:Error: boom");
  });

  it("hands the raw error instance to the normalize binding", async () => {
    const transport = makeFactory().createTransport({ baseUrl: "http://api" });
    const raw = new Error("raw-instance");
    unaryImpl.mockRejectedValue(raw);
    await expect(
      transport.unary(unaryMethod, undefined, undefined, new Headers(), {}, undefined),
    ).rejects.toThrow("normalized");
    expect(normalize).toHaveBeenCalledWith(raw);
  });

  it("flattens server-streaming messages and normalizes mid-iteration failures", async () => {
    const transport = makeFactory().createTransport({ baseUrl: "http://api" });
    streamImpl.mockResolvedValue({
      stream: true,
      service,
      method: streamMethod,
      header: new Headers(),
      trailer: new Headers(),
      message: (async function* () {
        yield create(Output, {});
        throw new Error("mid-stream");
      })(),
    });
    const res = await transport.stream(
      streamMethod,
      undefined,
      undefined,
      new Headers(),
      (async function* () {})(),
      undefined,
    );
    const seen: unknown[] = [];
    await expect(async () => {
      for await (const message of res.message) {
        seen.push(message);
      }
    }).rejects.toThrow("normalized:Error: mid-stream");
    expect(seen).toEqual([{}]);
  });

  it("normalizes stream call-time failures", async () => {
    const transport = makeFactory().createTransport({ baseUrl: "http://api" });
    streamImpl.mockRejectedValue(new Error("call-failed"));
    await expect(
      transport.stream(
        streamMethod,
        undefined,
        undefined,
        new Headers(),
        (async function* () {})(),
        undefined,
      ),
    ).rejects.toThrow("normalized:Error: call-failed");
  });

  it("passes the method descriptor through untouched", async () => {
    const transport = makeFactory().createTransport({ baseUrl: "http://api" });
    okUnary();
    await transport.unary(unaryMethod, undefined, undefined, new Headers(), {}, undefined);
    expect(unaryImpl.mock.calls.at(-1)?.[0]).toBe(unaryMethod);
  });

  it("passes signal, timeout, header, input, and context values to the unary leg", async () => {
    const transport = makeFactory().createTransport({ baseUrl: "http://api" });
    okUnary();
    const signal = new AbortController().signal;
    const header = new Headers({ "x-marker": "1" });
    const contextValues = { marker: true } as never;
    await transport.unary(unaryMethod, signal, 1234, header, { page: 1 }, contextValues);
    const args = unaryImpl.mock.calls.at(-1);
    expect(args?.[1]).toBe(signal);
    expect(args?.[2]).toBe(1234);
    expect(args?.[3]).toBe(header);
    expect(args?.[4]).toEqual({ page: 1 });
    expect(args?.[5]).toBe(contextValues);
  });

  it("passes signal, timeout, header, and context values to the stream leg", async () => {
    const transport = makeFactory().createTransport({ baseUrl: "http://api" });
    streamImpl.mockResolvedValue({
      stream: true,
      service,
      method: streamMethod,
      header: new Headers(),
      trailer: new Headers(),
      message: (async function* () {})(),
    });
    const signal = new AbortController().signal;
    const header = new Headers({ "x-marker": "2" });
    const input = (async function* () {
      yield create(Input, {});
    })();
    const contextValues = { marker: true } as never;
    await transport.stream(streamMethod, signal, 4321, header, input, contextValues);
    const args = streamImpl.mock.calls.at(-1);
    expect(args?.[1]).toBe(signal);
    expect(args?.[2]).toBe(4321);
    expect(args?.[3]).toBe(header);
    expect(args?.[4]).toBe(input);
    expect(args?.[5]).toBe(contextValues);
  });
});

describe("transport flattening law (toJson防腐层)", () => {
  it("renders StringValue wrappers as plain strings in unary responses", async () => {
    const transport = makeFactory().createTransport({ baseUrl: "http://api" });
    unaryImpl.mockResolvedValue({
      header: new Headers(),
      message: create(UserSchema, { avatar: "钢枪少年" }),
    });
    const res = await transport.unary(
      wktMethod,
      undefined,
      undefined,
      new Headers(),
      {},
      undefined,
    );
    const flat = res.message as unknown as Record<string, unknown>;
    expect(flat.avatar).toBe("钢枪少年");
  });

  it("renders Timestamp fields as RFC3339 strings in unary responses", async () => {
    const transport = makeFactory().createTransport({ baseUrl: "http://api" });
    unaryImpl.mockResolvedValue({
      header: new Headers(),
      message: create(UserSchema, {
        createdAt: create(TimestampSchema, { seconds: 0n }),
      }),
    });
    const res = await transport.unary(
      wktMethod,
      undefined,
      undefined,
      new Headers(),
      {},
      undefined,
    );
    const flat = res.message as unknown as Record<string, unknown>;
    expect(flat.createdAt).toMatch(/^1970-01-01T00:00:00(?:\.000)?Z$/);
  });

  it("omits unset wrapper fields client-side (accepted WKT rendering split)", async () => {
    // The server renders unset wrappers as null (EmitUnpopulated), the
    // client flatten omits them — the cross-protocol WKT inconsistency is
    // a ruled acceptance; this pins the client half so consumers rely on
    // undefined-as-unset, never on null.
    const transport = makeFactory().createTransport({ baseUrl: "http://api" });
    unaryImpl.mockResolvedValue({
      header: new Headers(),
      message: create(UserSchema, {}),
    });
    const res = await transport.unary(
      wktMethod,
      undefined,
      undefined,
      new Headers(),
      {},
      undefined,
    );
    const flat = res.message as unknown as Record<string, unknown>;
    expect(Object.hasOwn(flat, "avatar")).toBe(false);
  });

  it("flattens each streamed message independently", async () => {
    const transport = makeFactory().createTransport({ baseUrl: "http://api" });
    streamImpl.mockResolvedValue({
      stream: true,
      service,
      method: { ...streamMethod, output: UserSchema },
      header: new Headers(),
      trailer: new Headers(),
      message: (async function* () {
        yield create(UserSchema, { avatar: "甲" });
        yield create(UserSchema, { avatar: "乙" });
      })(),
    });
    const res = await transport.stream(
      { ...streamMethod, output: UserSchema } as unknown as DescMethod & {
        methodKind: "server_streaming";
      },
      undefined,
      undefined,
      new Headers(),
      (async function* () {})(),
      undefined,
    );
    const seen: unknown[] = [];
    for await (const message of res.message) {
      seen.push(message);
    }
    expect(seen.map((m) => (m as Record<string, unknown>).avatar)).toEqual(["甲", "乙"]);
  });
});

describe("stream dialect routing", () => {
  it("routes bidi methods to the ws dialect, bypassing connect-web stream", async () => {
    const transport = makeFactory().createTransport({ baseUrl: "http://api" });
    wsImpl.mockResolvedValue({
      stream: true,
      service,
      method: bidiMethod,
      header: new Headers(),
      trailer: new Headers(),
      message: (async function* () {})(),
    });
    await transport.stream(
      bidiMethod,
      undefined,
      undefined,
      new Headers(),
      (async function* () {})(),
      undefined,
    );
    expect(wsImpl).toHaveBeenCalledTimes(1);
    expect(streamImpl).not.toHaveBeenCalled();
    expect(wsImpl.mock.calls.at(-1)?.[0]).toBe("http://api");
  });

  it("routes server-streaming methods to the connect protocol stream", async () => {
    const transport = makeFactory().createTransport({ baseUrl: "http://api" });
    streamImpl.mockResolvedValue({
      stream: true,
      service,
      method: streamMethod,
      header: new Headers(),
      trailer: new Headers(),
      message: (async function* () {})(),
    });
    await transport.stream(
      streamMethod,
      undefined,
      undefined,
      new Headers(),
      (async function* () {})(),
      undefined,
    );
    expect(streamImpl).toHaveBeenCalledTimes(1);
    expect(wsImpl).not.toHaveBeenCalled();
  });
});

describe("stream response boundaries", () => {
  it("preserves header and trailer references through flattening", async () => {
    const transport = makeFactory().createTransport({ baseUrl: "http://api" });
    const markerHeader = new Headers({ "x-trace": "abc" });
    const markerTrailer = new Headers({ "x-end": "yes" });
    streamImpl.mockResolvedValue({
      stream: true,
      service,
      method: streamMethod,
      header: markerHeader,
      trailer: markerTrailer,
      message: (async function* () {})(),
    });
    const res = await transport.stream(
      streamMethod,
      undefined,
      undefined,
      new Headers(),
      (async function* () {})(),
      undefined,
    );
    expect(res.header).toBe(markerHeader);
    expect(res.trailer).toBe(markerTrailer);
  });

  it("ends cleanly on an empty stream", async () => {
    const transport = makeFactory().createTransport({ baseUrl: "http://api" });
    streamImpl.mockResolvedValue({
      stream: true,
      service,
      method: streamMethod,
      header: new Headers(),
      trailer: new Headers(),
      message: (async function* () {})(),
    });
    const res = await transport.stream(
      streamMethod,
      undefined,
      undefined,
      new Headers(),
      (async function* () {})(),
      undefined,
    );
    const seen: unknown[] = [];
    for await (const message of res.message) {
      seen.push(message);
    }
    expect(seen).toEqual([]);
  });

  it("normalizes a failure thrown before the first message", async () => {
    const transport = makeFactory().createTransport({ baseUrl: "http://api" });
    streamImpl.mockResolvedValue({
      stream: true,
      service,
      method: streamMethod,
      header: new Headers(),
      trailer: new Headers(),
      // biome-ignore lint/correctness/useYield: failing before the first yield is the fixture
      message: (async function* () {
        throw new Error("pre-first");
      })(),
    });
    const res = await transport.stream(
      streamMethod,
      undefined,
      undefined,
      new Headers(),
      (async function* () {})(),
      undefined,
    );
    await expect(async () => {
      for await (const _ of res.message) {
        // drain
      }
    }).rejects.toThrow("normalized:Error: pre-first");
  });
});

describe("cookie forwarding (承重墙)", () => {
  it("forwards the request cookie verbatim on the server branch", async () => {
    branch.mode = "server";
    serverCookie.mockReturnValue("session_token=abc; other=1");
    const transport = makeFactory().createTransport({ baseUrl: "http://api" });
    okUnary();
    await transport.unary(unaryMethod, undefined, undefined, new Headers(), {}, undefined);
    const req = { header: new Headers() };
    await runInterceptors(req);
    expect(req.header.get("cookie")).toBe("session_token=abc; other=1");
  });

  it("leaves the header unset when the server branch has no cookie", async () => {
    branch.mode = "server";
    serverCookie.mockReturnValue(undefined);
    const transport = makeFactory().createTransport({ baseUrl: "http://api" });
    okUnary();
    await transport.unary(unaryMethod, undefined, undefined, new Headers(), {}, undefined);
    const req = { header: new Headers() };
    await runInterceptors(req);
    expect(req.header.get("cookie")).toBeNull();
  });

  it("never touches the server module on the client branch", async () => {
    branch.mode = "client";
    const transport = makeFactory().createTransport({ baseUrl: "http://api" });
    okUnary();
    await transport.unary(unaryMethod, undefined, undefined, new Headers(), {}, undefined);
    await runInterceptors({ header: new Headers() });
    expect(serverCookie).not.toHaveBeenCalled();
  });
});

describe("server transport resolution", () => {
  it("caches one transport per resolved origin", async () => {
    branch.mode = "server";
    okUnary();
    const factory = makeFactory();
    await factory
      .createTransport()
      .unary(unaryMethod, undefined, undefined, new Headers(), {}, undefined);
    await factory
      .createTransport()
      .unary(unaryMethod, undefined, undefined, new Headers(), {}, undefined);
    const baseUrls = vi.mocked(createConnectTransport).mock.calls.map((call) => call[0]?.baseUrl);
    expect(baseUrls.filter((url) => url === "http://origin/connect").length).toBe(1);
  });

  it("builds a fresh transport when the origin changes", async () => {
    branch.mode = "server";
    okUnary();
    const factory = makeFactory();
    await factory
      .createTransport()
      .unary(unaryMethod, undefined, undefined, new Headers(), {}, undefined);
    originHolder.origin = "http://second";
    await factory
      .createTransport()
      .unary(unaryMethod, undefined, undefined, new Headers(), {}, undefined);
    const baseUrls = vi.mocked(createConnectTransport).mock.calls.map((call) => call[0]?.baseUrl);
    expect(baseUrls).toContain("http://origin/connect");
    expect(baseUrls).toContain("http://second/connect");
  });

  it("accepts a synchronous site-url binding", async () => {
    branch.mode = "server";
    okUnary();
    const factory = createTransportFactory({
      getLocale: () => "zh-CN",
      getSiteUrl: () => "http://sync-origin",
      proxyPath: "/connect",
      getSessionNamespace: () => "member",
      normalize,
    });
    await factory
      .createTransport()
      .unary(unaryMethod, undefined, undefined, new Headers(), {}, undefined);
    expect(vi.mocked(createConnectTransport).mock.calls.at(-1)?.[0].baseUrl).toBe(
      "http://sync-origin/connect",
    );
  });

  it("binds the client default transport to the proxy path", () => {
    branch.mode = "client";
    makeFactory().createTransport();
    expect(vi.mocked(createConnectTransport).mock.calls.at(-1)?.[0].baseUrl).toBe("/connect");
  });

  it("routes stream calls through the cached server transport too", async () => {
    branch.mode = "server";
    streamImpl.mockResolvedValue({
      stream: true,
      service,
      method: streamMethod,
      header: new Headers(),
      trailer: new Headers(),
      message: (async function* () {})(),
    });
    const factory = makeFactory();
    await factory
      .createTransport()
      .stream(
        streamMethod,
        undefined,
        undefined,
        new Headers(),
        (async function* () {})(),
        undefined,
      );
    await factory
      .createTransport()
      .stream(
        streamMethod,
        undefined,
        undefined,
        new Headers(),
        (async function* () {})(),
        undefined,
      );
    const baseUrls = vi.mocked(createConnectTransport).mock.calls.map((call) => call[0]?.baseUrl);
    expect(baseUrls.filter((url) => url === "http://origin/connect").length).toBe(1);
  });
});
