# Slice 1: Audit Service — 完整垂直闭环

> **闭环标准**: 服务可启动，通过 gRPC 提交审计事件，通过 HTTP 查询审计日志。

## 目标

构建 Audit 服务——yggdrasil 中最简单的服务（无 RBAC、无 condition、无 DataScope）。此 Slice 验证整个技术栈：Proto 生成、Ent ORM、Wire DI、多传输层（HTTP/gRPC/ConnectRPC）。

后续所有服务的代码模式都以本 Slice 为模板。

## 前置条件

- Slice 0 完成（app 骨架 + error_reason.proto 生成）
- `shared-go/` 模块可用
- `contracts/` 共享类型可用
- PostgreSQL 和 Redis 在本地可用（docker compose）

## 端口

| Transport | Address |
|-----------|---------|
| HTTP | `0.0.0.0:11001` |
| gRPC | `0.0.0.0:12001` |
| ConnectRPC | `0.0.0.0:13001` |
| Ops | `0.0.0.0:14001` |

---

## Step 1: 目录结构 + Nx 配置

```bash
mkdir -p apps/yggdrasil/services/audit/{cmd/app,configs,internal/{conf,data/ent/schema,biz,server/{locales},service}}
```

**文件**: `apps/yggdrasil/services/audit/project.json`

```json
{
  "name": "yggdrasil_audit",
  "$schema": "../../../../node_modules/nx/schemas/project-schema.json",
  "implicitDependencies": ["yggdrasil_api", "shared-go"],
  "targets": {
    "proto:conf": {
      "executor": "nx:run-commands",
      "options": {
        "command": "buf generate --template apps/yggdrasil/services/audit/buf.gen.conf.yaml --path apps/yggdrasil/services/audit/internal/conf"
      }
    },
    "generate": {
      "dependsOn": ["yggdrasil_api:proto:api", "proto:conf"],
      "executor": "nx:run-commands",
      "options": {
        "command": "cd apps/yggdrasil/services/audit && wire ./cmd/app/..."
      }
    },
    "dev": {
      "executor": "nx:run-commands",
      "options": {
        "command": "cd apps/yggdrasil/services/audit && go run ./cmd/app/... -conf ./configs"
      }
    },
    "build": {
      "executor": "nx:run-commands",
      "options": {
        "command": "cd apps/yggdrasil/services/audit && go build -o ./bin/audit ./cmd/app/..."
      }
    },
    "ent:new": {
      "executor": "nx:run-commands",
      "options": {
        "command": "cd apps/yggdrasil/services/audit && go run -mod=mod entgo.io/ent/cmd/ent new --target internal/data/ent/schema"
      }
    }
  }
}
```

**文件**: `apps/yggdrasil/services/audit/buf.gen.conf.yaml`

```yaml
version: v2
plugins:
  - local: [go, tool, protoc-gen-go]
    out: .
    opt: paths=source_relative
```

验证: `cat apps/yggdrasil/services/audit/project.json`

---

## Step 2: API Proto 定义

### 2a: audit_log.proto — 管理查询 API

**文件**: `apps/yggdrasil/api/v1/audit_log.proto`

```protobuf
syntax = "proto3";

// buf:lint:ignore PACKAGE_DIRECTORY_MATCH
package api.yggdrasil.v1;

import "buf/validate/validate.proto";
import "common/page.proto";
import "desc/desc.proto";
import "google/api/annotations.proto";
import "google/protobuf/timestamp.proto";
import "google/protobuf/wrappers.proto";

option go_package = "cyber-ecosystem/apps/yggdrasil/gen/go/v1;yggdrasilV1";

// AuditLogService
service AuditLogService {
  option (desc.service_comment) = "审计日志查询服务";

  // GetAuditLog
  rpc GetAuditLog(GetAuditLogRequest) returns (GetAuditLogResponse) {
    option (desc.method_comment) = "查询审计日志详情";
    option (google.api.http) = {get: "/audit/log/{id}"};
  }

  // QueryAuditLog
  rpc QueryAuditLog(QueryAuditLogRequest) returns (QueryAuditLogResponse) {
    option (desc.method_comment) = "查询审计日志列表";
    option (google.api.http) = {post: "/audit/log/query" body: "*"};
  }
}

message GetAuditLogRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
}

message GetAuditLogResponse {
  string id = 1;
  string actor_id = 2;
  google.protobuf.StringValue actor_name = 3;
  google.protobuf.StringValue department_id = 4;
  string action = 5;
  string resource_type = 6;
  google.protobuf.StringValue resource_id = 7;
  string service_name = 8;
  string result = 9;
  google.protobuf.StringValue ip = 10;
  google.protobuf.StringValue user_agent = 11;
  google.protobuf.StringValue detail = 12;
  google.protobuf.Timestamp created_at = 13;
}

message QueryAuditLogRequest {
  common.PageRequest page = 1;
  optional string actor_id = 2;
  optional string action = 3;
  optional string resource_type = 4;
  optional string resource_id = 5;
  optional string service_name = 6;
  optional string department_id = 7;
  optional string result = 8;

  repeated string order_by = 100 [(buf.validate.field).cel = {
    id: "QueryAuditLogRequest.order_by"
    message: ""
    expression: "this.all(item, size(item) == 0 || item.matches('^(created_at):(asc|desc)$'))"
  }];
}

message QueryAuditLogResponse {
  common.PageResponse page = 1;
  repeated GetAuditLogResponse list = 2;
}
```

