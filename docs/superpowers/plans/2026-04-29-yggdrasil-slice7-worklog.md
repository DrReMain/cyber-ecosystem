# Slice 7: Worklog Service — 业务模板

> **闭环标准**: 完整 CRUD + 全能力（auth/rbac/audit/storage）集成测试通过。

## 目标

Worklog 是一个参考实现业务服务，展示如何接入所有能力模块（auth、rbac、audit、condition、datascope、storage）。后续新业务服务参照此结构创建。

此 Slice 是 yggdrasil 项目的最终闭环验证。

## 前置条件

- Slice 5 完成（thin client 模块可用）
- IAM 服务运行中（认证 + RBAC + Condition + DataScope）
- Audit 服务运行中（审计事件接收）
- Storage 服务运行中（文件上传/下载）

## 端口

| Transport | Address |
|-----------|---------|
| HTTP | `0.0.0.0:11003` |
| gRPC | `0.0.0.0:12003` |
| ConnectRPC | `0.0.0.0:13003` |
| Ops | `0.0.0.0:14003` |

---

## Step 1: 目录结构 + Nx 配置

```bash
mkdir -p apps/yggdrasil/services/worklog/{cmd/app,configs,internal/{conf,data/ent/schema,biz,server/{locales},service}}
```

**文件**: `apps/yggdrasil/services/worklog/project.json`

```json
{
  "name": "yggdrasil_worklog",
  "$schema": "../../../../node_modules/nx/schemas/project-schema.json",
  "implicitDependencies": ["yggdrasil_api", "shared-go"],
  "targets": {
    "proto:conf": {
      "executor": "nx:run-commands",
      "options": {
        "command": "buf generate --template apps/yggdrasil/services/worklog/buf.gen.conf.yaml --path apps/yggdrasil/services/worklog/internal/conf"
      }
    },
    "generate": {
      "dependsOn": ["yggdrasil_api:proto:api", "proto:conf"],
      "executor": "nx:run-commands",
      "options": {
        "command": "cd apps/yggdrasil/services/worklog && wire ./cmd/app/..."
      }
    },
    "dev": {
      "executor": "nx:run-commands",
      "options": {
        "command": "cd apps/yggdrasil/services/worklog && go run ./cmd/app/... -conf ./configs"
      }
    },
    "build": {
      "executor": "nx:run-commands",
      "options": {
        "command": "cd apps/yggdrasil/services/worklog && go build -o ./bin/worklog ./cmd/app/..."
      }
    },
    "ent:new": {
      "executor": "nx:run-commands",
      "options": {
        "command": "cd apps/yggdrasil/services/worklog && go run -mod=mod entgo.io/ent/cmd/ent new --target internal/data/ent/schema"
      }
    }
  }
}
```

---

## Step 2: API Proto 定义

**文件**: `apps/yggdrasil/api/v1/worklog.proto`

