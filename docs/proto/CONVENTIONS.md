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
| rpc | `<Verb><Entity>`; verb order Create/Update/Delete/List → Get → GetByXxx/Other | `Create<Entity>`, `List<Entities>`, `Get<Entity>` |
| message | `<Entity>`, `<Verb><Entity>Request`/`<Verb><Entity>Response` | `<Entity>`, `Create<Entity>Request` |
| field | snake_case | `created_at`, `page_size` |
| enum | `UPPER_SNAKE`; first value `<ENUM>_UNSPECIFIED = 0` | `<ENUM>_UNSPECIFIED` |
| Go import alias | `<scope>pb` (see kratos CONVENTIONS) | `<service>pb`, `commonpb`, `errorspb` |

---

## 3. Field types — optionality & the null/zero contract

Field shape is chosen by **direction**:

- **Request/input fields → proto3 `optional`** (`optional <type>` → Go `*T`). Presence-based input: the caller sends only what it sets. Express "required" via `(buf.validate.field).required` (a validation rule), **not** by dropping `optional`. Add `string.min_len`/`max_len`/`pattern` and `cel` (cross-field) as needed.
- **Response/entity fields → WKT wrapper** (`google.protobuf.StringValue` / `BoolValue` / `Int32Value` … → Go `*wrapperspb.StringValue` etc.). Entities (`User`, `Dept`, `Item`, …) and their response messages flow *out* as wrappers.

**Why the split — the null/zero contract.** In proto3 JSON, an unset `optional` scalar renders as the zero value (`""`/`0`/`false`), indistinguishable from a value that is genuinely zero. An unset wrapper renders as `null`. So entity/response fields use wrappers, making every field present in the payload and the two states unambiguous:

| payload | meaning |
|---|---|
| `null` | no value (empty / not set) |
| `0`, `false`, `""` | has a value, and it is the zero value |

Requests don't need this (input is presence-based), so they use the lighter `optional`.

**Backend enabler — `EmitUnpopulated`.** `shared-go/kratos/jsoncodec` registers a protojson codec with `EmitUnpopulated: true` for **both** HTTP (`jsoncodec.Register()`) and Connect (`connect.WithCodec`). This is what makes the contract hold on the wire: unset wrappers emit `null`, zero-value scalars emit their value. Kratos's stock codec would otherwise honor struct-tag `omitempty` and drop zero-value scalars (`count=0`, `status=0`, …).

- **Time:** `google.protobuf.Timestamp`.
- **Path-bound fields** (bound from `{xxx}` in `google.api.http`): MUST be plain (non-`optional`, non-wrapper). Always present from the URL; `optional`/wrapper generates a proto oneof that conflicts with `body:"*"` decode ("field already set for oneof").
- **Validation:** `(buf.validate.field)` rules inline, correct and reasonable — `required` for mandatory, `string.min_len`/`max_len`/`pattern` for format, `cel` for cross-field. Validate at the boundary (server inbound); do not re-validate in biz.

> **Client note.** On the client, empty is represented per transport: Connect (protobuf-es) → `undefined`; HTTP (hey-api SDK) → `null` (wire value preserved). Both type sources are proto-derived. The two-layer value rule is set in ADR-0002 D7.

---

## 4. Errors

Two layers of error enums, both using the kratos `errors.proto` option (`errors.code` = HTTP status):

- **Shared generic** (`cyber.shared.errors.v1`): `GeneralError` (enum values mirror HTTP status: 4xx client / 5xx server), `InfraError` (3xxx db/cache/storage/mq/network), `FlowError`. Reused by every service — do not redefine generic codes per service.
- **Service business** (`cyber.<service>/v1/error_reason.proto`): `ErrorReason` — domain-specific failures (6xxx range), e.g. invalid state transition. Service-owned.

Go side consumes generated `errorspb.ErrorXxx("")` (see kratos CONVENTIONS §5/§6).

The `(errors.code)` option comes from the vendored `proto/errors/errors.proto` (§1); service enums `import "errors/errors.proto"` to annotate each value. That proto is **excluded from Go codegen** by the generation config — never generate or hand-write Go for it; kratos's bundled `errors.pb.go` provides the option at runtime.

---

## 5. HTTP & annotations

- **REST mapping:** every RPC has `option (google.api.http)` — `/api/v1/<service>/<resources>[/{id}]`, verb by CRUD (POST create / GET list-or-get / PUT update / DELETE delete).
- **Method description:** `option (cyber.ext.v1.method_comment)` — human-readable one-liner (Chinese OK); used by codegen/docs.
- **Access policy:** `option (cyber.ext.v1.access)` — `ACCESS_PUBLIC` / `ACCESS_ADMIN` / `ACCESS_APP`. **MUST be declared on every business RPC — there is no default level.** An unannotated RPC resolves to `ACCESS_UNSPECIFIED`, which servers wire to the deny-by-default guard: it fails loud at runtime (`MISSING_ANNOTATION`, 503) rather than silently falling back to ADMIN. Only framework built-ins outside `cyber.*` bypass the guard. See `proto/ext/v1/access.proto`.

---

## 6. Shared types

- **Pagination:** `cyber.shared.common.v1.PageRequest` / `PageResponse` (page_no / page_size / all / created_at_a-z / updated_at_a-z range filters). Reuse for every list RPC — do not redefine pagination per service.
- Any type needed by ≥2 services goes to `cyber/shared/...`; a type used by one service stays in its own `cyber/<service>/v1/`.

---

## 7. go_package & generation

- `option go_package = "cyber-ecosystem/gen/go/cyber/<scope>/v1"` — matches the gen tree. (The vendored `errors/errors.proto` keeps its own `github.com/go-kratos/kratos/v3/errors;errors`.)
- `cyber/shared/errors/v1/error_detail.proto` is an **anchor file** (no messages, by design): its `google/rpc/error_details.proto` import is what pulls `error_details_pb` into `gen/ts` via `--include-imports`. Excluded from Go codegen (`--exclude-path` in `proto/project.json`); its `go_package` exists only to satisfy the package-wide buf lint rule. Do not delete.
- `gen/` is **derived**: change proto → regenerate via the owning Nx target (e.g. `./nx run proto:generate`), never hand-edit gen.
- Proto is the single source of truth; Go struct fields / JSON tags follow generated code.

---

## 8. Checklist — adding to a proto

1. Edit the source `.proto` under the right scope (`<service>/v1/` if service-owned, `shared/...` if reused).
2. New RPC: `<Verb><Entity>` + `Request`/`Response` messages + `google.api.http` + `method_comment` + `access` (required on every business RPC — see §5). Follow the verb order.
3. New error: generic → `cyber.shared.errors.v1`; service-specific → `cyber.<service>/v1/error_reason.proto` (6xxx). Both with `errors.code`.
4. Regenerate (`./nx run proto:generate`) and build affected service. Verify no hand-edits in `gen/`.

---

## 9. Versioning & breaking changes (not yet defined)

v1 is the only version so far. Conventions for v2 / breaking changes (field removal, type changes, deprecation policy, wire-compat) are **not yet defined** — add before a second version or a breaking change lands.