### 2b: audit_sink.proto — 内部事件接收 API

**文件**: `apps/yggdrasil/api/v1/audit_sink.proto`

```protobuf
syntax = "proto3";

// buf:lint:ignore PACKAGE_DIRECTORY_MATCH
package api.yggdrasil.v1;

import "buf/validate/validate.proto";
import "google/protobuf/timestamp.proto";

option go_package = "cyber-ecosystem/apps/yggdrasil/gen/go/v1;yggdrasilV1";

// AuditSinkService — 内部 API，接收其他服务的审计事件。
service AuditSinkService {
  // SubmitEvents 批量接收审计事件。
  rpc SubmitEvents(SubmitEventsRequest) returns (SubmitEventsResponse);
}

message AuditEvent {
  string actor_id = 1 [(buf.validate.field).string.min_len = 1];
  string actor_name = 2;
  string department_id = 3;
  string action = 4 [(buf.validate.field).string.min_len = 1];
  string resource_type = 5 [(buf.validate.field).string.min_len = 1];
  string resource_id = 6;
  string service_name = 7 [(buf.validate.field).string.min_len = 1];
  string result = 8 [(buf.validate.field).string.min_len = 1];
  string ip = 9;
  string user_agent = 10;
  string detail = 11;
  google.protobuf.Timestamp timestamp = 12;
}

message SubmitEventsRequest {
  repeated AuditEvent events = 1 [(buf.validate.field).repeated = {
    min_items: 1
    max_items: 100
  }];
}

message SubmitEventsResponse {
  int32 accepted = 1;
}
```

### 2c: 生成 Proto 代码

```bash
buf lint apps/yggdrasil/api/v1/audit_log.proto
buf lint apps/yggdrasil/api/v1/audit_sink.proto
./nx run yggdrasil_api:proto:api
```

验证: `ls apps/yggdrasil/gen/go/v1/audit_*` — 应有 10+ 个生成文件

提交:
```bash
git add apps/yggdrasil/api/v1/audit_log.proto apps/yggdrasil/api/v1/audit_sink.proto apps/yggdrasil/gen/
git commit -m "feat(yggdrasil): add audit service API proto definitions"
```

---

## Step 3: Conf Proto + 生成

**文件**: `apps/yggdrasil/services/audit/internal/conf/conf.proto`

```protobuf
syntax = "proto3";

// buf:lint:ignore PACKAGE_DIRECTORY_MATCH
// buf:lint:ignore PACKAGE_VERSION_SUFFIX
package cyber_ecosystem.audit_conf;

import "google/protobuf/duration.proto";

option go_package = "cyber-ecosystem/apps/yggdrasil/services/audit/internal/conf;conf";

message Bootstrap {
  Server server = 1;
  Auth auth = 2;
  Log log = 3;
  Data data = 4;
  Trace trace = 5;
  Ops ops = 6;
}

message Server {
  message HTTP {
    string network = 1;
    string addr = 2;
    google.protobuf.Duration timeout = 3;
  }
  message GRPC {
    string network = 1;
    string addr = 2;
    google.protobuf.Duration timeout = 3;
  }
  message Connect {
    string network = 1;
    string addr = 2;
    google.protobuf.Duration timeout = 3;
  }
  HTTP http = 1;
  GRPC grpc = 2;
  Connect connect = 3;
}

message Auth {
  string secret = 1;
}

message Log {
  message Cache {
    bool enabled = 1;
    string level = 2;
    bool slow_query = 3;
    google.protobuf.Duration slow_query_threshold = 4;
  }
  message Ent {
    bool enabled = 1;
    string level = 2;
    bool slow_query = 3;
    google.protobuf.Duration slow_query_threshold = 4;
  }
  message Console {
    bool enabled = 1;
    bool color = 2;
    string format = 3;
  }
  message File {
    bool enabled = 1;
    string path = 2;
    int32 max_size = 3;
    int32 max_backups = 4;
    int32 max_age = 5;
    bool compress = 6;
  }
  message Loki {
    bool enabled = 1;
    string url = 2;
    map<string, string> labels = 3;
    google.protobuf.Duration batch_wait = 4;
    int32 batch_size = 5;
  }
  string level = 1;
  Cache cache = 2;
  Ent ent = 3;
  Console console = 4;
  File file = 5;
  Loki loki = 6;
}

message Data {
  message Database {
    string driver = 1;
    string host = 2;
    int32 port = 3;
    string user = 4;
    string password = 5;
    string db_name = 6;
    int32 max_open_conns = 7;
    int32 max_idle_conns = 8;
    google.protobuf.Duration conn_max_lifetime = 9;
    bool migrate = 10;
  }
  message Cache {
    message Memory { bool otel_enabled = 1; }
    message Redis {
      string network = 1;
      string addr = 2;
      google.protobuf.Duration read_timeout = 3;
      google.protobuf.Duration write_timeout = 4;
      int32 pool_size = 5;
      int32 min_idle_conns = 6;
      google.protobuf.Duration conn_max_lifetime = 7;
      string password = 8;
      int32 db = 9;
      bool otel_enabled = 10;
    }
    string type = 1;
    Memory memory = 2;
    Redis redis = 3;
  }
  Database database = 1;
  Cache cache = 2;
}

message Trace {
  string endpoint = 1;
  bool insecure = 2;
}

message Ops {
  message Pprof {
    bool enabled = 1;
    bool cpu_enabled = 2;
    bool heap_enabled = 3;
    bool goroutine_enabled = 4;
    bool mutex_enabled = 5;
    bool thread_enabled = 6;
    bool trace_enabled = 7;
  }
  bool enabled = 1;
  string network = 2;
  string addr = 3;
  string metrics = 4;
  Pprof pprof = 5;
}
```