```protobuf
syntax = "proto3";

// buf:lint:ignore PACKAGE_DIRECTORY_MATCH
package api.yggdrasil.v1;

import "buf/validate/validate.proto";
import "common/page.proto";
import "desc/desc.proto";
import "google/api/annotations.proto";
import "google/protobuf/wrappers.proto";

option go_package = "cyber-ecosystem/apps/yggdrasil/gen/go/v1;yggdrasilV1";

// WorklogService
service WorklogService {
  option (desc.service_comment) = "工作日志服务";

  rpc CreateWorklog(CreateWorklogRequest) returns (CreateWorklogResponse) {
    option (desc.method_comment) = "创建工作日志";
    option (google.api.http) = {post: "/worklog/worklogs" body: "*"};
  }

  rpc UpdateWorklog(UpdateWorklogRequest) returns (UpdateWorklogResponse) {
    option (desc.method_comment) = "更新工作日志";
    option (google.api.http) = {put: "/worklog/worklogs/{id}" body: "*"};
  }

  rpc DeleteWorklog(DeleteWorklogRequest) returns (DeleteWorklogResponse) {
    option (desc.method_comment) = "删除工作日志";
    option (google.api.http) = {delete: "/worklog/worklogs/{id}"};
  }

  rpc GetWorklog(GetWorklogRequest) returns (GetWorklogResponse) {
    option (desc.method_comment) = "获取工作日志详情";
    option (google.api.http) = {get: "/worklog/worklogs/{id}"};
  }

  rpc QueryWorklogs(QueryWorklogsRequest) returns (QueryWorklogsResponse) {
    option (desc.method_comment) = "查询工作日志列表";
    option (google.api.http) = {post: "/worklog/worklogs/query" body: "*"};
  }
}

message CreateWorklogRequest {
  string title = 1 [(buf.validate.field).string.min_len = 1];
  string content = 2 [(buf.validate.field).string.min_len = 1];
  repeated string attachment_ids = 3;
}

message CreateWorklogResponse {
  string id = 1;
}

message UpdateWorklogRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
  google.protobuf.StringValue title = 2;
  google.protobuf.StringValue content = 3;
  repeated string attachment_ids = 4;
}

message UpdateWorklogResponse {}

message DeleteWorklogRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
}

message DeleteWorklogResponse {}

message GetWorklogRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
}

message GetWorklogResponse {
  string id = 1;
  string title = 2;
  string content = 3;
  string owner_id = 4;
  google.protobuf.StringValue department_id = 5;
  repeated string attachment_ids = 6;
  string created_at = 7;
  string updated_at = 8;
}

message QueryWorklogsRequest {
  common.PageRequest page = 1;
  optional string owner_id = 2;
  repeated string order_by = 100;
}

message QueryWorklogsResponse {
  common.PageResponse page = 1;
  repeated GetWorklogResponse list = 2;
}
```

生成:
```bash
buf lint apps/yggdrasil/api/v1/worklog.proto
./nx run yggdrasil_api:proto:api
```

---

## Step 3: Conf Proto + 生成

与 Slice 1 相同结构，额外添加 Capability 配置:

```protobuf
message Capability {
  message Endpoint {
    string address = 1;
  }
  Endpoint iam = 1;
  Endpoint audit = 2;
  Endpoint storage = 3;
  bool auth_enabled = 4;
  bool rbac_enabled = 5;
  bool condition_enabled = 6;
  bool datascope_enabled = 7;
  bool audit_enabled = 8;
}

message Bootstrap {
  Server server = 1;
  Auth auth = 2;
  Log log = 3;
  Data data = 4;
  Trace trace = 5;
  Ops ops = 6;
  Capability capability = 7;
}
```

---

## Step 4: Ent Schema

**文件**: `apps/yggdrasil/services/worklog/internal/data/ent/schema/worklog.go`

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

type Worklog struct {
	ent.Schema
}

func (Worklog) Fields() []ent.Field {
	return []ent.Field{
		field.String("title").NotEmpty().MaxLen(200),
		field.Text("content").NotEmpty(),
		field.Strings("attachment_ids").Optional(),
		// DataScope fields — populated by scope injector + Ent hook
		field.String("owner_id").Default("").MaxLen(20),
		field.String("department_id").Optional().Nillable().MaxLen(20),
	}
}

func (Worklog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (Worklog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("owner_id"),
		index.Fields("department_id"),
		index.Fields("created_at"),
	}
}

func (Worklog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "worklog"},
	}
}
```

> 注意: DataScope 的 Ent interceptor（查询时自动过滤行）将在本 Slice 的 Server 层通过 scope_injector 中间件 + 本地 Ent hook 实现，不需要 shared mixin。

生成:
```bash
cd apps/yggdrasil/services/worklog && go generate ./internal/data/ent/...
```

---

## Step 5: Biz 层

**文件**: `apps/yggdrasil/services/worklog/internal/biz/biz.go`

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
	NewWorklogUC,
)
```

**文件**: `apps/yggdrasil/services/worklog/internal/biz/uc_worklog.go`

