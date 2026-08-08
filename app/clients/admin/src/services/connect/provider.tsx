import { TransportProvider as ConnectTransportProvider } from "@connectrpc/connect-query";
import type { PropsWithChildren } from "react";
import { transport } from "./transport";

export function TransportProvider({ children }: Readonly<PropsWithChildren>) {
  return <ConnectTransportProvider transport={transport}>{children}</ConnectTransportProvider>;
}
