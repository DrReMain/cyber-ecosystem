# ADR-0002: Admin Client — Data Path & Auth Architecture

- **Status:** Direction converged. **Implemented & validated**: services layer (D8–D9), error-handling chain (D10 part A), auth full loop (login/logout/guard/401-recovery, all browser+bff×connect+http quadrants). **Prototype-validated**: error content humanization/i18n (D10 part B, experiment matrix in section). Temporary ADR — folds into `docs/client/CONVENTIONS.md` once fully validated, then removed (same lifecycle as ADR-0001).
- **Date:** 2026-08-07 (revised 2026-08-13: D8–D10 chain implemented; 2026-08-16: D10-B prototype experiments, reason contract finalized as runtime-verified)

## Context

`app/clients/admin` is a TanStack Start (SSR) client to the `system` backend. The backend (ADR-0001) exposes three transports (gRPC / HTTP / Connect); **HTTP and Connect are exposed externally via traefik in both compose and k3s/k8s**, and kratos v3 serves both unary and non-unary (streaming) on HTTP and our Connect transport. Backend auth is **opaque-token** (access + refresh in Redis, `Bearer`).

Two hard constraints shape this ADR:

1. **The client must be dual-protocol** — Connect primary, HTTP secondary. HTTP stays first-class for special cases and for a future **open platform** that exposes capabilities externally (HTTP is the universal choice there).
2. **Follow official TanStack Start guidance where viable** — server functions for server-concerns (secrets, auth, SSR), TanStack Query on the client for interactive data.

The official TanStack Start model is BFF-via-server-functions. Naively applied, this **conflicts** with constraint 1: under pure BFF the browser sees only server functions and the Connect/HTTP choice collapses into a server-fn detail — killing the browser dual-track and connect-query's client-side value, and forcing streaming through two layers.

## Decision principle

**Hybrid, split per concern — not pure BFF, not pure browser-direct.** Route each concern by what suits it: auth/session/aggregation/SSR-prefetch go through server functions (BFF); business CRUD and streaming go browser-direct over the dual-protocol transport. The token model is split so both paths authenticate without exposing long-lived secrets to the browser.

## Decisions

### D1 — Hybrid data path (BFF + browser-direct, per concern)

| Concern | Path |
|---|---|
| Auth/session: login / logout / getCurrentUser / refresh | **BFF** (server fn) |
| Aggregation / shaping / multi-call composition | **BFF** (server fn) |
| Route first-paint SSR prefetch (loader) | **BFF** (server fn, server-side context) |
| Business CRUD: list/get/create/update/delete | **browser-direct** (connect-query) |
| Streaming (non-unary) | **browser-direct** (native Connect stream) |

Pure BFF loses browser dual-protocol + connect-query value + forces two-layer streaming; pure browser-direct loses SSR, weakens security, and needs CORS.

Default stance (clarified 2026-08-17): **browser-direct is the default; SSR prefetch (loader) is per-page opt-in** for SEO/first-paint. Admin pages behind the login wall rarely need SSR data — the stance mainly serves future shared-ts reuse on public-facing sites, and keeps the SSR recovery branch the rare path.

### D2 — Auth boundary is the server function, not `beforeLoad`

`beforeLoad` is **route UX only**. **Every server function that reads/writes private data authorizes itself** (session lookup → reject if absent/invalid). Mirrors the official rule and ADR-0001's data-boundary stance.

### D3 — Token split: refresh server-side, access short-lived to browser

- **refresh token** → httpOnly session cookie, **server-side only**.
- **access token** → short TTL, issued to browser by the login/refresh server fn; held **in memory**; transport interceptor attaches `Authorization: Bearer <access>`.
- **401 on a direct call** → refresh server fn (rotates via httpOnly refresh) → new access → retry once.
- **SSR prefetch** (BFF): server fn obtains access server-side; token never leaves server.

Bounds exposure (short access TTL), keeps refresh off-browser (XSS-proof), dissolves the "BFF session vs direct Bearer" tension.

### D4 — Dual-protocol at both layers

- **Browser-direct**: `connect-query` (Connect) **primary** + `@cyber-ecosystem/gen-openapi-ts` (HTTP SDK, hey-api / OAS→TS) **secondary**.
- **BFF**: server-side Connect client + HTTP `fetch`.