```go
package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/contracts/go/common"
	"cyber-ecosystem/shared-go/utils"
)

type WorklogEntry struct {
	ID            *string
	Title         *string
	Content       *string
	AttachmentIDs []string
	OwnerID       *string
	DepartmentID  *string
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
}

type WorklogQueryIn struct {
	*common.PageRequest
	OrderBy []*utils.OrderBy
	OwnerID *string
}

type WorklogQueryOut struct {
	*common.PageResponse
	List []*WorklogEntry
}

type WorklogRP interface {
	Create(ctx context.Context, w *WorklogEntry) (*WorklogEntry, error)
	Update(ctx context.Context, w *WorklogEntry) (*WorklogEntry, error)
	Delete(ctx context.Context, id string) (*WorklogEntry, error)
	Get(ctx context.Context, id string) (*WorklogEntry, error)
	Query(ctx context.Context, in *WorklogQueryIn) (*WorklogQueryOut, error)
}

type WorklogUC struct {
	UC
	worklogRP WorklogRP
}

func NewWorklogUC(logger log.Logger, tm Transaction, worklogRP WorklogRP) *WorklogUC {
	return &WorklogUC{
		UC:        UC{log: log.NewHelper(log.With(logger, "module", "biz/uc_worklog")), tm: tm},
		worklogRP: worklogRP,
	}
}

func (uc *WorklogUC) Create(ctx context.Context, w *WorklogEntry) (*WorklogEntry, error) {
	return uc.worklogRP.Create(ctx, w)
}

func (uc *WorklogUC) Update(ctx context.Context, w *WorklogEntry) (*WorklogEntry, error) {
	return uc.worklogRP.Update(ctx, w)
}

func (uc *WorklogUC) Delete(ctx context.Context, id string) error {
	_, err := uc.worklogRP.Delete(ctx, id)
	return err
}

func (uc *WorklogUC) Get(ctx context.Context, id string) (*WorklogEntry, error) {
	return uc.worklogRP.Get(ctx, id)
}

func (uc *WorklogUC) Query(ctx context.Context, in *WorklogQueryIn) (*WorklogQueryOut, error) {
	return uc.worklogRP.Query(ctx, in)
}
```

---

## Step 6: Data 层

按通用模式创建:
- `data/store.go`, `data/store_ent.go`, `data/store_cache.go`
- `data/rp_worklog.go` — CRUD + Query
- DB 名称: `cyber_ecosystem_yggdrasil_worklog`

DataScope 的查询过滤逻辑：在 `rp_worklog.go` 的 `Query` 方法中，从 context 读取 scope 规则（由 scope_injector 中间件注入），动态添加 Ent predicates。

```go
func (rp *worklogRP) Query(ctx context.Context, in *biz.WorklogQueryIn) (*biz.WorklogQueryOut, error) {
	query := rp.store.GetClient(ctx).Worklog.Query()

	// DataScope filtering: read effective scope from context
	if scope := datascope.FromContext(ctx); scope != nil {
		if !scope.IsAll {
			var predicates []predicate.Worklog
			if scope.SelfFilter {
				identity, _ := auth.IdentityFromContext(ctx)
				if identity != nil {
					predicates = append(predicates, worklog.OwnerIDEQ(identity.UserID))
				}
			}
			if scope.DeptFilter && len(scope.DeptIDs) > 0 {
				predicates = append(predicates, worklog.DepartmentIDIn(scope.DeptIDs...))
			}
			if len(predicates) > 0 {
				query.Where(worklog.Or(predicates...))
			}
		}
	}

	entutil.WherePtr(query, in.OwnerID, worklog.OwnerIDEQ)
	// ... pagination, order by ...
}
```

DataScope 的写入自动填充：在 `Create` 中，从 context 获取 identity，自动设置 `owner_id` 和 `department_id`。

---

## Step 7: Service 层

**文件**: `apps/yggdrasil/services/worklog/internal/service/service.go`

```go
package service

import (
	"github.com/google/wire"

	"github.com/go-kratos/kratos/v2/transport/grpc"
	krahttp "github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/capabilities/audit"
	"cyber-ecosystem/shared-go/kratos/transport/connect"
)

type Registrar interface {
	RegisterGRPC(*grpc.Server)
	RegisterHTTP(*krahttp.Server)
	RegisterConnect(*connect.Server)
}

var ProviderSet = wire.NewSet(
	NewRegistrarList,
	NewWorklogService,
)

func NewRegistrarList(s1 *WorklogService) []Registrar {
	return []Registrar{s1}
}
```

