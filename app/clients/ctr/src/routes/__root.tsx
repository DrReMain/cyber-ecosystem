import { RouterProgress } from "@cyber-ecosystem/shared-router-progress";
import { collectPersistedStores, JotaiProvider } from "@cyber-ecosystem/shared-store";
import { TanStackDevtools } from "@tanstack/react-devtools";
import type { QueryClient } from "@tanstack/react-query";
import { ReactQueryDevtoolsPanel } from "@tanstack/react-query-devtools";
import { createRootRouteWithContext, HeadContent, Scripts } from "@tanstack/react-router";
import { TanStackRouterDevtoolsPanel } from "@tanstack/react-router-devtools";
import { createServerFn } from "@tanstack/react-start";
import { getCookie } from "@tanstack/react-start/server";
import type { PropsWithChildren } from "react";
import { APP_NAME, HOME_PATH } from "#/config";
import { LocaleSelect } from "#/domains/i18n";
import { ErrorPage, FeedbackSonner, NotFoundPage } from "#/features/app";
import { sessionQuery } from "#/features/auth/session.fn";
import { AppLayout } from "#/features/layout-app/app-layout";
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

const getStoreCookies = createServerFn({ method: "GET" }).handler(async () =>
  collectPersistedStores((key) => getCookie(key)),
);

export const Route = createRootRouteWithContext<MyRouterContext>()({
  head: () => ({
    meta: [
      { charSet: "utf-8" },
      {
        name: "viewport",
        content:
          "width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no, viewport-fit=cover",
      },
      { name: "theme-color", content: "#000000" },
      { title: APP_NAME },
    ],
    links: [{ rel: "stylesheet", href: style }],
  }),
  beforeLoad: async ({ context }) => {
    if (typeof document !== "undefined") {
      document.documentElement.setAttribute("lang", getLocale());
      document.documentElement.setAttribute("dir", getTextDirection());
    }
    await context.queryClient.query(sessionQuery).catch(() => undefined);
  },
  component: AppLayout,
  loader: async () => {
    const storeData = await getStoreCookies();
    return { storeData };
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
  const { storeData } = Route.useLoaderData();
  return (
    <html
      dir={getTextDirection()}
      lang={getLocale()}
      suppressHydrationWarning={import.meta.env.PROD}
    >
      <head>
        <HeadContent />
      </head>
      <body>
        <JotaiProvider initialData={storeData}>
          <RouterProgress color="linear-gradient(90deg, #2563eb, #06b6d4, #34d399, #2563eb)" />
          <FeedbackSonner />
          <TransportProvider>
            <div className="flex min-h-svh flex-col">{children}</div>
          </TransportProvider>
        </JotaiProvider>
        <TanStackDevtools
          config={{ position: "bottom-right", panelLocation: "bottom" }}
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
