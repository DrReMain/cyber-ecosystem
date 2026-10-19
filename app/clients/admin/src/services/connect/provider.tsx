import { TransportProvider as ConnectTransportProvider } from "@connectrpc/connect-query";
import type { PropsWithChildren } from "react";
import { connectTransport } from "./transport";

export function TransportProvider({ children }: Readonly<PropsWithChildren>) {
  return (
    <ConnectTransportProvider transport={connectTransport}>{children}</ConnectTransportProvider>
  );
}
