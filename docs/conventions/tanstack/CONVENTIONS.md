# TanStack Start Client Conventions

**Status:** ACTIVE

> Rules here win over precedent in code when they conflict. Keep this file in
> sync when a rule changes (see root AGENTS.md §7 and docs/README.md).

Applies to web clients built on TanStack Start (`app/clients/*`).

---

## 1) Layering model

| Layer | Location | Rule |
|---|---|---|
| Route declaration | `src/routes/` | Only `createFileRoute` wiring — no components, no logic (§3) |
| App shell utilities | `src/features/app/` | The base's generic subpath: shell screens, boundaries, shared helpers; may import domains (§7) |
| Areas | `src/features/layout-<name>/`, `src/features/login-<name>/` | An area = one layout directory (the owner) + one login UI kit (§2) |
| Formal pages | `src/features/modules/<route-name>/` | One directory per route, `page.tsx` entry |
| App-level components | `src/components/` | Self-contained only — no project-internal dependencies; depended on, never depending |
| Domains | `src/domains/` | Admission-gated (§4) |
| Infra | `src/stores/`, `src/services/`, `src/libs/` | As conventional — `libs` holds app-agnostic utilities (zero app semantics, import nothing app-internal) |

Dependency direction: `routes → features → domains → services/stores/lib`.
Features may import each other only along the area-pair direction (§2).

**Transport responses are JSON-shaped, not message instances.** The connect
transport (`services/connect/transport.ts`) re-serializes every unary
response through proto-es `toJson`: WKT wrappers arrive flattened to scalars,
proto `map` fields as plain objects, enums as value-name strings. The
generated TS types claim Message/Map shapes the runtime does not deliver —
consume accordingly (`Object.entries` over maps, string/index access for
enums).

## 2) Area model

An **area** is a private zone of the app (e.g. `/dashboard`). It consists of
a layout directory plus a login UI kit:

- `features/layout-<name>/` — the area's owner: `chrome/` (visual
  furniture; families graduate to nested subpaths like `chrome/sider/`),
  `protocol/` (route-meta, nav, first-leaf-redirect), `auth/` (custody,
  session fn, guard, watcher, logout, session-kick), `login/` (server fn,
  mutation hook, the assembled login page), and a root `area.ts` — the
  single source of the area's binding (namespace, paths, login route id).
  The layout mounts the area's UI-library provider itself. Stores live in
  `src/stores/` under the single eager-registration barrel — load-bearing:
  the defineStore registry drives server-side cookie collection, and a
  late code-split registration would miss hydration. One store per file;
  an area-scoped store carries the area namespace in directory, variable,
  and cookie-key names (`dashboard-preferences/` →
  `dashboardPreferencesStore` → `store_dashboard_preferences`) so sibling
  layouts never collide.
- `features/login-<name>/` — a pure UI kit: `login-shell.tsx` (canvas +
  panel slots), `ui/` (form, brand pieces, stub affordances),
  `redirect.ts` (open-redirect hygiene, zero area knowledge). No layout
  code, no session logic — only base infra (config constants, i18n
  widgets) + framework deps. `<name>` is a skin family: a second layout
  reuses the kit with its own wiring.

Rules:

- **Layouts are not designed for reuse.** Each layout is high-cohesion;
  mechanisms may differ between layouts. Never introduce a cross-layout
  common layer. Adding an area = one layout directory + (optionally) a new
  login skin + route files + one `AREAS` registry line in `src/config.ts` —
  no edits to existing areas.
- **Dependency direction: layout → login kit.** The login route mounts the
  layout's wired login face
  (`#/features/layout-<name>/login/login-page`), which composes the UI
  kit; the kit never imports back. The login bundle pulls the layout's
  `login/` + `auth/` + `area.ts` by deep path — never its barrel, never
  `chrome/`. This bundle separation is a discipline, not a physical guard:
  review enforces it.
- An area is deletable as a whole: drop the layout directory, the login
  kit if now unreferenced, the route files, the `AREAS` entry, and the
  area's message directory (§6).

## 3) Route discipline

Three tiers:

1. **Formal pages**: route files contain only `createFileRoute(...)` options
   (`head`, `validateSearch`, `beforeLoad`, `component`, `notFoundComponent`).
   All React components live in `features/modules/<route-name>/`.
2. **`__root`**: exception — it is the composition root; inline server
   functions, providers, and the root document may live there.
3. **Temporary routes**: exception — keep them minimal, in-place, and
   tagged `// TODO: delete`. Keep the count low: retain only enough to cover
   distinct route shapes (nested groups, leaves, dynamic segments) and
   in-flight work fronts. Menu entries derive from route
   `staticData`, so deleting a route file removes its menu item
   automatically; drop the menu message keys alongside.