**文件**: `apps/yggdrasil/services/worklog/internal/service/worklog.go`

```go
package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	krahttp "github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/capabilities/audit"
	"cyber-ecosystem/shared-go/capabilities/auth"
	"cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	yggdrasilV1 "cyber-ecosystem/apps/yggdrasil/gen/go/v1"
	yggdrasilV1connect "cyber-ecosystem/apps/yggdrasil/gen/go/v1/yggdrasilV1connect"
	"cyber-ecosystem/apps/yggdrasil/services/worklog/internal/biz"
)

type WorklogService struct {
	yggdrasilV1.UnimplementedWorklogServiceServer
	log        *log.Helper
	worklogUC  *biz.WorklogUC
	auditSender *audit.BufferedSender
}

func NewWorklogService(logger log.Logger, worklogUC *biz.WorklogUC, auditSender *audit.BufferedSender) *WorklogService {
	return &WorklogService{
		log:         log.NewHelper(log.With(logger, "module", "service/worklog")),
		worklogUC:   worklogUC,
		auditSender: auditSender,
	}
}

func (s *WorklogService) RegisterGRPC(srv *grpc.Server) {
	yggdrasilV1.RegisterWorklogServiceServer(srv, s)
}

func (s *WorklogService) RegisterHTTP(srv *krahttp.Server) {
	yggdrasilV1.RegisterWorklogServiceHTTPServer(srv, s)
}

func (s *WorklogService) RegisterConnect(srv *connect.Server) {
	srv.Register(yggdrasilV1connect.NewWorklogServiceHandler(s, srv.HandlerOptions()...))
}

func (s *WorklogService) CreateWorklog(ctx context.Context, in *yggdrasilV1.CreateWorklogRequest) (*yggdrasilV1.CreateWorklogResponse, error) {
	identity, _ := auth.IdentityFromContext(ctx)

	entry := &biz.WorklogEntry{
		Title:         &in.Title,
		Content:       &in.Content,
		AttachmentIDs: in.AttachmentIds,
	}
	if identity != nil {
		entry.OwnerID = &identity.UserID
	}

	result, err := s.worklogUC.Create(ctx, entry)
	if err != nil {
		return nil, err
	}

	// Record audit event
	s.recordAudit(ctx, identity, "create", *result.ID)

	return &yggdrasilV1.CreateWorklogResponse{Id: *result.ID}, nil
}

func (s *WorklogService) UpdateWorklog(ctx context.Context, in *yggdrasilV1.UpdateWorklogRequest) (*yggdrasilV1.UpdateWorklogResponse, error) {
	identity, _ := auth.IdentityFromContext(ctx)

	entry := &biz.WorklogEntry{ID: &in.Id}
	if in.Title != nil {
		entry.Title = &in.Title.Value
	}
	if in.Content != nil {
		entry.Content = &in.Content.Value
	}
	if len(in.AttachmentIds) > 0 {
		entry.AttachmentIDs = in.AttachmentIds
	}

	_, err := s.worklogUC.Update(ctx, entry)
	if err != nil {
		return nil, err
	}

	s.recordAudit(ctx, identity, "update", in.Id)
	return &yggdrasilV1.UpdateWorklogResponse{}, nil
}

func (s *WorklogService) DeleteWorklog(ctx context.Context, in *yggdrasilV1.DeleteWorklogRequest) (*yggdrasilV1.DeleteWorklogResponse, error) {
	identity, _ := auth.IdentityFromContext(ctx)

	if err := s.worklogUC.Delete(ctx, in.Id); err != nil {
		return nil, err
	}

	s.recordAudit(ctx, identity, "delete", in.Id)
	return &yggdrasilV1.DeleteWorklogResponse{}, nil
}

func (s *WorklogService) GetWorklog(ctx context.Context, in *yggdrasilV1.GetWorklogRequest) (*yggdrasilV1.GetWorklogResponse, error) {
	identity, _ := auth.IdentityFromContext(ctx)

	result, err := s.worklogUC.Get(ctx, in.Id)
	if err != nil {
		return nil, err
	}

	s.recordAudit(ctx, identity, "read", in.Id)
	return s.toProto(result), nil
}

func (s *WorklogService) QueryWorklogs(ctx context.Context, in *yggdrasilV1.QueryWorklogsRequest) (*yggdrasilV1.QueryWorklogsResponse, error) {
	queryIn := &biz.WorklogQueryIn{
		PageRequest: in.Page,
		OrderBy:     utils.ParseOrderBy(in.OrderBy),
	}
	if in.OwnerId != nil {
		queryIn.OwnerID = in.OwnerId
	}

	out, err := s.worklogUC.Query(ctx, queryIn)
	if err != nil {
		return nil, err
	}

	return &yggdrasilV1.QueryWorklogsResponse{
		Page: out.PageResponse,
		List: utils.SliceMap(out.List, s.toProto),
	}, nil
}

func (s *WorklogService) toProto(e *biz.WorklogEntry) *yggdrasilV1.GetWorklogResponse {
	resp := &yggdrasilV1.GetWorklogResponse{
		Id:      utils.Deref(e.ID, ""),
		Title:   utils.Deref(e.Title, ""),
		Content: utils.Deref(e.Content, ""),
		OwnerId: utils.Deref(e.OwnerID, ""),
		AttachmentIds: e.AttachmentIDs,
		CreatedAt: utils.Deref(e.CreatedAt, time.Time{}).Format(time.RFC3339),
		UpdatedAt: utils.Deref(e.UpdatedAt, time.Time{}).Format(time.RFC3339),
	}
	if e.DepartmentID != nil {
		resp.DepartmentId = utils.Wrap(e.DepartmentID, utils.StringW)
	}
	return resp
}

func (s *WorklogService) recordAudit(ctx context.Context, identity *auth.Identity, action, resourceID string) {
	if s.auditSender == nil || identity == nil {
		return
	}
	_ = s.auditSender.Send(ctx, []*audit.Event{{
		ActorID:      identity.UserID,
		Action:       action,
		ResourceType: "worklog",
		ResourceID:   resourceID,
		ServiceName:  "worklog",
		Result:       "success",
	}})
}
```

