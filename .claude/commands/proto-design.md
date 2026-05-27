---
name: proto-design
description: Use when creating or modifying proto files, setting up buf generation, designing gRPC service definitions, or debugging proto/generation issues
---

# Proto Design & Generation

Guide for designing, organizing, and generating Protobuf/gRPC definitions in this monorepo.

---

## When to Use

- Creating new proto files (messages, services, errors)
- Setting up buf generation for a new app or service
- Designing BFF HTTP annotations
- Debugging generation failures or import issues

## Key Rules

- File name prefix required in service directories to avoid Go generation collisions
- `messages/` = pure data, no service blocks, no HTTP annotations
- `service_base/` = gRPC-only, NO HTTP annotations
- `bff_*/` = client-facing, WITH `google.api.http` annotations
- Use `StringValue` (not `optional string`) in response/model messages
- Use `optional string` in request messages with `(buf.validate.field)` constraints
- BFF HTTP paths must include BFF prefix (`/api/v1/admin/`, `/api/v1/mobile/`)
- All files in an app share one `package` and one `go_package`
- `PACKAGE_DIRECTORY_MATCH` lint ignore required due to subdirectory layout

---

## Directory Structure

```
apps/<app>/api/v1/
  messages/            # Shared messages (request/response pairs, models)
    article.proto
  service_base/        # Core gRPC services (NO HTTP annotations)
    base_article.proto
  bff_admin/           # Admin BFF services (WITH HTTP annotations)
    admin_article.proto
  bff_mobile/          # Mobile BFF services (WITH HTTP annotations)
    mobile_article.proto
  error/               # App-specific error enums
    error_reason.proto
```

- **messages/** — pure data shared by all services. No `service` blocks.
- **service_base/** — core domain, gRPC-only. BFFs call it internally.
- **bff_*/** — external-facing. Include `google.api.http` annotations. Each BFF selects the methods it needs.
- **error/** — app-specific errors. General/infra errors live in `contracts/errors/`.

---

## File Naming

All proto files in service directories MUST use a unique prefix:

```
service_base/base_article.proto    → prefix: "base_"
bff_admin/admin_article.proto      → prefix: "admin_"
bff_mobile/mobile_article.proto    → prefix: "mobile_"
messages/article.proto             → no prefix (shared)
error/error_reason.proto           → no prefix (errors)
```

The prefix applies to the file name, not the proto service name. This prevents Go generation collisions since buf generates into a single `gen/go/v1/` directory.

---

## Package & Import Conventions

```protobuf
// buf:lint:ignore PACKAGE_DIRECTORY_MATCH
package api.<app>.v1;

option go_package = "cyber-ecosystem/apps/<app>/gen/go/v1;<app>V1";
```

Imports resolve relative to `contracts/` or `apps/` (defined in root `buf.yaml`):

```protobuf
import "genesis/api/v1/messages/article.proto";  // apps/ → genesis/api/v1/...
import "common/page.proto";                       // contracts/common/page.proto
import "errors/errors.proto";                     // contracts/errors/errors.proto
import "buf/validate/validate.proto";
import "google/api/annotations.proto";
import "desc/desc.proto";
```

---

## Message Design Patterns

### Model message (in messages/)

```protobuf
message Article {
  google.protobuf.StringValue id = 1;
  google.protobuf.Timestamp created_at = 2;
  google.protobuf.Timestamp updated_at = 3;
  google.protobuf.StringValue title = 4;
  google.protobuf.StringValue content = 5;
  google.protobuf.StringValue status = 6;
}
```

Use `StringValue` for nullable fields in models — makes nil-vs-empty unambiguous.

### Request message

```protobuf
message CreateArticleRequest {
  optional string title = 1 [
    (buf.validate.field).required = true,
    (buf.validate.field).string.min_len = 1
  ];
  optional string content = 2;
}
```

Use `optional string` in requests. Required fields get `(buf.validate.field).required = true`.

### ID validation (xid 20-char)

```protobuf
optional string id = 1 [
  (buf.validate.field).required = true,
  (buf.validate.field).string.len = 20
];
```

### Pagination

```protobuf
import "common/page.proto";

message QueryArticleRequest {
  common.PageRequest page = 1;
  repeated string order_by = 100 [(buf.validate.field).cel = {
    id: "QueryArticleRequest.order_by"
    message: ""
    expression: "this.all(item, size(item) == 0 || item.matches('^(createdAt|updatedAt|sort):(asc|desc)$'))"
  }];
}

message QueryArticleResponse {
  common.PageResponse page = 1;
  repeated GetArticleResponse list = 2;
}
```

`order_by` uses field number 100 (high to avoid conflicts). CEL validates `fieldName:asc` or `fieldName:desc` format.

### Field mask (partial update)

```protobuf
message UpdateArticleRequest {
  optional string id = 1 [(buf.validate.field).required = true, (buf.validate.field).string.len = 20];
  optional string title = 2 [(buf.validate.field).required = true, (buf.validate.field).string.min_len = 1];
  optional string content = 3;
  repeated string fields_mask = 100 [(buf.validate.field).repeated = {
    min_items: 1
    unique: true
  }];
}
```

### Status transition

```protobuf
message UpdateArticleStatusRequest {
  optional string id = 1 [(buf.validate.field).required = true, (buf.validate.field).string.len = 20];
  optional string status = 2 [
    (buf.validate.field).required = true,
    (buf.validate.field).string = { in: ["draft", "published", "archived"] }
  ];
}
```

### Sort (fractional indexing)

```protobuf
message SortArticleRequest {
  optional string id = 1 [(buf.validate.field).required = true, (buf.validate.field).string.len = 20];
  optional string prev_id = 2;
  optional string next_id = 3;

  option (buf.validate.message).cel = {
    id: "SortArticleRequest.prev_id-next_id"
    message: "at least one of prev_id or next_id is required"
    expression: "has(this.prev_id) || has(this.next_id)"
  };
}
```

---

## Service Definitions

### Core service (service_base/) — gRPC only

```protobuf
service ArticleService {
  option (desc.service_comment) = "文章服务";

  rpc CreateArticle(CreateArticleRequest) returns (CreateArticleResponse) {
    option (desc.method_comment) = "创建文章";
  }
  rpc GetArticle(GetArticleRequest) returns (GetArticleResponse) {
    option (desc.method_comment) = "查询文章详情";
  }
}
```

### BFF service (bff_*/) — with HTTP