**Page read-set contract.** Every formal-page route declares
`staticData.operations` = the page's **read set** — the operations of every
query the page actually issues for data. Menu visibility and route guards
judge the **whole set**: all granted → page reachable; a page is never
designed for partial-grant degradation. Mutations are not in the read set —
an unauthorized action surfaces through the error domain (toast). The only
touched surface is the route config; page code carries zero permission
logic. Menu filtering reads session `permissions` (the session source of
truth).

## 4) Domains admission

A directory earns `src/domains/<name>/` only by being either:

- **Base-owned**: consumed by the shell, which always exists (error
  vocabulary, i18n wiring, SEO server routes), or
- **Genuinely reusable across areas** (e.g. a UI-library binding: provider,
  locale registry, SSR style extraction, skins).

Area-specific mechanics (session guards, permission checks, custody,
logout for a particular login flow) are **not** domain material — they live
in the owning layout's `auth/` subdirectory. The split for permissions:
judgment stays with the area (`op`, `isOperationAllowed`,
`requireOperation`), but the error it raises is a plain word in the error
domain's table (`forbidden`) — guards throw a pre-normalized plain DTO, so
every consumer (boundary, feedback, telemetry) shares the one error
pipeline and specializes by checking `normalized.status`. Mount position is
not code position: a provider mounted inside a layout may still be
implemented in a shared domain.

Inside a domain, directories follow volatility: `core/` holds the stable
mechanism (accessors included), `config/` holds content expected to
iterate frequently (word-table rows, copy tables), and transport
integrations live in their own directory (e.g. `connect/`) as
application-level extensions — the domain stays pluggable into whatever
transport the app actually uses.

## 5) Component file discipline

- **One React component per `.tsx` file** (Fast Refresh boundaries are
  files; co-located components widen every hot-reload blast radius). Helper
  components get their own file and an `export`.
- **Exported component name = file name** in PascalCase
  (`tabbar.tsx` → `Tabbar`, no legacy area prefixes).
- **UI/logic separation** — a formal-page module splits three ways:
  `page.tsx` wires the page (connect-query hooks, URL-search parsing,
  interaction state, composition); `ui/` holds the presentational components —
  props-only (data + callbacks in), no server calls, no navigation; and
  framework-free logic lives in flat module-root `.ts` files (`search.ts` URL
  schema, `types.ts` row types/builders, domain models such as `grants.ts`) —
  pure and directly testable.
- Route-option component factories (`component: firstLeafRedirect(path)`)
  are allowed with a reasoned `biome-ignore noComponentHookFactories`:
  route options evaluate once at module scope, so component identity is
  stable.

**Naming grammar** (directories are `role-variant` (`layout-dashboard`);
entry components follow component-name=file-name (`DashboardLayout` in
`dashboard-layout.tsx`) — mechanical, not inconsistency):

| Suffix / shape | Meaning | Examples |
|---|---|---|
| `<component>.tsx` | one React component, kebab of the PascalCase export | `tabbar.tsx` → `Tabbar` |
| `*.fn.ts` | server-function modules (`createServerFn` lives here) | `session.fn.ts` |
| `*.server.ts` | server-only helpers, imported only by server code | `custody.server.ts` |
| `use-*.ts(x)` | hooks | `use-viewport.ts` |
| `*-store.ts` | state/model modules — pure model + persistence, no React | `tab-session.ts` |
| `index.ts` | feature barrel (keep exports minimal) | `features/<name>/index.ts` |

**Layout-feature grammar**: an area's layout directory keeps its root
minimal (barrel, `area.ts`, the layout entry) and organizes by role into
`auth/` (session logic), `login/` (wired login face), `protocol/` (route
protocol), and `chrome/` (all visual furniture, families graduating to
nested subpaths per the ≥4-graduation rule below).

- Prop types are named `<Component>Props` — never `IProps`.
- Same-family files share a kebab prefix; the prefix is the future
  subdirectory name (a family of ≥4 files graduates to a subdirectory).

## 6) i18n ownership

Message keys are globally flat (paraglide); **the directory is the
ownership unit**, and **every key's prefix must equal its directory name in
snake_case** (`layout-dashboard/` → `layout_dashboard_*`, `not-found/` →
`not_found_*`). A key is `namespace + "_" + semantic segments` — when keys
move between directories, replace the namespace segment and keep the
semantic tail (`menu_orders` → `layout_dashboard_orders`).

**Key order in these files is deliberate, not sorted** — entries are grouped
by semantic domain (STATUS titles first, then the reason vocabulary by proto
domain and code range). Insert a new key inside its domain group; never
re-sort the file.

