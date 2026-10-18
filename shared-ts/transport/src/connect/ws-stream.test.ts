import { PageRequestSchema } from "@cyber-ecosystem/gen-connect-ts/cyber/shared/common/v1/pagination_pb";
import { afterAll, afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { wsStreamCall } from "./ws-stream";

const Frame = PageRequestSchema;

const method = {
  kind: "server_streaming",
  name: "Chat",
  parent: { typeName: "test.v1.ChatService" },
  input: Frame,
  output: Frame,
} as never;

const REAL_WEBSOCKET = globalThis.WebSocket;

class FakeSocket {
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSING = 2;
  static readonly CLOSED = 3;

  readyState = FakeSocket.CONNECTING;
  url: string;
  sent: string[] = [];
  closed = false;
  onopen: (() => void) | null = null;
  onmessage: ((event: { data: string }) => void) | null = null;
  onclose: ((event: { code: number; reason: string }) => void) | null = null;
  onerror: (() => void) | null = null;

  constructor(url: string) {
    this.url = url;
  }

  send(frame: string): void {
    if (this.readyState === FakeSocket.OPEN) {
      this.sent.push(frame);
    } else if (this.readyState === FakeSocket.CONNECTING) {
      this.sent.push(frame);
    }
  }

  close(): void {
    this.closed = true;
    this.readyState = FakeSocket.CLOSED;
  }
}

const sockets: FakeSocket[] = [];

beforeEach(() => {
  sockets.length = 0;
  vi.stubGlobal(
    "WebSocket",
    class extends FakeSocket {
      constructor(url: string) {
        super(url);
        sockets.push(this);
      }
    },
  );
});

afterEach(() => {
  vi.unstubAllGlobals();
});

afterAll(() => {
  globalThis.WebSocket = REAL_WEBSOCKET;
});

async function* oneShot(pageNo: number) {
  yield { pageNo };
}

function socket(): FakeSocket {
  const s = sockets[0];
  if (s === undefined) {
    throw new Error("no socket was constructed");
  }
  return s;
}

function openSocket(socket: FakeSocket): void {
  socket.readyState = FakeSocket.OPEN;
  socket.onopen?.();
}

describe("wsStreamCall", () => {
  it("rejects with Unimplemented when WebSocket is absent (boundary negative)", async () => {
    vi.unstubAllGlobals();
    // biome-ignore lint/suspicious/noExplicitAny: absence is the subject
    (globalThis as any).WebSocket = undefined;
    await expect(
      wsStreamCall("http://localhost:5174/connect", method, undefined, oneShot(1)),
    ).rejects.toThrow(/WebSocket is not available/);
  });

  it("builds the ws url from the base and the procedure path", async () => {
    const promise = wsStreamCall("http://localhost:5174/connect", method, undefined, oneShot(1));
    const response = await promise;
    expect(socket().url).toBe("ws://localhost:5174/connect/test.v1.ChatService/Chat");
    void response.message;
  });

  it("half-closes with the end control frame when the input exhausts", async () => {
    const response = await wsStreamCall("http://x/connect", method, undefined, oneShot(2));
    const drained = (async () => {
      for await (const _message of response.message) {
        // drain
      }
    })();
    await new Promise((resolve) => setTimeout(resolve, 0));
    openSocket(socket());
    // The client's own half-close does not finish the response side; the
    // server acknowledges with its own end frame.
    socket().onmessage?.({ data: "\x1eend" });
    await drained;
    expect(socket().sent.at(-1)).toBe("\x1eend");
  });

  it("encodes input frames as protojson of the input schema", async () => {
    const response = await wsStreamCall("http://x/connect", method, undefined, oneShot(2));
    const drained = (async () => {
      for await (const _message of response.message) {
        // drain
      }
    })();
    await new Promise((resolve) => setTimeout(resolve, 0));
    openSocket(socket());
    socket().onmessage?.({ data: "\x1eend" });
    await drained;
    const dataFrame = socket().sent.find((f) => !f.startsWith("\x1e"));
    expect(JSON.parse(dataFrame as string)).toEqual({ pageNo: 2 });
  });

  it("decodes response frames and completes on the end frame", async () => {
    const response = await wsStreamCall("http://x/connect", method, undefined, oneShot(2));
    const collected: unknown[] = [];
    const draining = (async () => {
      for await (const message of response.message) {
        collected.push(message);
      }
    })();
    openSocket(socket());
    socket().onmessage?.({ data: JSON.stringify({ pageNo: 7 }) });
    socket().onmessage?.({ data: "\x1eend" });
    await draining;
    expect(collected).toEqual([expect.objectContaining({ pageNo: 7 })]);
  });

  it("surfaces the error control frame as a ConnectError during iteration", async () => {
    const response = await wsStreamCall("http://x/connect", method, undefined, oneShot(2));
    const draining = (async () => {
      for await (const _message of response.message) {
        // drain
      }
    })();
    openSocket(socket());
    socket().onmessage?.({ data: "\x1eerror:boom" });
    await expect(draining).rejects.toThrow(/boom/);
  });

  it("treats an abnormal close as Unavailable (boundary negative)", async () => {
    const response = await wsStreamCall("http://x/connect", method, undefined, oneShot(2));
    const draining = (async () => {
      for await (const _message of response.message) {
        // drain
      }
    })();
    openSocket(socket());
    socket().onclose?.({ code: 1006, reason: "abnormal" });
    await expect(draining).rejects.toThrow(/websocket closed: 1006/);
  });

  it("ends with Canceled when the signal aborts (boundary negative)", async () => {
    const controller = new AbortController();
    const response = await wsStreamCall(
      "http://x/connect",
      method,
      controller.signal,
      (async function* () {
        yield { pageNo: 1 };
        await new Promise(() => {});
      })(),
    );
    const draining = (async () => {
      for await (const _message of response.message) {
        // drain
      }
    })();
    openSocket(socket());
    controller.abort();
    await expect(draining).rejects.toThrow(/aborted/);
    expect(socket().closed).toBe(true);
  });

  it("derives the ws scheme from location for relative base urls (https)", async () => {
    vi.stubGlobal("location", { protocol: "https:", host: "h5.example.test" });
    const draining = wsStreamCall("/connect", method, undefined, oneShot(1));
    openSocket(socket());
    socket().onclose?.({ code: 1000, reason: "" });
    await draining;
    expect(socket().url).toBe("wss://h5.example.test/connect/test.v1.ChatService/Chat");
  });

  it("derives the ws scheme from location for relative base urls (http)", async () => {
    vi.stubGlobal("location", { protocol: "http:", host: "localhost:5174" });
    const draining = wsStreamCall("/connect", method, undefined, oneShot(1));
    openSocket(socket());
    socket().onclose?.({ code: 1000, reason: "" });
    await draining;
    expect(socket().url).toBe("ws://localhost:5174/connect/test.v1.ChatService/Chat");
  });

  it("completes cleanly on a 1000 close without an end frame", async () => {
    const draining = wsStreamCall("http://api", method, undefined, oneShot(1));
    openSocket(socket());
    socket().onmessage?.({ data: JSON.stringify({ pageNo: 7 }) });
    socket().onclose?.({ code: 1000, reason: "" });
    const seen: unknown[] = [];
    for await (const message of (await draining).message) {
      seen.push(message);
    }
    expect(seen).toEqual([expect.objectContaining({ pageNo: 7 })]);
  });

  it("fails with Unavailable when the handshake errors before open (boundary negative)", async () => {
    const pending = wsStreamCall("http://api", method, undefined, oneShot(1));
    socket().onerror?.();
    await expect(
      (async () => {
        for await (const _ of (await pending).message) {
          // drain
        }
      })(),
    ).rejects.toThrow(/websocket connection failed/);
  });

  it("fails with Internal when the input pump throws mid-stream (boundary negative)", async () => {
    async function* brokenInput(): AsyncIterable<{ pageNo: number }> {
      yield { pageNo: 1 };
      throw new Error("producer died");
    }
    const pending = wsStreamCall("http://api", method, undefined, brokenInput());
    openSocket(socket());
    await expect(
      (async () => {
        for await (const _ of (await pending).message) {
          // drain
        }
      })(),
    ).rejects.toThrow(/input stream failed/);
    expect(socket().closed).toBe(true);
  });
});
