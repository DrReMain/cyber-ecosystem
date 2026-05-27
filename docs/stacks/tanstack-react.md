# TanStack + React Stack Guide

Guide for React applications built with TanStack Start, TanStack Router,
TanStack Query, Ant Design, Paraglide, Vite, pnpm, and Nx in this monorepo.

---

## 1. Scope

This guide is stack-level only. Product, client, page, and domain conventions
belong near the owning application.

Applies to React clients that use:

- React
- TanStack Start
- TanStack Router
- TanStack Query
- Ant Design
- Paraglide i18n
- Vite
- pnpm
- Nx

---

## 2. Toolchain

The toolchain has clear ownership boundaries:

| Tool | Responsibility |
|------|----------------|
| Nx | Monorepo orchestration and declared project workflows |
| pnpm | Package manager and script runner |
| Vite | Dev server, bundling, and plugin pipeline |
| TanStack Start | React application runtime and SSR entry |
| TanStack Router | Type-safe routing, route loading, and URL state |
| TanStack Query | Server-state caching and synchronization |
| Ant Design | Component system and design tokens |
| Paraglide | Type-safe messages, locale runtime, and localized URLs |
| Biome | Formatting and static checks |

### pnpm Isolated Mode

The monorepo uses pnpm's default isolated mode (`node-linker=isolated`). Each workspace package has its own `node_modules/` with only its declared dependencies. This eliminates the phantom dependency and React version collision problems that occur with hoisted mode in React Native + React web monorepos.

Expo SDK 54+ officially supports pnpm isolated mode.

`pnpm.overrides` in the root `package.json` is reserved for transitive dependency conflicts that cannot be resolved at the package level (e.g., `lightningcss` version pinning for Tailwind/Vite compatibility). Overrides are root-only — sub-projects cannot define them.

Use Nx for recurring workflows. Read the owning `project.json` and run declared
targets:

```bash
./nx run <project>:build
./nx run <project>:biome:check
./nx run <project>:messages:check
```

Do not bypass Nx with direct pnpm or Vite commands when a target exists. If a
workflow is repeated, expose it as an Nx target.

---

## 3. Architecture Principles

### Server-first when visible

Anything visible in the initial document should be resolved before render when
possible: locale, text direction, theme preference, metadata, and critical route
data. Browser effects are for synchronization after hydration, not for correcting
SSR output.

### Route owns page composition

Routes own URL state, loaders, page metadata, pending/error/not-found boundaries,
and page composition. Shared components, API clients, query helpers, and design
system code live outside `src/routes/`.

### Separate server state from UI state

Use TanStack Query for remote server state. Use Jotai stores for persisted or
shared client state. Keep transient interaction state in React.

### Centralize cross-cutting concerns

Theme, locale, direction, Ant Design config, CSS-in-JS registry, error handling,
and typed environment access are stack concerns. Configure them once in the shell/provider
layer instead of repeating them in feature components.

---

## 4. Application Shell

The root route is the document shell. It owns:

- `<html>` attributes (lang, dir)
- `<head>` metadata and stylesheet links
- root loaders for cookie-based provider hydration (theme, stores)
- provider composition
- scripts and devtools

### Shell vs Route Component

TanStack Start splits the root route into a `shellComponent` (renders `<html>`,
`<head>`, `<body>`) and the route content. The shell handles document-level
rendering; the route handles page content.

### Provider Nesting Order

```
JotaiProvider(initialData=storeData)    // Store hydration first (lowest level)
  ThemeProvider(initialTheme=themeData)  // Theme before UI libraries
    TransportProvider                    // API clients before data consumers
      RouterProgress                     // Progress bar watches router
        AntdProvider                     // Ant Design last (consumes theme/locale)
          AntdErrorFeedbackAdapter       // Error UI adapter
            {children}
```

Order reflects dependency: Jotai has no deps, theme depends on nothing,
transport depends on env, Antd depends on theme+locale.

### Root Loader (SSR Hydration)

The root loader reads cookies and passes initial data to providers:

- `themeData` — cookie-stored theme preference (skin, mode, compact)
- `storeData` — cookie-stored Jotai store values

This ensures SSR output matches client state (no flash).

---

## 5. Routing

Use TanStack Router file-based routes. Keep router setup typed and centralized.

### Router Configuration

```ts
// Key defaults:
scrollRestoration: true
defaultPreload: "intent"
defaultPreloadStaleTime: 0
defaultPendingMs: 200
defaultPendingMinMs: 200
defaultPendingComponent: Pending
defaultViewTransition: true
search: { strict: true }
```

