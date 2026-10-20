# TanStack Start Client Conventions

**Status:** ACTIVE

> Rules here win over precedent in code when they conflict. Keep this file in
> sync when a rule changes (see root AGENTS.md §7 and docs/README.md).

Applies to web clients built on TanStack Start (`app/clients/*`).

---

## 0. Execution contract

In force for every change under `app/clients/*` and `shared-ts/**`; rules here win over precedent in code.

1. Before writing a file below, load the sections in its dispatch row — they are the law for that artifact.
2. Before reporting done, run the §12 gate over the combined diff and emit its report table in the final message. No gate table = not done, whatever the build says.
3. Gate = exit code 0 or the Nx "Successfully ran" banner — a log tail is not a gate.
4. The §12 close runs for every change regardless of dispatch row — the chain row T-VER-1 is always in force.

| About to write | Load first | Gate rows |
|---|---|---|
| `src/routes/**` | §3 | T-STR-3 T-RTE-1 T-RTE-2 T-RTE-3 T-RTE-4 |
| `features/modules/<m>/page.tsx` | §1 transport shapes, §3 read-set, §5 UI/logic split | T-LAY-1 T-RTE-1 T-REA-2 |
| `features/modules/<m>/ui/**`, flat module `.ts` logic | §5 | T-STR-1 T-STR-2 T-REA-1 T-REA-2 |
| `features/layout-*/**`, `features/login-*/**` | §2, §5 grammar, §10 | T-LAY-2 T-STR-4 T-ERR-1 |
| `messages/**` (any locale file) | §6 | T-I18N-1 T-I18N-2 T-I18N-3 T-I18N-5 T-I18N-6 |
| `domains/**` | §4, §9 | T-LAY-1 T-ERR-1 |
| anything else under `app/clients/*` or `shared-ts/**` | §1, the touched layer's section, §7 for `shared-ts/<domain>/src/**`, §11 | full §12 |

---

## 1. Layering model

| Layer | Location | Rule |
|---|---|---|
| Route declaration | `src/routes/` | Only `createFileRoute` wiring — no components, no logic (§3) |
| App shell utilities | `src/features/app/` | The base's generic subpath: shell screens, boundaries, shared helpers; may import domains (§7) |
| Areas | `src/features/layout-<name>/`, `src/features/login-<name>/` | An area = one layout directory (the owner) + one login UI kit (§2) |
| Formal pages | `src/features/modules/<route-name>/` | One directory per route, `page.tsx` entry |
| App-level components | `src/components/` | Self-contained only — no project-internal dependencies; depended on, never depending |
| Domains | `src/domains/` | Admission-gated (§4) |
| Infra | `src/stores/`, `src/services/`, `src/libs/` | As conventional — `libs` holds app-agnostic utilities (zero app semantics, import nothing app-internal) |

Dependency direction: `routes → features → domains → services/stores/lib`. One
recorded counter-flow: `services/connect` mounts `domains/error`'s `normalize`
into the transport factory — the transport seam, and the sole exception the
conformance guard enforces.
Features may import each other only along the area-pair direction (§2), plus one
hosting direction: a formal-page module (`features/modules/*`,
`features/app/*`) imports its area's layout utilities — guard types, route
context, `permissions`/`op`, area search stores — and base helpers. The
layout never imports a module back; modules of different areas never import
each other.

**Transport responses are JSON-shaped, not message instances.** The connect
transport (`services/connect/transport.ts`) re-serializes every unary
response through proto-es `toJson`: WKT wrappers arrive flattened to scalars,
proto `map` fields as plain objects, enums as value-name strings. The
generated TS types claim Message/Map shapes the runtime does not deliver —
consume accordingly (`Object.entries` over maps, string/index access for
enums).

**Server-streaming is consumed imperatively, outside the query cache.** The
transport's stream entry takes a generated method descriptor (`*_pb.ts`) plus
one `AbortController` per stream; a feature hook drives it directly and
accumulates frames into component state or a page-injected store interface.
Frame intake normalizes protojson omissions (`frame.contentDelta ?? ""` —
unset scalars are absent, not ""). Abort is the normal close: drop the abort
error, release the controller in `finally`; a controller map keyed by stream
id guards against double-fire.

