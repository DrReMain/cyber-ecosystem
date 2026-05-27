# Expo + React Native Stack Guide

Guide for coding agents working on consumer-facing mobile applications built with
Expo and React Native in this monorepo.

---

## Quick Reference

### Connect RPC Client

```tsx
import { createClient } from "@connectrpc/connect"
import { useTransport } from "@connectrpc/connect-query"

const transport = useTransport()
const client = createClient(GeneratedServiceType, transport)
```

**NOT** `createPromiseClient` or `createConnectPromiseClient`. Use TanStack Query directly, NOT `@connectrpc/connect-query` wrapper hooks.

### Paginated List

```tsx
const { data, fetchNextPage, hasNextPage } = useInfiniteQuery({
  queryKey: ["articles"],
  queryFn: ({ pageParam }) =>
    client.queryArticle({ page: { pageNo: pageParam, pageSize: PAGE_SIZE } }),
  initialPageParam: 1,
  getNextPageParam: (lastPage) => lastPage.page?.more ? lastPage.page.pageNo + 1 : undefined,
})
const items = data?.pages.flatMap((page) => page.list) ?? []
```

### Mutation

```tsx
const mutation = useMutation({
  mutationFn: () => client.createArticle({ title, content }),
  onSuccess: () => queryClient.invalidateQueries({ queryKey: ["articles"] }),
})
```

### FlatList Infinite Scroll (with ref guard)

```tsx
const fetchingRef = useRef(false)
const handleEndReached = useCallback(() => {
  if (hasNextPage && !fetchingRef.current) {
    fetchingRef.current = true
    fetchNextPage().finally(() => { fetchingRef.current = false })
  }
}, [hasNextPage, fetchNextPage])
```

### Protobuf Timestamp

```tsx
new Date(Number(item.createdAt.seconds) * 1000).toISOString()
```

### i18n (Lingui)

```tsx
i18n._("article.title")              // explicit ID only, never msg`` macros
i18n._("article.time.minutesAgo", { n: 5 })
```

### State Management

| State type | Tool | Example |
|-----------|------|---------|
| Server data | TanStack Query | Article list |
| Persisted client state | Jotai (defineStore) | Theme, locale |
| Transient UI state | React useState | Modal, form inputs |

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
libraries on demand.

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
        AppShell                   // Direction + Stack
```

Order reflects dependency: Jotai has no deps, I18n depends on locale atom,
Connect depends on env, Query depends on Connect transport.

### Stack Re-render on Locale Change

Pass `key={locale}` to `<Stack>` to force a full re-render when locale changes.
Lingui's `I18nProvider` does not trigger re-renders for already-mounted
components on locale change. The key trick forces a complete remount:

```tsx
<Stack key={locale} screenOptions={{ headerShown: false }} />
```

### Init Calls

Store initialization runs at module level in `_layout.tsx`:

```tsx
import { initTheme } from "@/stores/theme/store"
import { initI18n, localeAtom } from "@/stores/i18n/store"

initTheme()
initI18n()
```

These must run before the component tree mounts.

---

## 5. Routing

Use expo-router file-based routes in `src/app/`.

| File | Path | Purpose |
|------|------|---------|
| `_layout.tsx` | — | Root layout with providers |
| `index.tsx` | `/` | Home screen |
| `article/[id].tsx` | `/article/:id` | Dynamic route |
| `article/create.tsx` | `/article/create` | Static route |

Rules:

- Layout files (`_layout.tsx`) own provider composition for their subtree.
- Do not import between route files — use shared components and hooks.
- `useLocalSearchParams` generic type requires casting: `as unknown as YourType`.
- Route params are strings — encode/decode with `encodeURIComponent` for special characters.

---

## 6. API Client (Connect RPC)

Type-safe API client generated from Protobuf definitions.

| Channel | Transport | Code Gen |
|---------|-----------|----------|
| Connect RPC | `@connectrpc/connect-web` | `protoc-gen-es` + `protoc-gen-connect-es` + `protoc-gen-connect-query` |

### Transport (`services/connect-provider.tsx`)

```tsx
import { createConnectTransport } from "@connectrpc/connect-web"

const transport = createConnectTransport({
  baseUrl: env.EXPO_PUBLIC_CONNECT_API_URL,
  useBinaryFormat: !__DEV__,
})
```

Binary format in production, JSON in development. Wrapped in `TransportProvider`
from `@connectrpc/connect-query`.

### Client Creation

Use `createClient` from `@connectrpc/connect` — NOT `createPromiseClient`:

```tsx
import { createClient } from "@connectrpc/connect"
import { useTransport } from "@connectrpc/connect-query"

const transport = useTransport()
const client = createClient(GeneratedServiceType, transport)
```

The `GenService` type from `*_pb.ts` files works directly with `createClient`.
Do NOT use connect-query wrapper hooks — use TanStack Query directly.

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

### Paginated List Pattern

```tsx
const { data, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
  queryKey: ["articles"],
  queryFn: ({ pageParam }) =>
    client.queryArticle({ page: { pageNo: pageParam, pageSize: PAGE_SIZE } }),
  initialPageParam: 1,
  getNextPageParam: (lastPage) => {
    if (!lastPage.page?.more) return undefined
    return lastPage.page.pageNo + 1
  },
})

