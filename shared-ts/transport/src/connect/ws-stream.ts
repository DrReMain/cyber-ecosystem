import type {
  DescMessage,
  DescMethodStreaming,
  MessageInitShape,
  MessageShape,
} from "@bufbuild/protobuf";
import { create, fromJson, toJson } from "@bufbuild/protobuf";
import type { StreamResponse } from "@connectrpc/connect";
import { Code, ConnectError } from "@connectrpc/connect";

// Client leg of the kratos WS stream dialect (server side: shared-go/kratos/
// transport/connect/ws.go). Same port and procedure paths as the connect
// protocol: every data frame is one WS text message holding a protojson-
// encoded message, the client half-closes with the control frame "\x1eend",
// the server signals errors with a "\x1eerror:" frame followed by an abnormal
// close, and a clean close ends the stream. Browsers cannot set WS handshake
// headers, so codec negotiation falls through to the server default
// (protojson) on both sides.

const CONTROL_END = "\x1eend";
const CONTROL_ERROR = "\x1eerror:";

function wsUrl(baseUrl: string, method: DescMethodStreaming): string {
  const path = `/${method.parent.typeName}/${method.name}`;
  if (/^https?:\/\//.test(baseUrl)) {
    return `${baseUrl.replace(/^http/, "ws").replace(/\/+$/, "")}${path}`;
  }
  const scheme = location.protocol === "https:" ? "wss" : "ws";
  return `${scheme}://${location.host}${baseUrl.replace(/\/+$/, "")}${path}`;
}

// Resolves as done once the signal aborts, so the input pump stops pulling
// from a producer (e.g. a Pushable) that nobody will end anymore.
function nextInput<T>(
  iterator: AsyncIterator<T>,
  signal: AbortSignal | undefined,
): Promise<IteratorResult<T>> {
  if (!signal) {
    return iterator.next();
  }
  return Promise.race([
    iterator.next(),
    new Promise<IteratorResult<T>>((resolve) => {
      signal.addEventListener("abort", () => resolve({ value: undefined, done: true }), {
        once: true,
      });
    }),
  ]);
}

// Implements Transport["stream"] for client-streaming and bidi methods over
// the WS dialect, returning the official StreamResponse shape. timeoutMs and
// header are not carried: streams are long-lived (the connect timeout does
// not apply, mirroring the server) and browsers cannot set WS handshake
// headers. Errors surface through the message iterable.
export function wsStreamCall<I extends DescMessage, O extends DescMessage>(
  baseUrl: string,
  method: DescMethodStreaming<I, O>,
  signal: AbortSignal | undefined,
  input: AsyncIterable<MessageInitShape<I>>,
): Promise<StreamResponse<I, O>> {
  if (typeof WebSocket === "undefined") {
    return Promise.reject(
      new ConnectError("WebSocket is not available in this environment", Code.Unimplemented),
    );
  }

  const socket = new WebSocket(wsUrl(baseUrl, method));

  // Pending frames produced before the socket opens; frames aimed at a dead
  // (CLOSED) socket are dropped silently.
  let sendBuffer: string[] = [];
  const sendFrame = (frame: string) => {
    if (socket.readyState === WebSocket.OPEN) {
      socket.send(frame);
    } else if (socket.readyState === WebSocket.CONNECTING) {
      sendBuffer.push(frame);
    }
  };

  // Response queue: the async iterator pulls, onmessage pushes.
  const queue: MessageShape<O>[] = [];
  let notify: (() => void) | null = null;
  let finished = false;
  let failure: ConnectError | null = null;
  // Wakes the parked response iterator; finish routes through it.
  const wake = () => {
    const n = notify;
    notify = null;
    n?.();
  };
  const finish = (err: ConnectError | null) => {
    if (finished) return;
    finished = true;
    failure = err;
    wake();
  };

  socket.onopen = () => {
    // sendFrame dispatches directly now that the socket is OPEN.
    const pending = sendBuffer;
    sendBuffer = [];
    for (const frame of pending) {
      sendFrame(frame);
    }
  };
  socket.onmessage = (event) => {
    if (typeof event.data !== "string" || finished) return;
    if (event.data === CONTROL_END) {
      finish(null);
      return;
    }
    if (event.data.startsWith(CONTROL_ERROR)) {
      // The dialect carries only the error text; the connect code is lost.
      finish(new ConnectError(event.data.slice(CONTROL_ERROR.length), Code.Internal));
      return;
    }
    queue.push(fromJson(method.output, JSON.parse(event.data)));
    wake();
  };
  socket.onclose = (event) => {
    if (finished) return;
    if (event.code === 1000 || event.code === 1001) {
      finish(null);
    } else {
      finish(new ConnectError(`websocket closed: ${event.code} ${event.reason}`, Code.Unavailable));
    }
  };
  socket.onerror = () => {
    if (!finished && socket.readyState !== WebSocket.OPEN) {
      finish(new ConnectError("websocket connection failed", Code.Unavailable));
    }
  };

  signal?.addEventListener(
    "abort",
    () => {
      socket.close();
      finish(new ConnectError("this operation was aborted", Code.Canceled));
    },
    { once: true },
  );

  // Pumps the input iterable into frames. If the producer never ends after
  // the stream finished, the pump parks on next() - a pending promise nothing
  // can resolve anymore, collectible once the producer itself is
  // unreachable. The abort race covers the unmount case.
  const pumpInput = async () => {
    const iterator = input[Symbol.asyncIterator]();
    try {
      for (;;) {
        const result = await nextInput(iterator, signal);
        if (result.done) {
          // A naturally exhausted input half-closes; the server sees Recv EOF.
          sendFrame(CONTROL_END);
          return;
        }
        if (finished) return;
        // create() normalizes the init shape into a full runtime message
        // (defaults + $typeName) - toJson requires it.
        sendFrame(JSON.stringify(toJson(method.input, create(method.input, result.value))));
      }
    } catch (err) {
      finish(new ConnectError(`input stream failed: ${String(err)}`, Code.Internal));
      socket.close();
    }
  };
  void pumpInput();

  // Response iterator: yields queued messages, wakes on every push, and
  // throws the stream failure (if any) when the stream ends - errors surface
  // through iteration, the official transport contract.
  async function* messages(): AsyncGenerator<MessageShape<O>> {
    for (;;) {
      const msg = queue.shift();
      if (msg !== undefined) {
        yield msg;
        continue;
      }
      if (finished) {
        if (failure) {
          throw failure;
        }
        return;
      }
      await new Promise<void>((resolve) => {
        notify = () => resolve();
      });
    }
  }

  // The dialect carries no response headers or trailers; both stay empty.
  return Promise.resolve({
    stream: true,
    service: method.parent,
    method,
    header: new Headers(),
    trailer: new Headers(),
    message: messages(),
  });
}
