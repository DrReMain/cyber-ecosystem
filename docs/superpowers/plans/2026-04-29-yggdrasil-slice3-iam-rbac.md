# Slice 3: IAM RBAC + Event Bus

> **闭环标准**: 删除角色 → 事件级联清理关联策略成功。

## 目标

在 Slice 2 的 IAM 认证基础上，构建 RBAC 能力和 Event Bus 系统：

1. **Event Bus** — in-process 同步事件总线，替代 singleton 的 RoleCascadeRegistry 和 ScopeCacheInvalidator
2. **RBAC** — 角色、角色绑定、权限绑定、Casbin 策略管理
3. **资源发现** — 服务/方法自省
4. **Authorizer 中间件** — 请求级鉴权决策

核心改进：所有 UC 间协调通过事件订阅，UC 不直接引用其他 UC。

## 前置条件

- Slice 2 完成（IAM 认证 + JWT + Session + 超级管理员种子）
- Audit 服务运行中（Slice 1）

---

## Step 1: Event Bus 实现

**文件**: `apps/yggdrasil/services/iam/internal/event/bus.go`

```go
package event

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-kratos/kratos/v2/log"
)

// Topic identifies an event type.
type Topic string

const (
	TopicRoleDeleted          Topic = "role.deleted"
	TopicRoleBindingCreated   Topic = "role_binding.created"
	TopicRoleBindingDeleted   Topic = "role_binding.deleted"
	TopicPermBindingCreated   Topic = "permission_binding.created"
	TopicPermBindingDeleted   Topic = "permission_binding.deleted"
	TopicDeptBindingChanged   Topic = "department_binding.changed"
	TopicUserAttributeChanged Topic = "user_attribute.changed"
	TopicUserDeleted          Topic = "user.deleted"
)

// Event carries a topic and arbitrary payload.
type Event struct {
	Topic   Topic
	Payload any
}

// Handler processes an event.
type Handler func(ctx context.Context, evt Event) error

// Bus is the in-process event bus interface.
type Bus interface {
	Subscribe(topic Topic, handler Handler)
	Publish(ctx context.Context, evt Event) error
}

// SyncBus implements Bus with synchronous, in-order handler invocation.
type SyncBus struct {
	mu       sync.RWMutex
	handlers map[Topic][]Handler
	log      *log.Helper
}

func NewSyncBus(logger log.Logger) *SyncBus {
	return &SyncBus{
		handlers: make(map[Topic][]Handler),
		log:      log.NewHelper(log.With(logger, "module", "event/bus")),
	}
}

func (b *SyncBus) Subscribe(topic Topic, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[topic] = append(b.handlers[topic], handler)
}

func (b *SyncBus) Publish(ctx context.Context, evt Event) error {
	b.mu.RLock()
	handlers := b.handlers[evt.Topic]
	b.mu.RUnlock()

	for _, h := range handlers {
		if err := h(ctx, evt); err != nil {
			b.log.Errorf("event handler failed: topic=%s error=%v", evt.Topic, err)
			// Continue processing other handlers
		}
	}
	return nil
}
```

**文件**: `apps/yggdrasil/services/iam/internal/event/event.go`

```go
package event

import "github.com/google/wire"

var ProviderSet = wire.NewSet(NewSyncBus)
```

---

## Step 2: API Proto 定义

### 2a: iam_role.proto

**文件**: `apps/yggdrasil/api/v1/iam_role.proto`

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

// RoleService
service RoleService {
  option (desc.service_comment) = "角色管理服务";

  rpc CreateRole(CreateRoleRequest) returns (CreateRoleResponse) {
    option (desc.method_comment) = "创建角色";
    option (google.api.http) = {post: "/iam/roles" body: "*"};
  }
  rpc UpdateRole(UpdateRoleRequest) returns (UpdateRoleResponse) {
    option (desc.method_comment) = "更新角色";
    option (google.api.http) = {put: "/iam/roles/{id}" body: "*"};
  }
  rpc DeleteRole(DeleteRoleRequest) returns (DeleteRoleResponse) {
    option (desc.method_comment) = "删除角色";
    option (google.api.http) = {delete: "/iam/roles/{id}"};
  }
  rpc GetRole(GetRoleRequest) returns (GetRoleResponse) {
    option (desc.method_comment) = "获取角色详情";
    option (google.api.http) = {get: "/iam/roles/{id}"};
  }
  rpc QueryRole(QueryRoleRequest) returns (QueryRoleResponse) {
    option (desc.method_comment) = "查询角色列表";
    option (google.api.http) = {post: "/iam/roles/query" body: "*"};
  }
}

message CreateRoleRequest {
  string name = 1 [(buf.validate.field).string.min_len = 1];
  string code = 2 [(buf.validate.field).string.min_len = 1];
  int32 sort = 3;
  string description = 4;
}
message CreateRoleResponse { string id = 1; }

message UpdateRoleRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
  google.protobuf.StringValue name = 2;
  google.protobuf.Int32Value sort = 3;
  google.protobuf.StringValue description = 4;
}
message UpdateRoleResponse {}

message DeleteRoleRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
}
message DeleteRoleResponse {}

message GetRoleRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
}
message GetRoleResponse {
  string id = 1;
  string name = 2;
  string code = 3;
  int32 sort = 4;
  string description = 5;
  string created_at = 6;
  string updated_at = 7;
}

message QueryRoleRequest {
  common.PageRequest page = 1;
  repeated string order_by = 100;
}
message QueryRoleResponse {
  common.PageResponse page = 1;
  repeated GetRoleResponse list = 2;
}
```

### 2b: iam_role_binding.proto

**文件**: `apps/yggdrasil/api/v1/iam_role_binding.proto`

```protobuf
syntax = "proto3";

// buf:lint:ignore PACKAGE_DIRECTORY_MATCH
package api.yggdrasil.v1;

import "buf/validate/validate.proto";
import "common/page.proto";
import "desc/desc.proto";
import "google/api/annotations.proto";

option go_package = "cyber-ecosystem/apps/yggdrasil/gen/go/v1;yggdrasilV1";

// RoleBindingService
service RoleBindingService {
  option (desc.service_comment) = "角色绑定服务";

  rpc CreateRoleBinding(CreateRoleBindingRequest) returns (CreateRoleBindingResponse) {
    option (desc.method_comment) = "创建角色绑定";
    option (google.api.http) = {post: "/iam/role-bindings" body: "*"};
  }
  rpc DeleteRoleBinding(DeleteRoleBindingRequest) returns (DeleteRoleBindingResponse) {
    option (desc.method_comment) = "删除角色绑定";
    option (google.api.http) = {delete: "/iam/role-bindings/{id}"};
  }
  rpc QueryRoleBinding(QueryRoleBindingRequest) returns (QueryRoleBindingResponse) {
    option (desc.method_comment) = "查询角色绑定";
    option (google.api.http) = {post: "/iam/role-bindings/query" body: "*"};
  }
}

message CreateRoleBindingRequest {
  string user_id = 1 [(buf.validate.field).string.min_len = 1];
  string role_id = 2 [(buf.validate.field).string.min_len = 1];
}
message CreateRoleBindingResponse { string id = 1; }

message DeleteRoleBindingRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
}
message DeleteRoleBindingResponse {}

message QueryRoleBindingRequest {
  common.PageRequest page = 1;
  optional string user_id = 2;
  optional string role_id = 3;
}
message QueryRoleBindingResponse {
  common.PageResponse page = 1;
  repeated RoleBindingItem list = 2;
}

message RoleBindingItem {
  string id = 1;
  string user_id = 2;
  string role_id = 3;
  string role_code = 4;
  string role_name = 5;
}
```

### 2c: iam_permission_binding.proto

**文件**: `apps/yggdrasil/api/v1/iam_permission_binding.proto`

```protobuf
syntax = "proto3";

// buf:lint:ignore PACKAGE_DIRECTORY_MATCH
package api.yggdrasil.v1;

import "buf/validate/validate.proto";
import "common/page.proto";
import "desc/desc.proto";
import "google/api/annotations.proto";

option go_package = "cyber-ecosystem/apps/yggdrasil/gen/go/v1;yggdrasilV1";

// PermissionBindingService
service PermissionBindingService {
  option (desc.service_comment) = "权限绑定服务";

  rpc CreatePermissionBinding(CreatePermissionBindingRequest) returns (CreatePermissionBindingResponse) {
    option (desc.method_comment) = "创建权限绑定";
    option (google.api.http) = {post: "/iam/permission-bindings" body: "*"};
  }
  rpc DeletePermissionBinding(DeletePermissionBindingRequest) returns (DeletePermissionBindingResponse) {
    option (desc.method_comment) = "删除权限绑定";
    option (google.api.http) = {delete: "/iam/permission-bindings/{id}"};
  }
  rpc QueryPermissionBinding(QueryPermissionBindingRequest) returns (QueryPermissionBindingResponse) {
    option (desc.method_comment) = "查询权限绑定";
    option (google.api.http) = {post: "/iam/permission-bindings/query" body: "*"};
  }
}

message CreatePermissionBindingRequest {
  string role_id = 1 [(buf.validate.field).string.min_len = 1];
  string resource = 2 [(buf.validate.field).string.min_len = 1];
  string effect = 3 [(buf.validate.field).string = {in: ["allow", "deny"]}];
}
message CreatePermissionBindingResponse { string id = 1; }

message DeletePermissionBindingRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
}
message DeletePermissionBindingResponse {}

message QueryPermissionBindingRequest {
  common.PageRequest page = 1;
  optional string role_id = 2;
  optional string resource = 3;
}
message QueryPermissionBindingResponse {
  common.PageResponse page = 1;
  repeated PermissionBindingItem list = 2;
}

message PermissionBindingItem {
  string id = 1;
  string role_id = 2;
  string role_code = 3;
  string resource = 4;
  string effect = 5;
}
```

### 2d: iam_resource.proto

**文件**: `apps/yggdrasil/api/v1/iam_resource.proto`

```protobuf
syntax = "proto3";

// buf:lint:ignore PACKAGE_DIRECTORY_MATCH
package api.yggdrasil.v1;

import "desc/desc.proto";
import "google/api/annotations.proto";

option go_package = "cyber-ecosystem/apps/yggdrasil/gen/go/v1;yggdrasilV1";

// ResourceService
service ResourceService {
  option (desc.service_comment) = "资源发现服务";

  rpc ListResources(ListResourcesRequest) returns (ListResourcesResponse) {
    option (desc.method_comment) = "列出所有已注册的服务和方法";
    option (google.api.http) = {get: "/iam/resources"};
  }
}

message ListResourcesRequest {}

message ResourceMethod {
  string name = 1;
  string path = 2;
  bool public_access = 3;
}

message ResourceService {
  string name = 1;
  repeated ResourceMethod methods = 2;
}

message ListResourcesResponse {
  repeated ResourceService services = 1;
}
```

### 2e: iam_internal_authorization.proto

**文件**: `apps/yggdrasil/api/v1/iam_internal_authorization.proto`

```protobuf
syntax = "proto3";

// buf:lint:ignore PACKAGE_DIRECTORY_MATCH
package api.yggdrasil.v1;

import "buf/validate/validate.proto";

option go_package = "cyber-ecosystem/apps/yggdrasil/gen/go/v1;yggdrasilV1";

// InternalAuthorizationService — 内部 API，供业务服务鉴权。
service InternalAuthorizationService {
  // CheckPermission 检查用户是否有指定资源的权限。
  rpc CheckPermission(CheckPermissionRequest) returns (CheckPermissionResponse);
  // GetRolesForUser 获取用户的所有角色。
  rpc GetRolesForUser(GetRolesForUserRequest) returns (GetRolesForUserResponse);
}

message CheckPermissionRequest {
  string user_id = 1 [(buf.validate.field).string.min_len = 1];
  string resource = 2 [(buf.validate.field).string.min_len = 1];
}

message CheckPermissionResponse {
  bool allowed = 1;
}

message GetRolesForUserRequest {
  string user_id = 1 [(buf.validate.field).string.min_len = 1];
}

message GetRolesForUserResponse {
  repeated string roles = 1;
}
```

### 2f: 生成 Proto 代码

```bash
buf lint apps/yggdrasil/api/v1/iam_role.proto
buf lint apps/yggdrasil/api/v1/iam_role_binding.proto
buf lint apps/yggdrasil/api/v1/iam_permission_binding.proto
buf lint apps/yggdrasil/api/v1/iam_resource.proto
buf lint apps/yggdrasil/api/v1/iam_internal_authorization.proto
./nx run yggdrasil_api:proto:api
```

验证: `ls apps/yggdrasil/gen/go/v1/iam_role* apps/yggdrasil/gen/go/v1/iam_permission_binding*`

提交:
```bash
git add apps/yggdrasil/api/v1/iam_*.proto apps/yggdrasil/gen/
git commit -m "feat(yggdrasil): add IAM RBAC API proto definitions"
```

---

## Step 3: Ent Schema 补充

### 3a: role_binding.go

**文件**: `apps/yggdrasil/services/iam/internal/data/ent/schema/role_binding.go`

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

type RoleBinding struct {
	ent.Schema
}

func (RoleBinding) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").NotEmpty().MaxLen(20),
		field.String("role_id").NotEmpty().MaxLen(20),
	}
}

func (RoleBinding) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (RoleBinding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "role_id").Unique(),
		index.Fields("user_id"),
		index.Fields("role_id"),
	}
}

func (RoleBinding) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "role_binding"},
	}
}
```

### 3b: permission_binding.go

**文件**: `apps/yggdrasil/services/iam/internal/data/ent/schema/permission_binding.go`

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

type PermissionBinding struct {
	ent.Schema
}

func (PermissionBinding) Fields() []ent.Field {
	return []ent.Field{
		field.String("role_id").NotEmpty().MaxLen(20),
		field.String("resource").NotEmpty().MaxLen(200),
		field.Enum("effect").Values("allow", "deny").Default("allow"),
	}
}

