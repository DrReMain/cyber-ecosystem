import { TanStackDevtools } from "@tanstack/react-devtools"
import type { QueryClient } from "@tanstack/react-query"
import { ReactQueryDevtoolsPanel } from "@tanstack/react-query-devtools"
import {
  createRootRouteWithContext,
  HeadContent,
  Scripts,
} from "@tanstack/react-router"
import { TanStackRouterDevtoolsPanel } from "@tanstack/react-router-devtools"
import type { PropsWithChildren } from "react"
import { TailwindIndicator } from "#/components/tailwind-indicator"
import { AntdProvider } from "#/domains/antd"
import {
  AntdErrorFeedbackAdapter,
  ErrorPage,
  initSentry,
  NotFound,
  registerErrorHandlers,
  reportError,
} from "#/domains/errors"
import { initTracing } from "#/domains/errors/tracing"
import { RouterProgress } from "#/domains/router-progress"
import { ThemeProvider } from "#/domains/theme"
import { getThemeFromServer } from "#/domains/theme/cookie"
import { m } from "#/paraglide/messages"
import { getLocale, getTextDirection } from "#/paraglide/runtime"
import { TransportProvider } from "#/services/connect-transport"
import { JotaiProvider } from "#/stores/_core/provider"
import { getStoreCookies } from "#/stores/_core/server"
import style from "#/styles/styles.css?url"

interface MyRouterContext {
  queryClient: QueryClient
}

export const Route = createRootRouteWithContext<MyRouterContext>()({
  beforeLoad: async () => {
    if (typeof document !== "undefined") {
      document.documentElement.setAttribute("lang", getLocale())
      document.documentElement.setAttribute("dir", getTextDirection())
    }
  },
  loader: async () => {
    const themeData = await getThemeFromServer()
    const storeData = await getStoreCookies()
    return { themeData, storeData }
  },
  notFoundComponent: NotFound,
  errorComponent: ({ error }: { error: Error }) => {
    reportError(error, { source: "route" })
    return <ErrorPage error={error} showDetails={!!import.meta.env.DEV} />
  },
  head: () => ({
    meta: [
      { charSet: "utf-8" },
      { name: "viewport", content: "width=device-width, initial-scale=1" },
      { title: m.app_title() },
    ],
    links: [{ rel: "stylesheet", href: style }],
  }),
  shellComponent: RootDocument,
})

function RootDocument({ children }: Readonly<PropsWithChildren>) {
  const { themeData, storeData } = Route.useLoaderData()

  initSentry()
  initTracing()
  registerErrorHandlers()

  return (
    <html
      className={themeData.preference === "dark" ? "dark" : ""}
      dir={getTextDirection()}
      lang={getLocale()}
    >
      <head>
        <HeadContent />
      </head>
      <body>
        <JotaiProvider initialData={storeData}>
          <ThemeProvider initialTheme={themeData}>
            <TransportProvider>
              <RouterProgress />
              <AntdProvider>
                <AntdErrorFeedbackAdapter />
                {children}
              </AntdProvider>
            </TransportProvider>
          </ThemeProvider>
        </JotaiProvider>
        <TanStackDevtools
          config={{
            position: "bottom-left",
            panelLocation: "bottom",
          }}
          plugins={[
            {
              name: "Tanstack Router",
              render: <TanStackRouterDevtoolsPanel />,
            },
            {
              name: "Tanstack Query",
              render: <ReactQueryDevtoolsPanel />,
            },
          ]}
        />
        <Scripts />
        <TailwindIndicator />
      </body>
    </html>
  )
}