**Usage-prefix rule for page namespaces** (`system-roles/`-style): position
copy is never reused across UI surfaces — a table header, a filter label
and a form label each get their own key (`…_col_status`, `…_filter_status`,
`…_form_status`). Equal wording today is coincidence, not contract, and
inflected locales (ar-SA) may need different forms per position; forking a
shared key later means touching every call site at once. Only domain
vocabulary — enum values, entity names (`…_status_on`) — is shared across
surfaces, and key groups order: page head → `filter_` → `col_` → cell
render vocabulary → status vocabulary → `act_` → empty states.

- `messages/common/`, `messages/error/`, `messages/not-found/` — base-owned
  (keep these lean: only what the shell renders). `messages/error/` mixes
  two key families by design: STATUS titles (`error_<word>_title`,
  lowercase) and the reason vocabulary mapped 1:1 from proto reasons
  (SCREAMING, e.g. `error_GENERAL_ERROR_INTERNAL`) — one directory, one
  semantic domain.
- `messages/<module>/` — owned by a formal module.
- `messages/layout-<name>/`, `messages/login-<name>/` — owned by an area;
  deleted with it.

Adding a message directory requires registering it in
`project.inlang/settings.json` `pathPattern` (explicit list — globs are not
supported). After renaming/moving/deleting keys, force regeneration
(`rm -rf src/paraglide`, rebuild) and grep the output: the build may skip
regeneration, and it does not type-check message references — typecheck
does, and it is the guard that catches stale key references (including
`staticData` title strings typed as `NoParamMessageKey`).

Adding a locale is a three-point registration: the `locales` array in
`project.inlang/settings.json` (the `{locale}` pathPattern entries pick the
files up automatically), the antd/dayjs entry in `src/domains/antd/locale.ts`,
and a native-name match arm in `common_locale_name` (the fallback prints the
raw tag). Every directory needs a complete file per locale before the
parity check (key order, key set, `{placeholder}` set against the base
locale) passes.

## 7) features/app (base utilities home)

Two resident kinds:

- **Presentation stubs** (`notfound-page`, `pending-fallback`): zero
  project-internal imports. Data comes in as props, behavior as function
  props, composition as ReactNode slots — wiring happens at the usage
  site (`__root.tsx` is the composition root). Framework/infrastructure
  imports (react, tanstack, sonner, paraglide, `shared-*`) do not count
  as external.
- **Base wiring screens and boundaries** (`error-page`, `error-fallback`,
  `base-error-boundary`): render through the error engine's `view`
  directly — no resolve-injection ceremony; catchers pass only
  `{error, onRetry?}`.

App-agnostic utilities with zero app semantics (`channel`,
`form-validator`) live in `src/libs/` (§1), not here. Error fallback
taxonomy (catch level × visual scale, base pairings) lives in §9.

Shared data hooks follow the batching discipline: a page-batching hook
(one call per visible list, e.g. avatar URLs) keys its query on a
sorted-unique signature, not array identity; when the payload is a
time-limited URL, `staleTime` stays under the server TTL so a cached entry
never outlives its signature.

## 8) Typed route constants

`AREAS` is declared `as const satisfies Readonly<Record<string, AreaWiring>>`
so literal types survive config indirection. Each login feature derives:

```ts
export const SELF_PATH = AREAS[namespace].login;          // "/login"
export const SELF_ROUTE_ID: `${typeof SELF_PATH}/` = `${SELF_PATH}/`; // "/login/"
```

`to` uses the path; `from`/`useSearch` use the id (index-route ids carry a
trailing slash). Because the literals flow from config, no `as never` casts
are needed at these sites. For fully dynamic navigation through factory
parameters (where literals are unresolvable), `redirect({ href })` plus
`URLSearchParams` is the accepted escape hatch.

## 9) Error surfaces

The error engine lives in `shared-ts/error` (zero-dependency: kind word
table `silent`/`http`, serializable envelope, three verbs — `normalize`
pure classification, `capture` dedup + policy ladder + sink fan-out,
`view` the single copy seam). Business code consumes it through
`domains/error` (connect and view-transition source adapters, sonner
feedback, sentry reporter, paraglide copy, react-query caches and
predicates) and is
zero-ceremony by default: queries stay quiet, mutations toast with the
reason copy, faults throw to the nearest boundary, silent words (canceled)
neither toast nor take over. Escape hatches: `meta: ErrorPolicy`
(`feedback`/`report`) per call site, `view(error)` for special-casing
UI, `capture(error)` for manual reporting. Renderers stay dumb — they
call `view(error)` internally and branch only on `silent`; catchers pass
`{error, onRetry?}` and nothing else.

UI surfaces split by scale:

