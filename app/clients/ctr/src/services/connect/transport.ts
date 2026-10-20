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
  normalize,
});

export const createCtrTransport = factory.createTransport;
export const connectTransport = factory.connectTransport;

// The transport seam: DEV + VITE_TEST_MOCK=1 intercepts only the test
// service's read face and passes everything else through the real
// transport (same "connect" static key, so connect-query cache keys stay
// identical). Production never activates it.
const providerTransport =
  import.meta.env.DEV && env.VITE_TEST_MOCK === "1"
    ? addStaticKeyToTransport(createTestMockTransport(factory.connectTransport), "connect")
    : factory.connectTransport;

export const TransportProvider = bindTransportProvider(providerTransport);
