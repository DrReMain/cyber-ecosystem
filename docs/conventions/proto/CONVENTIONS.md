# Proto Conventions

**Status:** ACTIVE

**Scope:** Protobuf contracts under `proto/`. These are the **source of truth**; `gen/` is derived (never hand-edit). Prescriptive — `MUST` / `SHOULD` / `MAY`.

For Go-side consumption (mapping, error handling, import aliases) see `docs/conventions/kratos/CONVENTIONS.md`.

---

## 1. Layout

- **Per-service domain:** `proto/cyber/<service>/v1/*.proto` — a service's own contracts (entities, service, service-specific errors).
- **Shared kernel:** `proto/cyber/shared/<area>/v1/*.proto` — cross-service types: `shared/common/v1` (pagination + shared value types), `shared/errors/v1` (generic error enums).
- **Custom annotations:** `proto/ext/v1/*.proto` — repo-specific proto options (`method` desc, `access`, `builtin`, `datascope`).
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
- **Response/entity fields → WKT wrapper** (`google.protobuf.StringValue` / `BoolValue` / `Int32Value` … → Go `*wrapperspb.StringValue` etc.). Entities (`User`, `Dept`, …) and their response messages flow *out* as wrappers.

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

> **Client note.** On the client, empty is represented per transport: Connect (protobuf-es) → `undefined`; HTTP (hey-api SDK) → `null` (wire value preserved). Known deviation: gnostic renders WKT wrappers as plain types without `nullable`, so hey-api types say `id?: string` while the wire may carry an explicit `null` — consumer code must not non-null-assert these fields. gnostic is upstream low-maintenance; do not expect this to be fixed there.
>
> **Carve-out:** a response scalar that exists whenever the call succeeds (e.g. `token` in `LoginResponse`) MAY be plain — the "success implies present" semantics justify it; note the justification in a comment.

**Batch-read responses key by id** — `map<string, Entry>`, not a repeated list: keyed entries match client-side without scanning, and an absent key carries skip semantics (row deleted / not ready since the listing) with no per-item error channel.

---

## 4. Errors

Two layers of error enums, both using the kratos `errors.proto` option (`errors.code` = HTTP status):

- **Shared generic** (`cyber.shared.errors.v1`): `GeneralError` (enum values mirror HTTP status: 4xx client / 5xx server), `InfraError` (3xxx db/cache/storage/mq/network), `FlowError`. Reused by every service — do not redefine generic codes per service.
- **Service business** (`cyber.<service>/v1/error_reason.proto`): a service-semantic enum — `System` for system, e.g. `SYSTEM_INVALID_STATUS_TRANSITION` (6xxx range). **The enum name carries the service identity**: buf's `ENUM_VALUE_PREFIX` then forces every value to start with the service prefix, so the wire reason is unique across services by construction. Do not name this enum `ErrorReason` (generic prefix collides the moment a second service copies the template).

**Wire reason shape & i18n.** The wire reason is the bare proto value name (e.g. `SYSTEM_INVALID_STATUS_TRANSITION`) — protobuf APIv2's `String()` emits no Go type prefix. `ENUM_VALUE_PREFIX` makes the value name unique across services, so it doubles as the client-side error identity: i18n catalogs key error copy as `error_<ValueName>` and the frontend looks it up directly, no transformation.

**Removing an error value reserves its number** (`reserved 6050;`, mirroring `reserved 6031;`): a deleted number left unreserved lets a future value silently reuse it, deserializing old wire data into the wrong semantics. The reserved line replaces the entry in place; the three i18n copies are deleted in the same change.

Go side consumes generated `errorspb.ErrorXxx("")` (see kratos CONVENTIONS §5/§6).

The `(errors.code)` option comes from the vendored `proto/errors/errors.proto` (§1); service enums `import "errors/errors.proto"` to annotate each value. That proto is **excluded from Go codegen** by the generation config — never generate or hand-write Go for it; kratos's bundled `errors.pb.go` provides the option at runtime.

---

## 5. HTTP & annotations

- **REST mapping:** every RPC has `option (google.api.http)` — `/api/v1/<service>/<resources>[/{id}]`, verb by CRUD (POST create / GET list-or-get / PUT update / DELETE delete).
- **Repeated input ⇒ POST + body, not GET query:** the transport's GET-query binding does not reliably carry `repeated` fields (they arrive empty and fail `min_items`). Batch reads marshal ids in a POST body — minting-style batch reads are action-shaped anyway.
- **Method desc (facts):** `option (cyber.ext.v1.method) = {comment, action}` — `comment` is the human-readable one-liner (Chinese OK), `action` (`ACTION_READ` / `ACTION_WRITE`) is the effect semantics the audit layer classifies on (allow side collects WRITE only). Facts only — policy knobs (audit toggles etc.) never live here. Mandatory on every business RPC, machine-enforced by `TestEveryBusinessRPCDeclaresAction` (`shared-go/kratos/security`, no exemptions). Consumed at runtime via proto reflection (`shared-go/helper/proto_reflect.go` reads the comment face; the Resource catalog transcribes it), not by codegen/docs. Custom option value types are compile-time contracts: reshaping one is a coordinated one-shot across all declarations and consumers, never a runtime-compat concern. See `proto/ext/v1/desc.proto`.
- **Access policy:** `option (cyber.ext.v1.access)` — `ACCESS_PUBLIC` / `ACCESS_ADMIN` / `ACCESS_APP`. **MUST be declared on every business RPC — there is no default level.** An unannotated RPC resolves to `ACCESS_UNSPECIFIED`, which servers wire to the deny-by-default guard: it fails loud at runtime (`MISSING_ANNOTATION`, 503) rather than silently falling back to ADMIN. Only framework built-ins outside `cyber.*` bypass the guard. Machine-enforced by `TestEveryBusinessRPCDeclaresAccess` (`shared-go/kratos/security`), whose exemption list holds one deliberate entry: `cyber.system.v1.ResourceService.ListResource` exercises the 503 path on purpose. See `proto/ext/v1/access.proto`.