### URL Rewriting for i18n

Paraglide's `localizeUrl`/`deLocalizeUrl` are wired into the router's
`rewrite.input`/`rewrite.output` hooks. URLs are localized in the browser,
internal matching uses base locale paths.

### Layout Routes

Pathless layout routes (`_app.tsx`) wrap child routes with shared UI (nav,
sidebar, footer). Use `viewTransitionName` on the `<Outlet>` wrapper for
animated transitions.

### SSR Query Integration

`setupRouterSsrQueryIntegration({ router, queryClient })` connects TanStack
Router's loader system with TanStack Query for SSR data fetching.

Rules:

- Provide `QueryClient` through router context.
- Keep `search` strict unless the local app has a deliberate compatibility reason.
- Treat generated route trees as derived output.

---

## 6. Dual API Backend

The app talks to two backend APIs with independent transport layers:

| Channel | Transport | Code Gen | Path |
|---------|-----------|----------|------|
| Connect RPC | `@connectrpc/connect-web` | `@connectrpc/protoc-gen-connect-query` | `/connect` |
| HTTP REST | `ky` via OpenAPI | `@hey-api/openapi-ts` | `/http` |

### Connect Transport (`services/connect-transport.tsx`)

- Creates a `ConnectTransport` with `@connectrpc/connect-web`
- Base URL: `resolveApiBaseUrl(env.CONNECT_API_URL, "/connect")` — full URL on server, relative on client
- Interceptors: locale injection (`Accept-Language`), error handling (extracts reason code)
- Wrapped in `ConnectTransportProvider` from `@connectrpc/connect-query` for TanStack Query integration

### HTTP Transport (`services/http-config.ts` + `services/http-client.ts`)

- OpenAPI-generated client configured via `createClientConfig`
- Base URL: `resolveApiBaseUrl(env.HTTP_API_URL, "/http")`
- Interceptors: locale header, error normalization
- `kyOptions: { throwHttpErrors: false }` — errors returned, not thrown

### File Client (`services/file-client.ts`)

- Standalone `ky` instance for file upload/download (not OpenAPI-generated)
- `uploadFile(file)` — FormData POST
- `downloadFile(id)` — returns Blob
- `getDownloadUrl(id)` — returns direct URL

### Pagination Builders (`lib/builder.ts`)

`buildConnectPage` and `buildHTTPPage` transform the same `PaginationInput`
into different shapes for each backend. `buildOrderBy` converts sort parameters.

---

## 7. Data Fetching

Use loaders for route-critical data and TanStack Query for cacheable server state.
Configure SSR query integration when the router is created.

### QueryClient Defaults

```ts
refetchOnWindowFocus: false
retry: false
gcTime: 0
```

Aggressive defaults: no auto-refetch, no retries, no garbage collection timer.
Each query controls its own stale time.

### Query/Mutation Error Pipeline

Both `QueryCache` and `MutationCache` have `onError` handlers:

```ts
onError: (error, query) => {
  if (query.meta?.silent) return  // suppress per-query
  emitApiError(toApiError(error), "query")
}
```

- `toApiError()` normalizes `ConnectError`, HTTP errors, and unknown errors into a unified `ApiError`
- `emitApiError()` fires an event that `AntdErrorFeedbackAdapter` listens to
- `meta?.silent` suppresses error feedback for specific queries/mutations

Rules:

- Create one router-scoped `QueryClient`.
- Keep query keys structured and stable.
- Own query keys near the API/query helper that defines the data contract.
- Do not create unrelated `QueryClient` instances inside feature components.

---

## 8. Error Pipeline

The error system has three layers: normalize → display → report.

### Normalization (`domains/errors/api-error.ts`)

`toApiError(error)` converts any error into a typed `ApiError` with `code`,
`reason`, `message`, and `metadata`. Handles Connect errors, HTTP errors,
and unknown errors.

### Display (`domains/errors/antd-feedback.tsx`)

`AntdErrorFeedbackAdapter` subscribes to `onApiError` events:

- **Mutation errors** → `message.error()` (toast)
- **Query errors** → `notification.error()` (notification panel)

Queries that should not show feedback set `meta: { silent: true }`.

### Error Handlers (`domains/errors/setup-handlers.ts`)

Domain-specific error handlers register by reason code:

```ts
registerErrorHandler("FLOW_ERROR_RATE_LIMITED", (error) => { /* custom handling */ })
```

