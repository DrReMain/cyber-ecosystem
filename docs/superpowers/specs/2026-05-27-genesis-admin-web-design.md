# Genesis Admin Web Client Design

## Summary

Copy the template admin client (`apps/template/clients/admin/`) to `apps/genesis/clients/admin/`, switch API connection to genesis admin_bff, replace Message entity with Article in playground pages.

## Scope

- Full TanStack Start stack (same as template): React 19, Ant Design 6, TanStack Query, Jotai, Paraglide i18n (5 locales), Vite 8
- Dual API backend: Connect RPC + HTTP REST
- No authentication
- No file service (genesis admin_bff has no file APIs)

## Changes from Template

### Remove
- `services/file-client.ts`
- Any file upload/download UI code in playground pages

### Regenerate
- Connect RPC clients from genesis proto → `src/services/connect/`
- OpenAPI clients from `genesis/gen/oas/openapi.yaml` → `src/services/openapi/`

### Modify
- API base URLs → point to genesis admin_bff (env.ts, connect-transport, http-config)
- Playground pages: Message → Article entity
- Paraglide messages: update entity labels

### Keep unchanged
- `domains/` (errors, antd, i18n, theme, router-progress, seo)
- `stores/` (all including examples)
- `lib/` (builder, cookie, use-utils)
- `components/`
- `services/` infrastructure (connect-transport.tsx, http-config.ts, http-client.ts)
- `__root.tsx`, `router.tsx`, `server.ts`

## API Surface

admin_bff exposes via both HTTP and Connect:

- Article CRUD: Create, Update, Delete, Get, Query (paginated), Sort, UpdateStatus
- Resource introspection: ListResource

## Routes

Same structure as template. Playground pages adapted for Article:

```
_app.playground.connect.*.tsx    # Article CRUD via Connect RPC
_app.playground.http.*.tsx       # Article CRUD via HTTP REST
```
