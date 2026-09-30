import { QueryClient } from "@tanstack/react-query";
import { createRouter as createTanStackRouter, useRouter } from "@tanstack/react-router";
import { setupRouterSsrQueryIntegration } from "@tanstack/react-router-ssr-query";
import {
  capture,
  errorMutationCache,
  errorQueryCache,
  shouldRetry,
  shouldThrow,
} from "#/domains/error";
import { ErrorFallback, PendingFallback } from "#/features/app";
import { deLocalizeUrl, localizeUrl } from "#/paraglide/runtime";
import { routeTree } from "./routeTree.gen";

export function getRouter() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { throwOnError: shouldThrow, retry: shouldRetry },
    },
    queryCache: errorQueryCache(),
    mutationCache: errorMutationCache(),
  });

  const router = createTanStackRouter({
    routeTree,
    context: { queryClient },
    defaultPreload: "intent",
    defaultPreloadStaleTime: 0,
    defaultViewTransition: true,
    defaultPendingComponent: PendingFallback,
    defaultErrorComponent: ({ error, reset }) => {
      const r = useRouter();
      return (
        <ErrorFallback
          error={error}
          onRetry={() => {
            reset();
            r.invalidate();
          }}
        />
      );
    },
    defaultOnCatch: (error) => capture(error, { feedback: false }),
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
