import type { DescMessage, MessageShape } from "@bufbuild/protobuf";
import { toJson } from "@bufbuild/protobuf";
import type { Interceptor, StreamResponse, Transport } from "@connectrpc/connect";
import { addStaticKeyToTransport } from "@connectrpc/connect-query";
import { createConnectTransport } from "@connectrpc/connect-web";
import { createIsomorphicFn } from "@tanstack/react-start";

import { wsStreamCall } from "./ws-stream";

export interface TransportBindings {
  getLocale: () => string;
  getSiteUrl: () => string | Promise<string>;
  proxyPath: string;
  getSessionNamespace: () => string;
  normalize: (error: unknown) => unknown;
}

export const SESSION_NAMESPACE_HEADER = "x-session-namespace";

export function sessionCookieNameFor(namespace: string): string {
  return namespace === "" ? "session_token" : `session_token_${namespace}`;
}

export interface TransportFactory {
  createTransport(options?: { baseUrl?: string; interceptors?: Interceptor[] }): Transport;
  connectTransport: Transport;
}

export function createTransportFactory(bindings: TransportBindings): TransportFactory {
  const localeInterceptor: Interceptor = (next) => {
    return async (req) => {
      req.header.set("Accept-Language", bindings.getLocale());
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

  const sessionNamespaceInterceptor: Interceptor = (next) => {
    return async (req) => {
      const namespace = bindings.getSessionNamespace();
      if (namespace !== "") {
        req.header.set(SESSION_NAMESPACE_HEADER, namespace);
      }
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

  function normalizedStream<I extends DescMessage, O extends DescMessage>(
    response: StreamResponse<I, O>,
  ): StreamResponse<I, O> {
    return {
      ...response,
      message: (async function* () {
        try {
          yield* response.message;
        } catch (error) {
          throw bindings.normalize(error);
        }
      })(),
    };
  }

  function headerCancellationFetch(fetchImpl: typeof globalThis.fetch): typeof globalThis.fetch {
    return (input, init) => {
      const outer = init?.signal;
      if (outer == null) {
        return fetchImpl(input, init);
      }
      const inner = new AbortController();
      const forward = () => inner.abort(outer.reason);
      if (outer.aborted) {
        inner.abort(outer.reason);
      } else {
        outer.addEventListener("abort", forward);
      }
      return fetchImpl(input, { ...init, signal: inner.signal }).finally(() => {
        outer.removeEventListener("abort", forward);
      });
    };
  }

  function singleConnectTransport(baseUrl: string, interceptors: Interceptor[]): Transport {
    const interceptorChain = [
      localeInterceptor,
      cookieForwarder,
      sessionNamespaceInterceptor,
      ...interceptors,
    ];
    const unaryConnect = createConnectTransport({
      baseUrl,
      fetch: headerCancellationFetch(globalThis.fetch),
      interceptors: interceptorChain,
    });
    const streamConnect = createConnectTransport({
      baseUrl,
      interceptors: interceptorChain,
    });
    return {
      unary: async (method, signal, timeoutMs, header, input, contextValues) => {
        try {
          const response = await unaryConnect.unary(
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
          throw bindings.normalize(error);
        }
      },
      stream: async (method, signal, timeoutMs, header, input, contextValues) => {
        try {
          const response =
            method.methodKind === "server_streaming"
              ? await streamConnect.stream(method, signal, timeoutMs, header, input, contextValues)
              : await wsStreamCall(baseUrl, method, signal, input);
          return flattenStream(method.output, normalizedStream(response));
        } catch (error) {
          throw bindings.normalize(error);
        }
      },
    };
  }

  const serverTransports = new Map<string, Transport>();

  async function serverConnectTransport(): Promise<Transport> {
    const origin = await bindings.getSiteUrl();
    let transport = serverTransports.get(origin);
    if (!transport) {
      transport = singleConnectTransport(`${origin}${bindings.proxyPath}`, []);
      serverTransports.set(origin, transport);
    }
    return transport;
  }

  function lazyServerConnectTransport(): Transport {
    return {
      unary: async (method, signal, timeoutMs, header, input, contextValues) => {
        const transport = await serverConnectTransport();
        return transport.unary(method, signal, timeoutMs, header, input, contextValues);
      },
      stream: async (method, signal, timeoutMs, header, input, contextValues) => {
        const transport = await serverConnectTransport();
        return transport.stream(method, signal, timeoutMs, header, input, contextValues);
      },
    };
  }

  const defaultConnectTransport = createIsomorphicFn()
    .server(() => lazyServerConnectTransport())
    .client(() => singleConnectTransport(bindings.proxyPath, []));

  function createTransport(
    options: { baseUrl?: string; interceptors?: Interceptor[] } = {},
  ): Transport {
    if (options.baseUrl !== undefined) {
      return singleConnectTransport(options.baseUrl, options.interceptors ?? []);
    }
    return defaultConnectTransport();
  }

  return {
    createTransport,
    connectTransport: addStaticKeyToTransport(createTransport(), "connect"),
  };
}
