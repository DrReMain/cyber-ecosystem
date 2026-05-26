# Expo + React Native Stack Guide

Guide for coding agents working on consumer-facing mobile applications built with
Expo and React Native in this monorepo.

---

## 1. Scope

This guide is stack-level only. Product, client, page, and domain conventions
belong near the owning application.

Applies to React Native clients that use:

- React Native
- Expo SDK
- expo-router
- NativeWind
- Lingui
- Jotai
- TanStack Query
- Connect RPC
- pnpm
- Nx

---

## 2. Toolchain

The toolchain has clear ownership boundaries:

| Tool | Responsibility |
|------|----------------|
| Nx | Monorepo orchestration and declared project workflows |
| pnpm | Package manager and script runner |
| Expo SDK | Native runtime, build pipeline, and dev server |
| expo-router | Type-safe file-based routing |
| NativeWind | Tailwind CSS utility classes for React Native |
| Lingui | Message extraction, compilation, and runtime i18n |
| Jotai | Atomic client state with persistence |
| TanStack Query | Server-state caching and synchronization |
| Connect RPC | Type-safe API client generated from Protobuf |
| Biome | Formatting and static checks |

Use Nx for recurring workflows. Read the owning `project.json` and run declared
targets:

```bash
./nx run <project>:biome:check
./nx run <project>:i18n:extract
./nx run <project>:i18n:compile
./nx run <project>:proto:connect
```

Do not bypass Nx with direct pnpm or Expo commands when a target exists. If a
workflow is repeated, expose it as an Nx target.

---

## 3. Architecture Principles

### Separate server state from UI state

Use TanStack Query for remote server state. Use Jotai stores for persisted or
shared client state. Keep transient interaction state in React.

### Centralize cross-cutting concerns

Theme, locale, direction, error handling, and typed environment access are stack
concerns. Configure them once in the layout/provider layer instead of repeating
them in feature components.

### No full UI component library

Consumer apps require high visual freedom. Build components with NativeWind
directly. For complex primitives (DatePicker, BottomSheet), add targeted
libraries on demand. Icons use `lucide-react-native` (consistent with web
client).

---

## 4. Application Shell

`src/app/_layout.tsx` is the root layout. It owns:

- CSS import (`import "@/global.css"`) for NativeWind context
- Provider composition
- RTL direction wrapping
- Jotai store provider

### Provider Nesting Order

```
AppProvider(Jotai)               // Store hydration first (lowest level)
  I18nProvider                     // Locale before UI rendering
    ConnectProvider                // API clients before data consumers
      QueryProvider                // Query cache
        Stack                      // expo-router navigation
```

Order reflects dependency: Jotai has no deps, I18n depends on locale atom,
Connect depends on env, Query depends on Connect transport.

---

## 5. Routing

Use expo-router file-based routes in `src/app/`.

- Convention: `app/` directory with file-based routes
- Supports layouts, tabs, stack navigation via file conventions
- Typed routes enabled via `experiments.typedRoutes`
- Deep linking support built-in

Rules:

- Treat generated route types as derived output.
- Layout files (`_layout.tsx`) own provider composition for their subtree.
- Do not import between route files — use shared components and hooks.

---

## 6. API Client (Connect RPC)

Type-safe API client generated from Protobuf definitions.

| Channel | Transport | Code Gen |
|---------|-----------|----------|
| Connect RPC | `@connectrpc/connect-web` | `@connectrpc/protoc-gen-connect-query` |

### Transport (`services/connect-transport.ts`)

- Creates a `ConnectTransport` with `@connectrpc/connect-web`
- Base URL from `env.EXPO_PUBLIC_CONNECT_API_URL`
- Binary format in production (`!__DEV__`), JSON in development
- Wrapped in `ConnectTransportProvider` for TanStack Query integration

### Client Generation

Proto files live in `apps/<app>/api/`. Generate Connect clients:

```bash
./nx run <project>:proto:connect
```

Generated code lands in `src/services/connect/`. This code is committed to git
but MUST NOT be manually edited. Fix proto sources or generator config, then
regenerate.

---

## 7. Data Fetching

Use TanStack Query for all server state. The QueryClient is created in
`lib/query-client.ts` and provided via `QueryProvider`.

### QueryClient Defaults

```ts
staleTime: 1000 * 60        // 60 seconds
gcTime: 1000 * 60 * 10      // 10 minutes
retry: 2
networkMode: "offlineFirst"
refetchOnWindowFocus: false
```

Offline-first defaults: queries work with cached data when offline.

Rules:

- Create one `QueryClient` in `lib/query-client.ts`.
- Keep query keys structured and stable.
- Use `@connectrpc/connect-query` hooks (`useQuery`, `useInfiniteQuery`) directly.
- Do not create unrelated `QueryClient` instances inside feature components.

---

## 8. Styling (NativeWind)

NativeWind v4 provides Tailwind CSS utility classes for React Native.

### Setup

- `metro.config.js` — `withNativeWind(config, { input: "./src/global.css" })`
- `tailwind.config.js` — `darkMode: "class"` + NativeWind preset
- `src/global.css` — `@tailwind base; @tailwind components; @tailwind utilities;`

### Dark Mode

Controlled via `useColorScheme` from `nativewind` + `Appearance.setColorScheme`
from React Native. Three modes: `light`, `dark`, `system`.

- `dark:` prefix on className for dark variant styles
- Persistence via MMKV storage
- Use `useAppTheme` hook from `lib/use-app-theme.ts` for theme control
- Icons need dynamic colors: `color={colorScheme === "dark" ? "#aaa" : "#666"}`

### Color System

Tailwind config extends with CSS custom properties for design tokens (primary,
error, success, warning, info, typography, outline, background). These are
NativeWind CSS variables — standard Tailwind colors (`gray-*`, `blue-*`) work
alongside them.