- **Base full page** (`features/app/error/error-page`): the brand-screen sibling
  of NotFoundPage — same deep-space canvas, rose reserved as the error
  semantics' only seat. The hero numeral is the envelope's wire `http`
  (kind-profile default when local); silent words render a quiet page
  without error chrome. Wired only at `__root`'s `errorComponent`.
- **Base embedded** (`features/app/error/error-fallback`): single-scale,
  zero-skin — one row (badge + title + optional retry link) on the
  currentColor ladder, zero motion; pairs with `pending-fallback` as the
  uncertainty pair (halted/live halves), style-neutral so any area can
  host them. Silent words collapse to a neutral inline state. Recovery is
  the catcher's business: `onRetry` is injected by the wiring layer,
  never defaulted. Height: the pair fills its mount slot — `size-full`
  where the host height is definite, `grow` where it is a flex column;
  slots that center them supply one of the two (the content-area
  watermark wrapper is a flex column, the root slot a min-h-svh rail).
- **Base component-level** (`features/app/error/base-error-boundary`):
  react-error-boundary with `resetKeys=[leafRouteId]` — TanStack Router
  swaps outlet children without remounting the boundary, so without a
  route-scoped reset key a rendered fallback sticks across navigation.
  Component boundaries also shield downstream route `errorComponent`s:
  child loader errors without their own errorComponent rethrow through
  the outlet and land on the nearest boundary, not the route error page.
- **Area level** (inside the layout): areas own no error UI of their own.
  Route-level catchers mount the base `error-fallback` directly;
  component boundaries mount `base-error-boundary` directly. A layout
  route's `errorComponent` renders in place of the whole layout (no
  provider, no chrome) — the style-neutral fallbacks survive that bare
  context by design, which is one reason area-dialect error faces stay
  out.

Verified behavior: SSR first paint of route errors renders server-side
with locale-correct copy; across the server-fn boundary connect's native
error JSON reconstructs a real `ConnectError` client-side, but every
other thrown value resolves as data (the boundary rethrows only
`instanceof Error`) — so server fns return status values instead of
throwing envelopes. Telemetry groups by `[kind, reason]` so error storms
aggregate into one issue. Route error recovery needs
`router.invalidate()` — the boundary's `reset` alone never re-runs
loaders.

Session-flow navigations (kick mirror, self-kick, logout, expired-param
strip) pass `viewTransition: false`: the View Transitions API skips the
**entire commit** in a hidden document (spec InvalidStateError), which
strands a background kick on the old page — a stranded kick then only
recovers via focus-refetch, which `staleTime` can suppress inside its
window. Any other hidden-tab transition abort is normalized to the
silent `canceled` word by the domain's view-transition source adapter.

An unreachable backend is a session verdict, not an auth one: the
session fn returns `unreachable` on `network`-kind failures, the guard
drops the cache entry and throws a `network` envelope (never a
redirect — the session is unverified, not absent), and the watcher
ignores the state so downtime never reads as a session flip.

## 10) RTL

Always Tailwind logical properties/variants (`ms-/me-/ps-/pe-`, `rtl:`
variant); never couple direction to locale strings. Mirror gradients and
directional offsets. Translatable text that changes length per locale needs
`rtl:tracking-normal` guards where letter-spacing is used decoratively.

## 11) Verification

Per change: `admin:typecheck`, `admin:check`, `admin:build` — all with
`--skip-nx-cache`, all exit codes checked. Order matters after route or
message edits: **build first** (regenerates `routeTree.gen.ts` and
paraglide output), then typecheck (build alone does not type-check message
or route-tree references). Regenerating those outputs does not dirty the
source files that import them, so tsc's incremental cache (`*.tsbuildinfo`)
skips them and reports stale greens — delete the tsbuildinfo before the
typecheck that follows a regeneration.

**Behavioral review for library seams** (typecheck cannot see these — they
are the recurring bug class): whenever code wires into, migrates between,
or adapts a library-owned behavior (dnd, wheel/scroll physics, menus,
toasts), enumerate the library's observable contract next to our
implementation and check each pair — semantics, axis, direction, ordering,
identity. Adapters that survive a runtime migration (e.g. HTML5 DnD →
dnd-kit) are the classic hiding place: the old semantics only agrees with
the new contract in one direction. Pure model functions at these seams get
unit tests pinning the semantics (`tab-session.test.ts` is the pattern).

**Reachable-surface rule**: any behavior change on a surface reachable
without login (login page, guard redirects, SSR entry) gets e2e-verified
before reporting done; short-lived UI is observed with a single
self-contained script installed before the event (a from-document-start
MutationObserver for one-shot notices), never with after-the-fact polling.
Auth-only surfaces (tabbar interactions, chrome) carry an explicit
user-verification list in the change report instead.
