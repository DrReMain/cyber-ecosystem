import { RouterProgress } from "@cyber-ecosystem/shared-router-progress";
import { collectPersistedStores, JotaiProvider } from "@cyber-ecosystem/shared-store";
import { resolveThemeData, THEME_COOKIE_KEY, ThemeProvider } from "@cyber-ecosystem/shared-theme";
import { TanStackDevtools } from "@tanstack/react-devtools";
import type { QueryClient } from "@tanstack/react-query";
import { ReactQueryDevtoolsPanel } from "@tanstack/react-query-devtools";
import { createRootRouteWithContext, HeadContent, Scripts } from "@tanstack/react-router";
import { TanStackRouterDevtoolsPanel } from "@tanstack/react-router-devtools";
import { createServerFn } from "@tanstack/react-start";
import { getCookie, getRequestHeader } from "@tanstack/react-start/server";
import type { PropsWithChildren } from "react";
import { AntdProvider } from "#/domains/antd";
import { ErrorFallback } from "#/domains/errors/ErrorFallback";
import { errorHandler } from "#/domains/errors/error-handler";
import { ErrorListeners } from "#/domains/errors/listen";
import { NotFoundFallback } from "#/domains/errors/NotFoundFallback";
import { getLocale, getTextDirection } from "#/paraglide/runtime";
import { TransportProvider } from "#/services/connect";
import { FeedbackToaster } from "../components/feedback-toaster";
import { TailwindIndicator } from "../components/tailwind-indicator";
import style from "../styles/styles.css?url";
import "#/stores";

interface MyRouterContext {
  queryClient: QueryClient;
}

const getThemeFromServer = createServerFn({ method: "GET" }).handler(async () =>
  resolveThemeData(getCookie(THEME_COOKIE_KEY), getRequestHeader("sec-ch-prefers-color-scheme")),
);

const getStoreCookies = createServerFn({ method: "GET" }).handler(async () =>
  collectPersistedStores((key) => getCookie(key)),
);

export const Route = createRootRouteWithContext<MyRouterContext>()({
  head: () => ({
    meta: [
      { charSet: "utf-8" },
      { name: "viewport", content: "width=device-width, initial-scale=1" },
      { title: "Cyber Ecosystem Admin" },
    ],
    links: [{ rel: "stylesheet", href: style }],
  }),
  notFoundComponent: NotFoundFallback,
  errorComponent: ({ error, reset }) => {
    errorHandler.handle(error, "render", { feedback: false });
    return (
      <ErrorFallback
        error={error}
        fullScreen
        resetErrorBoundary={reset}
        showDetails={!!import.meta.env.DEV}
      />
    );
  },
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
        <ErrorListeners />
        <JotaiProvider initialData={storeData}>
          <ThemeProvider initialTheme={themeData}>
            <RouterProgress />
            <FeedbackToaster />
            <TransportProvider>
              <AntdProvider>{children}</AntdProvider>
            </TransportProvider>
          </ThemeProvider>
        </JotaiProvider>
        <TanStackDevtools
          config={{ position: "bottom-left", panelLocation: "bottom" }}
          plugins={[
            { name: "Tanstack Router", render: <TanStackRouterDevtoolsPanel /> },
            { name: "TanStack Query", render: <ReactQueryDevtoolsPanel /> },
          ]}
        />
        <Scripts />
        <TailwindIndicator />
      </body>
    </html>
  );
}
