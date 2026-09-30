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
import { APP_NAME, HOME_PATH } from "#/config";
import { LocaleSelect } from "#/domains/i18n";
import { ErrorPage, FeedbackSonner, NotFoundPage } from "#/features/app";
import { getLocale, getTextDirection } from "#/paraglide/runtime";
import { TransportProvider } from "#/services/connect";
import { TailwindIndicator } from "../components/tailwind-indicator";
import style from "../styles/styles.css?url";
// Eager import: this is what registers every store. Route components are
// code-split, so each store's defineStore side effect would otherwise stay
// lazy and the persisted-store registry would be incomplete during SPA
// navigation.
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
      { title: APP_NAME },
    ],
    links: [{ rel: "stylesheet", href: style }],
  }),
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
  notFoundComponent: () => (
    <NotFoundPage appName={APP_NAME} headerTrailing={<LocaleSelect />} homePath={HOME_PATH} />
  ),
  errorComponent: ({ error }) => (
    <ErrorPage
      appName={APP_NAME}
      error={error}
      headerTrailing={<LocaleSelect />}
      homePath={HOME_PATH}
    />
  ),
  shellComponent: RootDocument,
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
            <RouterProgress color="linear-gradient(90deg, #c084fc, #60a5fa, #34d399, #c084fc)" />
            <FeedbackSonner />
            <TransportProvider>
              <div className="flex min-h-svh flex-col">{children}</div>
            </TransportProvider>
          </ThemeProvider>
        </JotaiProvider>
        <TanStackDevtools
          config={{ position: "bottom-right", panelLocation: "bottom" }}
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