生成:
```bash
./nx run yggdrasil_audit:proto:conf
```

验证: `ls apps/yggdrasil/services/audit/internal/conf/conf.pb.go`

提交:
```bash
git add apps/yggdrasil/services/audit/
git commit -m "feat(yggdrasil): add audit service config proto"
```

---

## Step 4: Ent Schema + 生成

**文件**: `apps/yggdrasil/services/audit/internal/data/ent/schema/audit_log.go`

```go
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

type AuditLog struct {
	ent.Schema
}

func (AuditLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("actor_id").NotEmpty().MaxLen(20),
		field.String("actor_name").Default("").MaxLen(200),
		field.String("department_id").Optional().Nillable().MaxLen(20),
		field.String("action").NotEmpty().MaxLen(50),
		field.String("resource_type").NotEmpty().MaxLen(50),
		field.String("resource_id").Optional().Nillable().MaxLen(20),
		field.String("service_name").NotEmpty().MaxLen(50),
		field.String("result").NotEmpty().MaxLen(20),
		field.String("ip").Default("").MaxLen(50),
		field.String("user_agent").Default("").MaxLen(500),
		field.Text("detail").Optional().Nillable(),
	}
}

func (AuditLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (AuditLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("actor_id"),
		index.Fields("action"),
		index.Fields("resource_type", "resource_id"),
		index.Fields("service_name"),
		index.Fields("department_id"),
		index.Fields("result"),
		index.Fields("created_at"),
		index.Fields("service_name", "created_at"),
		index.Fields("actor_id", "created_at"),
	}
}

func (AuditLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "audit_log"},
	}
}
```

**文件**: `apps/yggdrasil/services/audit/internal/data/ent/generate.go`

```go
package ent

//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate ./schema
```

生成:
```bash
cd apps/yggdrasil/services/audit && go generate ./internal/data/ent/...
```

验证: `ls apps/yggdrasil/services/audit/internal/data/ent/client.go`

提交:
```bash
git add apps/yggdrasil/services/audit/internal/data/ent/
git commit -m "feat(yggdrasil): add audit log Ent schema and generated code"
```

---

## Step 5: Biz 层

**文件**: `apps/yggdrasil/services/audit/internal/biz/biz.go`

```go
package biz

import (
	"context"

	"github.com/google/wire"

	"github.com/go-kratos/kratos/v2/log"
)

type Transaction interface {
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type UC struct {
	log *log.Helper
	tm  Transaction
}

var ProviderSet = wire.NewSet(
	NewAuditLogUC,
)
```

**文件**: `apps/yggdrasil/services/audit/internal/biz/uc_audit.go`

```go
package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/contracts/go/common"
	"cyber-ecosystem/shared-go/utils"
)

// Model

type AuditLog struct {
	ID           *string
	ActorID      *string
	ActorName    *string
	DepartmentID *string
	Action       *string
	ResourceType *string
	ResourceID   *string
	ServiceName  *string
	Result       *string
	IP           *string
	UserAgent    *string
	Detail       *string
	CreatedAt    *time.Time
}

type AuditLogQueryIn struct {
	*common.PageRequest
	OrderBy      []*utils.OrderBy
	ActorID      *string
	Action       *string
	ResourceType *string
	ResourceID   *string
	ServiceName  *string
	DepartmentID *string
	Result       *string
}

type AuditLogQueryOut struct {
	*common.PageResponse
	List []*AuditLog
}

// Port

type AuditLogRP interface {
	BatchCreate(ctx context.Context, events []*AuditLog) error
	Get(ctx context.Context, id string) (*AuditLog, error)
	Query(ctx context.Context, in *AuditLogQueryIn) (*AuditLogQueryOut, error)
}

// UC

type AuditLogUC struct {
	UC
	auditLogRP AuditLogRP
}

func NewAuditLogUC(logger log.Logger, tm Transaction, auditLogRP AuditLogRP) *AuditLogUC {
	return &AuditLogUC{
		UC: UC{
			log: log.NewHelper(log.With(logger, "module", "biz/uc_audit")),
			tm:  tm,
		},
		auditLogRP: auditLogRP,
	}
}

func (uc *AuditLogUC) BatchCreate(ctx context.Context, events []*AuditLog) error {
	return uc.auditLogRP.BatchCreate(ctx, events)
}

func (uc *AuditLogUC) Get(ctx context.Context, id string) (*AuditLog, error) {
	return uc.auditLogRP.Get(ctx, id)
}

func (uc *AuditLogUC) Query(ctx context.Context, in *AuditLogQueryIn) (*AuditLogQueryOut, error) {
	return uc.auditLogRP.Query(ctx, in)
}
```