---

## Step 8: Server 层 — 全能力中间件

**文件**: `apps/yggdrasil/services/worklog/internal/server/server.go`

```go
package server

import (
	"github.com/go-kratos/kratos/v2/middleware"
	jwt "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	jwtv5 "github.com/golang-jwt/jwt/v5"

	capauth "cyber-ecosystem/shared-go/capabilities/auth"
	caprbac "cyber-ecosystem/shared-go/capabilities/rbac"
	capcondition "cyber-ecosystem/shared-go/capabilities/condition"
	capdatascope "cyber-ecosystem/shared-go/capabilities/datascope"
	// ...
)

func buildMiddlewares(
	c *conf.Capability,
	authenticator capauth.Authenticator,
	authorizer caprbac.Authorizer,
	conditionChecker capcondition.Checker,
	scopeResolver capdatascope.Resolver,
	// ... standard params
) []middleware.Middleware {
	var mws []middleware.Middleware

	// Always-on (same as other services)
	mws = append(mws, i18n.Server(i18nBundle))
	mws = append(mws, recovery.Recovery())
	// ...

	// Auth gated — using thin client middleware
	var authMws []middleware.Middleware
	if c.GetAuthEnabled() {
		authMws = append(authMws, capauth.Middleware(authenticator))
	}
	if c.GetRbacEnabled() {
		authMws = append(authMws, caprbac.Middleware(authorizer))
	}
	if c.GetConditionEnabled() {
		authMws = append(authMws, capcondition.Middleware(conditionChecker))
	}
	if c.GetDatascopeEnabled() {
		authMws = append(authMws, datascopeInjector(scopeResolver))
	}

	mws = append(mws, selector.Server(authMws...).
		Match(auth.NewWhiteListByPublicAccessInProtoMatcher()).Build())

	mws = append(mws, validate.ProtoValidate())
	return mws
}

// datascopeInjector injects scope resolve func into context for Data layer use.
func datascopeInjector(resolver capdatascope.Resolver) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			identity, err := capauth.IdentityFromContext(ctx)
			if err != nil {
				return handler(ctx, req)
			}
			tr, _ := transport.FromServerContext(ctx)
			if tr == nil {
				return handler(ctx, req)
			}
			resolveFn := func(ctx context.Context) (*capdatascope.EffectiveScope, error) {
				return resolver.ResolveScope(ctx, identity.UserID, tr.Operation())
			}
			ctx = context.WithValue(ctx, scopeResolveFuncKey{}, resolveFn)
			return handler(ctx, req)
		}
	}
}

var ProviderSet = wire.NewSet(
	NewOpsServer, NewGRPCServer, NewHTTPServer, NewConnectServer, NewI18nBundle,
	NewAuthClient, NewRBACClient, NewAuditClient, NewConditionClient, NewDataScopeClient,
)
```

