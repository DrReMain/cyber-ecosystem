import type { DescMessage, MessageShape } from "@bufbuild/protobuf";
import { toJson } from "@bufbuild/protobuf";
import type { Interceptor, StreamResponse, Transport } from "@connectrpc/connect";
import { addStaticKeyToTransport } from "@connectrpc/connect-query";
import { createConnectTransport } from "@connectrpc/connect-web";
import { createIsomorphicFn } from "@tanstack/react-start";
import { normalize } from "#/domains/error";
import { resolveCONNECTBaseUrl } from "#/env";
import { getLocale } from "#/paraglide/runtime";
import { wsStreamCall } from "./ws-stream";

export const localeInterceptor: Interceptor = (next) => {
  return async (req) => {
    req.header.set("Accept-Language", getLocale());
    return await next(req);
  };
};

const forwardedCookie = createIsomorphicFn()
  .server(async () => (await import("@tanstack/react-start/server")).getRequestHeader("cookie"))
  .client((): string | null => null);

const cookieForwarder: Interceptor = (next) => {
  return async (req) => {
    const cookie = await forwardedCookie();
    if (cookie) req.header.set("cookie", cookie);
    return await next(req);
  };
};

function flattenStream<I extends DescMessage, O extends DescMessage>(
  output: O,
  response: StreamResponse<I, O>,
): StreamResponse<I, O> {
  return {
    ...response,
    message: (async function* () {
      for await (const message of response.message) {
        yield toJson(output, message) as unknown as MessageShape<O>;
      }
    })(),
  };
}

// Stream iteration errors normalize at the exit so consumers receive
// envelopes no matter where in the iteration the stream fails.
function normalizedStream<I extends DescMessage, O extends DescMessage>(
  response: StreamResponse<I, O>,
): StreamResponse<I, O> {
  return {
    ...response,
    message: (async function* () {
      try {
        yield* response.message;
      } catch (error) {
        throw normalize(error);
      }
    })(),
  };
}

export function createKratosTransport({
  baseUrl = resolveCONNECTBaseUrl(),
  interceptors,
}: {
  baseUrl?: string;
  interceptors?: Interceptor[];
} = {}): Transport {
  const connect = createConnectTransport({
    baseUrl,
    interceptors: [localeInterceptor, cookieForwarder, ...(interceptors ?? [])],
  });
  return {
    unary: async (method, signal, timeoutMs, header, input, contextValues) => {
      try {
        const response = await connect.unary(
          method,
          signal,
          timeoutMs,
          header,
          input,
          contextValues,
        );
        return {
          ...response,
          message: toJson(method.output, response.message) as unknown as typeof response.message,
        };
      } catch (error) {
        throw normalize(error);
      }
    },
    stream: async (method, signal, timeoutMs, header, input, contextValues) => {
      try {
        const response =
          method.methodKind === "server_streaming"
            ? await connect.stream(method, signal, timeoutMs, header, input, contextValues)
            : await wsStreamCall(baseUrl, method, signal, input);
        return flattenStream(method.output, normalizedStream(response));
      } catch (error) {
        throw normalize(error);
      }
    },
  };
}

export const connectTransport = addStaticKeyToTransport(
  createKratosTransport({ baseUrl: resolveCONNECTBaseUrl() }),
  "connect",
);
