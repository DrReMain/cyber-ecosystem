import type {
  DescMessage,
  DescMethodServerStreaming,
  MessageInitShape,
  MessageShape,
} from "@bufbuild/protobuf";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useCallback, useState } from "react";
import type { StreamStatus } from "./use-server-stream";
import { useServerStream } from "./use-server-stream";

// Server-streaming as a query: the stream writes projected messages into the
// query cache (the external-write paradigm - the canonical bridge between
// streams and react-query), so every consumer of the key reads the same live
// data. start() (re)opens the stream and clears accumulated data; pass
// enabled: true to open on mount instead.
export function useStreamQuery<I extends DescMessage, O extends DescMessage, T = MessageShape<O>>(
  method: DescMethodServerStreaming<I, O>,
  input: MessageInitShape<I>,
  {
    queryKey,
    project,
    enabled = false,
  }: {
    queryKey: readonly unknown[];
    project?: (message: MessageShape<O>) => T;
    enabled?: boolean;
  },
): { data: T[] | undefined; status: StreamStatus; error: unknown; start: () => void } {
  const queryClient = useQueryClient();
  const [runs, setRuns] = useState(0);

  const { data } = useQuery({
    queryKey,
    queryFn: () => [] as T[],
    staleTime: Infinity,
  });

  const { status, error } = useServerStream(method, input, {
    enabled: enabled || runs > 0,
    restart: runs,
    onMessage: (message) => {
      const value = project ? project(message) : (message as MessageShape<O> as T);
      queryClient.setQueryData<T[]>(queryKey, (old) => [...(old ?? []), value]);
    },
  });

  const start = useCallback(() => {
    queryClient.setQueryData<T[]>(queryKey, []);
    setRuns((r) => r + 1);
  }, [queryClient, queryKey]);

  return { data, status, error, start };
}