---

## 6. Shared types

- **Pagination:** `cyber.shared.common.v1.PageRequest` / `PageResponse` (page_no / page_size / all / created_at_a-z / updated_at_a-z range filters). Reuse for every list RPC — do not redefine pagination per service.
- Any type needed by ≥2 services goes to `cyber/shared/...`; a type used by one service stays in its own `cyber/<service>/v1/`.

---

## 7. go_package & generation

- `option go_package = "cyber-ecosystem/gen/go/cyber/<scope>/v1"` — matches the gen tree. (The vendored `errors/errors.proto` keeps its own `github.com/go-kratos/kratos/v3/errors;errors`.)
- `cyber/shared/errors/v1/error_detail.proto` is an **anchor file** (no messages, by design): its `google/rpc/error_details.proto` import is what pulls `error_details_pb` into `gen/connect-ts` via `--include-imports`. Excluded from Go codegen (`--exclude-path` in `proto/project.json`); its `go_package` exists only to satisfy the package-wide buf lint rule. Do not delete.
- `gen/` is **derived**: change proto → regenerate via the owning Nx target (e.g. `./nx run proto:generate`), never hand-edit gen.
- Proto is the single source of truth; Go struct fields / JSON tags follow generated code.
- **Pipeline tools are Go, under `proto/cmd/`** — the pipeline's own tooling lives with the pipeline, in the repo's single Go module: `cmd/catalog` (derives the per-package operation manifests below) and `cmd/clean-gen` (pre-regen wipe; hand-maintained root files and `node_modules` survive; all filesystem access goes through `os.Root`, sandbox-confined by construction). The Node scripts era is retired; pnpm/openapi-ts invocations stay in target commands — package-manager calls are orchestration, not tool logic.
- **Operation manifests (`gen/catalog/<package>.json`) are derived output** — `proto:generate:catalog` feeds `buf build`'s image to `cmd/catalog`, which emits one manifest per service-bearing proto package (services, methods, comment/access/builtin/datascope/http faces). The system service live-loads them as its federated grant catalog: a new service enters the grant tree when its manifest ships — no system code change, no restart. The JSON tags are a contract between `cmd/catalog` and system's resource module; both sides carry the sync note.

---

## 8. Checklist — changing contracts

1. Edit the source `.proto` under the right scope (`<service>/v1/` if service-owned, `shared/...` if reused). Never edit `gen/`.
2. New RPC: `<Verb><Entity>` + `Request`/`Response` messages + `google.api.http` + `method` desc (comment + action) + `access` (both required on every business RPC — see §5). Follow the verb order.
3. New error reason: generic → `cyber.shared.errors.v1`; service-specific → `cyber.<service>/v1/error_reason.proto` (6xxx). Both with `errors.code`.
4. Regenerate everything from the one source: `./nx run proto:generate` (gates on lint + format:check, then Go + connect-ts + OpenAPI/openapi-ts). Review the `gen/` diff — it must contain only what the source change implies; exclude unintended churn. `./nx run proto:check-drift` proves sync when in doubt.
5. A new error reason also means client copy: add `error_<ValueName>` to the client error vocabulary (`messages/error/`) in **every locale registered in the client's paraglide settings** (`project.inlang/settings.json` — the registered set is per-fork and that file is the truth) — a reason without copy renders raw on the client.
6. `./nx run proto:breaking` before merging (FILE rules vs `HEAD`). Blind spot: it cannot see protovalidate rule removals/weakenings — review validation rules explicitly (§9).
7. Build the affected service (`./nx run <service>:build`); if the client surface changed, run the client verification chain (`docs/conventions/tanstack/CONVENTIONS.md §11`).

Done = every command's exit code checked; breaking green (or weakenings consciously reviewed); gen diff clean; error copy present in every registered locale.

---

## 9. Versioning & breaking changes

v1 is the only version so far. `./nx run proto:breaking` diffs against **`HEAD`** (`--against .git#ref=HEAD,subdir=proto`) with FILE rules, the strongest tier — the gate answers "is this working tree a pure increment over the last commit's contracts?", which keeps it green-able on every change instead of drifting red against a stale remote branch. A deliberate breaking change (rename, migration) goes through the same gate consciously: verify every finding is the intended change, commit, and the gate re-arms. Cross-release checks against published baselines are a separate, deliberate invocation. Known blind spot: `buf breaking` does not diff custom option values, so **protovalidate rule removals/weakenings are invisible to it** — treat validation rules as behavior contract and review them explicitly (or cover them with per-rule rejection tests).

Conventions for v2 / deprecation policy are not yet defined — add before a second version lands.
