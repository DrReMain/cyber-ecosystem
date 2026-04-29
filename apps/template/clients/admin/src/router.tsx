import { MutationCache, QueryCache, QueryClient } from "@tanstack/react-query"
import { createRouter } from "@tanstack/react-router"
import { setupRouterSsrQueryIntegration } from "@tanstack/react-router-ssr-query"
import { emitApiError, Pending, toApiError } from "#/domains/errors"
import { deLocalizeUrl, localizeUrl } from "#/paraglide/runtime"
import { routeTree } from "#/routeTree.gen"

export function getRouter() {
  const queryClient = new QueryClient({
    queryCache: new QueryCache({
      onError: (error, query) => {
        if (query.meta?.silent) return
        emitApiError(toApiError(error), "query")
      },
    }),
    mutationCache: new MutationCache({
      onError: (error, _vars, _ctx, mutation) => {
        if (mutation.meta?.silent) return
        emitApiError(toApiError(error), "mutation")
      },
    }),
    defaultOptions: {
      queries: {
        refetchOnWindowFocus: false,
        retry: false,
        gcTime: 0,
      },
    },
  })

  const router = createRouter({
    routeTree,
    context: { queryClient },
    rewrite: {
      input: ({ url }) => deLocalizeUrl(url),
      output: ({ url }) => localizeUrl(url),
    },
    scrollRestoration: true,
    defaultPreload: "intent",
    defaultPreloadStaleTime: 0,
    defaultPendingMs: 200,
    defaultPendingMinMs: 200,
    defaultPendingComponent: Pending,
    defaultViewTransition: true,
    search: { strict: true },
  })

  setupRouterSsrQueryIntegration({ router, queryClient })

  return router
}

declare module "@tanstack/react-router" {
  interface Register {
    router: ReturnType<typeof getRouter>
  }
}