const items = data?.pages.flatMap((page) => page.list) ?? []
```

Pagination response from proto `PageResponse`: `{ pageNo, pageSize, total, more }`.

### Mutation Pattern

```tsx
const mutation = useMutation({
  mutationFn: () => client.createArticle({ title, content }),
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ["articles"] })
    router.back()
  },
  onError: (error: unknown) => {
    const msg = error instanceof ConnectError ? error.message : String(error)
    Alert.alert(i18n._("error.title"), msg)
  },
})
```

Key: `queryClient.invalidateQueries` refreshes list data. Check for `ConnectError`
in error handler. Use `mutation.isPending` for loading state.

### FlatList Infinite Scroll

FlatList's `onEndReached` fires immediately when content doesn't fill the
viewport. Use a ref guard to prevent rapid-fire requests:

```tsx
const fetchingRef = useRef(false)

const handleEndReached = useCallback(() => {
  if (hasNextPage && !fetchingRef.current) {
    fetchingRef.current = true
    fetchNextPage().finally(() => { fetchingRef.current = false })
  }
}, [hasNextPage, fetchNextPage])
```

The ref updates synchronously, preventing duplicate calls even when
`onEndReached` fires multiple times between renders.

Rules:

- Do NOT use connect-query wrapper hooks (`useQuery` from connect-query).
- Create one `QueryClient` in `lib/query-client.ts`.
- Keep query keys structured and stable.
- Do not create unrelated `QueryClient` instances inside feature components.

---

## 8. Styling (NativeWind)

NativeWind v4+ provides Tailwind CSS utility classes for React Native.

### Setup

- `metro.config.js` — `withNativeWind(config, { input: "./src/global.css" })`
- `tailwind.config.js` — `darkMode: "class"` + NativeWind preset
- `src/global.css` — Tailwind CSS entry point with NativeWind theme import

### Dark Mode

Controlled via `useColorScheme` from `nativewind` + `Appearance.setColorScheme`
from React Native. Three modes: `light`, `dark`, `system`.

- `dark:` prefix on className for dark variant styles
- Persistence via MMKV storage
- Icons need dynamic colors: `color={colorScheme === "dark" ? "#aaa" : "#666"}`

### className vs style

Use `className` for NativeWind (Tailwind) styles. Use `style` prop only for
dynamic values (e.g., `direction: "rtl"`). Do not mix for the same property.

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

### Locale Store Pattern

Locale uses separate read/write atoms:

```ts
export const localeAtom = atom<Locale>(getStoredLocale())
export const setLocaleAtom = atom(null, (_get, set, locale: Locale) => {
  storage.set("locale", locale)
  i18n.activate(locale)
  set(localeAtom, locale)
})
```

### When to Use

| State type | Tool | Example |
|-----------|------|---------|
| Server data | TanStack Query | Article list, user profile |
| Persisted client state | Jotai (defineStore) | Theme, locale |
| Transient UI state | React useState | Modal visibility, form inputs |

---

## 10. Internationalization (Lingui)

Lingui provides message extraction, compilation, and runtime translation.

### Key Convention: Explicit IDs

All i18n calls use explicit string IDs with `i18n._()`:

```tsx
import { i18n } from "@/lib/i18n"

i18n._("article.title")
i18n._("article.time.minutesAgo", { n: 5 })
```

Do NOT use macro calls (`msg\`text\``) which produce hash IDs.

### PO File Format

Every message in `.po` files needs a `#. js-lingui-explicit-id` comment:

```po
#. js-lingui-explicit-id
#: src/app/article/create.tsx
msgid "article.cancel"
msgstr "Cancel"
```

Without this comment, `lingui extract` replaces the ID with a hash.

### Workflow

```bash
./nx run <project>:i18n:extract    # extract messages from source → .po files
./nx run <project>:i18n:compile    # compile .po → .js + .d.ts
```

- Source `.po` files and compiled `.js`/`.d.ts` are committed to git.
- When adding keys manually to `.po`, always include `#. js-lingui-explicit-id`.
- After compile, run biome format to fix class ordering in `.po` msgstr values.

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
  EXPO_PUBLIC_CONNECT_API_URL: z.string().default("http://localhost:13002"),
  EXPO_PUBLIC_HTTP_API_URL: z.string().default("http://localhost:11002"),
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

## 13. Protobuf Timestamp Handling

Protobuf timestamps are `{ seconds: bigint, nanos: number }`. Convert for display:

```tsx
new Date(Number(item.createdAt.seconds) * 1000).toISOString()
```

For relative time display, use `formatRelativeTime` from `lib/relative-time.ts` —
it handles both ISO strings and protobuf Timestamp objects.

---

## 14. Animations

**react-native-reanimated** — UI-thread animations.