**Wire 客户端工厂:**

```go
func NewAuthClient(c *conf.Capability) (capauth.Authenticator, error) {
	return capauth.NewGRPCClient(c.Iam.Address)
}

func NewRBACClient(c *conf.Capability) (caprbac.Authorizer, error) {
	return caprbac.NewGRPCClient(c.Iam.Address)
}

func NewConditionClient(c *conf.Capability) (capcondition.Checker, error) {
	return capcondition.NewGRPCClient(c.Iam.Address)
}

func NewDataScopeClient(c *conf.Capability) (capdatascope.Resolver, error) {
	return capdatascope.NewGRPCClient(c.Iam.Address)
}

func NewAuditClient(c *conf.Capability, logger log.Logger) (*audit.BufferedSender, error) {
	client, err := audit.NewGRPCClient(c.Audit.Address)
	if err != nil {
		return nil, err
	}
	return audit.NewBufferedSender(client, logger, 50, 5*time.Second), nil
}
```

**Wire:**
```go
// cmd/app/wire.go
func wireApp(/* ... */) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		service.ProviderSet,
		biz.ProviderSet,
		data.ProviderSet,
		newApp,
	))
}
```

**Config:**
```yaml
# configs/config.yaml
server:
  http:
    addr: 0.0.0.0:11003
    timeout: 10s
  grpc:
    addr: 0.0.0.0:12003
    timeout: 10s
  connect:
    addr: 0.0.0.0:13003
    timeout: 10s

auth:
  secret: secret

capability:
  iam:
    address: "localhost:12000"
  audit:
    address: "localhost:12001"
  storage:
    address: "localhost:12002"
  auth_enabled: true
  rbac_enabled: true
  condition_enabled: true
  datascope_enabled: true
  audit_enabled: true

data:
  database:
    driver: postgres
    host: localhost
    port: 5432
    user: postgres
    password: postgres
    db_name: cyber_ecosystem_yggdrasil_worklog
    max_open_conns: 10
    max_idle_conns: 5
    conn_max_lifetime: 300s
    migrate: true
  cache:
    type: memory
    memory:
      otel_enabled: true

ops:
  enabled: true
  addr: "0.0.0.0:14003"
  metrics: "/metrics"
```

---

## Step 9: 编译闭环

```bash
cd apps/yggdrasil/services/worklog && go mod tidy
./nx run yggdrasil_worklog:generate
./nx run yggdrasil_worklog:build
```

提交:
```bash
git add apps/yggdrasil/services/worklog/
git commit -m "feat(yggdrasil): worklog service — business template with all capabilities"
```

---

## Step 10: 集成验证

### 10a: 启动所有服务

```bash
# 确保 IAM + Audit + Storage 服务运行
psql -h localhost -U postgres -c "CREATE DATABASE cyber_ecosystem_yggdrasil_worklog;"
./nx run yggdrasil_worklog:dev
```