### D5 — Streaming (non-unary) is always browser-direct

Native browser → traefik → backend. Never proxied through a server fn (two-layer stream, latency/cancellation cost).

### D6 — Edge injection (traefik ForwardAuth) is optional, deferred hardening

Production hardening only: traefik `ForwardAuth` validates session + injects Bearer upstream, so even the access token stays server-side. Deferred — paradigm-identical, heavy for validation phase.

### D7 — Two-layer value semantics: wire nullable; client per-transport

Wire = the contract (WKT wrappers, protojson `EmitUnpopulated`, every field present: unset→`null`, zero→typed zero). Client = two type sources: **Connect** (protobuf-es, absence=`undefined`) vs **HTTP** (hey-api, unset stays `null`). Both correct for their transport. `==` banned; judge presence with transport-appropriate `===`. Unifying would drop the generated HTTP SDK — rejected.

### D8 — Services layer: clients & interceptors (implemented)

`services/{connect,http}/` — one subdir per protocol, client + interceptors self-contained; the extension point for future transports (streaming/seaweedfs/livekit/centrifugo, each reusing Connect transport or own client). **Cross-cutting lives in the domain layer** (locale via paraglide today; access-token source `domains/auth` rebuilt when auth returns) — services only adapt & compose, no duplicated logic.

### D9 — HTTP cross-cutting via client interceptors (implemented)

`services/http/client.ts`: `createClient({ throwOnError:true })` + `interceptors.request`(locale) + `interceptors.error`(response-body → `Error`). `throwOnError` is a **sdk contract switch** — must stay `true` (`false` returns `{error}` and silently breaks the `.data` pattern). **Auth/401 retry**: the error interceptor runs in `catch` and cannot re-fetch → candidate seam is a custom `fetch` wrapper (when auth returns). Business code stays a plain call, zero auth/header/retry awareness, mirroring connect-query.

### D10 — Error handling

#### A. Capture → normalize → collect → feedback (validated 2026-08-16, reverted with the baseline; formal rebuild = S1)

- **capture**: route `errorComponent` (boundary — render/loader throw, `throwOnError:true`) + `window` listen (handler/async JS) + `Query/MutationCache.onError` (request).
- **normalize**(error)→`ApiError`; **collect** (WeakSet dedupe by `original` → 1× not 2–3×); **feedback** (mutation → injected antd `message`, `meta:{silent:true}` override; seam `feedback.ts` + antd `feedback-holder.tsx`, errors domain imports no UI lib).
- **throwOnError**: query global `true` (page-data fail → `ErrorPage`, 整页 **replacement**); mutation `false` (in-place + toast). ErrorBoundary **replaces** the subtree — "保留原 UI + 半透明遮罩" is **not** ErrorBoundary; a query needing that uses `throwOnError:false` + `query.error` business render.
- **Presentation default legislated 2026-08-17 (render-error-first)**: query → boundary / mutation → toast is the law, inline `query.error` is opt-in; SSR loader failures are accepted render errors (no loader-catch convention). Granularity comes from a **component-level ErrorBoundary primitive** (same projection as route ErrorPage, smaller) landing in S1 — route boundary stays the backstop.

#### B. Error content — humanization + i18n (prototype experimentally validated 2026-08-16)

Four orthogonal concerns, kept decoupled:

| concern | owner | note |
|---|---|---|
| ① **i18n** | independent layer (serves all UI, not just errors) | paraglide now; lingui (RN) / remote (hot-update) remain future interchangeable impls |
| ② **errors domain** | consumes ① via `errorMessage` in i18n domain | normalize → kind/reason → projection → feedback; binds no i18n lib, no UI lib |
| ③ **auth** | transport + guard (independent of errors feedback) | 401 split by reason |
| ④ **transport** | standardizes `Error{reason,code}` | flattens connect/http shape difference |

