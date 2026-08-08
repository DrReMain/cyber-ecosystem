import { MutationCache, QueryCache, QueryClient } from "@tanstack/react-query";
import { createRouter as createTanStackRouter } from "@tanstack/react-router";
import { setupRouterSsrQueryIntegration } from "@tanstack/react-router-ssr-query";
import { PendingFallback } from "#/domains/errors/PendingFallback";
import { onMutationError, onQueryError } from "#/domains/errors/request-listen";
import { deLocalizeUrl, localizeUrl } from "#/paraglide/runtime";
import { routeTree } from "./routeTree.gen";

export function getRouter() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { throwOnError: true },
    },
    queryCache: new QueryCache({
      onError: onQueryError,
    }),
    mutationCache: new MutationCache({
      onError: (error, _variables, _onMutateResult, mutation) => onMutationError(error, mutation),
    }),
  });

  const router = createTanStackRouter({
    routeTree,
    context: { queryClient },
    scrollRestoration: true,
    defaultPreload: "intent",
    defaultPreloadStaleTime: 0,
    defaultViewTransition: true,
    defaultPendingComponent: PendingFallback,
    search: { strict: true },
    rewrite: {
      input: ({ url }) => deLocalizeUrl(url),
      output: ({ url }) => localizeUrl(url),
    },
  });

  setupRouterSsrQueryIntegration({ router, queryClient });

  return router;
}

declare module "@tanstack/react-router" {
  interface Register {
    router: ReturnType<typeof getRouter>;
  }
}
