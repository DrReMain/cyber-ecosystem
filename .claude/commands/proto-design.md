# Proto Design & Generation

Guide for designing, organizing, and generating Protobuf/gRPC definitions in this monorepo. Use this when creating or modifying proto files, setting up generation for a new app, or debugging buf/generation issues.

---

## 1. Directory Structure

Proto files live under `apps/<app>/api/v1/` with this layout:

```
apps/<app>/api/v1/
  messages/            # Shared messages (request/response pairs, models)
    article.proto
    resource.proto
  service_base/        # Core gRPC services (NO HTTP annotations)
    base_article.proto
    base_resource.proto
  bff_admin/           # Admin BFF services (WITH HTTP annotations)
    admin_article.proto
    admin_resource.proto
  bff_mobile/          # Mobile BFF services (WITH HTTP annotations)
    mobile_article.proto
  error/               # App-specific error enums
    error_reason.proto
```

Why this split:
- **messages/** — pure data definitions shared by all services. No `service` blocks, no HTTP annotations.
- **service_base/** — core domain service, gRPC-only. No HTTP annotations because the base service only exposes gRPC. Other services (BFFs) call it internally.
- **bff_admin/ and bff_mobile/** — BFF services that face external clients. Include `google.api.http` annotations. Each BFF selects the subset of methods it needs.
- **error/** — error enums specific to this app. General/infra errors live in `contracts/errors/`.

---

## 2. File Naming — Prefix Required

All proto files in service directories MUST use a unique prefix to avoid Go generation collisions:

```
service_base/base_article.proto    → prefix: "base_"
bff_admin/admin_article.proto      → prefix: "admin_"
bff_mobile/mobile_article.proto    → prefix: "mobile_"
messages/article.proto             → no prefix (shared messages)
error/error_reason.proto           → no prefix (error enums)
```

Why: buf generates Go into a single `gen/go/v1/` directory. If two files define `ArticleService`, the generated `article_service_grpc.pb.go` overwrites the other. The prefix makes each service name unique in the generated output.

The prefix applies to the **file name**, not the service name in proto. Service names can still be clean (e.g., `ArticleService` in `base_article.proto`, `AdminArticleService` in `admin_article.proto`).

---

## 3. Package & Import Conventions

### Package name

All proto files in an app share the same package:

```protobuf
// buf:lint:ignore PACKAGE_DIRECTORY_MATCH
package api.<app>.v1;
```

The `PACKAGE_DIRECTORY_MATCH` ignore is required because files are in subdirectories (`messages/`, `service_base/`, etc.) but share one package.

### Go package

```protobuf
option go_package = "cyber-ecosystem/apps/<app>/gen/go/v1;<app>V1";
```

All files in the same app use the same `go_package` — they generate into the same Go package.

### Imports

Import paths are relative to the buf module root. The `buf.yaml` defines two module roots: `contracts/` and `apps/`. So:

```protobuf
// Import from the same app
import "genesis/api/v1/messages/article.proto";

// Import shared contracts
import "common/page.proto";
import "errors/errors.proto";
import "buf/validate/validate.proto";
import "google/api/annotations.proto";
import "desc/desc.proto";
import "google/protobuf/timestamp.proto";
import "google/protobuf/wrappers.proto";
```

The `genesis/` prefix comes from the `apps/` module root — buf resolves `genesis/api/v1/...` as `apps/genesis/api/v1/...`.

---

## 4. Message Design Patterns

### Model message (in messages/)

Represents the domain entity for read responses:

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

Use `StringValue` (not `optional string`) for nullable fields in model/response messages — this makes nil-vs-empty unambiguous.

### Request messages

```protobuf
message CreateArticleRequest {
  optional string title = 1 [
    (buf.validate.field).required = true,
    (buf.validate.field).string.min_len = 1
  ];
  optional string content = 2;
}
```

Use `optional string` in requests — the validation middleware checks constraints. Required fields get `(buf.validate.field).required = true`.

### ID validation pattern

IDs use xid (20-char string):

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
  // ... filters
  repeated string order_by = 100 [(buf.validate.field).cel = {
    id: "QueryArticleRequest.order_by"
    message: ""
    expression: "this.all(item, size(item) == 0 || item.matches('^(createdAt|updatedAt):(asc|desc)$'))"
  }];
}

message QueryArticleResponse {
  common.PageResponse page = 1;
  repeated GetArticleResponse list = 2;
}
```

- `order_by` uses field number 100 (high number to avoid conflicts when adding new fields).
- The CEL expression validates format: `fieldName:asc` or `fieldName:desc`.

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

Cross-field validation uses message-level CEL.

---

## 5. Service Definitions

### Core service (service_base/) — gRPC only

No HTTP annotations. Methods have `desc.method_comment` for documentation:

```protobuf
service ArticleService {
  option (desc.service_comment) = "文章服务";

  rpc CreateArticle(CreateArticleRequest) returns (CreateArticleResponse) {
    option (desc.method_comment) = "创建文章";
  }
  // ... more methods
}
```

### BFF service (bff_admin/, bff_mobile/) — with HTTP

Each method gets `google.api.http` annotations:

```protobuf
service AdminArticleService {
  option (desc.service_comment) = "管理端文章服务";

  rpc CreateArticle(CreateArticleRequest) returns (CreateArticleResponse) {
    option (desc.method_comment) = "创建文章";
    option (google.api.http) = {
      post: "/api/v1/article"
      body: "*"
    };
  }
  rpc GetArticle(GetArticleRequest) returns (GetArticleResponse) {
    option (desc.method_comment) = "查询文章详情";
    option (google.api.http) = {get: "/api/v1/article/{id}"};
  }
  // ...
}
```

HTTP path conventions:
- POST for create: `post: "/api/v1/<entity>"` with `body: "*"`
- PUT for update: `put: "/api/v1/<entity>/{id}"` with `body: "*"`
- DELETE: `delete: "/api/v1/<entity>/{id}"`
- GET single: `get: "/api/v1/<entity>/{id}"`
- GET list: `get: "/api/v1/<entity>"`
- POST for special actions: `post: "/api/v1/<entity>/{id}/sort"` or `post: "/api/v1/<entity>/{id}/status"`

---

## 6. Error Enums

App-specific errors in `error/error_reason.proto`:

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

- First value must be `_UNSPECIFIED = 0`.
- Use app-specific codes starting from 6000 to avoid collision with shared error codes.
- The `(errors.code)` option sets the HTTP status code returned to clients.

General/infra errors are in `contracts/errors/` — do NOT duplicate them here.

---

## 7. buf Toolchain

### Module configuration

Root `buf.yaml` defines two module roots:

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

This means all import paths resolve relative to `contracts/` or `apps/`.

### Generation configs

Two generation configs per app:

**`apps/<app>/api/buf.gen.api.yaml`** — generates app-level proto stubs (message types + error enums):

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

Run via: `./nx run <app>_api:proto:api`

**`apps/<app>/services/<service>/buf.gen.conf.yaml`** — generates service-level code (gRPC, HTTP, Connect, OpenAPI):

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

Run via: `./nx run <project>:proto:conf`

The `module=cyber-ecosystem` prefix rewrites import paths from `apps/<app>/gen/go/v1/` to `cyber-ecosystem/apps/<app>/gen/go/v1/` in generated code.

### Config proto

Each service has `internal/conf/conf.proto` for config generation:

```bash
./nx run <project>:proto:conf
```

This generates `conf.pb.go` from the config proto definition.

---

## 8. i18n Integration

The `i18n.protos` file in `internal/i18n/` lists proto files for error key generation:

```
../../../../../../contracts/errors/codes_general.proto
../../../../../../contracts/errors/codes_infra.proto
../../../../../../contracts/errors/codes_flow.proto
../../../../../../contracts/errors/codes_auth.proto
../../../../api/v1/error/error_reason.proto
```

The path is relative from `internal/i18n/` to the proto file. Order matters: general contracts first, app-specific last.

After adding new error enum values:
1. Run `./nx run <project>:generate:i18n` to regenerate YAML stubs
2. Fill in translations in `locales/v1.zh-CN.yaml` and `locales/v1.en-US.yaml`

---

## 9. Generation Order

Run targets in this order (the `generate` target handles this automatically):

```
1. genesis_api:proto:api     → Generate app-level proto stubs (messages + errors)
2. <project>:proto:conf      → Generate config proto (conf.pb.go)
3. <project>:generate:i18n   → Generate i18n stubs from error enums
4. <project>:generate:ent    → Generate Ent ORM code
5. <project>:generate:wire   → Generate Wire DI
```

Or run all at once: `./nx run <project>:generate`

**Important:** proto changes MUST be generated before i18n, ent, or wire — other layers depend on generated proto types.

---

## 10. Common Pitfalls

### File name collision

Two service files with the same base name will collide in generated output:
- `service_base/article.proto` + `bff_admin/article.proto` → both generate into the same Go files
- Fix: use prefixes (`base_article.proto`, `admin_article.proto`)

### Missing HTTP registration

If a proto service has NO `google.api.http` annotations, the generated code will NOT include `RegisterXxxHTTPServer`. The service's `RegisterHTTP` method must be a no-op:

```go
func (s *ArticleService) RegisterHTTP(_ *http.Server) {}
```

### Wrong import path in i18n.protos

The path is relative from `internal/i18n/` to the proto file, not from the project root. Check the actual directory depth:
- From `services/base/internal/i18n/` to `api/v1/error/` = `../../../../api/v1/error/error_reason.proto`
- From `services/base/internal/i18n/` to `contracts/errors/` = `../../../../../../contracts/errors/codes_general.proto`

### buf generate path

Always run buf generate from the workspace root with `--path` to scope generation:

```bash
buf generate --template apps/<app>/api/buf.gen.api.yaml --path apps/<app>/api
```

Never cd into a subdirectory to run buf.

### Ent schema depends on generated code

Local mixins (`local_mixins/soft_delete.go`, `local_mixins/sort.go`) import generated Ent packages that don't exist on first generation. Use the three-stage bootstrap:

1. Generate without local_mixins and without intercept feature
2. Generate with intercept feature (adds interceptor support)
3. Add local_mixins back and generate again

---

## 11. Region Annotations in Proto

Proto files use the same region annotation pattern as Go:

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