func (PermissionBinding) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (PermissionBinding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("role_id"),
		index.Fields("resource"),
		index.Fields("role_id", "resource", "effect").Unique(),
	}
}

func (PermissionBinding) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "permission_binding"},
	}
}
```

### 3c: 重新生成 Ent

```bash
cd apps/yggdrasil/services/iam && go generate ./internal/data/ent/...
```

---

## Step 4: Biz 层 — RBAC UCs

**文件**: `apps/yggdrasil/services/iam/internal/biz/uc_role.go`

```go
package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/event"
)

type Role struct {
	ID          *string
	Name        *string
	Code        *string
	Sort        *int
	Description *string
}

type RoleQueryIn struct {
	*common.PageRequest
	OrderBy []*utils.OrderBy
}

type RoleQueryOut struct {
	*common.PageResponse
	List []*Role
}

type RoleRP interface {
	Create(ctx context.Context, role *Role) (*Role, error)
	Update(ctx context.Context, role *Role) (*Role, error)
	Delete(ctx context.Context, id string) (*Role, error)
	Get(ctx context.Context, id string) (*Role, error)
	GetByCode(ctx context.Context, code string) (*Role, error)
	Query(ctx context.Context, in *RoleQueryIn) (*RoleQueryOut, error)
}

type RoleUC struct {
	UC
	roleRP RoleRP
	bus    *event.SyncBus
}

func NewRoleUC(logger log.Logger, tm Transaction, roleRP RoleRP, bus *event.SyncBus) *RoleUC {
	return &RoleUC{
		UC: UC{
			log: log.NewHelper(log.With(logger, "module", "biz/uc_role")),
			tm:  tm,
		},
		roleRP: roleRP,
		bus:    bus,
	}
}

func (uc *RoleUC) Create(ctx context.Context, role *Role) (*Role, error) {
	return uc.roleRP.Create(ctx, role)
}

func (uc *RoleUC) Update(ctx context.Context, role *Role) (*Role, error) {
	return uc.roleRP.Update(ctx, role)
}

func (uc *RoleUC) Delete(ctx context.Context, id string) error {
	deleted, err := uc.roleRP.Delete(ctx, id)
	if err != nil {
		return err
	}
	// Publish event — subscribers (PolicyUC) will clean up
	uc.bus.Publish(ctx, event.Event{
		Topic:   event.TopicRoleDeleted,
		Payload: deleted.Code,
	})
	return nil
}

func (uc *RoleUC) Get(ctx context.Context, id string) (*Role, error) {
	return uc.roleRP.Get(ctx, id)
}

func (uc *RoleUC) Query(ctx context.Context, in *RoleQueryIn) (*RoleQueryOut, error) {
	return uc.roleRP.Query(ctx, in)
}
```

**文件**: `apps/yggdrasil/services/iam/internal/biz/uc_role_binding.go`

```go
package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/event"
)

type RoleBinding struct {
	ID       *string
	UserID   *string
	RoleID   *string
	RoleCode *string
	RoleName *string
}

type RoleBindingQueryIn struct {
	*common.PageRequest
	UserID *string
	RoleID *string
}

type RoleBindingQueryOut struct {
	*common.PageResponse
	List []*RoleBinding
}

type RoleBindingRP interface {
	Create(ctx context.Context, rb *RoleBinding) (*RoleBinding, error)
	Delete(ctx context.Context, id string) (*RoleBinding, error)
	Query(ctx context.Context, in *RoleBindingQueryIn) (*RoleBindingQueryOut, error)
}

type RoleBindingUC struct {
	UC
	roleBindingRP RoleBindingRP
	roleRP        RoleRP
	policyRP      PolicyRP
	bus           *event.SyncBus
}

func NewRoleBindingUC(
	logger log.Logger, tm Transaction,
	roleBindingRP RoleBindingRP, roleRP RoleRP, policyRP PolicyRP,
	bus *event.SyncBus,
) *RoleBindingUC {
	return &RoleBindingUC{
		UC:            UC{log: log.NewHelper(log.With(logger, "module", "biz/uc_role_binding")), tm: tm},
		roleBindingRP: roleBindingRP,
		roleRP:        roleRP,
		policyRP:      policyRP,
		bus:           bus,
	}
}

func (uc *RoleBindingUC) Create(ctx context.Context, rb *RoleBinding) (*RoleBinding, error) {
	role, err := uc.roleRP.Get(ctx, *rb.RoleID)
	if err != nil {
		return nil, err
	}
	result, err := uc.roleBindingRP.Create(ctx, rb)
	if err != nil {
		return nil, err
	}
	// Sync Casbin grouping
	_, syncFn, err := uc.policyRP.AddRoleForUser(ctx, *rb.UserID, *role.Code)
	if err != nil {
		return nil, err
	}
	syncFn()
	// Publish event
	uc.bus.Publish(ctx, event.Event{
		Topic:   event.TopicRoleBindingCreated,
		Payload: *rb.UserID,
	})
	return result, nil
}

