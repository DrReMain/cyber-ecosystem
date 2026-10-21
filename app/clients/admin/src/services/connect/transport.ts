import { bindTransportProvider, createTransportFactory } from "@cyber-ecosystem/shared-transport";
import { normalize } from "#/domains/error";
import { env, getSiteUrl } from "#/env";
import { getLocale } from "#/paraglide/runtime";

const factory = createTransportFactory({
  getLocale,
  getSiteUrl,
  proxyPath: env.VITE_CONNECT_API_PROXY,
  getSessionNamespace: () => env.VITE_POPULATION,
  normalize,
});

export const createKratosTransport = factory.createTransport;
export const connectTransport = factory.connectTransport;
export const TransportProvider = bindTransportProvider(factory.connectTransport);