验证: `gofmt -e apps/yggdrasil/services/audit/internal/biz/*.go`

---

## Step 6: Data 层

按「通用模式参考」（见 master plan）创建以下文件，只改 import 路径和 DB 名称：

**文件**: `apps/yggdrasil/services/audit/internal/data/store.go`
- import `ent` 路径: `cyber-ecosystem/apps/yggdrasil/services/audit/internal/data/ent`
- 其他与通用模式完全一致

**文件**: `apps/yggdrasil/services/audit/internal/data/store_ent.go`
- import `conf` 路径: `cyber-ecosystem/apps/yggdrasil/services/audit/internal/conf`
- import `ent`, `migrate`, `runtime` 路径: `cyber-ecosystem/apps/yggdrasil/services/audit/internal/data/ent/...`
- `defaultError` 使用 `yggdrasilV1.ErrorErrorReason*` 系列
- import alias: `yggdrasilV1 "cyber-ecosystem/apps/yggdrasil/gen/go/v1"`

**文件**: `apps/yggdrasil/services/audit/internal/data/store_cache.go`
- import `conf` 路径: `cyber-ecosystem/apps/yggdrasil/services/audit/internal/conf`
- 其他与通用模式完全一致

**文件**: `apps/yggdrasil/services/audit/internal/data/data.go`

```go
package data

import (
	"github.com/google/wire"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/yggdrasil/services/audit/internal/biz"
)

type RP struct {
	log   *log.Helper
	store *Store
}

var ProviderSet = wire.NewSet(
	NewStore,
	NewCache,
	NewEntClient,
	wire.Bind(new(biz.Transaction), new(*Store)),
	NewAuditLogRP,
)
```

**文件**: `apps/yggdrasil/services/audit/internal/data/rp_audit_log.go`