**Backend error shape (verified):** identity = **reason** (UPPER_SNAKE enum name) + **code** (HTTP/connect); **message always empty** (human text lives in private `cause`, not sent to client). Enums: `GeneralError`(100-503) / `InfraError`(3000-3402) / `FlowError`(4000-4003) / `ErrorReason`(6xxx, per-domain). Auth middleware errors are mapped by `server.go init()` onto the enums — the wire only ever carries `GENERAL_ERROR_TOKEN_EXPIRED` / `GENERAL_ERROR_UNAUTHENTICATED` (verified incl. real-expiry via redis key delete). Connect carries reason in `google.rpc.ErrorInfo` detail; http in `body.reason`.

- **B1 transport standardization (implemented & verified)** — connect `errorInterceptor`: `findDetails(ErrorInfoSchema)` → reason attached to a re-thrown `ConnectError` (must stay ConnectError: a plain `Error` gets re-wrapped by connect-es as `Unknown`, losing the code). http error interceptor: reason/code from body, **code falls back to `response.status`** (non-JSON bodies like 502 html carry no code — without the fallback, kind classification degrades to `unknown`; verified by unit probe).
  - **@connectrpc/connect v2 API facts (verified 2026-08-17, supersedes experiment-era notes)**: the code enum is **`Code`** (not `ConnectCode`); `findDetails` is a **ConnectError method** (not a standalone function); human text is **`rawMessage`** (`message` carries the `[unavailable] ` code prefix); `ErrorInfoSchema` lives at `gen-ts/google/rpc/error_details_pb`; message classes are types — construct via `create(Schema, init)`; `error.details` holds `OutgoingDetail {desc, value}` locally but **wire-arriving details are `IncomingDetail` matched by typeName via a registry** — locally-constructed unit fixtures exercise the outgoing branch only, so reason extraction must be confirmed against the real wire (this exact gap drove the 2026-08-17 T1 revoke: unit tests validated a fiction of the wire while green).