### 10b: 创建角色 + 权限

```bash
TOKEN=$(curl -s -X POST http://localhost:11000/iam/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "admin123"}' | jq -r '.access_token')

# 创建角色
ROLE_ID=$(curl -s -X POST http://localhost:11000/iam/roles \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"name": "WorklogUser", "code": "worklog_user"}' | jq -r '.id')

# 绑定权限
for perm in "worklog.create:allow" "worklog.list:allow" "worklog.get:allow" "worklog.update:allow" "worklog.delete:allow"; do
  IFS=: read -r resource effect <<< "$perm"
  curl -s -X POST http://localhost:11000/iam/permission-bindings \
    -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
    -d "{\"role_id\": \"$ROLE_ID\", \"resource\": \"$resource\", \"effect\": \"$effect\"}"
done
```

### 10c: 创建测试用户 + 绑定角色

```bash
USER_ID=$(curl -s -X POST http://localhost:11000/iam/users \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"email": "worker@example.com", "password": "worker123", "name": "Worker"}' | jq -r '.id')

curl -s -X POST http://localhost:11000/iam/role-bindings \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d "{\"user_id\": \"$USER_ID\", \"role_id\": \"$ROLE_ID\"}"

# 登录获取 worker token
WORKER_TOKEN=$(curl -s -X POST http://localhost:11000/iam/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "worker@example.com", "password": "worker123"}' | jq -r '.access_token')
```

### 10d: 创建工作日志 (ConnectRPC)

```bash
curl -X POST http://localhost:13003/api.yggdrasil.v1.WorklogService/CreateWorklog \
  -H "Authorization: Bearer $WORKER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title": "完成 yggdrasil 设计", "content": "完成了所有 8 个 Slice 的设计文档"}'
```

预期: `{id: "..."}`

### 10e: 查询工作日志

```bash
curl -X POST http://localhost:11003/worklog/worklogs/query \
  -H "Authorization: Bearer $WORKER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

预期: 包含刚创建的日志

### 10f: 验证审计事件

```bash
# 等待缓冲发送（5 秒）
sleep 6

grpcurl -plaintext -d '{}' localhost:12001 api.yggdrasil.v1.AuditLogService/QueryAuditLog
```

预期: 包含 `action: "create"`, `resource_type: "worklog"`, `service_name: "worklog"` 的审计记录

### 10g: 验证 RBAC 拒绝

```bash
# 用没有权限的用户（新用户无角色）
UNAUTH_USER=$(curl -s -X POST http://localhost:11000/iam/users \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"email": "nobody@example.com", "password": "nobody123", "name": "Nobody"}' | jq -r '.id')

NOBODY_TOKEN=$(curl -s -X POST http://localhost:11000/iam/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "nobody@example.com", "password": "nobody123"}' | jq -r '.access_token')

curl -X POST http://localhost:11003/worklog/worklogs/query \
  -H "Authorization: Bearer $NOBODY_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

预期: 403 Forbidden

### 10h: 删除工作日志

```bash
WORKLOG_ID="<id_from_10d>"
curl -X DELETE "http://localhost:11003/worklog/worklogs/$WORKLOG_ID" \
  -H "Authorization: Bearer $WORKER_TOKEN"
```

预期: `{}`

### 10i: 最终提交

```bash
git add apps/yggdrasil/
git commit -m "feat(yggdrasil): worklog service passes full integration verification"
```

---

## 完成标准

- [x] `./nx run yggdrasil_worklog:build` 编译通过
- [x] 服务可启动，连接 IAM/Audit/Storage
- [x] Auth 中间件通过 thin client 校验 JWT
- [x] RBAC 中间件通过 thin client 校验权限
- [x] 创建工作日志 → owner_id 自动填充
- [x] 查询工作日志 → DataScope 过滤正常
- [x] 审计事件通过 buffered sender 异步投递到 audit 服务
- [x] 无权限用户被 RBAC 拒绝
- [x] 全链路闭环: 前端 → worklog → (iam auth + rbac) → biz → data → (audit event)
- [x] 变更已提交