```go
package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/orm/ent/entutil"
	"cyber-ecosystem/shared-go/utils"

	yggdrasilV1 "cyber-ecosystem/apps/yggdrasil/gen/go/v1"
	"cyber-ecosystem/apps/yggdrasil/services/audit/internal/biz"
	"cyber-ecosystem/apps/yggdrasil/services/audit/internal/data/ent"
	entauditlog "cyber-ecosystem/apps/yggdrasil/services/audit/internal/data/ent/auditlog"
)

type auditLogRP struct {
	RP
}

func NewAuditLogRP(logger log.Logger, store *Store) biz.AuditLogRP {
	return &auditLogRP{
		RP: RP{
			log:   log.NewHelper(log.With(logger, "module", "data/rp_audit_log")),
			store: store,
		},
	}
}

func (rp *auditLogRP) BatchCreate(ctx context.Context, events []*biz.AuditLog) error {
	if len(events) == 0 {
		return nil
	}
	client := rp.store.GetClient(ctx)
	bulk := make([]*ent.AuditLogCreate, len(events))
	for i, e := range events {
		builder := client.AuditLog.Create().
			SetActorID(*e.ActorID).
			SetAction(*e.Action).
			SetResourceType(*e.ResourceType).
			SetServiceName(*e.ServiceName).
			SetResult(*e.Result)
		if e.ActorName != nil {
			builder.SetActorName(*e.ActorName)
		}
		if e.DepartmentID != nil {
			builder.SetDepartmentID(*e.DepartmentID)
		}
		if e.ResourceID != nil {
			builder.SetResourceID(*e.ResourceID)
		}
		if e.IP != nil {
			builder.SetIP(*e.IP)
		}
		if e.UserAgent != nil {
			builder.SetUserAgent(*e.UserAgent)
		}
		if e.Detail != nil {
			builder.SetDetail(*e.Detail)
		}
		bulk[i] = builder
	}
	if _, err := client.AuditLog.CreateBulk(bulk).Save(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *auditLogRP) Get(ctx context.Context, id string) (*biz.AuditLog, error) {
	result, err := rp.store.GetClient(ctx).AuditLog.Get(ctx, id)
	if err != nil {
		return nil, HandleError(err)
	}
	return mapAuditLog(result), nil
}

func (rp *auditLogRP) Query(ctx context.Context, in *biz.AuditLogQueryIn) (*biz.AuditLogQueryOut, error) {
	query := rp.store.GetClient(ctx).AuditLog.Query()
	entutil.WherePtr(query, in.ActorID, entauditlog.ActorIDEQ)
	entutil.WherePtr(query, in.Action, entauditlog.ActionEQ)
	entutil.WherePtr(query, in.ResourceType, entauditlog.ResourceTypeEQ)
	entutil.WherePtr(query, in.ResourceID, entauditlog.ResourceIDEQ)
	entutil.WherePtr(query, in.ServiceName, entauditlog.ServiceNameEQ)
	entutil.WherePtr(query, in.DepartmentID, entauditlog.DepartmentIDEQ)
	entutil.WherePtr(query, in.Result, entauditlog.ResultEQ)
	entutil.ApplyOrderBy(in.OrderBy, ent.Asc, ent.Desc, entutil.FOMapping{
		"created_at": func(sel entutil.SQLSelector) { query.Order(sel(entauditlog.FieldCreatedAt)) },
	})
	total, offset, limit, err := entutil.ApplyPagination(ctx, query, in.PageRequest,
		entutil.NewPageConfig(entutil.DefaultPageSize, entutil.DefaultPageSizeUnlimit),
		yggdrasilV1.ErrorErrorReasonPaginationInvalidArgument(""),
	)
	if err != nil {
		return nil, HandleError(err)
	}
	list, err := query.Offset(offset).Limit(limit).All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}
	return &biz.AuditLogQueryOut{
		PageResponse: entutil.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(list, mapAuditLog),
	}, nil
}

func mapAuditLog(r *ent.AuditLog) *biz.AuditLog {
	return &biz.AuditLog{
		ID:           &r.ID,
		ActorID:      &r.ActorID,
		ActorName:    &r.ActorName,
		DepartmentID: r.DepartmentID,
		Action:       &r.Action,
		ResourceType: &r.ResourceType,
		ResourceID:   r.ResourceID,
		ServiceName:  &r.ServiceName,
		Result:       &r.Result,
		IP:           &r.IP,
		UserAgent:    &r.UserAgent,
		Detail:       r.Detail,
		CreatedAt:    &r.CreatedAt,
	}
}
```

验证: `gofmt -e apps/yggdrasil/services/audit/internal/data/*.go`

提交:
```bash
git add apps/yggdrasil/services/audit/internal/
git commit -m "feat(yggdrasil): add audit biz and data layers"
```

---

## Step 7: Service 层

**文件**: `apps/yggdrasil/services/audit/internal/service/service.go`

```go
package service

import (
	"github.com/google/wire"

	"github.com/go-kratos/kratos/v2/transport/grpc"
	krahttp "github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"
)

type Registrar interface {
	RegisterGRPC(*grpc.Server)
	RegisterHTTP(*krahttp.Server)
	RegisterConnect(*connect.Server)
}

var ProviderSet = wire.NewSet(
	NewRegistrarList,
	NewAuditLogService,
	NewAuditSinkService,
)

func NewRegistrarList(
	s1 *AuditLogService,
	s2 *AuditSinkService,
) []Registrar {
	return []Registrar{s1, s2}
}
```

**文件**: `apps/yggdrasil/services/audit/internal/service/audit_log.go`

