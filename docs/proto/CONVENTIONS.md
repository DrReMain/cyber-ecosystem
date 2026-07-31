# Proto Conventions

**Scope:** Protobuf contracts under `proto/`. These are the **source of truth**; `gen/` is derived (never hand-edit). Prescriptive — `MUST` / `SHOULD` / `MAY`.

For Go-side consumption (mapping, error handling, import aliases) see `docs/kratos/CONVENTIONS.md`.

---

## 1. Layout

- **Per-service domain:** `proto/cyber/<service>/v1/*.proto` — a service's own contracts (entities, service, service-specific errors).
- **Shared kernel:** `proto/cyber/shared/<area>/v1/*.proto` — cross-service types: `shared/common/v1` (pagination + shared value types), `shared/errors/v1` (generic error enums).
- **Custom annotations:** `proto/ext/v1/*.proto` — repo-specific proto options (`method_comment`, `access`).
- **Kratos errors option:** `proto/errors/errors.proto` (vendored) — provides `errors.code` / `errors.default_code` for error enums.

A service MUST own its proto under `cyber/<service>/v1/`; shared types go under `cyber/shared/...`, never duplicated per service.

---

## 2. Naming

| Concept | Rule | Example |
|---|---|---|
| package | `cyber.<scope>.v1` | `cyber.<service>.v1`, `cyber.shared.common.v1` |
| service | `<Entity>Service` | `<Entity>Service` |
| rpc | `<Verb><Entity>`; verb order Create/Update/Delete/Read/Other | `Create<Entity>`, `List<Entities>`, `Get<Entity>` |
| message | `<Entity>`, `<Verb><Entity>Request`/`<Verb><Entity>Response` | `<Entity>`, `Create<Entity>Request` |
| field | snake_case | `created_at`, `page_size` |
| enum | `UPPER_SNAKE`; first value `<ENUM>_UNSPECIFIED = 0` | `<ENUM>_UNSPECIFIED` |
| Go import alias | `<scope>pb` (see kratos CONVENTIONS) | `<service>pb`, `commonpb`, `errorspb` |

---

## 3. Field types

- **Nullable scalar/string:** `optional <type>` (proto3 optional) for value semantics, or a **WKT wrapper** (`google.protobuf.StringValue` / `BoolValue` / `Int32Value` …) when the field must round-trip as a nullable pointer across JSON/codegen. Pick one per field and stay consistent.
- **Time:** `google.protobuf.Timestamp`.
- **Non-null scalar:** plain `string` / `int32` / …
- **Validation:** `buf.validate.field` rules inline (`required`, `string.min_len`, `int32.gt`/`lte`, …). Validate at the boundary (server inbound); do not re-validate in biz.

---

## 4. Errors

Two layers of error enums, both using the kratos `errors.proto` option (`errors.code` = HTTP status):

- **Shared generic** (`cyber.shared.errors.v1`): `GeneralError` (1xxx client / 2xxx server), `InfraError` (3xxx db/cache/storage/mq/network), `FlowError`. Reused by every service — do not redefine generic codes per service.
- **Service business** (`cyber.<service>/v1/error_reason.proto`): `ErrorReason` — domain-specific failures (6xxx range), e.g. invalid state transition. Service-owned.

Go side consumes generated `errorspb.ErrorXxx("")` (see kratos CONVENTIONS §5/§6).

---

## 5. HTTP & annotations

- **REST mapping:** every RPC has `option (google.api.http)` — `/api/v1/<service>/<resources>[/{id}]`, verb by CRUD (POST create / GET list-or-get / PUT update / DELETE delete).
- **Method description:** `option (cyber.ext.v1.method_comment)` — human-readable one-liner (Chinese OK); used by codegen/docs.
- **Access policy:** `option (cyber.ext.v1.access)` — `ACCESS_PUBLIC` / `ACCESS_ADMIN` (default). Drives the auth selector (deny-by-default; public endpoints opt out). See `proto/ext/v1/access.proto`.

---

## 6. Shared types

- **Pagination:** `cyber.shared.common.v1.PageRequest` / `PageResponse` (page_no / page_size / all / created_at_a-z / updated_at_a-z range filters). Reuse for every list RPC — do not redefine pagination per service.
- Any type needed by ≥2 services goes to `cyber/shared/...`; a type used by one service stays in its own `cyber/<service>/v1/`.

---

## 7. go_package & generation

- `option go_package = "cyber-ecosystem/gen/go/cyber/<scope>/v1"` — matches the gen tree. (The vendored `errors/errors.proto` keeps its own `github.com/go-kratos/kratos/v3/errors;errors`.)
- `gen/` is **derived**: change proto → regenerate via the owning Nx target (e.g. `./nx run proto:generate`), never hand-edit gen.
- Proto is the single source of truth; Go struct fields / JSON tags follow generated code.

---

## 8. Checklist — adding to a proto

1. Edit the source `.proto` under the right scope (`<service>/v1/` if service-owned, `shared/...` if reused).
2. New RPC: `<Verb><Entity>` + `Request`/`Response` messages + `google.api.http` + `method_comment` (+ `access` if non-default). Follow the verb order.
3. New error: generic → `cyber.shared.errors.v1`; service-specific → `cyber.<service>/v1/error_reason.proto` (6xxx). Both with `errors.code`.
4. Regenerate (`./nx run proto:generate`) and build affected service. Verify no hand-edits in `gen/`.

---

## 9. TODO: versioning & breaking changes

v1 is the only version so far. Conventions for v2 / breaking changes (field removal, type changes, deprecation policy, wire-compat) are **not yet defined** — add before a second version or a breaking change lands.
