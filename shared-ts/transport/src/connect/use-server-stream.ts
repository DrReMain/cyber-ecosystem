import type {
  DescMessage,
  DescMethodServerStreaming,
  MessageInitShape,
  MessageShape,
} from "@bufbuild/protobuf";
import { useTransport } from "@connectrpc/connect-query";
import { useEffect, useRef, useState } from "react";
import { runStream } from "./run-stream";

// Shared lifecycle of every stream hook: idle until opened, streaming while
// open, done on a clean end (or abort), error on a failure.
export type StreamStatus = "idle" | "streaming" | "done" | "error";

// input travels as its JSON string (inputKey) so the effect can depend on a
// stable primitive identity - a fresh object literal on every render would
// restart the stream on each render. The JSON round-trip back into a message
// is that mechanism, not a data transformation.
function singleMessage<I extends DescMessage>(inputKey: string) {
  return (async function* () {
    yield JSON.parse(inputKey) as MessageInitShape<I>;
  })();
}

export function useServerStream<I extends DescMessage, O extends DescMessage>(
  method: DescMethodServerStreaming<I, O>,
  input: MessageInitShape<I>,
  {
    enabled = true,
    restart,
    onMessage,
  }: {
    enabled?: boolean;
    restart?: number;
    onMessage: (message: MessageShape<O>) => void;
  },
): { status: StreamStatus; error: unknown } {
  const transport = useTransport();
  const [status, setStatus] = useState<StreamStatus>("idle");
  const [error, setError] = useState<unknown>(null);

  const onMessageRef = useRef(onMessage);
  onMessageRef.current = onMessage;

  const inputKey = JSON.stringify(input);

  // biome-ignore lint/correctness/useExhaustiveDependencies: restart is an intentional re-run trigger, not a value the effect reads
  useEffect(() => {
    if (!enabled) {
      setStatus("idle");
      return;
    }
    const controller = new AbortController();
    let cancelled = false;
    setStatus("streaming");
    setError(null);

    runStream(transport, method, controller.signal, singleMessage<I>(inputKey), (m) =>
      onMessageRef.current(m),
    )
      .then(() => {
        if (!cancelled) setStatus("done");
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err);
          setStatus("error");
        }
      });

    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [transport, method, enabled, restart, inputKey]);

  return { status, error };
}
