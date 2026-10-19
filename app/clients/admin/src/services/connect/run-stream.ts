import type {
  DescMessage,
  DescMethodStreaming,
  MessageInitShape,
  MessageShape,
} from "@bufbuild/protobuf";
import type { Transport } from "@connectrpc/connect";

export async function runStream<I extends DescMessage, O extends DescMessage>(
  transport: Transport,
  method: DescMethodStreaming<I, O>,
  signal: AbortSignal,
  input: AsyncIterable<MessageInitShape<I>>,
  onMessage: (message: MessageShape<O>) => void,
): Promise<void> {
  const response = await transport.stream(method, signal, undefined, undefined, input);
  for await (const message of response.message) {
    if (signal.aborted) return;
    onMessage(message);
  }
}
