import type { Transport } from "@connectrpc/connect";
import { describe, expect, it, vi } from "vitest";
import { runStream } from "./run-stream";

function fakeTransport(messages: string[], onStream?: (input: unknown) => void): Transport {
  return {
    stream: vi.fn(
      async (_method: unknown, _signal: unknown, _t: unknown, _h: unknown, input: unknown) => {
        onStream?.(input);
        return {
          message: (async function* () {
            for (const m of messages) {
              yield m;
            }
          })(),
        };
      },
    ),
  } as unknown as Transport;
}

describe("runStream", () => {
  it("delivers every message to onMessage in order and resolves", async () => {
    const seen: unknown[] = [];
    const input = (async function* () {})();
    await runStream(
      fakeTransport(["a", "b", "c"]) as Transport,
      { methodKind: "server_streaming" } as never,
      new AbortController().signal,
      input,
      (m) => seen.push(m),
    );
    expect(seen).toEqual(["a", "b", "c"]);
  });

  it("stops delivering once the signal is aborted (boundary negative)", async () => {
    const controller = new AbortController();
    const seen: unknown[] = [];
    const input = (async function* () {
      yield { pageNo: 1 };
    })();
    await runStream(
      fakeTransport(["one", "two"]) as Transport,
      { methodKind: "server_streaming" } as never,
      controller.signal,
      input,
      () => {
        seen.push("x");
        controller.abort();
      },
    );
    expect(seen).toEqual(["x"]);
  });

  it("passes the input iterable through to the transport", async () => {
    const input = (async function* () {
      yield { text: "q" };
    })();
    const captured: unknown[] = [];
    await runStream(
      fakeTransport([], (i) => captured.push(i)) as Transport,
      { methodKind: "server_streaming" } as never,
      new AbortController().signal,
      input,
      () => {},
    );
    expect(captured[0]).toBe(input);
  });

  it("propagates iteration failures to the caller", async () => {
    const transport = {
      stream: async () => ({
        message: (async function* () {
          yield "first";
          throw new Error("broken");
        })(),
      }),
    } as unknown as Transport;
    await expect(
      runStream(
        transport,
        { methodKind: "server_streaming" } as never,
        new AbortController().signal,
        (async function* () {})(),
        () => {},
      ),
    ).rejects.toThrow("broken");
  });
});
