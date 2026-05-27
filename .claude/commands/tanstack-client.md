# TanStack Start React Web Client

Guide for scaffolding and implementing React web clients using TanStack Start in this monorepo. Use this when creating a new web client, adding pages, connecting to backend APIs, or configuring i18n/theme/error handling.

For architecture reference, utility libraries, and project structure, see `docs/stacks/tanstack-react.md`.

---

## 1. Client Directory Structure

```
apps/<app>/clients/<client>/
  src/
    routes/                   # TanStack Router file-based routes
      __root.tsx              # Document shell, providers, root loader
      _app.tsx                # Layout route (nav, sidebar)
      _app.index.tsx          # Home page
      _app.<feature>.*.tsx    # Feature pages
    router.tsx                # Router setup, QueryClient config
    server.ts                 # TanStack Start server entry
    env.ts                    # Typed env (t3-env)
    domains/                  # Cross-cutting domain modules
      antd/                   # Ant Design provider stack
      errors/                 # Error handling system
      i18n/                   # i18n UI helpers
      router-progress/        # Route transition progress bar
      seo/                    # SEO utilities
      theme/                  # Theme state management
    services/                 # API client layer
      connect-transport.tsx   # Connect RPC transport
      http-config.ts          # OpenAPI client configuration
      http-client.ts          # OpenAPI HTTP client
      connect/                # Generated Connect code
      openapi/                # Generated OpenAPI code
    stores/                   # Jotai state management
    lib/                      # Shared utilities
    components/               # Shared UI components
    styles/                   # Global CSS
  i18n/
    messages/                 # Paraglide message source files
  project.inlang/            # Paraglide project config
  public/                    # Static assets
  vite.config.ts
  hey-api.config.ts          # OpenAPI code generation config
  buf.gen.client.yaml        # Connect code generation config
  project.json               # Nx targets
  package.json
  tsconfig.json
  biome.json
```

---

## 2. Adding a New Web Client

When adding a new web client (e.g., "admin", "portal") to an app:

### Step 1: Create Directory Structure

Create `apps/<app>/clients/<client>/` with the directory layout above. Copy an existing client as a starting point if one exists in the same app.

### Step 2: Configure package.json

Key dependencies:

```json
{
  "name": "<app>_client_<client>",
  "dependencies": {
    "@bufbuild/protobuf": "<version>",
    "@connectrpc/connect": "<version>",
    "@connectrpc/connect-query": "<version>",
    "@connectrpc/connect-web": "<version>",
    "@hey-api/openapi-ts": "<version>",
    "@shared/antd": "workspace:*",
    "@tanstack/react-query": "^5",
    "@tanstack/react-router": "^1",
    "@tanstack/react-start": "^1",
    "antd": "^6",
    "ky": "^2",
    "react": "<pinned-version>",
    "react-dom": "<pinned-version>"
  },
  "devDependencies": {
    "@inlang/paraglide-js": "^2",
    "@tailwindcss/vite": "^4",
    "@vitejs/plugin-react": "^6",
    "tailwindcss": "^4",
    "vite": "^8"
  }
}
```

### Step 3: Configure Code Generation

**Connect clients** (`buf.gen.client.yaml`):

```yaml
version: v2
plugins:
  - local: ./node_modules/.bin/protoc-gen-es
    out: apps/<app>/clients/<client>/src/services/connect
    opt: [target=ts]
  - local: ./node_modules/.bin/protoc-gen-connect-es
    out: apps/<app>/clients/<client>/src/services/connect
    opt: [target=ts]
  - local: ./node_modules/.bin/protoc-gen-connect-query
    out: apps/<app>/clients/<client>/src/services/connect
    opt: [target=ts]
```

**OpenAPI clients** (`hey-api.config.ts`):

```ts
import { defineConfig } from "@hey-api/openapi-ts"

export default defineConfig({
  input: "../../../<app>/gen/oas/openapi.yaml",
  output: "src/services/openapi",
  plugins: [
    {
      name: "@hey-api/client-ky",
      runtimeConfigPath: "./src/services/http-config",
    },
  ],
})
```

### Step 4: Configure Vite

`vite.config.ts` must include the plugin chain in this order:

```ts
plugins: [
  TanstackDevtools(),
  paraglideVitePlugin({
    project: "./project.inlang",
    outdir: "./src/paraglide",
    strategy: ["custom-smart-preferred", "url", "baseLocale"],
  }),
  tailwindcss(),
  tanstackStart(),
  viteReact(),
  babel({ presets: [reactCompilerPreset()] }),
]
```

Configure dev proxy for backend services:

```ts
server: {
  proxy: {
    "/connect": {
      target: env.CONNECT_API_URL || "http://localhost:<connect-port>",
      changeOrigin: true,
      rewrite: (path) => path.replace(/^\/connect/, ""),
    },
    "/http": {
      target: env.HTTP_API_URL || "http://localhost:<http-port>",
      changeOrigin: true,
      rewrite: (path) => path.replace(/^\/http/, ""),
    },
  },
}
```

### Step 5: Configure Nx Targets

In `project.json`, declare standard targets:

```json
{
  "targets": {
    "proto:connect": {
      "executor": "nx:run-commands",
      "options": {
        "cwd": "{workspaceRoot}",
        "command": "buf generate --template apps/<app>/clients/<client>/buf.gen.client.yaml --path apps/<app>/api --include-imports"
      }
    },
    "proto:openapi": {
      "executor": "nx:run-commands",
      "options": {
        "cwd": "{projectRoot}",
        "command": "openapi-ts -f hey-api.config.ts"
      }
    },
    "dev": { "executor": "nx:run-commands", "options": { "cwd": "{projectRoot}", "command": "vite dev --host" } },
    "build": { "executor": "nx:run-commands", "options": { "cwd": "{projectRoot}", "command": "vite build" } }
  }
}
```

Note: `proto:connect` runs from `{workspaceRoot}` because buf generators are installed as root devDependencies.

### Step 6: Wire Into the Workspace

1. `pnpm install` — pnpm workspace auto-discovers the new package via `apps/**` glob
2. Generate Connect clients: `./nx run <project>:proto:connect`
3. Generate OpenAPI clients: `./nx run <project>:proto:openapi`
4. Run dev: `./nx run <project>:dev`

---

## 3. Dual API Backend Pattern

Web clients connect to BFF services through two independent transport layers. Both layers must be configured for a fully functional client.

### Connect RPC (Primary)

Transport setup in `src/services/connect-transport.tsx`:
- Creates `ConnectTransport` with `@connectrpc/connect-web`
- Base URL: full URL on server, `/connect` on client (proxied by Vite dev server)
- Interceptors: locale injection, error handling

Generated query hooks via `@connectrpc/protoc-gen-connect-query`. Usage:

```ts
import { useQuery } from "@connectrpc/connect-query"
import { listArticles } from "#/services/connect/<app>V1-<service>Service-Connect"

const { data } = useQuery(listArticles, { pageSize: 10 })
```

### HTTP REST (Secondary)

Generated client via `@hey-api/openapi-ts` with `ky` runtime. Config in `http-config.ts`:
- Base URL: full URL on server, `/http` on client (proxied by Vite dev server)

Usage:

```ts
import { adminArticleServiceListArticle } from "#/services/openapi/CustomClient"

const { data } = await adminArticleServiceListArticle({ queries: { "page.size": 10 } })
```

### Pagination Builders

`lib/builder.ts` provides `buildConnectPage` and `buildHTTPPage` to transform pagination parameters for each transport, plus `buildOrderBy` for sort parameters.

---

## 4. i18n Workflow (Paraglide)

Message definitions in `i18n/messages/` are the source of truth. The `project.inlang/` directory configures the Paraglide project.

Workflow:
1. Add/edit messages in `i18n/messages/` source files
2. Run `./nx run <project>:messages:gen` to generate Paraglide code
3. Use generated functions from `#/paraglide/messages`
4. Run `./nx run <project>:messages:check` to validate

Generated files under `src/paraglide/` are derived output — never edit them directly.

---

## 5. Adding a New Page

### Route File Convention

TanStack Router uses file-based routing. Files under `src/routes/`:

| File | Purpose |
|------|---------|
| `__root.tsx` | Document shell (providers, metadata) |
| `_app.tsx` | Layout route (navigation, sidebar) |
| `_app.index.tsx` | Home page (matches `/`) |
| `_app.<feature>.tsx` | Feature page (matches `/<feature>`) |
| `_app.<feature>.<sub>.tsx` | Sub-feature page (matches `/<feature>/<sub>`) |

The `_app` prefix creates a pathless layout group — all `_app.*` routes share the `_app.tsx` layout.

### Typical Page Structure

```tsx
import { createFileRoute } from "@tanstack/react-router"

export const Route = createFileRoute("/_app/<feature>")({
  component: FeaturePage,
})

function FeaturePage() {
  // Page content
}
```

For data-fetching pages, use route loaders or TanStack Query hooks.

---

## 6. Error Handling

The error pipeline is: normalize → display → report. See `docs/stacks/tanstack-react.md` Section 8 for the full pipeline.

Key points:
- `toApiError()` normalizes Connect, HTTP, and unknown errors
- `AntdErrorFeedbackAdapter` shows mutation errors as toasts, query errors as notifications
- Domain-specific handlers register by error reason code
- Queries that should suppress feedback: `meta: { silent: true }`

---

## 7. Common Pitfalls

### Connect code generation runs from workspace root

The `proto:connect` target uses `cwd: "{workspaceRoot}"` because `protoc-gen-*` binaries are root devDependencies. The buf config path is relative to the workspace root, but `./node_modules/.bin/` in the config also resolves from the workspace root.

### OpenAPI client naming follows BFF path prefixes

When BFFs use path prefixes (`/api/v1/admin/`, `/api/v1/mobile/`), the OpenAPI generator produces function names prefixed by the BFF name (e.g., `adminArticleServiceCreateArticle`). The client code must import from the correct generated module.

### Dual transport requires dual pagination

Connect and HTTP backends have different pagination shapes. Use `buildConnectPage` / `buildHTTPPage` from `lib/builder.ts` — do not construct pagination parameters manually.

### Paraglide messages are generated

Never edit files under `src/paraglide/`. Edit `i18n/messages/` source files and regenerate. Run `messages:check` before committing to catch stale generated output.

### pnpm isolated mode

The monorepo uses pnpm's default isolated mode. Each package only sees its declared dependencies. If a package needs a tool (e.g., buf generators), it must either declare it as a dependency or rely on the root package's devDependencies with workspace-root-relative execution.