Both Connect and HTTP transports route errors through these handlers before
re-throwing.

### Reporting (`domains/errors/report.ts`)

`reportError(error, { source })` normalizes and sends to Sentry. Deduplicates
via `WeakSet`. Sentry is lazy-initialized from `VITE_GLITCHTIP_DSN`.

### Error Page

The root route's `errorComponent` renders a full-page error with "Back Home"
and "Retry" actions. Expandable error details are shown in dev mode only.

---

## 9. UI And Styling

### Ant Design Provider Stack

Three-layer nesting in `domains/antd/provider.tsx`:

```
AntdRegistry          // SSR CSS-in-JS extraction
  ConfigProvider      // Locale, direction, global defaults
    SkinSwitcher      // Theme skin (light/dark/compact)
      App             // Static methods (message, notification)
```

ConfigProvider sets global defaults:
- `form.requiredMark: "optional"` — marks optional instead of required
- `input.autoComplete: "off"`, `input.allowClear: true`
- Screen breakpoints mapped to Tailwind's 5-tier system

### Skin System

Skins are pluggable themes from `@shared/antd/skins`. Each skin defines a
complete Ant Design theme (algorithm, tokens, components). The `SkinSwitcher`
receives `skinId`, `isDark`, `compact` and wraps children with the appropriate
theme. Not just light/dark — full skin swaps with compact mode.

### Antd + Tailwind Integration

Antd CSS variables are available as Tailwind utilities:
`bg-antd-base`, `border-antd-border-secondary`, `hover:bg-antd-fill`, etc.

Styling defaults:

- Prefer Ant Design tokens and component APIs for Ant Design surfaces.
- Use Tailwind for layout and non-Ant Design surfaces.
- Keep global CSS limited to resets, variables, and stack-wide base behavior.

Universal Ant Design overrides belong in the provider/theme layer, not in feature
components.

---

## 10. State Management (Jotai)

Use `defineStore` from `stores/_core/define-store` for client state that needs
persistence or sharing across the component tree.

### Store Definition

```ts
const store = defineStore("store_counter", 0, {
  persist: true,       // write to cookie on every change
  debugLabel: "Counter",
  schema: z.number(),  // optional Zod validation
})
```

Each store creates:
- `immerAtom` — internal Immer-powered atom for draft-based updates
- `atom` — public atom that handles persistence (writes cookie if `persist: true`)

The public atom's `set` accepts either a new value or an Immer draft callback.

### Store Registry

All stores are registered in a global `registry`. The root loader iterates the
registry to read cookies, and the `JotaiProvider` hydrates initial values from
cookie data. This is the SSR hydration mechanism for stores.

### When to Use

- **TanStack Query** — server state (API data)
- **Jotai stores** — client state that persists (cookies) or shares across distant components
- **React state** — transient interaction state (modals, form inputs, hover)

---

## 11. Internationalization

Paraglide owns messages, locale runtime, text direction, and localized URL
behavior. Generated files under `src/paraglide/` are derived output.

Message workflow:

```bash
./nx run <project>:messages:gen
./nx run <project>:messages:check
```

Use generated message functions from `#/paraglide/messages`. Do not hardcode
document `lang`, document `dir`, or localized URL behavior in feature components.

---

## 12. Environment

Use `@t3-oss/env-core` for typed environment configuration.

### Variable Categories

```ts
shared: { HTTP_API_URL, CONNECT_API_URL }  // server + client
client: { VITE_SITE_URL, VITE_GLITCHTIP_DSN, VITE_OTEL_URL }  // client only (VITE_ prefix)
```

### SSR Proxy Pattern

`resolveApiBaseUrl(serverUrl, clientPath)` returns different values based on context:

- **Server** (`typeof window === "undefined"`) — full URL (e.g., `http://base:11000`)
- **Client** — relative path (e.g., `/connect`, `/http`) — proxied by Vite dev server or Caddy

Rules:

- Client-exposed variables must use the `VITE_` prefix.
- Server-only values must not be exposed through client-prefixed variables.
- Validation belongs in the environment schema, not at scattered call sites.

---

## 13. Theme

`domains/theme/` provides theme state with three dimensions:

- `skinId` — which skin to use
- `mode` — `"light"` | `"dark"` | `"system"`
- `compact` — boolean compact mode

System preference is tracked via `matchMedia("(prefers-color-scheme: dark)")`
with real-time sync. Theme state is persisted to cookies for SSR hydration.
Dark mode toggles `document.documentElement.classList.add("dark")` for Tailwind.

