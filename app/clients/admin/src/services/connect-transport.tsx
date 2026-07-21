import type { Interceptor } from "@connectrpc/connect";
import { TransportProvider as ConnectTransportProvider } from "@connectrpc/connect-query";
import { createConnectTransport } from "@connectrpc/connect-web";
import type { PropsWithChildren } from "react";
import { resolveCONNECTBaseUrl } from "#/env";
import { getLocale } from "#/paraglide/runtime";

const localeInterceptor: Interceptor = (next) => async (req) => {
  req.header.set("Accept-Language", getLocale());
  return next(req);
};

// Note: connect errors flow through QueryCache/MutationCache onError → emitApiError
// (see router.tsx), so no transport errorInterceptor here — it would duplicate feedback.
export const transport = createConnectTransport({
  baseUrl: resolveCONNECTBaseUrl(),
  interceptors: [localeInterceptor],
});

export function TransportProvider({ children }: Readonly<PropsWithChildren>) {
  return <ConnectTransportProvider transport={transport}>{children}</ConnectTransportProvider>;
}