func (uc *RoleBindingUC) Delete(ctx context.Context, id string) error {
	deleted, err := uc.roleBindingRP.Delete(ctx, id)
	if err != nil {
		return err
	}
	// Remove Casbin grouping
	_, syncFn, err := uc.policyRP.RemoveRoleForUser(ctx, *deleted.UserID, *deleted.RoleCode)
	if err != nil {
		return nil // best-effort
	}
	syncFn()
	// Publish event
	uc.bus.Publish(ctx, event.Event{
		Topic:   event.TopicRoleBindingDeleted,
		Payload: *deleted.UserID,
	})
	return nil
}

func (uc *RoleBindingUC) Query(ctx context.Context, in *RoleBindingQueryIn) (*RoleBindingQueryOut, error) {
	return uc.roleBindingRP.Query(ctx, in)
}
```

**文件**: `apps/yggdrasil/services/iam/internal/biz/uc_permission_binding.go`

```go
package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/event"
)

type PermissionBinding struct {
	ID       *string
	RoleID   *string
	RoleCode *string
	Resource *string
	Effect   *string
}

type PermissionBindingQueryIn struct {
	*common.PageRequest
	RoleID   *string
	Resource *string
}

type PermissionBindingQueryOut struct {
	*common.PageResponse
	List []*PermissionBinding
}

type PermBindingRP interface {
	Create(ctx context.Context, pb *PermissionBinding) (*PermissionBinding, error)
	Delete(ctx context.Context, id string) (*PermissionBinding, error)
	Query(ctx context.Context, in *PermissionBindingQueryIn) (*PermissionBindingQueryOut, error)
}

type PermissionBindingUC struct {
	UC
	permBindingRP PermBindingRP
	roleRP        RoleRP
	policyRP      PolicyRP
	bus           *event.SyncBus
}

func NewPermissionBindingUC(
	logger log.Logger, tm Transaction,
	permBindingRP PermBindingRP, roleRP RoleRP, policyRP PolicyRP,
	bus *event.SyncBus,
) *PermissionBindingUC {
	return &PermissionBindingUC{
		UC:            UC{log: log.NewHelper(log.With(logger, "module", "biz/uc_perm_binding")), tm: tm},
		permBindingRP: permBindingRP,
		roleRP:        roleRP,
		policyRP:      policyRP,
		bus:           bus,
	}
}

func (uc *PermissionBindingUC) Create(ctx context.Context, pb *PermissionBinding) (*PermissionBinding, error) {
	role, err := uc.roleRP.Get(ctx, *pb.RoleID)
	if err != nil {
		return nil, err
	}
	result, err := uc.permBindingRP.Create(ctx, pb)
	if err != nil {
		return nil, err
	}
	// Sync Casbin policy
	_, syncFn, err := uc.policyRP.AddPermissionForRole(ctx, *role.Code, *pb.Resource, *pb.Effect)
	if err != nil {
		return nil, err
	}
	syncFn()
	// Publish event
	uc.bus.Publish(ctx, event.Event{
		Topic:   event.TopicPermBindingCreated,
		Payload: *pb.RoleID,
	})
	return result, nil
}

func (uc *PermissionBindingUC) Delete(ctx context.Context, id string) error {
	deleted, err := uc.permBindingRP.Delete(ctx, id)
	if err != nil {
		return err
	}
	role, _ := uc.roleRP.Get(ctx, *deleted.RoleID)
	if role != nil {
		_, syncFn, _ := uc.policyRP.RemovePermissionForRole(ctx, *role.Code, *deleted.Resource, *deleted.Effect)
		if syncFn != nil {
			syncFn()
		}
	}
	// Publish event
	uc.bus.Publish(ctx, event.Event{
		Topic:   event.TopicPermBindingDeleted,
		Payload: *deleted.RoleID,
	})
	return nil
}

func (uc *PermissionBindingUC) Query(ctx context.Context, in *PermissionBindingQueryIn) (*PermissionBindingQueryOut, error) {
	return uc.permBindingRP.Query(ctx, in)
}
```

**文件**: `apps/yggdrasil/services/iam/internal/biz/uc_policy.go`

```go
package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/event"
)

type PolicyUC struct {
	UC
	policyRP PolicyRP
	bus      *event.SyncBus
}

