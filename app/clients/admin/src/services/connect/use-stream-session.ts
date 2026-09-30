import type {
  DescMessage,
  DescMethodBiDiStreaming,
  MessageInitShape,
  MessageShape,
} from "@bufbuild/protobuf";
import { useTransport } from "@connectrpc/connect-query";
import { useCallback, useEffect, useRef, useState } from "react";
import type { Pushable } from "./pushable";
import { createPushable } from "./pushable";
import { runStream } from "./run-stream";
import type { StreamStatus } from "./use-server-stream";

// Bidirectional streaming as a session: open() starts the call with an
// internally-owned pushable input (send pushes one message, end half-closes)
// and an AbortController that tears the socket down on unmount. Server frames
// arrive through onMessage; status follows the shared stream lifecycle.
export function useStreamSession<I extends DescMessage, O extends DescMessage>(
  method: DescMethodBiDiStreaming<I, O>,
  { onMessage }: { onMessage: (message: MessageShape<O>) => void },
): {
  status: StreamStatus;
  error: unknown;
  open: () => void;
  send: (message: MessageInitShape<I>) => void;
  end: () => void;
} {
  const transport = useTransport();
  const [status, setStatus] = useState<StreamStatus>("idle");
  const [error, setError] = useState<unknown>(null);
  const pushableRef = useRef<Pushable<MessageInitShape<I>> | null>(null);
  const controllerRef = useRef<AbortController | null>(null);

  const onMessageRef = useRef(onMessage);
  onMessageRef.current = onMessage;

  // Aborting tears the call down - the transport closes the socket.
  useEffect(() => () => controllerRef.current?.abort(), []);

  const open = useCallback(() => {
    if (controllerRef.current) return;
    const controller = new AbortController();
    controllerRef.current = controller;
    const pushable = createPushable<MessageInitShape<I>>();
    pushableRef.current = pushable;
    setStatus("streaming");
    setError(null);
    runStream(transport, method, controller.signal, pushable, (message) =>
      onMessageRef.current(message),
    )
      .then(() => setStatus("done"))
      .catch((err: unknown) => {
        if (controller.signal.aborted) {
          // An abort ends the session (unmount/teardown), not a failure.
          setStatus("done");
        } else {
          setError(err);
          setStatus("error");
        }
      })
      .finally(() => {
        controllerRef.current = null;
        pushableRef.current = null;
      });
  }, [transport, method]);

  const send = useCallback((message: MessageInitShape<I>) => {
    pushableRef.current?.push(message);
  }, []);

  const end = useCallback(() => {
    pushableRef.current?.end();
  }, []);

  return { status, error, open, send, end };
}
