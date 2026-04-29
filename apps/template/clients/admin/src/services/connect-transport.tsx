import type { Interceptor } from "@connectrpc/connect"
import { ConnectError } from "@connectrpc/connect"
import { TransportProvider as ConnectTransportProvider } from "@connectrpc/connect-query"
import { createConnectTransport } from "@connectrpc/connect-web"
import type { PropsWithChildren } from "react"
import {
  extractReason,
  getErrorHandler,
  grpcCodeToHttpStatus,
} from "#/domains/errors"
import { env, resolveApiBaseUrl } from "#/env"
import { getLocale } from "#/paraglide/runtime"

const localeInterceptor: Interceptor = (next) => async (req) => {
  req.header.set("Accept-Language", getLocale())
  return next(req)
}

const errorInterceptor: Interceptor = (next) => async (req) => {
  try {
    return await next(req)
  } catch (error) {
    if (error instanceof ConnectError) {
      const reason = extractReason(error)
      const handler = getErrorHandler(reason)
      if (handler) {
        handler({
          code: grpcCodeToHttpStatus[error.code] ?? 500,
          reason,
          message: error.rawMessage,
          metadata: Object.fromEntries(error.metadata.entries()),
        })
      }
    }
    throw error
  }
}

export const transport = createConnectTransport({
  baseUrl: resolveApiBaseUrl(env.CONNECT_API_URL, "/connect"),
  interceptors: [localeInterceptor, errorInterceptor],
})

export function TransportProvider({ children }: Readonly<PropsWithChildren>) {
  return (
    <ConnectTransportProvider transport={transport}>
      {children}
    </ConnectTransportProvider>
  )
}
