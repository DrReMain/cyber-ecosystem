import type {
  DescMessage,
  DescMethodClientStreaming,
  MessageInitShape,
  MessageShape,
} from "@bufbuild/protobuf";
import { Code, ConnectError } from "@connectrpc/connect";
import { useTransport } from "@connectrpc/connect-query";
import { useMutation } from "@tanstack/react-query";

// The input of a client-streaming call: any iterable of init-shaped messages
// - a plain array or an async generator.
export type StreamMutationInput<I extends DescMessage> =
  | Iterable<MessageInitShape<I>>
  | AsyncIterable<MessageInitShape<I>>;

async function* toAsyncIterable<I extends DescMessage>(
  input: StreamMutationInput<I>,
): AsyncGenerator<MessageInitShape<I>> {
  // for-await accepts sync iterables as well.
  for await (const message of input) {
    yield message;
  }
}

// Client-streaming as a mutation: a finite input stream producing a single
// response. Wrapping tanstack useMutation puts stream calls on the same
// error pipeline (mutation cache onError → toast) as unary mutations. The
// drain mirrors the official client's semantics: exactly one response
// message, anything else is a protocol error.
export function useStreamMutation<I extends DescMessage, O extends DescMessage>(
  method: DescMethodClientStreaming<I, O>,
) {
  const transport = useTransport();
  return useMutation({
    mutationFn: async (input: StreamMutationInput<I>): Promise<MessageShape<O>> => {
      const response = await transport.stream(
        method,
        undefined,
        undefined,
        undefined,
        toAsyncIterable(input),
      );
      let single: MessageShape<O> | undefined;
      let count = 0;
      for await (const message of response.message) {
        single = message;
        count++;
      }
      if (!single) {
        throw new ConnectError("protocol error: missing response message", Code.Unimplemented);
      }
      if (count > 1) {
        throw new ConnectError(
          "protocol error: received extra messages for client streaming method",
          Code.Unimplemented,
        );
      }
      return single;
    },
  });
}
