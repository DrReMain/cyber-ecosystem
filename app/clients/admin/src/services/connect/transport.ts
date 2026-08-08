import { createConnectTransport } from "@connectrpc/connect-web";
import { resolveCONNECTBaseUrl } from "#/env";

export const transport = createConnectTransport({
  baseUrl: resolveCONNECTBaseUrl(),
});
