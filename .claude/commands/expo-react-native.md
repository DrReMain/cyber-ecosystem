---
name: expo-react-native
description: Use when creating a new mobile client, adding screens to an Expo app, connecting to BFF services with Connect RPC, or configuring i18n/theme/navigation in React Native
---

# Expo React Native Mobile Client

Guide for scaffolding and implementing mobile clients using Expo SDK, React Native, and expo-router.

---

## When to Use

- Creating a new mobile client
- Adding screens, navigation, or routing
- Connecting to BFF services via Connect RPC
- Configuring i18n, theme, or state management

## Key Rules

- Use `createClient` (NOT `createPromiseClient` or `createConnectPromiseClient`)
- Use TanStack Query directly — NOT `@connectrpc/connect-query` wrapper hooks
- FlatList `onEndReached` needs a ref guard (`fetchingRef`) to prevent duplicate calls
- Lingui: explicit string IDs only (`i18n._("article.title")`), never macro calls (`msg`)
- PO files need `#. js-lingui-explicit-id` comment on every message
- Use `key={locale}` on `<Stack>` to force re-render on locale change
- expo-router params are strings — use `encodeURIComponent` for special characters

---

## Adding a New Mobile Client — Step by Step

### Step 1: Create Directory Structure

Create `apps/<app>/clients/<client>/` with the layout from `docs/stacks/expo-react-native.md` Section 18. Copy an existing mobile client as a starting point.

### Step 2: Configure package.json

```json
{
  "name": "<app>_client_<client>",
  "dependencies": {
    "@bufbuild/protobuf": "^2",
    "@connectrpc/connect": "^2",
    "@connectrpc/connect-query": "^2",
    "@connectrpc/connect-web": "^2",
    "@lingui/core": "^5",
    "@lingui/react": "^5",
    "@t3-oss/env-core": "^0.12",
    "@tanstack/react-query": "^5",
    "expo": "~56",
    "expo-router": "~5",
    "jotai": "^2",
    "jotai-immutable": "^0.3",
    "nativewind": "^5",
    "react": "^19",
    "react-native": "^0.85",
    "react-native-mmkv": "^3",
    "react-native-reanimated": "^3",
    "zod": "^3"
  },
  "devDependencies": {
    "@lingui/cli": "^5",
    "tailwindcss": "^4",
    "typescript": "^5"
  }
}
```

### Step 3: Configure Code Generation

`buf.gen.client.yaml`:

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

### Step 4: Configure Metro + NativeWind

`metro.config.js`:

```js
const { getDefaultConfig } = require("expo/metro-config")
const { withNativeWind } = require("nativewind/metro")
const { withExpoRouter } = require("expo-router/metro")

const config = getDefaultConfig(__dirname)
module.exports = withNativeWind(withExpoRouter(config), { input: "./src/global.css" })
```

### Step 5: Configure Nx Targets

```json
{
  "targets": {
    "proto:connect": {
      "executor": "nx:run-commands",
      "options": { "cwd": "{workspaceRoot}", "command": "buf generate --template apps/<app>/clients/<client>/buf.gen.client.yaml --path apps/<app>/api --include-imports" }
    },
    "i18n:extract": {
      "executor": "nx:run-commands",
      "options": { "cwd": "{projectRoot}", "command": "lingui extract --clean" }
    },
    "i18n:compile": {
      "executor": "nx:run-commands",
      "options": { "cwd": "{projectRoot}", "command": "lingui compile" }
    },
    "biome:check": {
      "executor": "nx:run-commands",
      "options": { "cwd": "{projectRoot}", "command": "biome check src/" }
    },
    "biome:format": {
      "executor": "nx:run-commands",
      "options": { "cwd": "{projectRoot}", "command": "biome check --unsafe --write src/" }
    },
    "typecheck": {
      "executor": "nx:run-commands",
      "options": { "cwd": "{projectRoot}", "command": "tsc --noEmit" }
    }
  }
}
```

---

## Connect RPC Client

```tsx
import { createClient } from "@connectrpc/connect"
import { useTransport } from "@connectrpc/connect-query"

const transport = useTransport()
const client = createClient(GeneratedServiceType, transport)
```

The `GenService` type from `*_pb.ts` files works directly with `createClient`. Binary format in production, JSON in dev.

---

## Data Fetching Patterns

### Paginated list

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

### Mutation

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

Key: `queryClient.invalidateQueries` refreshes list data. Check for `ConnectError` in error handler.