func NewPolicyUC(logger log.Logger, tm Transaction, policyRP PolicyRP, bus *event.SyncBus) *PolicyUC {
	uc := &PolicyUC{
		UC:       UC{log: log.NewHelper(log.With(logger, "module", "biz/uc_policy")), tm: tm},
		policyRP: policyRP,
		bus:      bus,
	}
	// Subscribe to role deletion events
	bus.Subscribe(event.TopicRoleDeleted, uc.onRoleDeleted)
	// Subscribe to binding events for policy reload
	bus.Subscribe(event.TopicRoleBindingCreated, uc.onBindingChanged)
	bus.Subscribe(event.TopicRoleBindingDeleted, uc.onBindingChanged)
	bus.Subscribe(event.TopicPermBindingCreated, uc.onBindingChanged)
	bus.Subscribe(event.TopicPermBindingDeleted, uc.onBindingChanged)
	return uc
}

func (uc *PolicyUC) Enforce(ctx context.Context, sub, obj string) (bool, error) {
	return uc.policyRP.Enforce(sub, Domain, obj)
}

func (uc *PolicyUC) GetRolesForUser(ctx context.Context, userID string) ([]string, error) {
	return uc.policyRP.GetRolesForUser(userID), nil
}

func (uc *PolicyUC) CheckPermission(ctx context.Context, userID, resource string) (bool, error) {
	return uc.Enforce(ctx, userID, resource)
}

// Event handlers

func (uc *PolicyUC) onRoleDeleted(ctx context.Context, evt event.Event) error {
	roleCode, ok := evt.Payload.(string)
	if !ok {
		return nil
	}
	// Clean up all policy rules for this role
	return uc.policyRP.RemoveRoleGroupings(ctx, roleCode)
}

func (uc *PolicyUC) onBindingChanged(ctx context.Context, evt event.Event) error {
	// Reload Casbin policy from DB
	uc.policyRP.ReloadPolicy()
	return nil
}
```

> `PolicyRP` 接口需要追加 `ReloadPolicy()` 和 `GetRolesForUser(userID string) []string`（已在 Slice 2 定义的基础上扩展）。

**文件**: `apps/yggdrasil/services/iam/internal/biz/uc_resource.go`

```go
package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type ResourceMethod struct {
	Name         string
	Path         string
	PublicAccess bool
}

type ResourceService struct {
	Name     string
	Methods  []ResourceMethod
}

type ResourceUC struct {
	UC
}

func NewResourceUC(logger log.Logger) *ResourceUC {
	return &ResourceUC{
		UC: UC{log: log.NewHelper(log.With(logger, "module", "biz/uc_resource"))},
	}
}

func (uc *ResourceUC) ListResources(ctx context.Context) ([]*ResourceService, error) {
	// Discover services/methods from proto registry
	// This is a simplified version — full implementation uses protoregistry
	return nil, nil // TODO: implement
}
```

**文件**: `apps/yggdrasil/services/iam/internal/biz/adapter_authorizer.go`

```go
package biz

import (
	"context"
	"fmt"

	pkgauth "cyber-ecosystem/apps/yggdrasil/services/iam/internal/pkg/auth"
)

// Authorizer adapts PolicyUC to the middleware interface.
// Used by server middleware to check RBAC permissions.
type authorizer struct {
	policyUC *PolicyUC
}

func NewAuthorizer(policyUC *PolicyUC) *authorizer {
	return &authorizer{policyUC: policyUC}
}

func (a *authorizer) Check(ctx context.Context, identity *pkgauth.Identity, operation string) error {
	allowed, err := a.policyUC.Enforce(ctx, identity.Subject, operation)
	if err != nil {
		return err
	}
	if !allowed {
		return fmt.Errorf("access denied: %s", operation)
	}
	return nil
}
```

更新 `biz.go` ProviderSet:

```go
var ProviderSet = wire.NewSet(
	NewAccountUC,
	NewSessionValidator,
	// Slice 3 additions:
	NewRoleUC,
	NewRoleBindingUC,
	NewPermissionBindingUC,
	NewPolicyUC,
	NewResourceUC,
	NewAuthorizer,
)
```

---

## Step 5: Data 层 — RBAC RPs

**文件**: `apps/yggdrasil/services/iam/internal/data/rp_role.go`

```go
package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/orm/ent/entutil"
	"cyber-ecosystem/shared-go/utils"

	yggdrasilV1 "cyber-ecosystem/apps/yggdrasil/gen/go/v1"
	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/biz"
	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/data/ent"
	entrole "cyber-ecosystem/apps/yggdrasil/services/iam/internal/data/ent/role"
)

type roleRP struct {
	RP
}

func NewRoleRP(logger log.Logger, store *Store) biz.RoleRP {
	return &roleRP{
		RP: RP{log: log.NewHelper(log.With(logger, "module", "data/rp_role")), store: store},
	}
}