- Declarative `useAnimatedStyle` and `withSpring`/`withTiming` APIs
- Pair with `react-native-gesture-handler` for gesture-driven interactions
- Do NOT rely on CSS/Tailwind animations

---

## 15. Generated Files

Do not manually edit generated output, including:

- `src/services/connect/` (Connect/Protobuf generated clients)
- `locale/*/messages.js` (compiled Lingui catalogs)
- `locale/*/messages.d.ts` (compiled Lingui types)

Fix the proto file, message source, generator configuration, or Nx target,
then regenerate.

---

## 16. Nx Targets

```bash
./nx run <project>:biome:check      # Lint and format check
./nx run <project>:biome:format     # Lint and format auto-fix
./nx run <project>:i18n:extract     # Extract translatable messages
./nx run <project>:i18n:compile     # Compile message catalogs
./nx run <project>:proto:connect    # Regenerate Connect RPC clients
```

Always use `./nx run` — never run `lingui`, `buf`, or `biome` directly.

---

## 17. Biome Configuration

Biome replaces ESLint for this project. Key rules:

- `useSortedClasses` — CSS class sorting (NativeWind)
- `noLeakedRender` — prevents promise leaks in JSX
- `useSortedAttributes` — JSX prop sorting
- Generated code excluded: `services/connect/`, `locale/**/*.js`

Run with `--unsafe` for class sorting and prop sorting auto-fixes.

---

## 18. Project Structure Reference

```
src/
  app/                               # expo-router file-based routes
    _layout.tsx                      # Root layout, provider composition
    index.tsx                        # Home screen
    <feature>/
      [id].tsx                       # Dynamic route (detail)
      create.tsx                     # Static route (creation)
  components/                        # Shared UI components
    <Entity>Card.tsx                 # List item card
    <Entity>List.tsx                 # FlatList with infinite scroll
    EmptyState.tsx                   # Empty list placeholder
    LoadingFooter.tsx                # Loading indicator
    LocalePicker.tsx                 # Language selection bottom sheet
  lib/                               # Shared utilities
    env.ts                           # Typed env (t3-env + Zod)
    i18n.ts                          # Lingui setup, locale map, RTL detection
    mmkv.ts                          # MMKV storage singleton
    query-client.ts                  # TanStack Query client
    query-provider.tsx               # QueryClientProvider wrapper
    relative-time.ts                 # Protobuf timestamp → relative time
  services/                          # API client layer
    connect-provider.tsx             # TransportProvider wrapper
    connect/                         # Generated Connect/Protobuf code
  stores/                            # Jotai state management
    _core/                           # Store infrastructure
      define-store.ts                # defineStore factory (persist + schema)
      provider.tsx                   # AppProvider (JotaiProvider wrapper)
    i18n/
      store.ts                       # Locale atom + initI18n
    theme/
      store.ts                       # Theme atom + initTheme
  global.css                         # Tailwind CSS entry point
locale/                              # Lingui translation catalogs
  en-US/messages.po                  # Source catalog
  en-US/messages.js                  # Compiled catalog
  zh-CN/
  ar-SA/                             # RTL locale
  ja-JP/
  ko-KR/
```

---

## 19. Testing

### Conventions

- Co-located test files: `{file}.test.tsx` or `{file}.spec.tsx`
- Recommended setup (configure per client — there is no `test` target yet): `vitest` runner, `@testing-library/react-native` for components

### Running Tests

```bash
./nx run <project>:typecheck          # TypeScript check (fastest validation)
./nx run <project>:biome:check        # Lint check
# Test target may not be declared yet — use `pnpm vitest` if needed
```

### What to Test

| Layer | What to Test | Tool |
|-------|-------------|------|
| Components | Rendering, NativeWind classes | vitest + testing-library |
| Hooks (queries) | Data loading, pagination | vitest |
| Stores (Jotai) | State, persistence | vitest |
| Navigation | Route params, deep links | vitest |

---

## 20. Generation Troubleshooting

### Lingui produces hash IDs instead of explicit IDs

**Cause**: Missing `#. js-lingui-explicit-id` comment in `.po` file.

**Fix**: Add the comment before every `msgid` entry. Re-run: `./nx run <project>:i18n:extract`

### Connect code generation runs from workspace root

**Expected**: The `proto:connect` target uses `cwd: "{workspaceRoot}"` because `protoc-gen-*` binaries are root devDependencies. This is correct.

### FlatList triggers `onEndReached` immediately

**Cause**: Content is shorter than viewport. Use a ref guard (`fetchingRef`) to prevent duplicate calls. Do NOT rely on `isFetchingNextPage` alone — it's async state that doesn't update fast enough.

### i18n not reactive on locale switch

**Fix**: Pass `key={locale}` to `<Stack>` to force full remount. Lingui's `I18nProvider` doesn't trigger re-renders for already-mounted components.

### pnpm "cannot find module"

**Cause**: pnpm isolated mode. Each package only sees declared dependencies.

**Fix**: Add the dependency to the package's `package.json`, then `pnpm install`.
