import type { DescMessage, MessageShape } from "@bufbuild/protobuf";
import { toJson } from "@bufbuild/protobuf";
import type { Interceptor, StreamResponse, Transport } from "@connectrpc/connect";
import { addStaticKeyToTransport } from "@connectrpc/connect-query";
import { createConnectTransport } from "@connectrpc/connect-web";
import { createIsomorphicFn } from "@tanstack/react-start";

import { wsStreamCall } from "./ws-stream";

// Client-specific bindings the mechanism cannot own: the locale getter
// (paraglide runtime), the SSR origin resolver (client env), the proxy
// path, and the error-domain normalizer. Injecting them keeps this package
// free of client imports while both clients share one full-form transport.
export interface TransportBindings {
  getLocale: () => string;
  // Isomorphic by nature: the server resolves the origin asynchronously,
  // the browser returns it synchronously; the transport awaits either.
  getSiteUrl: () => string | Promise<string>;
  proxyPath: string;
  // The client world's credential slot, riding every request so the server
  // reads the right namespaced session cookie. An untrusted selection among
  // cookies the browser already sent; mirrors shared-go's
  // SessionNamespaceHeader — rename the two together.
  getSessionNamespace: () => string;
  // Returns the client error domain's envelope; thrown in place of the
  // raw transport error so consumers always receive the normalized shape.
  normalize: (error: unknown) => unknown;
}

// The session credential's wire contract, mirroring shared-go's
// krauth.SessionNamespaceHeader / SessionCookieNameFor — the one place this
// language spells them; rename the two languages together.
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

  // Server-side authenticated calls forward the session cookie verbatim;
  // browser calls rely on same-origin cookies riding automatically.
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
          throw bindings.normalize(error);
        }
      })(),
    };
  }

  function singleConnectTransport(baseUrl: string, interceptors: Interceptor[]): Transport {
    const connect = createConnectTransport({
      baseUrl,
      interceptors: [
        localeInterceptor,
        cookieForwarder,
        sessionNamespaceInterceptor,
        ...interceptors,
      ],
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
          throw bindings.normalize(error);
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