func (rp *roleRP) Create(ctx context.Context, r *biz.Role) (*biz.Role, error) {
	builder := rp.store.GetClient(ctx).Role.Create().
		SetName(*r.Name).SetCode(*r.Code)
	if r.Sort != nil {
		builder.SetSort(*r.Sort)
	}
	if r.Description != nil {
		builder.SetDescription(*r.Description)
	}
	result, err := builder.Save(ctx)
	if err != nil {
		return nil, HandleError(err)
	}
	return mapRole(result), nil
}

func (rp *roleRP) Update(ctx context.Context, r *biz.Role) (*biz.Role, error) {
	builder := rp.store.GetClient(ctx).Role.UpdateOneID(*r.ID)
	if r.Name != nil {
		builder.SetName(*r.Name)
	}
	if r.Sort != nil {
		builder.SetSort(int(*r.Sort))
	}
	if r.Description != nil {
		builder.SetDescription(*r.Description)
	}
	result, err := builder.Save(ctx)
	if err != nil {
		return nil, HandleError(err)
	}
	return mapRole(result), nil
}

func (rp *roleRP) Delete(ctx context.Context, id string) (*biz.Role, error) {
	result, err := rp.store.GetClient(ctx).Role.Get(ctx, id)
	if err != nil {
		return nil, HandleError(err)
	}
	if err := rp.store.GetClient(ctx).Role.DeleteOneID(id).Exec(ctx); err != nil {
		return nil, HandleError(err)
	}
	return mapRole(result), nil
}

func (rp *roleRP) Get(ctx context.Context, id string) (*biz.Role, error) {
	result, err := rp.store.GetClient(ctx).Role.Get(ctx, id)
	if err != nil {
		return nil, HandleError(err)
	}
	return mapRole(result), nil
}

func (rp *roleRP) GetByCode(ctx context.Context, code string) (*biz.Role, error) {
	result, err := rp.store.GetClient(ctx).Role.Query().Where(entrole.Code(code)).Only(ctx)
	if err != nil {
		return nil, HandleError(err)
	}
	return mapRole(result), nil
}

