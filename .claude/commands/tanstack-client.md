---
name: tanstack-client
description: Use when creating a new React web client, adding pages or routes, connecting to BFF APIs via Connect RPC or OpenAPI, or configuring i18n/theme/error handling in TanStack Start
---

# TanStack Start React Web Client

Guide for scaffolding and implementing React web clients using TanStack Start in this monorepo.

---

## When to Use

- Creating a new web client (admin, portal, etc.)
- Adding pages, routes, or navigation
- Connecting to BFF services (Connect RPC or HTTP REST)
- Configuring i18n (Paraglide), theme, or error handling

## Key Rules

- Dual transport: Connect RPC (primary) + HTTP REST via OpenAPI (secondary)
- Use `buildConnectPage` / `buildHTTPPage` from `lib/builder.ts` for pagination — never construct manually
- Paraglide messages are generated — never edit `src/paraglide/`, edit `i18n/messages/` source files
- `proto:connect` runs from `{workspaceRoot}` because buf generators are root devDependencies
- BFF path prefixes (`/api/v1/admin/`, `/api/v1/mobile/`) produce prefixed OpenAPI function names
- Vite plugin order matters: devtools → paraglide → tailwind → start → react → babel
- pnpm isolated mode: packages only see declared dependencies

---

## Adding a New Web Client — Step by Step

### Step 1: Create Directory Structure

Create `apps/<app>/clients/<client>/` with the layout from `docs/stacks/tanstack-react.md` Section 19. Copy an existing client as a starting point.

### Step 2: Configure package.json

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
  plugins: [{ name: "@hey-api/client-ky", runtimeConfigPath: "./src/services/http-config" }],
})
```

### Step 4: Configure Vite

Plugin chain order matters:

```ts
plugins: [
  TanstackDevtools(),
  paraglideVitePlugin({ project: "./project.inlang", outdir: "./src/paraglide", strategy: ["custom-smart-preferred", "url", "baseLocale"] }),
  tailwindcss(),
  tanstackStart(),
  viteReact(),
  babel({ presets: [reactCompilerPreset()] }),
]
```

Dev proxy for backend:

```ts
server: {
  proxy: {
    "/connect": { target: env.CONNECT_API_URL || "http://localhost:<connect-port>", changeOrigin: true, rewrite: (path) => path.replace(/^\/connect/, "") },
    "/http": { target: env.HTTP_API_URL || "http://localhost:<http-port>", changeOrigin: true, rewrite: (path) => path.replace(/^\/http/, "") },
  },
}
```

### Step 5: Configure Nx Targets

```json
{
  "targets": {
    "proto:connect": {
      "executor": "nx:run-commands",
      "options": { "cwd": "{workspaceRoot}", "command": "buf generate --template apps/<app>/clients/<client>/buf.gen.client.yaml --path apps/<app>/api --include-imports" }
    },
    "proto:openapi": {
      "executor": "nx:run-commands",
      "options": { "cwd": "{projectRoot}", "command": "openapi-ts -f hey-api.config.ts" }
    },
    "messages:gen": {
      "executor": "nx:run-commands",
      "options": { "cwd": "{projectRoot}", "command": "pnpm paraglide compile" }
    },
    "messages:check": {
      "executor": "nx:run-commands",
      "options": { "cwd": "{projectRoot}", "command": "pnpm paraglide compile --dry-run" }
    },
    "biome:check": {
      "executor": "nx:run-commands",
      "options": { "cwd": "{projectRoot}", "command": "biome check src/" }
    },
    "biome:format": {
      "executor": "nx:run-commands",
      "options": { "cwd": "{projectRoot}", "command": "biome check --unsafe --write src/" }
    }
  }
}
```

### Step 6: Wire Into the Workspace

1. `pnpm install` — pnpm workspace auto-discovers via `apps/**` glob
2. Generate Connect clients: `./nx run <project>:proto:connect`
3. Generate OpenAPI clients: `./nx run <project>:proto:openapi`
4. Generate messages: `./nx run <project>:messages:gen`

---

## Dual API Backend

Web clients connect to BFF services through two independent transport layers.

### Connect RPC (primary)

```ts
import { useQuery } from "@connectrpc/connect-query"
import { listArticles } from "#/services/connect/<app>V1-<service>Service-Connect"
const { data } = useQuery(listArticles, { pageSize: 10 })
```

### HTTP REST (secondary)

```ts
import { adminArticleServiceListArticle } from "#/services/openapi/CustomClient"
const { data } = await adminArticleServiceListArticle({ queries: { "page.size": 10 } })
```

Use `buildConnectPage` / `buildHTTPPage` from `lib/builder.ts` for pagination — never construct manually.

---

## i18n Workflow

1. Add/edit messages in `i18n/messages/` source files
2. `./nx run <project>:messages:gen` — generate Paraglide code
3. Use generated functions from `#/paraglide/messages`
4. `./nx run <project>:messages:check` — validate

Generated files under `src/paraglide/` are derived output — never edit directly.

---

## Adding a New Page

| File | Purpose |
|------|---------|
| `__root.tsx` | Document shell (providers, metadata) |
| `_app.tsx` | Layout route (navigation, sidebar) |
| `_app.index.tsx` | Home page (matches `/`) |
| `_app.<feature>.tsx` | Feature page (matches `/<feature>`) |

```tsx
import { createFileRoute } from "@tanstack/react-router"

export const Route = createFileRoute("/_app/<feature>")({
  component: FeaturePage,
})

function FeaturePage() {
  // Page content
}
```

---

## Error Handling

Three-layer pipeline: normalize → display → report.

- `toApiError()` normalizes Connect, HTTP, and unknown errors into a unified `ApiError`
- `AntdErrorFeedbackAdapter`: mutation errors → `message.error()` (toast), query errors → `notification.error()`
- Domain handlers register by reason code: `registerErrorHandler("CODE", handler)`
- Suppress per-query: `meta: { silent: true }`
- `reportError()` sends to Sentry (lazy-initialized from `VITE_GLITCHTIP_DSN`)

---

## Common Pitfalls

### Connect code generation runs from workspace root

The `proto:connect` target uses `cwd: "{workspaceRoot}"` because `protoc-gen-*` binaries are root devDependencies.

### OpenAPI client naming follows BFF path prefixes

BFF prefixes produce prefixed function names: `adminArticleServiceCreateArticle`.

### Dual transport requires dual pagination

Use `buildConnectPage` / `buildHTTPPage` from `lib/builder.ts` — do not construct pagination manually.

### Paraglide messages are generated

Never edit `src/paraglide/`. Edit `i18n/messages/` and regenerate. Run `messages:check` before committing.

### pnpm isolated mode

Each package only sees its declared dependencies. Missing module → add to `package.json` → `pnpm install`.

### Vite plugin order

Must be: devtools → paraglide → tailwind → start → react → babel. Changing order breaks SSR or HMR.

---

## Nx Targets

```bash
./nx run <project>:proto:connect    # Generate Connect clients
./nx run <project>:proto:openapi    # Generate OpenAPI clients
./nx run <project>:messages:gen     # Generate Paraglide messages
./nx run <project>:messages:check   # Validate Paraglide (dry-run)
./nx run <project>:biome:check      # Lint check
./nx run <project>:biome:format     # Lint fix
```

---

For deep architecture details, see `docs/stacks/tanstack-react.md`.