- **B2 reason contract: runtime-verified, NOT code-generated** — the full wire names live in the generated Schema descriptors (`GeneralErrorSchema.values[i].name`, byte-exact; verified). A hand-written descriptor parser (post-gen) was tried and **rejected**: one silent skip-field bug dropped 3/4 enum families without failing the build — the risk class is real, and the only hard consumer of a compile-time `Reason` union (copy completeness) is equally served by a **CI reconciliation test** (`reason-coverage.test.ts`: core reasons hard-asserted, full set soft-reported by name; new backend reason without copy fails CI naming what's missing). See `docs/client/CONVENTIONS.md §5`.
- **B3 copy projection (implemented & verified)** — `errorMessage(ApiError)` in the i18n domain (type-only dep on errors): reason tier `m["error_" + reason]` (paraglide emits large-case keys via ESM string re-exports — dynamic lookup is a supported mechanism, verified) → kind tier `error_kind_*` (always present — never renders a raw reason) → `e.message` last resort. Verified matrix: http/connect × en/zh all render human copy ("凭证无效或未登录" etc.).
- **B4 kind taxonomy (implemented, simplified from design)** — `auth` (login-state semantics: UNAUTHENTICATED/TOKEN_EXPIRED) / `biz` (any other reason) / `network` (connect Code 14 or HTTP ≥500) / `unknown`. Drives projection + future retry short-circuit.
- **B5 auth 401 split (belongs to auth spike / Slice 3; verified two-way, not four)** — `TOKEN_EXPIRED` → refresh+retry (transparent, browser AND SSR); `UNAUTHENTICATED` → relogin (`forceRelogin`). Refresh itself is the final arbiter: a misjudged reason still self-heals (refresh validates session server-side). Rejected kind=error (transient network) forcing logout — `rejected`/`error` split in `RefreshResult`.
- **B6 field-level validation** — deferred (backend does not emit `BadRequest.FieldViolation` yet).
- **B7 fallback ladder (simplified, verified)** — reason hit → copy; miss → kind tier (always present). No 16-row Code table needed in practice.
- **B8 server-fn failures MUST use result objects (structural rule, discovered by experiment)** — `throw` from a server fn crosses the RPC boundary with **custom Error properties stripped** (only message survives). `reason` attached to an Error does not reach the browser. Login now returns `LoginResult {ok} | {ok:false, reason}` and the component re-throws a reason-carrying Error browser-side; `access.ts` `RefreshResult` was the first instance of the same wall. Rule: **BFF failure paths return result objects; never rely on throw-carried identity**.
- **B9 query-error timing vs `throwOnError` (mechanism clarified, mitigation verified)** — `throwOnError` governs only what happens **after the hook has the error in render**. Upstream of that: SSR-phase queries fetch during render (failure = SSR render error → ErrorPage), and client navigations await queries via the ssr-query integration (failure = route error) — both bypass the option entirely (verified both ways with backend down; `throwOnError:false` never engaged). The designed path is **loader → initialData**: the loader prefetches explicitly and may catch and fall back; with initialData present, SSR never runs the query, and subsequent client refetch failures land in `query.error` where `throwOnError:false` works. Verified end-to-end: loader catch → empty `initialData` → backend down → page alive, inline `ConnectError: [unavailable] HTTP 502`. **Resolved 2026-08-17** (render-error-first): the formal build accepts SSR loader failure as a render error by default — no loader-catch convention; loader stays `initialData` passthrough. Inline error forms remain opt-in per page.

#### C. Token recovery architecture (Slice 3 — built, validated, reverted for formal rebuild)

**This section preserves the full design knowledge from the prototype experiments. All code was reverted 2026-08-16; the formal build reconstructs from this record.**

##### C1. Token lifecycle (three-file design in `domains/auth/`)

```
constants.ts   LOGIN_PATH / HOME_PATH / REASON_* (literals only in decision modules)
core.server.ts session read / establish / clear (server-only, httpOnly)
access.ts      server fns: hasSessionFn / getAccessTokenFn / refreshAccessFn / clearSessionFn
               + authTransport (bare, no interceptors)
token.ts       browser memory cache: setAccessToken / clearAccessToken / getAccessToken
recovery.ts    recoverUnauthorized (唯一 401 决策) + forceRelogin + safeRedirectPath
```

**Token source is environment-adaptive** — `getAccessToken()` is the ONLY call-site for interceptors:
- browser: memory cache, miss → `getAccessTokenFn()` RPC backfill (concurrent misses merged via a single `hydrating` promise — verified: page refresh fires 3 queries, 1 backfill hop).
- bff (server): every call reads session directly — **never cache in a module variable** (multi-user shared process would leak A's token to B).
- Session (httpOnly) is the single truth; memory is a hot-path cache only (ADR D3).
- Token expiry needs no active invalidation — stale attach → 401 → recovery chain corrects.

##### C2. 401 recovery decision (`recoverUnauthorized`)

```
recoverUnauthorized(reason) → Promise<string | null>
  reason == REASON_UNAUTHENTICATED → forceRelogin(); return null (session dead, refresh必败)
  else (TOKEN_EXPIRED / unknown)    → refreshAccessFn()
    ok      → setAccessToken(new); return new (retry once)
    rejected → forceRelogin(); return null
    error    → return null (network transient; session preserved, do NOT logout)
```

- **Retry exactly once** — second 401 is terminal (prevents loops).
- **`rejected` vs `error` split is load-bearing**: backend 401 rejection clears session server-side; network/5xx preserves it. Without the split, backend jitter logs users out (was a real bug before the split).
- `forceRelogin()` uses `location.assign` (interceptor is outside React tree, no router instance); redirect param preserves current page+query.

##### C3. Refresh single-flight (server-side, per-refreshToken)

`refreshAccessFn` maintains `Map<refreshToken, Promise>` — the backend is **single-use rotation** (old refresh token reuse = 401, verified): concurrent 401s MUST share one rotation or the second caller fails. This dedup lives in the server fn (not browser) because:
- both protocols (connect + http) share the same token → dedup must be cross-protocol
- multiple tabs hit the same server process

##### C4. authTransport (bare, no interceptors) — three reasons

1. **Breaks circular dependency**: business transport's authInterceptor imports auth's server fns; auth calling the backend through business transport would cycle.
2. **Prevents self-triggering recovery**: login password-error 401 (`UNAUTHENTICATED`) would trigger `forceRelogin` from inside the login page — wrong.
3. **The http client skips auth for login URL** (`base.url.includes("/api/v1/system/auth/login")`) for the same reason; refresh goes through `authTransport`, never the http client.

##### C5. B9 loader fallback — three forms tested

| Form | Loader behavior | Page state | Error display | Best for |
|---|---|---|---|---|
| A (empty) | catch → empty initialData | alive, empty list | none until refetch | silent degradation |
| **B (result)** | catch → `{ok:false}` + empty initialData | alive | explicit error state + retry button | **preferred: user knows what happened** |
| C (throw) | no catch | full-page ErrorPage | route errorComponent | hard failure, page unusable |

Form B verified end-to-end: loader failed → page alive → inline "加载失败" + retry → backend revived → retry succeeded → data rendered.

##### C6. Logout transparent recovery (verified)

Access token expired (redis key deleted) → click logout → `Logout 401` → recovery chain: `Refresh 200` → `Logout 200`. The best-effort catch in the logout handler was never reached — the authInterceptor transparently healed the expiry. Sequence from backend logs:
```
Logout 401 → Refresh 200 → Logout 200
```

##### C7. Topology enforcement (dep-scan, built and validated then reverted)

`tools/dep-scan.mjs` (Nx target `scan:deps`): module graph from import statements, L1–L6 check + cycle detection (Kahn). Validated:
- baseline scan matched manual audit exactly (2 violations = the 2 known backlog items)
- L3 fix (feedback-holder → `__root.tsx` FeedbackWiring) → 0 violations, feedback still works
- L4 fix was **reclassification not code change**: `route-transition` moved from UI domains to mechanism layer (theme suppressing route VT is mechanism cooperation, not UI composition)

##### C8. Verified experiment matrix (all 2026-08-16)

| Domain | Scenario | Result |
|---|---|---|
| auth injection | connect/http × browser/bff | all 4 quadrants carry Bearer ✓ |
| auth recovery | SSR access expired | 401 → Refresh 200 → retry 200, 17ms, transparent ✓ |
| auth recovery | SPA session dead | query 401 → refresh rejected → forceRelogin → login?redirect=full ✓ |
| auth recovery | logout with expired access | Logout 401 → Refresh → Logout 200 (transparent heal) ✓ |
| error copy | wrong password, http, en | "Invalid credentials or not signed in" ✓ |
| error copy | wrong password, http, zh | "凭证无效或未登录" ✓ |
| error copy | wrong password, connect, zh | same (findDetails fix required) ✓ |
| error copy | backend down, connect mutation | "网络异常，请检查连接后重试" ✓ |
| error copy | backend down, http mutation | same (response.status fallback required) ✓ |
| error copy | unknown reason | kind tier fallback (never raw reason) ✓ |
| retry | network kind, backend down | 3 requests [1ms, 1015ms, 3034ms] = 1+2 retries ✓ |
| coverage | reconciliation test | 3/3 pass; missing copy listed by name ✓ |
| topology | dep-scan after fixes | 0 violations, 0 cycles ✓ |

##### C9. Environment lessons (for future verification work)

- **React Compiler**: programmatic `fill`/`dispatchEvent` does NOT reach controlled state — use native tool events or change defaults in code.
- **`duration:0` toasts**: never auto-close; clear all `[role="alert"]` before asserting new toasts.
- **performance buffer**: 250 entries max — call `performance.clearResourceTimings()` before counting requests.
- **console-pipe echo storm**: with backend down, TanStack Devtools' console-pipe creates recursive log amplification (vite log → browser console → pipe → vite log...) — vite log counts are unreliable; use browser-side `performance` entries.
- **nx background tasks**: 3-minute default timeout kills dev servers; run `npx vite dev` directly or `go build && ./binary` for long sessions.
- **`nx run system:dev` forks**: the nx wrapper exits quickly but the child `app` process survives — check `lsof` not just the task status.
- **Go backend binary**: must run from `cmd/app/` directory (relative `../../configs` path).

## Request flow

```
AUTH (BFF):                         DATA (browser-direct, dual-protocol):
browser                             browser
  │ login / logout / refresh          │ list / get / create / update / stream
  ▼                                   ▼ (Bearer <short access> + session cookie)
Start server fn ──► backend          traefik ──► backend
(httpOnly refresh)                   (future D6: traefik injects Bearer,
                                     browser holds only session cookie)
```

## Consequences

- **+** Token exposure bounded (short access TTL; refresh never in browser).
- **+** Streaming native; dual-protocol preserved at both layers.
- **+** Aligns with the official TanStack server-fn / client-Query split.
- **+** No CORS (browser same-origin via traefik; BFF is server-to-server).
- **−** Two layers to maintain; each new backend capability may need a path decision.
- **−** Short-lived access token briefly in browser memory until D6 (mitigated by short TTL).
- **−** Refresh is a server-fn round-trip on 401.

## Validation (auth vertical spike)

1. `login` server fn → backend `Login` → refresh in httpOnly session → short access to browser.
2. `getCurrentUser` server fn → session → user (for `beforeLoad` guard + auth context).
3. One **direct** protected RPC via connect-query, Bearer attached by transport interceptor.
4. Force expiry → 401 → refresh server fn → new access → retry succeeds.
5. Same RPC over **HTTP** secondary client (dual-protocol check).

Green = paradigm validated → promote to `docs/client/CONVENTIONS.md`, delete this ADR.

## Open questions

- ~~**SSR query error strategy (B9)**~~ — **resolved 2026-08-17**: render-error-first (see D10-A); SSR loader failures are accepted render errors by default, loader stays initialData passthrough.
- **Kind semantics granularity (B4)**: `UNAUTHENTICATED` currently maps to kind `auth` — conflates login-form rejects with session death. Both toast correctly today; revisit if feedback needs to differ.
- **Toast duration**: `message.error(msg, 0)` (never auto-close) stacks and polluted even our own test observations — pick a sane default in the formal build.
- **Remote i18n (hot-update)**: paraglide is compile-time; a remote value source would slot behind `errorMessage` — keep the seam, design later.
- **Streaming use cases**: concrete admin need for non-unary? If none short-term, D5 stays policy-only.
- **Edge-injection timing (D6)**: when does access-in-browser risk justify ForwardAuth?
- **Observability sink**: collect taps exist; real sink (OTel/log/error-reporting) TBD. Severity gating (≥500 only, per cal.com) when it lands.
- **Component ErrorBoundary**: deferred; route-level covers query throwOnError for now.
- **Feedback library**: errors feedback decoupled (injected antd `message`); if antd is ever dropped, **Sonner** is the 2026 community-favored replacement.

## Alternatives considered

- **Pure BFF** — rejected: collapses Connect/HTTP choice, kills browser dual-track + connect-query value, forces two-layer streaming.
- **Pure browser-direct** — rejected: weak security, no SSR prefetch, backend must CORS, refresh fully client-side.
- **Edge-injection-first** — deferred (D6): correct end state, too heavy for validation.
- **Hand-written error dictionary** (D10.B) — rejected: drifts from backend enums.
- **Post-gen `Reason` union codegen** — tried and rejected (2026-08-16): a hand-written protobuf wire parser dropped 3/4 enum families silently; the completeness guarantee it bought (compile-time Record keys) is equally delivered by the CI reconciliation test at far lower risk. Contract layer for reasons = generated Schemas at runtime + paraglide keys + the coverage test.
- **Runtime-only REASON table without reconciliation** — rejected: a new backend reason would silently fall to the kind tier; the coverage test is what makes the runtime contract safe.

## References

- ADR-0001 — backend opaque-token auth, token lifecycle, data-boundary stance.
- Official TanStack Start — Server Functions, Authentication.
- `archive1` tag admin — browser dual-track reference.
- Traefik unified-edge / ForwardAuth decision (memory: `deploy-traefik-unified-edge`).
- Error model: [kratos Errors](https://go-kratos.dev/docs/component/errors/), [Connect Errors](https://connectrpc.com/docs/web/errors/), [Google AIP-193](https://google.aip.dev/193), [google/rpc/error_details.proto](https://github.com/googleapis/googleapis/blob/master/google/rpc/error_details.proto).
- React Query error handling: [TkDodo](https://tkdodo.eu/blog/react-query-error-handling).
- i18n: paraglide (TS), lingui (RN); remote hot-update TBD.
