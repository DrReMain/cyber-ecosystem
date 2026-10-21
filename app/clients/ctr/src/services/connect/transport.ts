import { addStaticKeyToTransport } from "@connectrpc/connect-query";
import { bindTransportProvider, createTransportFactory } from "@cyber-ecosystem/shared-transport";
import { normalize } from "#/domains/error";
import { env, getSiteUrl } from "#/env";
import { getLocale } from "#/paraglide/runtime";
import { createTestMockTransport } from "./test-mock";

const factory = createTransportFactory({
  getLocale,
  getSiteUrl,
  proxyPath: env.VITE_CONNECT_API_PROXY,
  getSessionNamespace: () => env.VITE_POPULATION,
  normalize,
});

export const createCtrTransport = factory.createTransport;
export const connectTransport = factory.connectTransport;

// The transport seam: DEV + VITE_TEST_MOCK=1 intercepts only the test
// service's read face and passes everything else through the real
// transport (same "connect" static key, so connect-query cache keys stay
// identical). Production never activates it.
// TODO(mock-removal): temporary gate — revert to the plain
// factory.connectTransport passthrough when the mock goes (test-mock.ts).
const providerTransport =
  import.meta.env.DEV && env.VITE_TEST_MOCK === "1"
    ? addStaticKeyToTransport(createTestMockTransport(factory.connectTransport), "connect")
    : factory.connectTransport;

export const TransportProvider = bindTransportProvider(providerTransport);
