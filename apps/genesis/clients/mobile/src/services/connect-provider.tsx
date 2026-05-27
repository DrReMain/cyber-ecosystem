import { TransportProvider as ConnectTransportProvider } from "@connectrpc/connect-query"
import { createConnectTransport } from "@connectrpc/connect-web"
import type { PropsWithChildren } from "react"
import { env } from "@/env"

const transport = createConnectTransport({
  baseUrl: env.EXPO_PUBLIC_CONNECT_API_URL,
  useBinaryFormat: !__DEV__,
})

export function ConnectProvider({ children }: Readonly<PropsWithChildren>) {
  return (
    <ConnectTransportProvider transport={transport}>
      {children}
    </ConnectTransportProvider>
  )
}