```protobuf
service AdminArticleService {
  option (desc.service_comment) = "管理端文章服务";

  rpc CreateArticle(CreateArticleRequest) returns (CreateArticleResponse) {
    option (desc.method_comment) = "创建文章";
    option (google.api.http) = { post: "/api/v1/admin/article", body: "*" };
  }
  rpc GetArticle(GetArticleRequest) returns (GetArticleResponse) {
    option (google.api.http) = { get: "/api/v1/admin/article/{id}" };
  }
}
```

**BFF path prefix convention:**

| BFF | Prefix | Example |
|-----|--------|---------|
| Admin | `/api/v1/admin/` | `/api/v1/admin/article` |
| Mobile | `/api/v1/mobile/` | `/api/v1/mobile/article` |

HTTP path conventions: POST create, PUT update, DELETE delete, GET single, GET list, POST for special actions (sort, status).

---

## Error Enums

```protobuf
syntax = "proto3";
package api.<app>.v1;

import "errors/errors.proto";

option go_package = "cyber-ecosystem/apps/<app>/gen/go/v1;<app>V1";

enum ErrorReason {
  option (errors.default_code) = 500;

  ERROR_REASON_UNSPECIFIED = 0 [(errors.code) = 500];
  ERROR_REASON_<ENTITY>_<DESCRIPTION> = <code> [(errors.code) = <http_status>];
}
```

- First value must be `_UNSPECIFIED = 0`
- Use codes starting from 6000 to avoid collision with shared error codes
- General/infra errors are in `contracts/errors/` — do NOT duplicate

---

## buf Toolchain

### Module configuration (root `buf.yaml`)

```yaml
version: v2
modules:
  - path: contracts
  - path: apps
deps:
  - buf.build/googleapis/googleapis
  - buf.build/bufbuild/protovalidate
  - buf.build/gnostic/gnostic
```

### Generation configs

**App-level** (`apps/<app>/api/buf.gen.api.yaml`) — message types + error enums:

```yaml
version: v2
plugins:
  - local: [go, tool, protoc-gen-go]
    out: ./go
    opt: paths=source_relative
  - local: [go, tool, protoc-gen-go-errors]
    out: ./go
    opt: paths=source_relative
```

**Service-level** (`apps/<app>/services/<service>/buf.gen.conf.yaml`) — gRPC, HTTP, Connect, OpenAPI:

```yaml
version: v2
plugins:
  - local: [go, tool, protoc-gen-go]
    out: .
    opt: [module=cyber-ecosystem]
  - local: [go, tool, protoc-gen-go-grpc]
    out: .
    opt: [module=cyber-ecosystem]
  - local: [go, tool, protoc-gen-go-http]
    out: .
    opt: [module=cyber-ecosystem]
  - local: [go, tool, protoc-gen-connect-go]
    out: .
    opt: [module=cyber-ecosystem, simple]
  - local: [go, tool, protoc-gen-go-errors]
    out: .
    opt: [module=cyber-ecosystem]
  - local: [go, tool, protoc-gen-openapi]
    strategy: all
    out: ./apps/<app>/gen/oas
    opt: [paths=source_relative, enum_type=string, fq_schema_naming=true, default_response=false, naming=json]
```

---

## i18n Integration

The `i18n.protos` file in `internal/i18n/` lists proto files for error key generation:

```
../../../../../../contracts/errors/codes_general.proto
../../../../../../contracts/errors/codes_infra.proto
../../../../../../contracts/errors/codes_flow.proto
../../../../../../contracts/errors/codes_auth.proto
../../../../api/v1/error/error_reason.proto
```

Paths are relative from `internal/i18n/` to the proto file. Order: shared contracts first, app-specific last.

After adding new error enum values: `./nx run <project>:generate:i18n` → fill translations in `locales/v1.zh-CN.yaml`.

---

## Generation Order

```
1. genesis_api:proto:api     → App-level proto stubs (messages + errors)
2. <project>:proto:conf      → Config proto (conf.pb.go)
3. <project>:generate:i18n   → i18n stubs from error enums
4. <project>:generate:ent    → Ent ORM code
5. <project>:generate:wire   → Wire DI
6. <project>:proto:connect   → Connect clients (BFF services with HTTP annotations)
7. <project>:proto:openapi   → OpenAPI spec (web clients)
```

Or run all at once: `./nx run <project>:generate`

Proto changes MUST be generated before i18n, ent, or wire — other layers depend on generated proto types.

---

## Region Annotations in Proto

```protobuf
// region[rgba(239,83,80,0.15)] 🔴 Model

message Article { ... }

// endregion

// region[rgba(186,104,200,0.15)] 🟣 Messages

message CreateArticleRequest { ... }

// endregion

// region[rgba(52,152,219,0.2)] 🔵 Service

service ArticleService { ... }

// endregion
```

Section order: Model → Messages → Service.

---

## Common Pitfalls

### File name collision

Two service files with the same base name collide in generated output. Fix: use prefixes (`base_article.proto`, `admin_article.proto`).

### Missing HTTP registration

If a proto service has NO `google.api.http` annotations, generated code will NOT include `RegisterXxxHTTPServer`. `RegisterHTTP` and `RegisterConnect` must be no-ops.

### Wrong import path in i18n.protos

Path is relative from `internal/i18n/` to the proto file, not from the project root.

### buf generate path

Always run from workspace root with `--path` to scope: `buf generate --template apps/<app>/api/buf.gen.api.yaml --path apps/<app>/api`. Never cd into a subdirectory.

### Ent schema depends on generated code

Local mixins import generated packages that don't exist on first generation. Use the three-stage bootstrap (see `/kratos-scaffold`).

### OpenAPI path collision across BFFs

Two BFFs sharing `GET /api/v1/article` causes the OpenAPI generator to overwrite. Fix: use BFF-specific path prefixes.

### StringValue vs optional string confusion

Use `StringValue` in response/model messages (nil-safe). Use `optional string` in request messages (for validation constraints). Never mix the two in the same message type.

---

## Nx Targets

```bash
./nx run <app>_api:proto:api     # App-level proto stubs
./nx run <project>:proto:conf    # Config proto (conf.pb.go)
./nx run <project>:generate      # All generation targets
```

---

For deep architecture details, see `docs/stacks/kratos-go.md`.