---

## 9. State Management (Jotai)

Use `defineStore` from `stores/_core/define-store.ts` for client state that needs
persistence or sharing across the component tree.

### Store Definition

```ts
const store = defineStore("store_key", initialValue, {
  persist: true,       // write to MMKV on every change
  schema: z.number(),  // optional Zod validation on read-back
})
```

Each store creates:
- `immerAtom` — internal Immer-powered atom for draft-based updates
- `atom` — public atom that handles persistence

The public atom's `set` accepts either a new value or an Immer draft callback.

### When to Use

- **TanStack Query** — server state (API data)
- **Jotai stores** — client state that persists (MMKV) or shares across distant components
- **React state** — transient interaction state (modals, form inputs, toggles)

---

## 10. Internationalization (Lingui)

Lingui provides message extraction, compilation, and runtime translation.

### Workflow

```bash
./nx run <project>:i18n:extract    # extract messages from source → .po files
./nx run <project>:i18n:compile    # compile .po → .js + .d.ts
```

- Source `.po` files and compiled `.js`/`.d.ts` are committed to git.
- Use `_(msg\`text\`)` from `useLingui()` for all user-visible text.
- Import `msg` from `@lingui/core/macro` (babel plugin transforms at build time).

### Locale Configuration

- 5 locales: `en-US`, `zh-CN`, `ar-SA`, `ja-JP`, `ko-KR`
- Source locale: `en-US`
- Fallback: `en-US`
- RTL detection: `LOCALES[locale].rtl`
- Device locale detection via `expo-localization`

### RTL Support

```tsx
const direction = LOCALES[locale].rtl ? "rtl" : "ltr"
<View style={{ direction, flex: 1 }}>
```

Uses React Native's `direction` style prop (immediate effect, no restart).

---

## 11. Environment

Use `@t3-oss/env-core` for typed environment configuration.

```ts
client: {
  EXPO_PUBLIC_CONNECT_API_URL: z.string().default("http://localhost:13000"),
  EXPO_PUBLIC_HTTP_API_URL: z.string().default("http://localhost:11000"),
  EXPO_PUBLIC_GLITCHTIP_DSN: z.string().optional(),
  EXPO_PUBLIC_OTEL_URL: z.string().optional(),
}
```

Rules:

- Client-exposed variables must use the `EXPO_PUBLIC_` prefix.
- Validation belongs in the environment schema, not at scattered call sites.
- Defaults provide working values for local development.

---

## 12. Storage (MMKV)

`react-native-mmkv` provides fast, synchronous key-value storage.

```ts
import { storage } from "@/lib/mmkv"
storage.getString("key")
storage.set("key", "value")
```

Used for:
- Theme preference persistence
- Locale preference persistence
- Jotai store persistence (via `defineStore`)

---

## 13. Animations

**react-native-reanimated** — UI-thread animations.

- Declarative `useAnimatedStyle` and `withSpring`/`withTiming` APIs
- Pair with `react-native-gesture-handler` for gesture-driven interactions
- Do NOT rely on CSS/Tailwind animations

---

## 14. Generated Files

Do not manually edit generated output, including:

- `src/services/connect/` (Connect/Protobuf generated clients)
- `locale/*/messages.js` (compiled Lingui catalogs)
- `locale/*/messages.d.ts` (compiled Lingui types)

Fix the proto file, message source, generator configuration, or Nx target,
then regenerate.

---

## 15. Nx Targets

```bash
./nx run <project>:biome:check      # Lint and format check
./nx run <project>:biome:format     # Lint and format auto-fix
./nx run <project>:i18n:extract     # Extract translatable messages
./nx run <project>:i18n:compile     # Compile message catalogs
./nx run <project>:proto:connect    # Regenerate Connect RPC clients
```

Always use `./nx run` — never run `lingui`, `buf`, or `biome` directly.

---

## 16. Biome Configuration

Biome replaces ESLint for this project. Key rules:

- `useSortedClasses` — CSS class sorting (NativeWind)
- `noLeakedRender` — prevents promise leaks in JSX
- `useSortedAttributes` — JSX prop sorting
- Generated code excluded: `services/connect/`, `locale/**/*.js`

Run with `--unsafe` for class sorting and prop sorting auto-fixes.

---

## 17. Project Structure Reference

```
src/
  app/                               # expo-router file-based routes
    _layout.tsx                      # Root layout, provider composition
    index.tsx                        # Home screen
  components/                        # Shared UI components
    locale-picker.tsx                # Language selection bottom sheet
    theme-toggle.tsx                 # Theme mode toggle button
  hooks/                             # Custom React hooks
  lib/                               # Shared utilities
    env.ts                           # Typed env (t3-env + Zod)
    i18n.ts                          # Lingui setup, locale map, RTL detection
    mmkv.ts                          # MMKV storage singleton
    query-client.ts                  # TanStack Query client
    query-provider.tsx               # QueryClientProvider wrapper
    use-app-theme.ts                 # Theme hook (light/dark/system)
  services/                          # API client layer
    connect-transport.ts             # Connect RPC transport setup
    connect-provider.tsx             # TransportProvider wrapper
    connect/                         # Generated Connect/Protobuf code
  stores/                            # Jotai state management
    _core/                           # Store infrastructure
      define-store.ts                # defineStore factory (persist + schema)
      provider.tsx                   # JotaiProvider wrapper
    i18n.ts                          # Locale atom + MMKV persistence
  global.css                         # Tailwind CSS entry point
locale/                              # Lingui translation catalogs
  en-US/messages.po                  # Source catalog
  en-US/messages.js                  # Compiled catalog
  ...                                # Other locales (zh-CN, ar-SA, ja-JP, ko-KR)
```