func (rp *roleRP) Query(ctx context.Context, in *biz.RoleQueryIn) (*biz.RoleQueryOut, error) {
	query := rp.store.GetClient(ctx).Role.Query()
	entutil.ApplyOrderBy(in.OrderBy, ent.Asc, ent.Desc, entutil.FOMapping{
		"created_at": func(sel entutil.SQLSelector) { query.Order(sel(entrole.FieldCreatedAt)) },
		"sort":       func(sel entutil.SQLSelector) { query.Order(sel(entrole.FieldSort)) },
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
	return &biz.RoleQueryOut{
		PageResponse: entutil.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(list, mapRole),
	}, nil
}

func mapRole(r *ent.Role) *biz.Role {
	sort := int32(r.Sort)
	return &biz.Role{
		ID:          &r.ID,
		Name:        &r.Name,
		Code:        &r.Code,
		Sort:        &sort,
		Description: &r.Description,
	}
}
```

**文件**: `apps/yggdrasil/services/iam/internal/data/rp_role_binding.go`
- 与 rp_role.go 结构相同，使用 `ent.RoleBinding` / `entrolebinding` 包
- Query 方法 JOIN role 表获取 role_code 和 role_name

**文件**: `apps/yggdrasil/services/iam/internal/data/rp_permission_binding.go`
- 与 rp_role.go 结构相同，使用 `ent.PermissionBinding` / `entpermissionbinding` 包
- Query 方法 JOIN role 表获取 role_code

**文件**: `apps/yggdrasil/services/iam/internal/data/rp_policy.go`
- 在 Slice 2 基础上扩展，添加完整 Casbin adapter 实现:
  - `RemoveRoleGroupings(ctx, roleCode)` — 删除 Casbin `g` 行
  - `RemoveRolePermissions(ctx, roleCode)` — 删除 Casbin `p` 行
  - `ReloadPolicy()` — 从 DB 重新加载策略
  - `GetRolesForUser(userID)` — 查询用户角色
  - Redis pub/sub 通知（如果使用 Redis cache）

更新 `data.go` ProviderSet:

```go
var ProviderSet = wire.NewSet(
	NewStore, NewCache, NewEntClient,
	wire.Bind(new(biz.Transaction), new(*Store)),
	NewUserRP, NewSessionRP, NewPolicyRP,
	// Slice 3 additions:
	NewRoleRP, NewRoleBindingRP, NewPermissionBindingRP,
)
```

---

## Step 6: Service 层 — RBAC Services

按 Slice 1 模式创建:

- `service/role.go` — 实现 `RoleServiceServer`，映射 proto ↔ biz model
- `service/role_binding.go` — 实现 `RoleBindingServiceServer`
- `service/permission_binding.go` — 实现 `PermissionBindingServiceServer`
- `service/resource.go` — 实现 `ResourceServiceServer`
- `service/internal_authorization.go` — 实现 `InternalAuthorizationServiceServer`

更新 `service/service.go` RegistrarList:

```go
func NewRegistrarList(
	s1 *AccountAuthService,
	s2 *InternalAuthService,
	s3 *RoleService,
	s4 *RoleBindingService,
	s5 *PermissionBindingService,
	s6 *ResourceService,
	s7 *InternalAuthorizationService,
) []Registrar {
	return []Registrar{s1, s2, s3, s4, s5, s6, s7}
}
```

---

## Step 7: Server 层 — 添加 Authorizer 中间件

**文件**: `apps/yggdrasil/services/iam/internal/server/middleware/authorizer.go`

```go
package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	pkgauth "cyber-ecosystem/apps/yggdrasil/services/iam/internal/pkg/auth"
	yggdrasilV1 "cyber-ecosystem/apps/yggdrasil/gen/go/v1"
)

// Authorizer checks RBAC permissions.
func Authorizer(authorizer *biz.authorizer, logger log.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			identity, err := pkgauth.IdentityFromContext(ctx)
			if err != nil {
				return handler(ctx, req)
			}
			// Extract operation from transport info
			tr, _ := transport.FromServerContext(ctx)
			if tr == nil {
				return handler(ctx, req)
			}
			operation := tr.Operation()
			if err := authorizer.Check(ctx, identity, operation); err != nil {
				return nil, yggdrasilV1.ErrorErrorReasonForbidden("access denied")
			}
			return handler(ctx, req)
		}
	}
}
```

更新 `server/server.go` 中的 `buildMiddlewares()`，在 `SessionValidator` 之后添加:

```go
mws = append(mws, selector.Server(
	jwt.Server(...),
	mw.SessionValidator(sessionValidator, logger),
	mw.Authorizer(authorizer, logger),
	// Condition, DataScope will be added in Slice 4
).Match(auth.NewWhiteListByPublicAccessInProtoMatcher()).Build())
```

更新 Wire 注入: `NewGRPCServer` / `NewHTTPServer` / `NewConnectServer` 的参数中添加 `authorizer *biz.authorizer`。

---

## Step 8: 编译闭环

```bash
cd apps/yggdrasil/services/iam && go mod tidy
./nx run yggdrasil_iam:generate
./nx run yggdrasil_iam:build
```

验证: `ls apps/yggdrasil/services/iam/bin/iam`

提交:
```bash
git add apps/yggdrasil/
git commit -m "feat(yggdrasil): IAM RBAC + event bus — first full build"
```

---

## Step 9: 集成验证

### 9a: 启动服务

```bash
# 确保 audit 服务和 IAM 服务都运行
./nx run yggdrasil_audit:dev &
./nx run yggdrasil_iam:dev
```

### 9b: 创建角色

```bash
TOKEN=$(curl -s -X POST http://localhost:11000/iam/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "admin123"}' | jq -r '.access_token')

curl -X POST http://localhost:11000/iam/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Editor", "code": "editor", "sort": 2}'
```

### 9c: 创建权限绑定

```bash
ROLE_ID="<id_from_9b>"

curl -X POST http://localhost:11000/iam/permission-bindings \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"role_id\": \"$ROLE_ID\", \"resource\": \"worklog.create\", \"effect\": \"allow\"}"
```

### 9d: 删除角色 → 验证事件级联

```bash
# 删除角色
curl -X DELETE "http://localhost:11000/iam/roles/$ROLE_ID" \
  -H "Authorization: Bearer $TOKEN"

# 验证: Casbin 策略已自动清理
grpcurl -plaintext -d "{\"user_id\": \"admin_id\", \"resource\": \"worklog.create\"}" \
  localhost:12000 api.yggdrasil.v1.InternalAuthorizationService/CheckPermission
```

预期: 删除角色后，关联的 Casbin 策略通过事件订阅自动清理。

### 9e: 测试 RBAC 中间件

```bash
# 超级管理员应该能访问所有接口（因为种子策略 p = super_admin, self, /*, allow）
curl http://localhost:11000/iam/roles/query \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

预期: 返回角色列表

### 9f: 停止服务并提交

```bash
git add apps/yggdrasil/
git commit -m "feat(yggdrasil): IAM RBAC + event bus passes integration verification"
```

---

## 完成标准

- [x] `./nx run yggdrasil_iam:build` 编译通过
- [x] Event Bus 实现并可订阅/发布事件
- [x] 角色创建/查询/删除正常
- [x] 角色绑定创建/删除正常（自动同步 Casbin）
- [x] 权限绑定创建/删除正常（自动同步 Casbin）
- [x] 删除角色 → 事件级联清理关联策略
- [x] Authorizer 中间件正确拒绝无权限请求
- [x] 内部 CheckPermission API 正常工作
- [x] 变更已提交
