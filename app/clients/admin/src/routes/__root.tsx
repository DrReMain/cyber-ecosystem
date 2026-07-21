import { TanStackDevtools } from "@tanstack/react-devtools";
import type { QueryClient } from "@tanstack/react-query";
import { ReactQueryDevtoolsPanel } from "@tanstack/react-query-devtools";
import { createRootRouteWithContext, HeadContent, Scripts } from "@tanstack/react-router";
import { TanStackRouterDevtoolsPanel } from "@tanstack/react-router-devtools";
import type { PropsWithChildren } from "react";
import { AntdProvider } from "#/domains/antd";
import { ErrorPage, NotFound } from "#/domains/errors";
import { RouterProgress } from "#/domains/router-progress";
import { getThemeFromServer, ThemeProvider } from "#/domains/theme";
import { getLocale, getTextDirection } from "#/paraglide/runtime";
import { TransportProvider } from "#/services/connect-transport";
import { JotaiProvider } from "#/stores/_core/provider";
import { getStoreCookies } from "#/stores/_core/server";
import { TailwindIndicator } from "../components/tailwind-indicator";
import style from "../styles/styles.css?url";
import "#/stores";

interface MyRouterContext {
  queryClient: QueryClient;
}

export const Route = createRootRouteWithContext<MyRouterContext>()({
  head: () => ({
    meta: [
      { charSet: "utf-8" },
      { name: "viewport", content: "width=device-width, initial-scale=1" },
      { title: "Cyber Ecosystem Admin" },
    ],
    links: [{ rel: "stylesheet", href: style }],
  }),
  notFoundComponent: NotFound,
  errorComponent: ({ error }: { error: Error }) => (
    <ErrorPage error={error} showDetails={!!import.meta.env.DEV} />
  ),
  shellComponent: RootDocument,
  beforeLoad: () => {
    if (typeof document !== "undefined") {
      document.documentElement.setAttribute("lang", getLocale());
      document.documentElement.setAttribute("dir", getTextDirection());
    }
  },
  loader: async () => {
    const [themeData, storeData] = await Promise.all([getThemeFromServer(), getStoreCookies()]);
    return { themeData, storeData };
  },
});

function RootDocument({ children }: Readonly<PropsWithChildren>) {
  const { themeData, storeData } = Route.useLoaderData();
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
            <RouterProgress />
            <TransportProvider>
              <AntdProvider>{children}</AntdProvider>
            </TransportProvider>
          </ThemeProvider>
        </JotaiProvider>
        <TanStackDevtools
          config={{ position: "bottom-left", panelLocation: "bottom" }}
          plugins={[
            { name: "Tanstack Router", render: <TanStackRouterDevtoolsPanel /> },
            { name: "Tanstack Query", render: <ReactQueryDevtoolsPanel /> },
          ]}
        />
        <Scripts />
        <TailwindIndicator />
      </body>
    </html>
  );
}
