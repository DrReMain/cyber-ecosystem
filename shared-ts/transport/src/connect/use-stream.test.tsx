import type { Transport } from "@connectrpc/connect";
import { TransportProvider } from "@connectrpc/connect-query";
import {
  PageRequestSchema,
  PageResponseSchema,
} from "@cyber-ecosystem/gen-connect-ts/cyber/shared/common/v1/pagination_pb";
import { renderHook, waitFor } from "@testing-library/react";
import type { PropsWithChildren } from "react";
import { describe, expect, it, vi } from "vitest";
import { useServerStream } from "./use-server-stream";

const service = { typeName: "test.v1.StreamService" };

const serverMethod = {
  methodKind: "server_streaming",
  name: "Watch",
  parent: service,
  input: PageRequestSchema,
  output: PageResponseSchema,
} as unknown as Parameters<typeof useServerStream>[0];

type StreamImpl = NonNullable<Transport["stream"]>;

// The fake streams yield plain objects; the generics of the real signature
// demand message shapes, so the impl crosses one deliberate cast.
function makeTransport(streamImpl: unknown): Transport {
  return {
    unary: async () => {
      throw new Error("unary is not under test here");
    },
    stream: streamImpl as StreamImpl,
  } as Transport;
}

function wrapper(transport: Transport) {
  // biome-ignore lint/correctness/noComponentHookFactories: one Wrapper per test transport at test setup, not a render-path component.
  return function Wrapper({ children }: Readonly<PropsWithChildren>) {
    return <TransportProvider transport={transport}>{children}</TransportProvider>;
  };
}

describe("useServerStream", () => {
  it("delivers every server frame, then settles on done", async () => {
    const onMessage = vi.fn();
    const transport = makeTransport(async () => ({
      stream: true,
      service,
      method: serverMethod,
      header: new Headers(),
      trailer: new Headers(),
      message: (async function* () {
        yield { pageNo: 1 };
        yield { pageNo: 2 };
      })(),
    }));
    const { result } = renderHook(
      () => useServerStream(serverMethod, { pageNo: 1 }, { onMessage }),
      { wrapper: wrapper(transport) },
    );
    await waitFor(() => expect(result.current.status).toBe("done"));
    expect(onMessage).toHaveBeenCalledTimes(2);
    expect(onMessage).toHaveBeenNthCalledWith(1, { pageNo: 1 });
    expect(result.current.error).toBeNull();
  });

  it("lands on error when the stream call fails", async () => {
    const onMessage = vi.fn();
    const transport = makeTransport(async () => {
      throw new Error("stream refused");
    });
    const { result } = renderHook(
      () => useServerStream(serverMethod, { pageNo: 1 }, { onMessage }),
      { wrapper: wrapper(transport) },
    );
    await waitFor(() => expect(result.current.status).toBe("error"));
    expect(result.current.error).toMatchObject({ message: "stream refused" });
    expect(onMessage).not.toHaveBeenCalled();
  });

  it("stays idle and never calls the transport while disabled", async () => {
    const streamImpl = vi.fn(async () => ({
      stream: true,
      service,
      method: serverMethod,
      header: new Headers(),
      trailer: new Headers(),
      message: (async function* () {})(),
    }));
    const transport = makeTransport(streamImpl);
    const { result } = renderHook(
      () => useServerStream(serverMethod, { pageNo: 1 }, { enabled: false, onMessage: vi.fn() }),
      { wrapper: wrapper(transport) },
    );
    expect(result.current.status).toBe("idle");
    expect(streamImpl).not.toHaveBeenCalled();
  });

  it("aborts the in-flight call on unmount (boundary negative)", async () => {
    let capturedSignal: AbortSignal | undefined;
    const transport = makeTransport(async (_method: unknown, signal?: AbortSignal) => {
      capturedSignal = signal;
      return {
        stream: true,
        service,
        method: serverMethod,
        header: new Headers(),
        trailer: new Headers(),
        message: (async function* () {
          yield { pageNo: 1 };
          await new Promise(() => {});
        })(),
      };
    });
    const { result, unmount } = renderHook(
      () => useServerStream(serverMethod, { pageNo: 1 }, { onMessage: vi.fn() }),
      { wrapper: wrapper(transport) },
    );
    await waitFor(() => expect(result.current.status).toBe("streaming"));
    unmount();
    expect(capturedSignal?.aborted).toBe(true);
  });

  it("restarts the stream when the restart token changes", async () => {
    let calls = 0;
    const transport = makeTransport(async () => {
      calls += 1;
      return {
        stream: true,
        service,
        method: serverMethod,
        header: new Headers(),
        trailer: new Headers(),
        message: (async function* () {})(),
      };
    });
    const { result, rerender } = renderHook(
      ({ restart }: { restart: number }) =>
        useServerStream(serverMethod, { pageNo: 1 }, { restart, onMessage: vi.fn() }),
      { wrapper: wrapper(transport), initialProps: { restart: 0 } },
    );
    await waitFor(() => expect(result.current.status).toBe("done"));
    rerender({ restart: 1 });
    await waitFor(() => expect(calls).toBe(2));
  });
});