---

## 14. Router Progress

`domains/router-progress/` provides a custom progress bar that subscribes to
router events:

- `onBeforeLoad` — starts animation (checks `pathChanged` to skip same-page)
- `onResolved` — finishes animation

Uses direct DOM manipulation with `requestAnimationFrame` for smooth animation.
Gradient rainbow bar with trickle effect.

---

## 15. SEO

`domains/seo/` provides server-side sitemap generation:

- Iterates `sitemapRoutes` to build XML `<url>` entries
- Uses `localizeUrl()` from Paraglide for `hreflang` alternate links
- `robots.ts` for robots.txt generation

---

## 16. Generated Files

Do not manually edit generated output, including:

- `src/routeTree.gen.ts`
- `src/paraglide/*`
- `src/services/connect/` (Connect/Protobuf generated clients)
- `src/services/openapi/` (OpenAPI generated clients)
- files explicitly marked as generated

Fix the source route, message schema, proto file, generator configuration,
or Nx target, then regenerate.

---

## 17. Utility Libraries

### `lib/use-utils.tsx`

Composable table column builders:
- `fieldTimestamp` — formatted time column with mono font
- `fieldAction` — action button column
- `fieldCopy` — copyable text column

SSR-safe time formatting via `useSyncExternalStore` for client detection.

### `lib/builder.ts`

Dual-pagination builders (`buildConnectPage`, `buildHTTPPage`) and sort
parameter conversion (`buildOrderBy`).

### `lib/cookie.ts`

Cookie read/write helpers for SSR hydration.

---

## 18. Validation

Before closing stack-level frontend changes, run the relevant declared Nx targets:

```bash
./nx run <project>:biome:check
./nx run <project>:messages:check
./nx run <project>:build
```

Run message generation first when i18n sources changed. If a target is missing,
use the closest owning workflow and record the gap.

---

## 19. Project Structure Reference

```
src/
  routes/                             # TanStack Router file-based routes
    __root.tsx                        # Document shell, providers, root loader
    _app.tsx                          # Layout route (nav, sidebar)
    _app.index.tsx                    # Home page
    _app.playground.*.tsx             # Feature pages
  router.tsx                          # Router setup, QueryClient config
  server.ts                           # TanStack Start server entry
  env.ts                              # Typed env (t3-env)
  domains/                            # Cross-cutting domain modules
    antd/                             # Ant Design configuration
      provider.tsx                    # Three-layer provider stack
      registry.tsx                    # SSR CSS-in-JS extraction
      skins.ts                        # Re-export from @shared/antd
      locale.ts                       # Locale mapping
    errors/                           # Error handling system
      api-error.ts                    # ApiError class, normalization
      error-events.ts                 # Event bus for error pipeline
      antd-feedback.tsx               # Mutation=toast, Query=notification
      setup-handlers.ts               # Reason-code handler registration
      report.ts                       # Sentry reporting (deduplicated)
      error.tsx                       # Full-page error component
      not-found.tsx                   # 404 page
      pending.tsx                     # Loading/pending component
    i18n/                             # i18n UI helpers
      locale-switcher.tsx             # Language switcher
    router-progress/                  # Route transition progress bar
      progress.tsx                    # Custom animated progress bar
      config.ts                       # Animation configuration
    seo/                              # SEO utilities
      sitemap.ts                      # XML sitemap with hreflang
      robots.ts                       # robots.txt
    theme/                            # Theme state management
      provider.tsx                    # Theme context + system sync
      hooks.ts                        # Theme hooks
      cookie.ts                       # Cookie persistence
      toggle.tsx                      # Theme toggle button
  services/                           # API client layer
    connect-transport.tsx             # Connect RPC transport + interceptors
    http-client.ts                    # OpenAPI HTTP client
    http-config.ts                    # OpenAPI client configuration
    file-client.ts                    # File upload/download (ky)
    connect/                          # Generated Connect/Protobuf code
    openapi/                          # Generated OpenAPI client code
  stores/                             # Jotai state management
    _core/                            # Store infrastructure
      define-store.ts                 # defineStore factory
      provider.tsx                    # JotaiProvider with SSR hydration
      server.ts                       # Server-side store utilities
    counter/store.ts                  # Example store
    todolist/store.ts                  # Example store
  lib/                                # Shared utilities
    builder.ts                        # Pagination + sort builders
    cookie.ts                         # Cookie helpers
    use-utils.tsx                     # Table column builders, time formatting
  components/                         # Shared UI components
```
