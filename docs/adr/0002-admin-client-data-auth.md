# ADR-0002: Admin Client — Data Path & Auth Architecture

- **Status:** Direction converged; validation via the **auth vertical spike** pending. This ADR is the spike's design basis and is **temporary** — it folds into `docs/client/CONVENTIONS.md` once the paradigm is validated in code, then is removed (same lifecycle as ADR-0001).
- **Date:** 2026-08-07

## Context

`app/clients/admin` is a TanStack Start (SSR) client to the `system` backend. The backend (ADR-0001) exposes three transports (gRPC / HTTP / Connect); **HTTP and Connect are exposed externally via traefik in both compose and k3s/k8s**, and kratos v3 serves both unary and non-unary (streaming) on HTTP and our Connect transport. Backend auth is **opaque-token** (access + refresh in Redis, `Bearer`).

Two hard constraints shape this ADR:

1. **The client must be dual-protocol** — Connect primary, HTTP secondary. HTTP stays first-class for special cases and for a future **open platform** that exposes capabilities externally (HTTP is the universal choice there). The `archive1` tag's admin already ran a browser-side dual-track (Connect + OpenAPI/hey-api + ky); this ADR keeps that property.
2. **Follow official TanStack Start guidance where viable** — server functions for server-concerns (secrets, auth, SSR), TanStack Query on the client for interactive data.

The official TanStack Start model is BFF-via-server-functions (browser → server fn → backend; session in httpOnly cookie; token never in browser). Naively applied, this **conflicts** with constraint 1: under pure BFF the browser sees only server functions and the Connect/HTTP choice collapses into a server-fn implementation detail — killing the browser dual-track and connect-query's client-side value, and forcing streaming through two layers.

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

Rationale: this *is* the official TanStack split ("static/auth → loader + serverFn; interactive/cached/streaming → client TanStack Query"). Pure BFF loses browser dual-protocol + connect-query value + forces two-layer streaming; pure browser-direct loses SSR, weakens security, and needs CORS.

### D2 — Auth boundary is the server function, not `beforeLoad`

`beforeLoad` is **route UX only** (keep users out of screens they can't use; avoid work that would 401 anyway). **Every server function that reads/writes private data authorizes itself** (session lookup → reject if absent/invalid). Mirrors the official rule and ADR-0001's "data boundary" stance.

### D3 — Token split: refresh server-side, access short-lived to browser

- **refresh token** → httpOnly session cookie, **server-side only**. Browser JS can never read it.
- **access token** → short TTL, issued to the browser by the login/refresh server fn. Browser holds it **in memory**; the dual-protocol transport interceptor attaches `Authorization: Bearer <access>` for direct calls.
- **401 on a direct call** → call the refresh server fn (server uses the httpOnly refresh to rotate) → new access → retry once. The reactive 401-interceptor contract is retained, but the refresh hop is a server fn, not a browser-direct call to the backend.
- **SSR prefetch** (BFF path): the server fn obtains an access token server-side (from the httpOnly refresh) and calls the backend; the token never leaves the server.

Rationale: bounds exposure (short access TTL), keeps the refresh off-browser (XSS cannot steal it), and avoids edge-injection complexity during validation. This is the hinge that makes the hybrid work — it dissolves the "BFF session vs direct Bearer" tension.

### D4 — Dual-protocol at both layers

- **Browser-direct layer**: `connect-query` (Connect, generated from proto) **primary** + `hey-api`/`ky` (HTTP, generated from OpenAPI) **secondary**. archive1's dual-track is the reference.
- **BFF layer**: server-side Connect client + HTTP `fetch`.

Rationale: backend exposes HTTP+Connect via traefik (both unary + non-unary); the client must keep both viable (Connect primary, HTTP for special cases + future open platform).

### D5 — Streaming (non-unary) is always browser-direct

Connect stream runs native browser → traefik → backend. **Never proxied through a server function** (that would be a two-layer stream with extra latency and cancellation/backpressure complexity). kratos v3 serves non-unary on both HTTP and Connect, so the direct path covers both protocols.

### D6 — Edge injection (traefik ForwardAuth) is optional, deferred hardening

Production hardening, **not** required for validation: traefik `ForwardAuth` calls Start to validate the session and **injects** the Bearer token to the upstream backend call. With this, even the access token stays server-side and the browser holds only the session cookie. Deferred because (a) it doesn't change the client paradigm — only *where the token lives* — and (b) ForwardAuth + header-injection config is heavy for the validation phase. It composes with the existing "traefik unified edge / ForwardAuth central auth" decision.

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
- **+** No CORS (browser goes same-origin via traefik; BFF path is server-to-server).
- **+** Composes with the traefik ForwardAuth decision (D6 hardens it later).
- **−** Two layers to maintain (BFF server fns + browser-direct clients); each new backend capability may need a decision on which path.
- **−** Short-lived access token still briefly in browser memory until D6 lands (mitigated by short TTL; D6 eliminates it).
- **−** Refresh path is a server-fn round-trip on 401 (one extra hop vs browser-direct refresh).

## Validation (auth vertical spike)

A minimal end-to-end slice validates the paradigm before codification:

1. `login` server fn → calls backend `Login` → stores refresh in httpOnly session → issues short-lived access to browser.
2. `getCurrentUser` server fn → reads session → returns user (for `beforeLoad` guard + auth context).
3. One **direct** protected RPC (e.g. `ListUsers`) via connect-query, Bearer attached by transport interceptor.
4. Force token expiry → 401 → refresh server fn → new access → retry succeeds.
5. Confirm the same RPC also works over the **HTTP** secondary client (dual-protocol check).

Green = paradigm validated → promote to `docs/client/CONVENTIONS.md`, delete this ADR.

## Open questions

- **Streaming use cases**: concrete admin business need for non-unary (real-time / progress)? If none short-term, D5 stays as policy but unexercised.
- **Edge-injection timing (D6)**: when does the access-token-in-browser risk justify landing ForwardAuth + header injection?
- **Aggregation BFF shape**: which screens actually need multi-call aggregation/shaping vs direct connect-query?
- **SSR prefetch granularity**: loader prefetch via BFF for which routes (vs client-only fetch)?
- **Composition with L0/L1/L2**: in L0 (single process, `nx run dev`) Start and backend are separate processes; in L1/L2 the traefik topology + session-cookie domain + ForwardAuth need pinning.

## Alternatives considered

- **Pure BFF (all data via server fn)** — rejected: collapses Connect/HTTP choice into a server-fn detail, kills the browser dual-track and connect-query client value, and forces two-layer streaming.
- **Pure browser-direct (token in browser, no BFF)** — rejected: weak security (long-lived token in browser), no SSR data prefetch, backend must open CORS to browsers, refresh logic fully client-side.
- **Edge-injection-first (browser holds only session cookie from day one)** — deferred (D6): correct end state but too heavy for the validation phase; paradigm-identical, so no reason to block on it.

## References

- ADR-0001 — backend opaque-token auth (D5), token lifecycle (D13), data-boundary stance.
- Official TanStack Start — Server Functions, Authentication, "Calling an external API" tutorial.
- `archive1` tag admin — browser dual-track (Connect + OpenAPI/hey-api + ky).
- Traefik unified-edge / ForwardAuth decision (memory: `deploy-traefik-unified-edge`).