---

## FlatList Infinite Scroll

```tsx
const fetchingRef = useRef(false)

const handleEndReached = useCallback(() => {
  if (hasNextPage && !fetchingRef.current) {
    fetchingRef.current = true
    fetchNextPage().finally(() => { fetchingRef.current = false })
  }
}, [hasNextPage, fetchNextPage])
```

The ref updates synchronously, preventing duplicate calls even when `onEndReached` fires multiple times. Do NOT use `isFetchingNextPage` alone — it's async state that doesn't update fast enough.

---

## Routing

| File | Path | Purpose |
|------|------|---------|
| `_layout.tsx` | — | Root layout with providers |
| `index.tsx` | `/` | Home screen |
| `article/[id].tsx` | `/article/:id` | Dynamic route |
| `article/create.tsx` | `/article/create` | Static route |

Rules:
- `useLocalSearchParams` generic type requires casting: `as unknown as YourType`
- Route params are strings — encode/decode with `encodeURIComponent`
- Do not import between route files — use shared components and hooks

---

## i18n Workflow

All calls use explicit string IDs:

```tsx
import { i18n } from "@/lib/i18n"

i18n._("article.title")
i18n._("article.time.minutesAgo", { n: 5 })
```

Do NOT use macro calls (`msg\`text\``) which produce hash IDs.

PO file format — every message needs `#. js-lingui-explicit-id`:

```po
#. js-lingui-explicit-id
#: src/app/article/create.tsx
msgid "article.cancel"
msgstr "Cancel"
```

5 locales: `en-US`, `zh-CN`, `ar-SA`, `ja-JP`, `ko-KR`. Source locale: `en-US`.

---

## State Management

| State type | Tool | Example |
|-----------|------|---------|
| Server data | TanStack Query | Article list, user profile |
| Persisted client state | Jotai (defineStore) | Theme, locale |
| Transient UI state | React useState | Modal visibility, form inputs |

Use `defineStore` from `stores/_core/define-store.ts` for persisted state:

```ts
const store = defineStore("store_key", initialValue, {
  persist: true,
  schema: z.number(),
})
```

---

## Protobuf Timestamp Handling

```tsx
new Date(Number(item.createdAt.seconds) * 1000).toISOString()
```

Use `formatRelativeTime` from `lib/relative-time.ts` for relative time display.

---

## Modal Bottom Sheet

```tsx
<Modal animationType="fade" transparent visible={visible} onRequestClose={onClose}>
  <Pressable className="flex-1 bg-black/40" onPress={onClose}>
    <View className="absolute right-0 bottom-0 left-0 rounded-t-2xl bg-white pt-3 pb-8 dark:bg-gray-900">
      <View className="mb-3 h-1 w-10 self-center rounded-full bg-gray-300 dark:bg-gray-600" />
      {/* Content */}
    </View>
  </Pressable>
</Modal>
```

---

## Common Pitfalls

### createClient, not createPromiseClient

`@connectrpc/connect` v2 exports `createClient` for promise-based clients. `createPromiseClient` does not exist.

### useLocalSearchParams generic type error

Cast with `as unknown as YourType` to bypass expo-router's type constraint.

### FlatList onEndReached fires without scrolling

Use a ref guard (`fetchingRef`). Do NOT rely on `isFetchingNextPage` alone.

### Lingui hash IDs

If `lingui extract` produces hash IDs, `.po` file is missing `#. js-lingui-explicit-id` comments.

### i18n not reactive on locale switch

Use `key={locale}` on `<Stack>` to force full remount.

### Route params encoding

expo-router passes params as strings. Use `encodeURIComponent` / `decodeURIComponent`.

### Connect code generation runs from workspace root

The `proto:connect` target uses `cwd: "{workspaceRoot}"` because `protoc-gen-*` binaries are root devDependencies.

### NativeWind className vs style

Use `className` for NativeWind styles. Use `style` only for dynamic values (e.g., `direction: "rtl"`). Do not mix for the same property.

---

## Nx Targets

```bash
./nx run <project>:proto:connect    # Generate Connect clients
./nx run <project>:i18n:extract     # Extract → .po files
./nx run <project>:i18n:compile     # Compile .po → .js + .d.ts
./nx run <project>:biome:check      # Lint check
./nx run <project>:biome:format     # Lint fix
./nx run <project>:typecheck        # TypeScript check
```

---

For deep architecture details, see `docs/stacks/expo-react-native.md`.