```go
package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	krahttp "github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	yggdrasilV1 "cyber-ecosystem/apps/yggdrasil/gen/go/v1"
	yggdrasilV1connect "cyber-ecosystem/apps/yggdrasil/gen/go/v1/yggdrasilV1connect"
	"cyber-ecosystem/apps/yggdrasil/services/audit/internal/biz"
)

type AuditLogService struct {
	yggdrasilV1.UnimplementedAuditLogServiceServer
	log        *log.Helper
	auditLogUC *biz.AuditLogUC
}

func NewAuditLogService(logger log.Logger, auditLogUC *biz.AuditLogUC) *AuditLogService {
	return &AuditLogService{
		log:        log.NewHelper(log.With(logger, "module", "service/audit_log")),
		auditLogUC: auditLogUC,
	}
}

func (s *AuditLogService) RegisterGRPC(srv *grpc.Server) {
	yggdrasilV1.RegisterAuditLogServiceServer(srv, s)
}

func (s *AuditLogService) RegisterHTTP(srv *krahttp.Server) {
	yggdrasilV1.RegisterAuditLogServiceHTTPServer(srv, s)
}

func (s *AuditLogService) RegisterConnect(srv *connect.Server) {
	srv.Register(yggdrasilV1connect.NewAuditLogServiceHandler(s, srv.HandlerOptions()...))
}

func (s *AuditLogService) GetAuditLog(ctx context.Context, in *yggdrasilV1.GetAuditLogRequest) (*yggdrasilV1.GetAuditLogResponse, error) {
	result, err := s.auditLogUC.Get(ctx, in.Id)
	if err != nil {
		return nil, err
	}
	return s.auditLogToProto(result), nil
}

func (s *AuditLogService) QueryAuditLog(ctx context.Context, in *yggdrasilV1.QueryAuditLogRequest) (*yggdrasilV1.QueryAuditLogResponse, error) {
	queryIn := &biz.AuditLogQueryIn{
		PageRequest:  in.Page,
		OrderBy:      utils.ParseOrderBy(in.OrderBy),
		ActorID:      optStr(in.ActorId),
		Action:       optStr(in.Action),
		ResourceType: optStr(in.ResourceType),
		ResourceID:   optStr(in.ResourceId),
		ServiceName:  optStr(in.ServiceName),
		DepartmentID: optStr(in.DepartmentId),
		Result:       optStr(in.Result),
	}
	out, err := s.auditLogUC.Query(ctx, queryIn)
	if err != nil {
		return nil, err
	}
	list := utils.SliceMap(out.List, s.auditLogToProto)
	return &yggdrasilV1.QueryAuditLogResponse{
		Page: out.PageResponse,
		List: list,
	}, nil
}

func (s *AuditLogService) auditLogToProto(e *biz.AuditLog) *yggdrasilV1.GetAuditLogResponse {
	return &yggdrasilV1.GetAuditLogResponse{
		Id:           utils.Deref(e.ID, ""),
		ActorId:      utils.Deref(e.ActorID, ""),
		ActorName:    utils.Wrap(e.ActorName, utils.StringW),
		DepartmentId: utils.Wrap(e.DepartmentID, utils.StringW),
		Action:       utils.Deref(e.Action, ""),
		ResourceType: utils.Deref(e.ResourceType, ""),
		ResourceId:   utils.Wrap(e.ResourceID, utils.StringW),
		ServiceName:  utils.Deref(e.ServiceName, ""),
		Result:       utils.Deref(e.Result, ""),
		Ip:           utils.Wrap(e.IP, utils.StringW),
		UserAgent:    utils.Wrap(e.UserAgent, utils.StringW),
		Detail:       utils.Wrap(e.Detail, utils.StringW),
		CreatedAt:    utils.ToTimestamp(e.CreatedAt),
	}
}

func optStr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
```

**文件**: `apps/yggdrasil/services/audit/internal/service/audit_sink.go`

```go
package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	krahttp "github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"

	yggdrasilV1 "cyber-ecosystem/apps/yggdrasil/gen/go/v1"
	yggdrasilV1connect "cyber-ecosystem/apps/yggdrasil/gen/go/v1/yggdrasilV1connect"
	"cyber-ecosystem/apps/yggdrasil/services/audit/internal/biz"
)

type AuditSinkService struct {
	yggdrasilV1.UnimplementedAuditSinkServiceServer
	log        *log.Helper
	auditLogUC *biz.AuditLogUC
}

func NewAuditSinkService(logger log.Logger, auditLogUC *biz.AuditLogUC) *AuditSinkService {
	return &AuditSinkService{
		log:        log.NewHelper(log.With(logger, "module", "service/audit_sink")),
		auditLogUC: auditLogUC,
	}
}

func (s *AuditSinkService) RegisterGRPC(srv *grpc.Server) {
	yggdrasilV1.RegisterAuditSinkServiceServer(srv, s)
}

func (s *AuditSinkService) RegisterHTTP(srv *krahttp.Server) {
	yggdrasilV1.RegisterAuditSinkServiceHTTPServer(srv, s)
}

func (s *AuditSinkService) RegisterConnect(srv *connect.Server) {
	srv.Register(yggdrasilV1connect.NewAuditSinkServiceHandler(s, srv.HandlerOptions()...))
}

func (s *AuditSinkService) SubmitEvents(ctx context.Context, in *yggdrasilV1.SubmitEventsRequest) (*yggdrasilV1.SubmitEventsResponse, error) {
	events := make([]*biz.AuditLog, len(in.Events))
	for i, e := range in.Events {
		events[i] = &biz.AuditLog{
			ActorID:      &e.ActorId,
			ActorName:    strPtr(e.ActorName),
			DepartmentID: strPtrNonEmpty(e.DepartmentId),
			Action:       &e.Action,
			ResourceType: &e.ResourceType,
			ResourceID:   strPtrNonEmpty(e.ResourceId),
			ServiceName:  &e.ServiceName,
			Result:       &e.Result,
			IP:           strPtr(e.Ip),
			UserAgent:    strPtr(e.UserAgent),
			Detail:       strPtrNonEmpty(e.Detail),
		}
	}
	if err := s.auditLogUC.BatchCreate(ctx, events); err != nil {
		return nil, err
	}
	return &yggdrasilV1.SubmitEventsResponse{
		Accepted: int32(len(events)),
	}, nil
}

func strPtr(s string) *string       { return &s }
func strPtrNonEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
```

验证: `gofmt -e apps/yggdrasil/services/audit/internal/service/*.go`

---