**Persisted store state is superjson-wrapped.** Cookie-persisted stores
serialize through superjson (a `{"json":…}` envelope, `shared-ts/store`); a raw
`JSON.parse` on the payload silently yields `undefined` — hydration goes through
the store's own deserialize path, never a hand-rolled parse.

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
  `src/stores/` under the single eager-registration barrel (the registry
  drives server-side cookie collection — a late code-split registration
  misses hydration); one store per file; an area-scoped store carries the
  area namespace in directory, variable, and cookie-key names
  (`dashboard-preferences` → `dashboardPreferencesStore` →
  `store_dashboard_preferences`) so sibling layouts never collide.
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

**`staticData.title` is a typed message key** — `MessageKey` = `keyof typeof m`
(`route-meta.ts`): the bare key literal is the only legal form; typos,
template strings, and variables (type `string`) all fail typecheck. No
constant table — the key set has no second source.

**Dot-named nesting.**

- A dot-named route file (`login.wechat.callback.tsx`) is a child of each shorter prefix (`/login/wechat`, `/login`): every parent on that chain must be a layout route — its component renders `<Outlet />`, and the parent's own page lives in `login.index.tsx`.
- A parent with a direct `component` never mounts the child while the child's `head` still applies — a correct SSR title over the parent's view is the drift symptom (right URL, parent's render), not a router bug.

**Page read-set contract.**

- Every formal-page route declares `staticData.operations` = the page's **read set** — the operations of every grant-gated query the page actually issues for data.
- Builtin faces (`ACCESS_BUILTIN`) never enter a read set — authentication-guaranteed, the builtin chain never consults grants; a builtin member would make the page unreachable for every non-wildcard user (an upload probe's `ListUploadedParts` is the cautionary example).
- Menu visibility and route guards judge the **whole set**: all granted → page reachable. A page is never designed for partial-grant degradation.
- Mutations are not in the read set — an unauthorized action surfaces through the error domain (toast).
- The only touched surface is the route config; page code carries zero permission logic. Menu filtering reads session `permissions` (the session source of truth).

**Tab identity (dashboard chrome).**

- A route owns a tabbar tab iff its `staticData` declares a `title`; title-less routes render inside the nearest titled ancestor's tab (the profile group pattern). A detail page joins its list's tab by nesting under the titled route as a `menu.hide` child with no `title` of its own (its `head` carries the page title); the titled route is then a layout whose own page lives in `index.tsx` — and a menu-bearing layout whose children are all deliberately hidden stays a leaf nav entry (children dropped for grants still prune it — hide-by-default holds).
- `key` = resolved pathname (search excluded; one tab per param-route instance, e.g. each `/order/$id` id); `href` = full href including search; `pattern` = route `fullPath`, kept for permission-filtered restore.
- Activation, close, and href capture compare the **exact key** from `tabFromMatches` — path-prefix ownership is forbidden.
- Router `location` and `matches` lag each other by one commit during navigations; every matches-derived tab decision gates behind `matchesAgreeWith(matches, pathname)` — torn frames can neither capture a foreign href nor re-append a just-closed tab.
- Session restore resolves `pattern` against the permission-filtered nav index; hidden routes (`menu.hide`) restore via their persisted title; a pattern present in the unfiltered index is a permission-dropped visible route and stays closed.
- The active tab seeds through the same pattern rule — a direct URL load lands in the bar even with no persisted record; the restore effect rebuilds membership after the append effect's mount run and must not drop it.
- The home tab is seeded only when a titled leaf exists at the area root — apps without a workbench start with an empty bar.
- Default post-login landing is `defaultLanding(permissions)` — the first permission-surviving leaf, never a hardcoded area root.
- Parameterized titles interpolate `match.params` (`compactParams`, `text(key, params)`).

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
- **Render-path state reads go through state, never a ref.** React Compiler
  caches render output — a render that dereferences a ref renders that ref's
  stale value until an unrelated re-render (ghost state). Refs are read in
  effects, callbacks, and event handlers only.

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

**UI-library integration laws** (each pinned by a live regression): antd
`Watermark` wraps children in an inline `overflow: hidden` — sticky descendants
never stick inside one; keep sticky elements outside the watermark. antd v6
`Drawer` sizes via `size` presets — numeric `width` is the retired v5
vocabulary. antdx composition pieces (`Creation` regions, `Sender` attachments)
mount through their declared slots, never children injection. Tailwind v4's
class extractor tokenizes on dots — arbitrary values containing decimal points
split into dead classes; prefer scale values or CSS variables.

### Skin presets (`shared-ts/antd/src/skin-presets/`)

Law for adding or modifying skins — the shared-ts capability directory, recorded here as the client-stack home.

**System map**

- A skin is a `SkinPlugin` (`../skins/types.ts`): `{ id, name, bodyBg: { light, dark }, Provider }`. `bodyBg` feeds Storybook's canvas; the app reads surface colors from tokens, not from this field.
- One file per skin, default-exported (`skin-presets/<id>.tsx`). Registration is the only wiring step — the import + entry in `SKINS` (`../skins/registry.tsx`); every consumer (settings drawer grid, Storybook toolbar, `SkinSwitcher`) derives from `getAllSkins()`.
- Providers stay pure and cheap: props (`isDark`, `compact`, `children`) only — the settings drawer renders every thumbnail inside that skin's own Provider, so expensive providers multiply across the grid.

**Three styling channels — escalate in order**

1. `theme.token` (global) and `theme.components.<C>` — cheapest, cascades to every instance; prefer for colors, radius, motion, fonts.
2. Semantic style slots on ConfigProvider (`<component>: { styles: { <slot>: CSSProperties } }`, v6 semantic DOM) — for values antd hardcodes on real elements (Switch track radius, popup surfaces, panel paddings); slot names are typed — typecheck is the drift alarm when antd upgrades rename slots.
3. `createStyles` selectors (`antd-style`) — the only channel reaching pseudo-elements and internal descendants (`.ant-switch-handle::before`, `.ant-steps-item-icon`); React style objects cannot target `::before`; `cssVar.*` works only inside createStyles, static `CSSProperties` use `var(--ant-*)` strings.

**Token risk classes** — colors are layout-inert; geometry and metric tokens change box sizes and therefore layouts: `borderRadius*`, `controlHeight`, `fontSize`/`lineHeight`, paddings, `fontFamily` (mono stacks run ~10% wider), motion durations. Every geometry token change demands a Storybook sweep, not a spot look.

**Chrome token contract** — the app chrome consumes a fixed contract each skin fills in its own palette: `Layout: { bodyBg, headerBg: <mode container>, headerHeight: 56, siderBg + lightSiderBg: <mode container> }`, `Menu: { itemHeight: 40, darkItemBg/darkSubMenuItemBg: "transparent" }`. `skin-presets/default.tsx` plus lyra/luma are the checklist; a skin missing any slot falls back to antd defaults (wrong hue or geometry) somewhere in the layout.

**Anti-patterns** (symptom → cause → fix):

- Default skin's thumbnail ghosts the active skin — nested ConfigProviders merge; a skin with an empty `token` renders as a tinted clone of whatever is active → the default skin passes its full resolved map (`LIGHT_TOKENS | DARK_TOKENS`), the others set every seed the visible surfaces derive from; never `theme.inherit: false` here (it drops the domain provider's breakpoint alignment tokens).
- Blue-black patch on the dark-mode sider — antd's generic dark-menu background paints over the skin → `Menu: { darkItemBg/darkSubMenuItemBg: "transparent" }`.
- Topbar 64px under one skin, 56px under another — `Layout.Header` height is the `headerHeight` token (antd default 64) → pin 56 in every skin.
- Floating sider painted with generic antd colors in one mode — `Layout.Sider` reads `lightSiderBg` under `theme="light"` but `siderBg` under `theme="dark"` → pin both.
- Switch stays round in a square skin — geometry-intrinsic shapes (Switch, Steps icons, Slider handles) ignore the global `borderRadius` → reshape via all three channels at once (root/content slots for track and clip, a createStyles selector for `.ant-switch-handle::before` — the semantic `indicator` slot is the invisible positioning wrapper).
- Blue hover/selection in Select options — option tokens don't inherit `colorPrimary` for backgrounds → pin `optionSelectedColor/optionSelectedBg/optionActiveBg` to the skin primary.
- Button overrides misfire on colored/danger buttons — variant + color + type interact; the `utils.pickButtonVariant` dispatch encodes tested per-variant decisions → keep the dispatch structure verbatim from lyra/luma, restyle only the mapped style objects; dedupe keys only on byte-identical CSS.
- Text truncates, rows wrap, scrollbars appear under one skin — a geometry token moved a box; keep `borderRadiusXS` under half the smallest chip height; the compact algorithm compounds tightness — check normal × compact; diagnose on the Storybook density surfaces, not in the app.
- Layout regressions under a skin — providers render exactly one sanctioned DOM wrapper (the Steps/Slider patch); more wrappers break `h-svh`/flex chains.

Procedure (new-skin recipe, tuning loop): docs/conventions/tanstack/RUNBOOKS.md §1 "New skin".

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
`project.inlang/settings.json` `pathPattern` (explicit list — no globs).
After renaming/moving/deleting keys, force regeneration (`rm -rf
src/paraglide`, rebuild) and verify: the build may skip regeneration, and
compiled files are named **lowercased id + numeric dedup suffix**
(`error_SYSTEM_TENANT_DISABLED` → `error_system_tenant_disabled20.js`), so an
exact-case grep of the output is always a false negative. The build does not
type-check message references — typecheck does, and it is the guard that
catches stale key references (including `staticData` titles typed as
`NoParamMessageKey`).

Adding a locale is a three-point registration: the `locales` array in
`project.inlang/settings.json` (the `{locale}` pathPattern entries pick the
files up automatically), the antd/dayjs entry in `src/domains/antd/locale.ts`,
and a native-name match arm in `common_locale_name` (the fallback prints the
raw tag). Every directory needs a complete file per locale before the
parity check (key order, key set, `{placeholder}` set against the base
locale) passes.

## 7) features/app (base utilities home)

Four resident directories:

- **`error/`** — the error-engine faces: `error-page` (brand-screen full
  page), `error-fallback` (single-scale embedded row), `base-error-boundary`
  (route-scoped boundary) render through the error engine's `view` directly —
  no resolve-injection ceremony; catchers pass only `{error, onRetry?}`.
  `notfound-page` is the brand-screen sibling — props-only, zero
  project-internal imports.
- **`feedback/`** — `feedback-sonner` (toast wiring) and `pending-fallback`
  (the uncertainty pair's live half): presentation stubs — props-only,
  zero project-internal imports. Framework/infrastructure imports (react,
  tanstack, sonner, paraglide, `shared-*`) do not count as internal.
- **`file/`** — the upload surface (`put-url.ts` XHR direct PUT, presign
  confirm, `use-file-url` id gating, the proxy/upload hooks).
- **`user/`** — identity display utilities (directory hydration,
  deleted-tag, user-ref).

App-agnostic utilities with zero app semantics (`channel`,
`form-validator`) live in `src/libs/` (§1), not here. Error fallback
taxonomy (catch level × visual scale, base pairings) lives in §9.

Shared data hooks follow the batching discipline: a page-batching hook
(one call per visible list, e.g. avatar URLs) keys its query on a
sorted-unique signature, not array identity; when the payload is a
time-limited URL, `staleTime` stays under the server TTL so a cached entry
never outlives its signature. A singular URL hook (one object's URL,
`use-file-url.ts`) gates on id presence before hydrating — an absent id is
render state, not a fetch error.

**Browser direct uploads PUT through XHR** (`put-url.ts`), never fetch:
upload progress rides `xhr.upload`, and the ETag header multipart confirm
needs is only readable through a CORS-exposed XHR response. `Content-Type`
is declared client-side — the presign does not sign it and the backend
stores the header verbatim.

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

Verified behavior (pinned laws, not observations):

- SSR first paint of route errors renders server-side with locale-correct copy.
- Across the server-fn boundary: connect's native error JSON reconstructs a real `ConnectError` client-side; every other thrown value resolves as data (the boundary rethrows only `instanceof Error`) — so server fns return status values instead of throwing envelopes.
- Telemetry groups by `[kind, reason]` — error storms aggregate into one issue.
- Route error recovery needs `router.invalidate()` — the boundary's `reset` alone never re-runs loaders.

Session-flow navigations (kick mirror, self-kick, logout, expired-param
strip) pass `viewTransition: false`: the View Transitions API skips the
**entire commit** in a hidden document (spec InvalidStateError), stranding a
background kick on the old page — its only recovery is focus-refetch, which
`staleTime` can suppress. Any other hidden-tab transition abort normalizes
to the silent `canceled` word. Every session exit — guard redirect, watcher
kick, deliberate logout — hands the current href to the login page as
`?redirect=` so re-login returns to the origin.

View-transition geometry is percentage-based, never px (px resolves against
the snapshot containing block and drifts); DOM mutations in transition
callbacks flush with `flushSync` before capture; morphing frames get
`transition: none` — the three recurring VT bugs.

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

Per change: for each client under `app/clients/*` affected by the change,
its `<client>:typecheck`, `<client>:check`, `<client>:build` — all with
`--skip-nx-cache`, all exit codes checked. Order matters after route or
message edits: **build first** (regenerates `routeTree.gen.ts` and
paraglide output), then typecheck (build alone does not type-check message
or route-tree references). The clients run tsc without `incremental` — there
is no `*.tsbuildinfo` to go stale; every typecheck is a full check. The
stale-green risk lives one level up, in the Nx task cache — hence
`--skip-nx-cache` on every gate leg.

Per change, also run `./nx run tools:paraglide:check` — the build compiles a key
missing from one locale from the base locale and exits 0, so locale parity has
no other gate.

**Behavioral review for library seams** (typecheck cannot see these — they
are the recurring bug class): whenever code wires into, migrates between,
or adapts a library-owned behavior (dnd, wheel/scroll physics, menus,
toasts), enumerate the library's observable contract next to our
implementation and check each pair — semantics, axis, direction, ordering,
identity. Adapters that survive a runtime migration (e.g. HTML5 DnD →
dnd-kit) are the classic hiding place: the old semantics only agrees with
the new contract in one direction. Pure model functions at these seams get
unit tests pinning the semantics (the `domains/error` source-adapter
tests are the pattern).

**Reachable-surface rule**: any behavior change on a surface reachable
without login (login page, guard redirects, SSR entry) gets e2e-verified
before reporting done; short-lived UI is observed with a single
self-contained script installed before the event (a from-document-start
MutationObserver for one-shot notices), never with after-the-fact polling.
Auth-only surfaces (tabbar interactions, chrome) carry an explicit
user-verification list in the change report instead.

---

## 12. Pre-completion gate

Run over the combined diff (working tree plus the commits being amended into), per owning client under `app/clients/*`. One row per item, emitted verbatim in the final report:

| item | verdict | evidence |
|---|---|---|
| T-VER-1 | PASS | `<client>` chain exit 0 (build → typecheck → check, `--skip-nx-cache`) |

PASS / FAIL / N/A are legal verdicts; an absent row is not. N/A names its glob and shows it empty. `lint:` rows scan the added lines of the combined diff; `eye:` rows cite file:line.

Posture (root `AGENTS.md` §6.1): declare the change-kind and close with the counted-evidence line. Kinds here — surface: this table plus browser closed loop (viewports, locales, zero console; no browser-automation tool → `docs/environment.md`, manual walkthrough is the degraded form); logic / refactor: this table, refactor with cache bypass and counterexample replay. Absorb-time mechanical checks: deleted tokens/keys grepped for remaining consumers; new utility classes traced to their token source.

### Structure — law §5

- **T-STR-1** — one React component per `.tsx` file; helper components get their own file and export. eye.
- **T-STR-2** — exported component name = file name, PascalCase export, kebab file. eye.
- **T-STR-3** — route files hold only `createFileRoute` options (§3 tiers); `__root` and tagged temporary routes are the recorded exceptions. eye.
- **T-STR-4** — file-suffix grammar honored (`*.fn.ts`, `*.server.ts`, `use-*.ts`, `*-store.ts`); prop types `<Component>Props`, never `IProps`. eye.

### Layering — law §1, §2, §7

- **T-LAY-1** — dependency direction `routes → features → domains → services/stores/libs`; feature-to-feature imports only along the area-pair and hosting directions; the transport seam is the sole counter-flow. eye: import list per added file.
- **T-LAY-2** — area isolation: the layout never imports a module back; no cross-layout imports; the login kit imports no layout code; area-scoped stores namespace directory, variable, and cookie key. eye.
- **T-LAY-3** — `libs/` stays app-agnostic — local imports reach only siblings and paraglide. eye.

### React law — law §5

- **T-REA-1** — render-path state reads go through state, never a ref (React Compiler ghost-state). eye: added render bodies.
- **T-REA-2** — UI/logic split: `ui/` components props-only (no server calls, no navigation); framework-free logic in flat module-root `.ts` files. eye.

### Routes — law §3

- **T-RTE-1** — every formal-page route declares `staticData.operations` = the page's full read set; reachability judges the whole set; page code carries zero permission logic. eye.
- **T-RTE-2** — builtin faces (`ACCESS_BUILTIN`) never enter a read set. eye.
- **T-RTE-3** — `staticData.title` is a bare typed `MessageKey` literal; typos and variables fail typecheck. lint: `<client>:typecheck` green after regeneration.
- **T-RTE-4** — dot-named nesting: every parent on the chain is a layout route rendering `<Outlet />`; the parent's own page lives in `index.tsx`; a detail page joins its list's tab as a `menu.hide` child without its own `title`. eye.

### i18n — law §6

- **T-I18N-1** — every new key's prefix equals its owning directory name in snake_case. lint: `./nx run tools:paraglide:check`.
- **T-I18N-2** — new keys inserted inside their semantic group; files never re-sorted. eye.
- **T-I18N-3** — every touched directory holds a complete file per registered locale; key set, key order, and `{placeholder}` parity against the base locale. lint: `./nx run tools:paraglide:check`.
- **T-I18N-4** — a new locale registers at all three points (inlang settings, antd/dayjs entry, native-name arm). eye.
- **T-I18N-5** — after key moves/renames/deletes, paraglide output force-regenerated (`rm -rf src/paraglide` + build) before the typecheck. lint: typecheck green post-regen.
- **T-I18N-6** — position copy never reused across surfaces (`…_col_`/`…_filter_`/`…_form_` forks); only domain vocabulary shared. eye.

### Errors & sessions — law §9

- **T-ERR-1** — business code consumes the error engine zero-ceremony; catchers pass `{error, onRetry?}` only; renderers branch on `view` output; session-exit navigations pass `viewTransition: false` and hand the current href off as `?redirect=`. eye.

### Verification — law §11

- **T-VER-1** — per owning client: build → typecheck → check, all exit 0, `--skip-nx-cache` (the Nx cache replays stale greens). lint: gate outputs.
- **T-VER-3** — behavior changes on surfaces reachable without login e2e-verified before done; auth-only surfaces carry an explicit user-verification list. eye: links or the list.
- **T-VER-4** — library-seam changes get the §11 behavioral pairing review (the library's observable contract enumerated next to our implementation, each pair checked). eye.

T-VER-2 is retired (the tsbuildinfo ritual it named never existed — §11); IDs do not renumber.
