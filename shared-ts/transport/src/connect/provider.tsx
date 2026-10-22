import type { Transport } from "@connectrpc/connect";
import { TransportProvider as ConnectTransportProvider } from "@connectrpc/connect-query";
import type { PropsWithChildren } from "react";

// Binds a factory-built transport into the connect-query context so
// schema-first hooks ride the shared mechanism without per-client wiring.
export function bindTransportProvider(transport: Transport) {
  // biome-ignore lint/correctness/noComponentHookFactories: called once per transport instance at setup, never on a render path — the remount hazard this rule guards does not apply.
  return function TransportProvider({ children }: Readonly<PropsWithChildren>) {
    return <ConnectTransportProvider transport={transport}>{children}</ConnectTransportProvider>;
  };
}