## Step 8: Server 层

按「通用模式参考」创建以下文件，只改 import 路径：

**文件**: `apps/yggdrasil/services/audit/internal/server/server.go`
- init() 错误映射使用 `yggdrasilV1.ErrorErrorReason*`
- `ProviderSet = wire.NewSet(NewOpsServer, NewGRPCServer, NewHTTPServer, NewConnectServer, NewI18nBundle)`

**文件**: `apps/yggdrasil/services/audit/internal/server/grpc.go`
- 构造函数签名: `NewGRPCServer(c *conf.Server, ca *conf.Auth, logger log.Logger, registrar []service.Registrar, tp *trace.TracerProvider, _metricRequests metric.Int64Counter, _metricSeconds metric.Float64Histogram, i18nBundle *i18n.Bundle)`
- Audit 服务**无** Security 中间件。中间件链: `i18n → recovery → ratelimit → metrics → [tracing] → metadata → logging → selector(jwt) → validate`
- JWT key function: `func(token *jwtv5.Token) (any, error) { return []byte(ca.Secret), nil }`

**文件**: `apps/yggdrasil/services/audit/internal/server/http.go`
- 同 grpc.go 的构造函数签名 + CORS filter
- 同样的简化中间件链（无 session/rbac/condition/scope）

**文件**: `apps/yggdrasil/services/audit/internal/server/connect.go`
- 同 http.go 但使用 `connect.Server`

**文件**: `apps/yggdrasil/services/audit/internal/server/ops.go`
- 与通用模式完全一致，只改 conf import 路径

**文件**: `apps/yggdrasil/services/audit/internal/server/i18n.go`
- `//go:embed locales/*.yaml` + `i18n.NewBundleFS(locales, "locales", "v1", language.Make("zh-CN"))`

**文件**: `apps/yggdrasil/services/audit/internal/server/locales/v1.en-US.yaml`

```yaml
ERROR_REASON_UNSPECIFIED: "Unknown error"
ERROR_REASON_ENT_NOT_FOUND: "Audit log not found"
ERROR_REASON_ENT_VALIDATION: "Data parameter validation failed"
ERROR_REASON_ENT_NOT_SINGULAR: "Data is not singular"
ERROR_REASON_ENT_NOT_LOADED: "Data not loaded"
ERROR_REASON_ENT_CONSTRAINT: "Data constraint conflict"
ERROR_REASON_RATELIMIT: "Too many requests, please slow down"
ERROR_REASON_CIRCUITBREAKER: "Service temporarily unavailable"
ERROR_REASON_VALIDATOR: "Request parameter validation error"
ERROR_REASON_PAGINATION_INVALID_ARGUMENT: "Invalid pagination argument"
ERROR_REASON_UNAUTHORIZED: "Authentication failed"
ERROR_REASON_INVALID_ARGUMENT: "Invalid argument"
```

**文件**: `apps/yggdrasil/services/audit/internal/server/locales/v1.zh-CN.yaml`

```yaml
ERROR_REASON_UNSPECIFIED: "未知错误"
ERROR_REASON_ENT_NOT_FOUND: "审计日志未找到"
ERROR_REASON_ENT_VALIDATION: "数据参数验证失败"
ERROR_REASON_ENT_NOT_SINGULAR: "数据非单一"
ERROR_REASON_ENT_NOT_LOADED: "数据未加载"
ERROR_REASON_ENT_CONSTRAINT: "数据约束冲突"
ERROR_REASON_RATELIMIT: "请求过于频繁，请稍后再试"
ERROR_REASON_CIRCUITBREAKER: "服务暂时不可用，请稍后重试"
ERROR_REASON_VALIDATOR: "请求参数验证错误"
ERROR_REASON_PAGINATION_INVALID_ARGUMENT: "请求分页参数无效"
ERROR_REASON_UNAUTHORIZED: "身份认证失败"
ERROR_REASON_INVALID_ARGUMENT: "参数无效"
```

验证: `gofmt -e apps/yggdrasil/services/audit/internal/server/*.go`

提交:
```bash
git add apps/yggdrasil/services/audit/internal/
git commit -m "feat(yggdrasil): add audit service and server layers"
```

---

## Step 9: Main + Wire + Config

**文件**: `apps/yggdrasil/services/audit/cmd/app/main.go`

按通用模式创建。关键参数:
- `Name string = "yggdrasil_audit"`
- `flagConf` default: `../../configs`
- `wireApp` 参数: `*conf.Server, *conf.Auth, *conf.Log, *conf.Data, *conf.Ops, log.Logger, *tracesdk.TracerProvider, *metricsdk.MeterProvider, metric.Int64Counter, metric.Float64Histogram`
- 注意: 无 `*conf.Security` 和 `*conf.Super`（audit 不需要）

**文件**: `apps/yggdrasil/services/audit/cmd/app/wire.go`

```go
//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/yggdrasil/services/audit/internal/biz"
	"cyber-ecosystem/apps/yggdrasil/services/audit/internal/conf"
	"cyber-ecosystem/apps/yggdrasil/services/audit/internal/data"
	"cyber-ecosystem/apps/yggdrasil/services/audit/internal/server"
	"cyber-ecosystem/apps/yggdrasil/services/audit/internal/service"
)

func wireApp(
	*conf.Server,
	*conf.Auth,
	*conf.Log,
	*conf.Data,
	*conf.Ops,
	log.Logger,
	*tracesdk.TracerProvider,
	*metricsdk.MeterProvider,
	metric.Int64Counter,
	metric.Float64Histogram,
) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		service.ProviderSet,
		biz.ProviderSet,
		data.ProviderSet,
		newApp,
	))
}
```

**文件**: `apps/yggdrasil/services/audit/configs/config.yaml`

```yaml
server:
  http:
    addr: 0.0.0.0:11001
    timeout: 10s
  grpc:
    addr: 0.0.0.0:12001
    timeout: 10s
  connect:
    addr: 0.0.0.0:13001
    timeout: 10s

auth:
  secret: secret

log:
  level: debug
  ent:
    enabled: false
  cache:
    enabled: false
  console:
    enabled: true
    color: true
    format: "console"
  file:
    enabled: false
  loki:
    enabled: false

data:
  database:
    driver: postgres
    host: localhost
    port: 5432
    user: postgres
    password: postgres
    db_name: cyber_ecosystem_yggdrasil_audit
    max_open_conns: 10
    max_idle_conns: 5
    conn_max_lifetime: 300s
    migrate: true
  cache:
    type: memory
    memory:
      otel_enabled: true

trace:
  insecure: true
  endpoint: "http://localhost:4318/v1/traces"

ops:
  enabled: true
  addr: "0.0.0.0:14001"
  metrics: "/metrics"
  pprof:
    enabled: false
```

验证: `gofmt -e apps/yggdrasil/services/audit/cmd/app/main.go apps/yggdrasil/services/audit/cmd/app/wire.go`

---

## Step 10: 编译闭环

```bash
cd apps/yggdrasil/services/audit && go mod tidy
./nx run yggdrasil_audit:generate
./nx run yggdrasil_audit:build
```

验证: `ls apps/yggdrasil/services/audit/bin/audit`

如果失败，修复所有编译错误后再继续。

提交:
```bash
git add apps/yggdrasil/services/audit/
git commit -m "feat(yggdrasil): audit service — first full build"
```

---

## Step 11: 集成验证

### 11a: 启动基础设施

```bash
docker compose -f infra/docker/docker-compose.yml up -d postgres redis
```

### 11b: 创建数据库

```bash
psql -h localhost -U postgres -c "CREATE DATABASE cyber_ecosystem_yggdrasil_audit;"
```

### 11c: 启动服务

```bash
./nx run yggdrasil_audit:dev
```

等待启动日志显示 3 个传输层监听。

### 11d: 测试 SubmitEvents (gRPC)

```bash
grpcurl -plaintext -d '{
  "events": [{
    "actor_id": "test_user_001",
    "actor_name": "Test User",
    "action": "create",
    "resource_type": "worklog",
    "resource_id": "wl_001",
    "service_name": "worklog",
    "result": "success",
    "ip": "127.0.0.1",
    "user_agent": "test-agent"
  }]
}' localhost:12001 api.yggdrasil.v1.AuditSinkService/SubmitEvents
```

预期: `{"accepted": 1}`

### 11e: 测试 QueryAuditLog (gRPC)

```bash
grpcurl -plaintext -d '{}' localhost:12001 api.yggdrasil.v1.AuditLogService/QueryAuditLog
```

预期: 返回包含 1 条事件的列表

### 11f: 测试 GetAuditLog (HTTP)

```bash
curl http://localhost:11001/audit/log/<id_from_step_11e>
```

预期: JSON 格式的完整审计日志

### 11g: 测试 ConnectRPC

```bash
curl -X POST http://localhost:13001/api.yggdrasil.v1.AuditSinkService/SubmitEvents \
  -H "Content-Type: application/json" \
  -d '{
    "events": [{
      "actor_id": "connect_user",
      "action": "login",
      "resource_type": "session",
      "service_name": "iam",
      "result": "success"
    }]
  }'
```

预期: `{"accepted": 1}`

### 11h: 测试 Ops 端点

```bash
curl http://localhost:14001/metrics
```

预期: Prometheus 指标输出

### 11i: 停止服务并提交

```bash
git add apps/yggdrasil/
git commit -m "feat(yggdrasil): audit service passes integration verification"
```

---

## 完成标准

- [x] `./nx run yggdrasil_audit:build` 编译通过
- [x] 服务可启动，3 个传输层正常监听
- [x] gRPC SubmitEvents 提交事件成功
- [x] gRPC/HTTP QueryAuditLog 查询事件成功
- [x] ConnectRPC SubmitEvents 提交事件成功
- [x] Ops metrics 端点可访问
- [x] 变更已提交
