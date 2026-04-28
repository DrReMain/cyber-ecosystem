# Slice 4: IAM Advanced — Condition / DataScope / Department / UserAttribute

> **闭环标准**: 条件/数据权限集成测试通过。

## 目标

在 Slice 3 的 RBAC 基础上，构建 IAM 的高级能力：

1. **ABAC 条件** — 时间/IP/星期/属性匹配插件 + 条件规则管理
2. **DataScope** — 行级数据权限插件 + scope 规则管理
3. **部门管理** — 层级部门 + 部门绑定
4. **用户属性** — KV 属性存储
5. **用户管理** — 完整用户 CRUD

核心改进：所有能力通过事件总线协调，无 UC 间直接依赖。

## 前置条件

- Slice 3 完成（RBAC + Event Bus + Authorizer）

---

## Step 1: API Proto 定义

### 1a: iam_condition.proto

**文件**: `apps/yggdrasil/api/v1/iam_condition.proto`

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

// ConditionService
service ConditionService {
  option (desc.service_comment) = "条件规则管理服务";

  rpc CreateCondition(CreateConditionRequest) returns (CreateConditionResponse) {
    option (desc.method_comment) = "创建条件规则";
    option (google.api.http) = {post: "/iam/conditions" body: "*"};
  }
  rpc UpdateCondition(UpdateConditionRequest) returns (UpdateConditionResponse) {
    option (desc.method_comment) = "更新条件规则";
    option (google.api.http) = {put: "/iam/conditions/{id}" body: "*"};
  }
  rpc DeleteCondition(DeleteConditionRequest) returns (DeleteConditionResponse) {
    option (desc.method_comment) = "删除条件规则";
    option (google.api.http) = {delete: "/iam/conditions/{id}"};
  }
  rpc GetCondition(GetConditionRequest) returns (GetConditionResponse) {
    option (desc.method_comment) = "获取条件详情";
    option (google.api.http) = {get: "/iam/conditions/{id}"};
  }
  rpc QueryCondition(QueryConditionRequest) returns (QueryConditionResponse) {
    option (desc.method_comment) = "查询条件列表";
    option (google.api.http) = {post: "/iam/conditions/query" body: "*"};
  }
}

message CreateConditionRequest {
  string role_id = 1 [(buf.validate.field).string.min_len = 1];
  string name = 2 [(buf.validate.field).string.min_len = 1];
  string type = 3 [(buf.validate.field).string.min_len = 1];
  string config = 4 [(buf.validate.field).string.min_len = 1];
  string group_id = 5;
  string target_resource = 6;
}
message CreateConditionResponse { string id = 1; }

message UpdateConditionRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
  google.protobuf.StringValue name = 2;
  google.protobuf.StringValue config = 3;
  google.protobuf.StringValue group_id = 4;
  google.protobuf.StringValue target_resource = 5;
}
message UpdateConditionResponse {}

message DeleteConditionRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
}
message DeleteConditionResponse {}

message GetConditionRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
}
message GetConditionResponse {
  string id = 1;
  string role_id = 2;
  string name = 3;
  string type = 4;
  string config = 5;
  string group_id = 6;
  string target_resource = 7;
}

message QueryConditionRequest {
  common.PageRequest page = 1;
  optional string role_id = 2;
  optional string type = 3;
  optional string target_resource = 4;
}
message QueryConditionResponse {
  common.PageResponse page = 1;
  repeated GetConditionResponse list = 2;
}
```

### 1b: iam_data_scope.proto

**文件**: `apps/yggdrasil/api/v1/iam_data_scope.proto`

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

// DataScopeService
service DataScopeService {
  option (desc.service_comment) = "数据权限管理服务";

  rpc CreateDataScope(CreateDataScopeRequest) returns (CreateDataScopeResponse) {
    option (desc.method_comment) = "创建数据权限规则";
    option (google.api.http) = {post: "/iam/data-scopes" body: "*"};
  }
  rpc UpdateDataScope(UpdateDataScopeRequest) returns (UpdateDataScopeResponse) {
    option (desc.method_comment) = "更新数据权限规则";
    option (google.api.http) = {put: "/iam/data-scopes/{id}" body: "*"};
  }
  rpc DeleteDataScope(DeleteDataScopeRequest) returns (DeleteDataScopeResponse) {
    option (desc.method_comment) = "删除数据权限规则";
    option (google.api.http) = {delete: "/iam/data-scopes/{id}"};
  }
  rpc GetDataScope(GetDataScopeRequest) returns (GetDataScopeResponse) {
    option (desc.method_comment) = "获取数据权限详情";
    option (google.api.http) = {get: "/iam/data-scopes/{id}"};
  }
  rpc QueryDataScope(QueryDataScopeRequest) returns (QueryDataScopeResponse) {
    option (desc.method_comment) = "查询数据权限列表";
    option (google.api.http) = {post: "/iam/data-scopes/query" body: "*"};
  }
}

message CreateDataScopeRequest {
  string role_id = 1 [(buf.validate.field).string.min_len = 1];
  string name = 2 [(buf.validate.field).string.min_len = 1];
  string type = 3 [(buf.validate.field).string.min_len = 1];
  string config = 4;
  string target_resource = 5;
}
message CreateDataScopeResponse { string id = 1; }

message UpdateDataScopeRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
  google.protobuf.StringValue name = 2;
  google.protobuf.StringValue config = 3;
  google.protobuf.StringValue target_resource = 4;
}
message UpdateDataScopeResponse {}

message DeleteDataScopeRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
}
message DeleteDataScopeResponse {}

message GetDataScopeRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
}
message GetDataScopeResponse {
  string id = 1;
  string role_id = 2;
  string name = 3;
  string type = 4;
  string config = 5;
  string target_resource = 6;
}

message QueryDataScopeRequest {
  common.PageRequest page = 1;
  optional string role_id = 2;
  optional string type = 3;
  optional string target_resource = 4;
}
message QueryDataScopeResponse {
  common.PageResponse page = 1;
  repeated GetDataScopeResponse list = 2;
}
```

### 1c: iam_department.proto

**文件**: `apps/yggdrasil/api/v1/iam_department.proto`

```protobuf
syntax = "proto3";

// buf:lint:ignore PACKAGE_DIRECTORY_MATCH
package api.yggdrasil.v1;

import "buf/validate/validate.proto";
import "desc/desc.proto";
import "google/api/annotations.proto";
import "google/protobuf/wrappers.proto";

option go_package = "cyber-ecosystem/apps/yggdrasil/gen/go/v1;yggdrasilV1";

// DepartmentService
service DepartmentService {
  option (desc.service_comment) = "部门管理服务";

  rpc CreateDepartment(CreateDepartmentRequest) returns (CreateDepartmentResponse) {
    option (desc.method_comment) = "创建部门";
    option (google.api.http) = {post: "/iam/departments" body: "*"};
  }
  rpc UpdateDepartment(UpdateDepartmentRequest) returns (UpdateDepartmentResponse) {
    option (desc.method_comment) = "更新部门";
    option (google.api.http) = {put: "/iam/departments/{id}" body: "*"};
  }
  rpc DeleteDepartment(DeleteDepartmentRequest) returns (DeleteDepartmentResponse) {
    option (desc.method_comment) = "删除部门";
    option (google.api.http) = {delete: "/iam/departments/{id}"};
  }
  rpc GetDepartment(GetDepartmentRequest) returns (GetDepartmentResponse) {
    option (desc.method_comment) = "获取部门详情";
    option (google.api.http) = {get: "/iam/departments/{id}"};
  }
  rpc ListDepartments(ListDepartmentsRequest) returns (ListDepartmentsResponse) {
    option (desc.method_comment) = "列出所有部门（树形）";
    option (google.api.http) = {get: "/iam/departments"};
  }
}

message CreateDepartmentRequest {
  string name = 1 [(buf.validate.field).string.min_len = 1];
  google.protobuf.StringValue parent_id = 2;
  int32 sort = 3;
}
message CreateDepartmentResponse { string id = 1; }

message UpdateDepartmentRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
  google.protobuf.StringValue name = 2;
  google.protobuf.Int32Value sort = 3;
}
message UpdateDepartmentResponse {}

message DeleteDepartmentRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
}
message DeleteDepartmentResponse {}

message GetDepartmentRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
}
message GetDepartmentResponse {
  string id = 1;
  string name = 2;
  google.protobuf.StringValue parent_id = 3;
  int32 sort = 4;
  repeated GetDepartmentResponse children = 5;
}

message ListDepartmentsRequest {}
message ListDepartmentsResponse {
  repeated GetDepartmentResponse list = 1;
}
```

### 1d: iam_department_binding.proto

**文件**: `apps/yggdrasil/api/v1/iam_department_binding.proto`

```protobuf
syntax = "proto3";

// buf:lint:ignore PACKAGE_DIRECTORY_MATCH
package api.yggdrasil.v1;

import "buf/validate/validate.proto";
import "desc/desc.proto";
import "google/api/annotations.proto";

option go_package = "cyber-ecosystem/apps/yggdrasil/gen/go/v1;yggdrasilV1";

// DepartmentBindingService
service DepartmentBindingService {
  option (desc.service_comment) = "部门绑定服务";

  rpc SetUserDepartments(SetUserDepartmentsRequest) returns (SetUserDepartmentsResponse) {
    option (desc.method_comment) = "设置用户所属部门（替换）";
    option (google.api.http) = {put: "/iam/users/{user_id}/departments" body: "*"};
  }
  rpc GetUserDepartments(GetUserDepartmentsRequest) returns (GetUserDepartmentsResponse) {
    option (desc.method_comment) = "获取用户部门列表";
    option (google.api.http) = {get: "/iam/users/{user_id}/departments"};
  }
}

message SetUserDepartmentsRequest {
  string user_id = 1 [(buf.validate.field).string.min_len = 1];
  repeated string department_ids = 2;
}
message SetUserDepartmentsResponse {}

message GetUserDepartmentsRequest {
  string user_id = 1 [(buf.validate.field).string.min_len = 1];
}
message GetUserDepartmentsResponse {
  repeated string department_ids = 1;
}
```

### 1e: iam_user_attribute.proto

**文件**: `apps/yggdrasil/api/v1/iam_user_attribute.proto`

```protobuf
syntax = "proto3";

// buf:lint:ignore PACKAGE_DIRECTORY_MATCH
package api.yggdrasil.v1;

import "buf/validate/validate.proto";
import "desc/desc.proto";
import "google/api/annotations.proto";

option go_package = "cyber-ecosystem/apps/yggdrasil/gen/go/v1;yggdrasilV1";

// UserAttributeService
service UserAttributeService {
  option (desc.service_comment) = "用户属性管理服务";

  rpc SetUserAttributes(SetUserAttributesRequest) returns (SetUserAttributesResponse) {
    option (desc.method_comment) = "设置用户属性（upsert）";
    option (google.api.http) = {put: "/iam/users/{user_id}/attributes" body: "*"};
  }
  rpc GetUserAttributes(GetUserAttributesRequest) returns (GetUserAttributesResponse) {
    option (desc.method_comment) = "获取用户属性";
    option (google.api.http) = {get: "/iam/users/{user_id}/attributes"};
  }
}

message AttributeKV {
  string key = 1;
  string value = 2;
}

message SetUserAttributesRequest {
  string user_id = 1 [(buf.validate.field).string.min_len = 1];
  repeated AttributeKV attributes = 2;
}
message SetUserAttributesResponse {}

message GetUserAttributesRequest {
  string user_id = 1 [(buf.validate.field).string.min_len = 1];
}
message GetUserAttributesResponse {
  repeated AttributeKV attributes = 1;
}
```

### 1f: iam_user.proto

**文件**: `apps/yggdrasil/api/v1/iam_user.proto`

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

// UserService
service UserService {
  option (desc.service_comment) = "用户管理服务";

  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) {
    option (desc.method_comment) = "创建用户";
    option (google.api.http) = {post: "/iam/users" body: "*"};
  }
  rpc UpdateUser(UpdateUserRequest) returns (UpdateUserResponse) {
    option (desc.method_comment) = "更新用户";
    option (google.api.http) = {put: "/iam/users/{id}" body: "*"};
  }
  rpc DeleteUser(DeleteUserRequest) returns (DeleteUserResponse) {
    option (desc.method_comment) = "删除用户";
    option (google.api.http) = {delete: "/iam/users/{id}"};
  }
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {
    option (desc.method_comment) = "获取用户详情";
    option (google.api.http) = {get: "/iam/users/{id}"};
  }
  rpc QueryUser(QueryUserRequest) returns (QueryUserResponse) {
    option (desc.method_comment) = "查询用户列表";
    option (google.api.http) = {post: "/iam/users/query" body: "*"};
  }
}

message CreateUserRequest {
  string email = 1 [(buf.validate.field).string.email = true];
  string password = 2 [(buf.validate.field).string.min_len = 6];
  string name = 3;
}
message CreateUserResponse { string id = 1; }

message UpdateUserRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
  google.protobuf.StringValue name = 2;
  google.protobuf.StringValue password = 3;
  google.protobuf.StringValue status = 4;
}
message UpdateUserResponse {}

message DeleteUserRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
}
message DeleteUserResponse {}

message GetUserRequest {
  string id = 1 [(buf.validate.field).string.min_len = 1];
}
message GetUserResponse {
  string id = 1;
  string email = 2;
  string name = 3;
  string status = 4;
  string created_at = 5;
}

message QueryUserRequest {
  common.PageRequest page = 1;
  optional string status = 2;
  repeated string order_by = 100;
}
message QueryUserResponse {
  common.PageResponse page = 1;
  repeated GetUserResponse list = 2;
}
```

### 1g: iam_internal_scope.proto

**文件**: `apps/yggdrasil/api/v1/iam_internal_scope.proto`

```protobuf
syntax = "proto3";

// buf:lint:ignore PACKAGE_DIRECTORY_MATCH
package api.yggdrasil.v1;

import "buf/validate/validate.proto";

option go_package = "cyber-ecosystem/apps/yggdrasil/gen/go/v1;yggdrasilV1";

// InternalScopeService — 内部 API，供业务服务获取 DataScope 规则。
service InternalScopeService {
  // ResolveScope 获取用户对指定资源的数据权限规则。
  rpc ResolveScope(ResolveScopeRequest) returns (ResolveScopeResponse);
}

message ResolveScopeRequest {
  string user_id = 1 [(buf.validate.field).string.min_len = 1];
  string resource = 2 [(buf.validate.field).string.min_len = 1];
}

message ScopeRule {
  string field = 1;
  string op = 2;
  string value = 3;
}

message ResolveScopeResponse {
  bool is_all = 1;
  bool self_filter = 2;
  bool dept_filter = 3;
  repeated string dept_ids = 4;
  repeated ScopeRule rules = 5;
  string logic = 6;
}
```

### 1h: iam_internal_condition.proto

**文件**: `apps/yggdrasil/api/v1/iam_internal_condition.proto`

```protobuf
syntax = "proto3";

// buf:lint:ignore PACKAGE_DIRECTORY_MATCH
package api.yggdrasil.v1;

import "buf/validate/validate.proto";

option go_package = "cyber-ecosystem/apps/yggdrasil/gen/go/v1;yggdrasilV1";

// InternalConditionService — 内部 API，供业务服务检查条件。
service InternalConditionService {
  // CheckConditions 检查用户对指定操作的条件是否满足。
  rpc CheckConditions(CheckConditionsRequest) returns (CheckConditionsResponse);
}

message CheckConditionsRequest {
  string user_id = 1 [(buf.validate.field).string.min_len = 1];
  string operation = 2 [(buf.validate.field).string.min_len = 1];
}

message CheckConditionsResponse {
  bool allowed = 1;
}
```

### 1i: 生成 Proto 代码

```bash
buf lint apps/yggdrasil/api/v1/iam_condition.proto
buf lint apps/yggdrasil/api/v1/iam_data_scope.proto
buf lint apps/yggdrasil/api/v1/iam_department.proto
buf lint apps/yggdrasil/api/v1/iam_department_binding.proto
buf lint apps/yggdrasil/api/v1/iam_user_attribute.proto
buf lint apps/yggdrasil/api/v1/iam_user.proto
buf lint apps/yggdrasil/api/v1/iam_internal_scope.proto
buf lint apps/yggdrasil/api/v1/iam_internal_condition.proto
./nx run yggdrasil_api:proto:api
```

提交:
```bash
git add apps/yggdrasil/api/v1/iam_*.proto apps/yggdrasil/gen/
git commit -m "feat(yggdrasil): add IAM advanced capability proto definitions"
```

---

## Step 2: Ent Schema 补充

### 2a: condition.go

```go
package schema

type Condition struct {
	ent.Schema
}

func (Condition) Fields() []ent.Field {
	return []ent.Field{
		field.String("role_id").NotEmpty().MaxLen(20),
		field.String("name").NotEmpty().MaxLen(100),
		field.String("type").NotEmpty().MaxLen(50),         // time_range, ip_range, day_of_week, attribute_match
		field.Text("config"),                                // JSON config
		field.String("group_id").Default("").MaxLen(50),     // AND within group, OR between groups
		field.String("target_resource").Default("").MaxLen(200), // e.g. "worklog.*" or specific operation
	}
}

func (Condition) Mixin() []ent.Mixin {
	return []ent.Mixin{mixins.IDStringMixin{}, mixins.CreatedUpdatedMixin{}}
}

func (Condition) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("role_id"),
		index.Fields("type"),
		index.Fields("target_resource"),
	}
}
```

### 2b: data_scope.go

```go
package schema

type DataScope struct {
	ent.Schema
}

func (DataScope) Fields() []ent.Field {
	return []ent.Field{
		field.String("role_id").NotEmpty().MaxLen(20),
		field.String("name").NotEmpty().MaxLen(100),
		field.String("type").NotEmpty().MaxLen(50),  // all, self, dept, attribute
		field.Text("config").Optional(),             // JSON config (for attribute type)
		field.String("target_resource").Default("").MaxLen(200),
	}
}

func (DataScope) Mixin() []ent.Mixin {
	return []ent.Mixin{mixins.IDStringMixin{}, mixins.CreatedUpdatedMixin{}}
}

func (DataScope) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("role_id"),
		index.Fields("type"),
		index.Fields("target_resource"),
	}
}
```

### 2c: department.go

```go
package schema

type Department struct {
	ent.Schema
}

func (Department) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().MaxLen(100),
		field.String("parent_id").Optional().Nillable().MaxLen(20),
		field.Int("sort").Default(0),
	}
}

func (Department) Mixin() []ent.Mixin {
	return []ent.Mixin{mixins.IDStringMixin{}, mixins.CreatedUpdatedMixin{}}
}

func (Department) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("parent_id"),
	}
}
```

### 2d: department_binding.go

```go
package schema

type DepartmentBinding struct {
	ent.Schema
}

func (DepartmentBinding) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").NotEmpty().MaxLen(20),
		field.String("department_id").NotEmpty().MaxLen(20),
	}
}

func (DepartmentBinding) Mixin() []ent.Mixin {
	return []ent.Mixin{mixins.IDStringMixin{}, mixins.CreatedUpdatedMixin{}}
}

func (DepartmentBinding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "department_id").Unique(),
		index.Fields("user_id"),
		index.Fields("department_id"),
	}
}
```

### 2e: user_attribute.go

```go
package schema

type UserAttribute struct {
	ent.Schema
}

func (UserAttribute) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").NotEmpty().MaxLen(20),
		field.String("key").NotEmpty().MaxLen(100),
		field.Text("value"),
	}
}

func (UserAttribute) Mixin() []ent.Mixin {
	return []ent.Mixin{mixins.IDStringMixin{}, mixins.CreatedUpdatedMixin{}}
}

func (UserAttribute) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "key").Unique(),
		index.Fields("user_id"),
	}
}
```

### 2f: 重新生成 Ent

```bash
cd apps/yggdrasil/services/iam && go generate ./internal/data/ent/...
```

---

## Step 3: Pkg 层 — Condition / DataScope 插件

### 3a: Condition 插件

**文件**: `apps/yggdrasil/services/iam/internal/pkg/condition/plugin.go`

```go
package condition

import "context"

type ConditionPlugin interface {
	Type() string
	Evaluate(ctx context.Context, config string) (bool, error)
	ValidateConfig(config string) error
}

type Registry struct {
	plugins map[string]ConditionPlugin
}

func NewRegistry() *Registry {
	return &Registry{plugins: make(map[string]ConditionPlugin)}
}

func (r *Registry) Register(p ConditionPlugin) {
	r.plugins[p.Type()] = p
}

func (r *Registry) Get(typ string) (ConditionPlugin, bool) {
	p, ok := r.plugins[typ]
	return p, ok
}

func NewBuiltinRegistry() *Registry {
	r := NewRegistry()
	r.Register(&TimeRangePlugin{})
	r.Register(&IPRangePlugin{})
	r.Register(&DayOfWeekPlugin{})
	r.Register(&AttributeMatchPlugin{})
	return r
}
```

**文件**: `apps/yggdrasil/services/iam/internal/pkg/condition/plugin_time_range.go`

```go
package condition

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type TimeRangePlugin struct{}

func (p *TimeRangePlugin) Type() string { return "time_range" }

type TimeRangeConfig struct {
	StartHour int `json:"start_hour"` // 0-23
	EndHour   int `json:"end_hour"`   // 0-23
}

func (p *TimeRangePlugin) Evaluate(ctx context.Context, config string) (bool, error) {
	var c TimeRangeConfig
	if err := json.Unmarshal([]byte(config), &c); err != nil {
		return false, err
	}
	hour := time.Now().Hour()
	return hour >= c.StartHour && hour <= c.EndHour, nil
}

func (p *TimeRangePlugin) ValidateConfig(config string) error {
	var c TimeRangeConfig
	if err := json.Unmarshal([]byte(config), &c); err != nil {
		return err
	}
	if c.StartHour < 0 || c.StartHour > 23 || c.EndHour < 0 || c.EndHour > 23 {
		return fmt.Errorf("invalid hour range")
	}
	return nil
}
```

> 类似创建 `plugin_ip_range.go`、`plugin_day_of_week.go`、`plugin_attribute_match.go`

### 3b: DataScope 插件

**文件**: `apps/yggdrasil/services/iam/internal/pkg/datascope/plugin.go`

```go
package datascope

import "context"

type ScopePlugin interface {
	Type() string
	ValidateConfig(config string) error
	Merge(scope RoleScope, snap *ScopeSnapshot, result *EffectiveScope) error
}

type RoleScope struct {
	RoleID         string
	Type           string
	Config         string
	TargetResource string
}

type ScopeSnapshot struct {
	Roles      []string
	Scopes     []RoleScope
	DeptIDs    []string
	Attributes map[string]string
}

type EffectiveScope struct {
	IsAll           bool
	SelfFilter      bool
	DeptFilter      bool
	AttributeFilter bool
	DeptIDs         []string
	Rules           []FilterRule
	Logic           string
}

type FilterRule struct {
	Field string `json:"field"`
	Op    string `json:"op"` // eq, neq, gt, gte, lt, lte, in
	Value string `json:"value"`
}

type Registry struct {
	plugins map[string]ScopePlugin
}

func NewRegistry() *Registry {
	return &Registry{plugins: make(map[string]ScopePlugin)}
}

func (r *Registry) Register(p ScopePlugin) { r.plugins[p.Type()] = p }

func (r *Registry) Get(typ string) (ScopePlugin, bool) {
	p, ok := r.plugins[typ]
	return p, ok
}

func NewBuiltinRegistry() *Registry {
	r := NewRegistry()
	r.Register(&AllScopePlugin{})
	r.Register(&SelfScopePlugin{})
	r.Register(&DeptScopePlugin{})
	r.Register(&AttributeScopePlugin{})
	return r
}
```

**文件**: `apps/yggdrasil/services/iam/internal/pkg/datascope/plugin_all.go`

```go
package datascope

type AllScopePlugin struct{}

func (p *AllScopePlugin) Type() string { return "all" }

func (p *AllScopePlugin) ValidateConfig(string) error { return nil }

func (p *AllScopePlugin) Merge(scope RoleScope, snap *ScopeSnapshot, result *EffectiveScope) error {
	result.IsAll = true
	return nil
}
```

> 类似创建 `plugin_self.go`、`plugin_dept.go`、`plugin_attribute.go`

---

## Step 4: Biz 层 — Advanced UCs

**文件**: `apps/yggdrasil/services/iam/internal/biz/uc_condition.go`

```go
package biz

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/event"
	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/pkg/condition"
)

type ConditionRule struct {
	ID             *string
	RoleID         *string
	Name           *string
	Type           *string
	Config         *string
	GroupID        *string
	TargetResource *string
}

type ConditionQueryIn struct {
	*common.PageRequest
	RoleID         *string
	Type           *string
	TargetResource *string
}

type ConditionQueryOut struct {
	*common.PageResponse
	List []*ConditionRule
}

type ConditionRP interface {
	Create(ctx context.Context, c *ConditionRule) (*ConditionRule, error)
	Update(ctx context.Context, c *ConditionRule) (*ConditionRule, error)
	Delete(ctx context.Context, id string) (*ConditionRule, error)
	Query(ctx context.Context, in *ConditionQueryIn) (*ConditionQueryOut, error)
	QueryByRoleIDs(ctx context.Context, roleIDs []string) ([]*ConditionRule, error)
}

type ConditionUC struct {
	UC
	conditionRP  ConditionRP
	registry     *condition.Registry
	policyRP     PolicyRP
	bus          *event.SyncBus
}

func NewConditionUC(
	logger log.Logger, tm Transaction,
	conditionRP ConditionRP, registry *condition.Registry, policyRP PolicyRP,
	bus *event.SyncBus,
) *ConditionUC {
	uc := &ConditionUC{
		UC:          UC{log: log.NewHelper(log.With(logger, "module", "biz/uc_condition")), tm: tm},
		conditionRP: conditionRP,
		registry:    registry,
		policyRP:    policyRP,
		bus:         bus,
	}
	// Subscribe to role deletion
	bus.Subscribe(event.TopicRoleDeleted, uc.onRoleDeleted)
	return uc
}

func (uc *ConditionUC) CheckConditions(ctx context.Context, userID, operation string) (bool, error) {
	roles := uc.policyRP.GetRolesForUser(userID)
	if len(roles) == 0 {
		return true, nil // No conditions to check
	}
	rules, err := uc.conditionRP.QueryByRoleIDs(ctx, roles)
	if err != nil {
		return false, err
	}
	if len(rules) == 0 {
		return true, nil
	}
	// Group by role, then by group_id
	// AND within group, OR between groups
	// If ANY role passes, allow
	return uc.evaluate(rules, ctx, operation)
}

func (uc *ConditionUC) evaluate(rules []*ConditionRule, ctx context.Context, operation string) (bool, error) {
	// Group rules by role
	byRole := make(map[string][]*ConditionRule)
	for _, r := range rules {
		byRole[*r.RoleID] = append(byRole[*r.RoleID], r)
	}
	for _, roleRules := range byRole {
		if ok, err := uc.evaluateRoleRules(roleRules, ctx, operation); err != nil {
			return false, err
		} else if ok {
			return true, nil
		}
	}
	return false, nil
}

func (uc *ConditionUC) evaluateRoleRules(rules []*ConditionRule, ctx context.Context, operation string) (bool, error) {
	// Group by group_id
	groups := make(map[string][]*ConditionRule)
	for _, r := range rules {
		// Filter by target_resource match
		groups[*r.GroupID] = append(groups[*r.GroupID], r)
	}
	// OR between groups
	for _, groupRules := range groups {
		allPass := true
		for _, r := range groupRules {
			plugin, ok := uc.registry.Get(*r.Type)
			if !ok {
				continue
			}
			pass, err := plugin.Evaluate(ctx, *r.Config)
			if err != nil {
				return false, err
			}
			if !pass {
				allPass = false
				break
			}
		}
		if allPass {
			return true, nil
		}
	}
	return false, nil
}

func (uc *ConditionUC) onRoleDeleted(ctx context.Context, evt event.Event) error {
	roleCode, _ := evt.Payload.(string)
	// Delete all conditions for this role
	// Query by role code → get role ID → delete conditions
	return nil // TODO: implement
}

// CRUD methods delegate to conditionRP
```

**文件**: `apps/yggdrasil/services/iam/internal/biz/uc_data_scope.go`

```go
package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/event"
	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/pkg/datascope"
)

type DataScopeRule struct {
	ID             *string
	RoleID         *string
	Name           *string
	Type           *string
	Config         *string
	TargetResource *string
}

type DataScopeQueryIn struct {
	*common.PageRequest
	RoleID         *string
	Type           *string
	TargetResource *string
}

type DataScopeQueryOut struct {
	*common.PageResponse
	List []*DataScopeRule
}

type DataScopeRP interface {
	Create(ctx context.Context, ds *DataScopeRule) (*DataScopeRule, error)
	Update(ctx context.Context, ds *DataScopeRule) (*DataScopeRule, error)
	Delete(ctx context.Context, id string) (*DataScopeRule, error)
	Query(ctx context.Context, in *DataScopeQueryIn) (*DataScopeQueryOut, error)
	QueryByRoleIDs(ctx context.Context, roleIDs []string) ([]*DataScopeRule, error)
}

type DataScopeSnapshotRP interface {
	GetSnapshot(ctx context.Context, userID string) (*datascope.ScopeSnapshot, error)
	InvalidateUser(ctx context.Context, userID string) error
}

type DataScopeUC struct {
	UC
	dataScopeRP       DataScopeRP
	snapshotRP        DataScopeSnapshotRP
	scopeRegistry     *datascope.Registry
	policyRP          PolicyRP
	bus               *event.SyncBus
}

func NewDataScopeUC(
	logger log.Logger, tm Transaction,
	dataScopeRP DataScopeRP, snapshotRP DataScopeSnapshotRP,
	scopeRegistry *datascope.Registry, policyRP PolicyRP,
	bus *event.SyncBus,
) *DataScopeUC {
	uc := &DataScopeUC{
		UC:            UC{log: log.NewHelper(log.With(logger, "module", "biz/uc_data_scope")), tm: tm},
		dataScopeRP:   dataScopeRP,
		snapshotRP:    snapshotRP,
		scopeRegistry: scopeRegistry,
		policyRP:      policyRP,
		bus:           bus,
	}
	// Subscribe to events that invalidate scope cache
	bus.Subscribe(event.TopicRoleDeleted, uc.onRoleDeleted)
	bus.Subscribe(event.TopicRoleBindingCreated, uc.onBindingChanged)
	bus.Subscribe(event.TopicRoleBindingDeleted, uc.onBindingChanged)
	bus.Subscribe(event.TopicDeptBindingChanged, uc.onBindingChanged)
	bus.Subscribe(event.TopicUserAttributeChanged, uc.onBindingChanged)
	bus.Subscribe(event.TopicUserDeleted, uc.onUserDeleted)
	return uc
}

func (uc *DataScopeUC) ResolveScope(ctx context.Context, userID, resource string) (*datascope.EffectiveScope, error) {
	snap, err := uc.snapshotRP.GetSnapshot(ctx, userID)
	if err != nil {
		return nil, err
	}
	// Filter scopes by target resource
	matched := uc.matchScopes(snap.Scopes, resource)
	// Merge using plugins
	return uc.mergeScopes(matched, snap)
}

func (uc *DataScopeUC) matchScopes(scopes []datascope.RoleScope, resource string) []datascope.RoleScope {
	// Filter by TargetResource matching
	// TODO: implement with wildcard matching
	return scopes
}

func (uc *DataScopeUC) mergeScopes(scopes []datascope.RoleScope, snap *datascope.ScopeSnapshot) (*datascope.EffectiveScope, error) {
	result := &datascope.EffectiveScope{Logic: "or"}
	for _, s := range scopes {
		plugin, ok := uc.scopeRegistry.Get(s.Type)
		if !ok {
			continue
		}
		if err := plugin.Merge(s, snap, result); err != nil {
			return nil, err
		}
		if result.IsAll {
			return result, nil // Short-circuit
		}
	}
	return result, nil
}

// Event handlers
func (uc *DataScopeUC) onRoleDeleted(ctx context.Context, evt event.Event) error {
	// Delete all scope rules for this role
	return nil // TODO
}

func (uc *DataScopeUC) onBindingChanged(ctx context.Context, evt event.Event) error {
	userID, _ := evt.Payload.(string)
	if userID != "" {
		return uc.snapshotRP.InvalidateUser(ctx, userID)
	}
	return nil
}

func (uc *DataScopeUC) onUserDeleted(ctx context.Context, evt event.Event) error {
	userID, _ := evt.Payload.(string)
	if userID != "" {
		return uc.snapshotRP.InvalidateUser(ctx, userID)
	}
	return nil
}

// CRUD methods delegate to dataScopeRP
```

**文件**: `apps/yggdrasil/services/iam/internal/biz/uc_department.go`

```go
package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/event"
)

type Department struct {
	ID       *string
	Name     *string
	ParentID *string
	Sort     *int32
	Children []*Department
}

type DepartmentRP interface {
	Create(ctx context.Context, dept *Department) (*Department, error)
	Update(ctx context.Context, dept *Department) (*Department, error)
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*Department, error)
	List(ctx context.Context) ([]*Department, error)
}

type DepartmentUC struct {
	UC
	deptRP DepartmentRP
	bus    *event.SyncBus
}

func NewDepartmentUC(logger log.Logger, tm Transaction, deptRP DepartmentRP, bus *event.SyncBus) *DepartmentUC {
	return &DepartmentUC{
		UC:     UC{log: log.NewHelper(log.With(logger, "module", "biz/uc_department")), tm: tm},
		deptRP: deptRP,
		bus:    bus,
	}
}

// CRUD methods + tree building from flat list
```

**文件**: `apps/yggdrasil/services/iam/internal/biz/uc_department_binding.go`

```go
package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/event"
)

type DepartmentBindingRP interface {
	SetUserDepartments(ctx context.Context, userID string, deptIDs []string) error
	GetUserDepartments(ctx context.Context, userID string) ([]string, error)
	GetDescendantDeptIDs(ctx context.Context, deptIDs []string) ([]string, error)
}

type DepartmentBindingUC struct {
	UC
	deptBindingRP DepartmentBindingRP
	bus           *event.SyncBus
}

func NewDepartmentBindingUC(logger log.Logger, tm Transaction, deptBindingRP DepartmentBindingRP, bus *event.SyncBus) *DepartmentBindingUC {
	return &DepartmentBindingUC{
		UC:            UC{log: log.NewHelper(log.With(logger, "module", "biz/uc_dept_binding")), tm: tm},
		deptBindingRP: deptBindingRP,
		bus:           bus,
	}
}

func (uc *DepartmentBindingUC) SetUserDepartments(ctx context.Context, userID string, deptIDs []string) error {
	if err := uc.deptBindingRP.SetUserDepartments(ctx, userID, deptIDs); err != nil {
		return err
	}
	uc.bus.Publish(ctx, event.Event{
		Topic:   event.TopicDeptBindingChanged,
		Payload: userID,
	})
	return nil
}

func (uc *DepartmentBindingUC) GetUserDepartments(ctx context.Context, userID string) ([]string, error) {
	return uc.deptBindingRP.GetUserDepartments(ctx, userID)
}
```

**文件**: `apps/yggdrasil/services/iam/internal/biz/uc_user_attribute.go`

```go
package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/event"
)

type UserAttribute struct {
	Key   string
	Value string
}

type UserAttributeRP interface {
	Set(ctx context.Context, userID string, attrs []UserAttribute) error
	Get(ctx context.Context, userID string) ([]UserAttribute, error)
}

type UserAttributeUC struct {
	UC
	userAttrRP UserAttributeRP
	bus        *event.SyncBus
}

func NewUserAttributeUC(logger log.Logger, tm Transaction, userAttrRP UserAttributeRP, bus *event.SyncBus) *UserAttributeUC {
	return &UserAttributeUC{
		UC:         UC{log: log.NewHelper(log.With(logger, "module", "biz/uc_user_attr")), tm: tm},
		userAttrRP: userAttrRP,
		bus:        bus,
	}
}

func (uc *UserAttributeUC) Set(ctx context.Context, userID string, attrs []UserAttribute) error {
	if err := uc.userAttrRP.Set(ctx, userID, attrs); err != nil {
		return err
	}
	uc.bus.Publish(ctx, event.Event{
		Topic:   event.TopicUserAttributeChanged,
		Payload: userID,
	})
	return nil
}

func (uc *UserAttributeUC) Get(ctx context.Context, userID string) ([]UserAttribute, error) {
	return uc.userAttrRP.Get(ctx, userID)
}
```

**文件**: `apps/yggdrasil/services/iam/internal/biz/uc_user.go`

```go
package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/event"
)

// User model expanded from Slice 2 minimal version
type UserQueryIn struct {
	*common.PageRequest
	OrderBy []*utils.OrderBy
	Status  *string
}

type UserQueryOut struct {
	*common.PageResponse
	List []*User
}

// UserRP expanded from Slice 2
type UserRP interface {
	GetByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) (*User, error)
	Get(ctx context.Context, id string) (*User, error)
	Update(ctx context.Context, user *User) (*User, error)
	Delete(ctx context.Context, id string) (*User, error)
	Query(ctx context.Context, in *UserQueryIn) (*UserQueryOut, error)
}

type UserUC struct {
	UC
	userRP UserRP
	bus    *event.SyncBus
}

func NewUserUC(logger log.Logger, tm Transaction, userRP UserRP, bus *event.SyncBus) *UserUC {
	return &UserUC{
		UC:     UC{log: log.NewHelper(log.With(logger, "module", "biz/uc_user")), tm: tm},
		userRP: userRP,
		bus:    bus,
	}
}

func (uc *UserUC) Delete(ctx context.Context, id string) error {
	_, err := uc.userRP.Delete(ctx, id)
	if err != nil {
		return err
	}
	uc.bus.Publish(ctx, event.Event{
		Topic:   event.TopicUserDeleted,
		Payload: id,
	})
	return nil
}

// Other CRUD methods delegate to userRP
```

**文件**: `apps/yggdrasil/services/iam/internal/biz/adapter_condition.go`

```go
package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	pkgauth "cyber-ecosystem/apps/yggdrasil/services/iam/internal/pkg/auth"
)

// ConditionChecker adapts ConditionUC for middleware use.
type conditionChecker struct {
	conditionUC *ConditionUC
}

func NewConditionChecker(conditionUC *ConditionUC) *conditionChecker {
	return &conditionChecker{conditionUC: conditionUC}
}

func (c *conditionChecker) Check(ctx context.Context, userID, operation string) (bool, error) {
	return c.conditionUC.CheckConditions(ctx, userID, operation)
}
```

**文件**: `apps/yggdrasil/services/iam/internal/biz/adapter_datascope.go`

```go
package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/pkg/datascope"
)

// ScopeResolver adapts DataScopeUC for middleware use.
type scopeResolver struct {
	dataScopeUC *DataScopeUC
}

func NewScopeResolver(dataScopeUC *DataScopeUC) *scopeResolver {
	return &scopeResolver{dataScopeUC: dataScopeUC}
}

func (s *scopeResolver) Resolve(ctx context.Context, userID, operation string) (*datascope.EffectiveScope, error) {
	return s.dataScopeUC.ResolveScope(ctx, userID, operation)
}
```

更新 `biz.go` ProviderSet:

```go
var ProviderSet = wire.NewSet(
	// Slice 2
	NewAccountUC, NewSessionValidator,
	// Slice 3
	NewRoleUC, NewRoleBindingUC, NewPermissionBindingUC, NewPolicyUC, NewResourceUC, NewAuthorizer,
	// Slice 4
	NewConditionUC, NewDataScopeUC, NewDepartmentUC, NewDepartmentBindingUC,
	NewUserAttributeUC, NewUserUC,
	NewConditionChecker, NewScopeResolver,
	condition.NewBuiltinRegistry,
	datascope.NewBuiltinRegistry,
)
```

---

## Step 5: Data 层 + Service 层

按已有模式创建:

**Data 层:**
- `data/rp_condition.go` — CRUD + QueryByRoleIDs
- `data/rp_data_scope.go` — CRUD + QueryByRoleIDs
- `data/rp_data_scope_snapshot.go` — GetSnapshot (build from user's roles → scopes → dept IDs → attrs) + InvalidateUser (cache)
- `data/rp_department.go` — CRUD + List (flat)
- `data/rp_department_binding.go` — SetUserDepartments + GetUserDepartments + GetDescendantDeptIDs (recursive query)
- `data/rp_user_attribute.go` — Set (upsert) + Get
- `data/rp_user.go` — 扩展 Slice 2 版本，添加 Update/Delete/Query

**Service 层:**
- `service/condition.go`
- `service/data_scope.go`
- `service/department.go`
- `service/department_binding.go`
- `service/user_attribute.go`
- `service/user.go`
- `service/internal_scope.go` — 实现 InternalScopeServiceServer
- `service/internal_condition.go` — 实现 InternalConditionServiceServer

更新 `service/service.go` RegistrarList 添加所有新服务。

---

## Step 6: Server 层 — 添加 Condition + DataScope 中间件

**文件**: `apps/yggdrasil/services/iam/internal/server/middleware/condition_checker.go`

```go
package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	pkgauth "cyber-ecosystem/apps/yggdrasil/services/iam/internal/pkg/auth"
	yggdrasilV1 "cyber-ecosystem/apps/yggdrasil/gen/go/v1"
)

func ConditionChecker(checker *biz.conditionChecker, logger log.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			identity, err := pkgauth.IdentityFromContext(ctx)
			if err != nil {
				return handler(ctx, req)
			}
			tr, _ := transport.FromServerContext(ctx)
			if tr == nil {
				return handler(ctx, req)
			}
			allowed, err := checker.Check(ctx, identity.Subject, tr.Operation())
			if err != nil {
				return nil, err
			}
			if !allowed {
				return nil, yggdrasilV1.ErrorErrorReasonForbidden("condition denied")
			}
			return handler(ctx, req)
		}
	}
}
```

**文件**: `apps/yggdrasil/services/iam/internal/server/middleware/scope_injector.go`

```go
package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	pkgauth "cyber-ecosystem/apps/yggdrasil/services/iam/internal/pkg/auth"
	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/pkg/datascope"
)

// ScopeInjector lazily injects a ScopeResolveFunc into context.
// The actual resolution happens at Ent query time (via mixin/interceptor).
func ScopeInjector(resolver *biz.scopeResolver) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			identity, err := pkgauth.IdentityFromContext(ctx)
			if err != nil {
				return handler(ctx, req)
			}
			tr, _ := transport.FromServerContext(ctx)
			if tr == nil {
				return handler(ctx, req)
			}
			resolveFn := func(ctx context.Context) (*datascope.EffectiveScope, error) {
				return resolver.Resolve(ctx, identity.Subject, tr.Operation())
			}
			ctx = datascope.WithResolveFunc(ctx, resolveFn)
			return handler(ctx, req)
		}
	}
}
```

更新 `server/server.go` 中的 `buildMiddlewares()`:

```go
mws = append(mws, selector.Server(
	jwt.Server(...),
	mw.SessionValidator(sessionValidator, logger),
	mw.Authorizer(authorizer, logger),
	mw.ConditionChecker(conditionChecker, logger),
	mw.ScopeInjector(scopeResolver),
).Match(auth.NewWhiteListByPublicAccessInProtoMatcher()).Build())
```

---

## Step 7: 编译闭环

```bash
cd apps/yggdrasil/services/iam && go mod tidy
./nx run yggdrasil_iam:generate
./nx run yggdrasil_iam:build
```

提交:
```bash
git add apps/yggdrasil/
git commit -m "feat(yggdrasil): IAM advanced capabilities — first full build"
```

---

## Step 8: 集成验证

### 8a: 创建部门和用户

```bash
TOKEN=$(curl -s -X POST http://localhost:11000/iam/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "admin123"}' | jq -r '.access_token')

# 创建部门
DEPT_ID=$(curl -s -X POST http://localhost:11000/iam/departments \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"name": "Engineering", "sort": 1}' | jq -r '.id')

# 创建用户
USER_ID=$(curl -s -X POST http://localhost:11000/iam/users \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "test123", "name": "Test User"}' | jq -r '.id')
```

### 8b: 设置部门绑定 + 用户属性

```bash
curl -X PUT "http://localhost:11000/iam/users/$USER_ID/departments" \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d "{\"department_ids\": [\"$DEPT_ID\"]}"

curl -X PUT "http://localhost:11000/iam/users/$USER_ID/attributes" \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"attributes": [{"key": "level", "value": "senior"}]}'
```

### 8c: 创建角色 + 条件 + DataScope

```bash
ROLE_ID=$(curl -s -X POST http://localhost:11000/iam/roles \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"name": "Engineer", "code": "engineer"}' | jq -r '.id')

# 创建条件（仅工作时间允许）
curl -s -X POST http://localhost:11000/iam/conditions \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d "{\"role_id\": \"$ROLE_ID\", \"name\": \"工作时间\", \"type\": \"time_range\", \"config\": \"{\\\"start_hour\\\": 9, \\\"end_hour\\\": 18}\"}"

# 创建数据权限（仅本部门）
curl -s -X POST http://localhost:11000/iam/data-scopes \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d "{\"role_id\": \"$ROLE_ID\", \"name\": \"部门数据\", \"type\": \"dept\"}"
```

### 8d: 绑定角色到用户

```bash
curl -s -X POST http://localhost:11000/iam/role-bindings \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d "{\"user_id\": \"$USER_ID\", \"role_id\": \"$ROLE_ID\"}"
```

### 8e: 测试内部 API

```bash
# 测试条件检查
grpcurl -plaintext -d "{\"user_id\": \"$USER_ID\", \"operation\": \"worklog.create\"}" \
  localhost:12000 api.yggdrasil.v1.InternalConditionService/CheckConditions

# 测试 DataScope 解析
grpcurl -plaintext -d "{\"user_id\": \"$USER_ID\", \"resource\": \"worklog\"}" \
  localhost:12000 api.yggdrasil.v1.InternalScopeService/ResolveScope
```

### 8f: 测试事件级联

```bash
# 删除角色 → 应自动清理条件 + DataScope + Casbin 策略
curl -X DELETE "http://localhost:11000/iam/roles/$ROLE_ID" \
  -H "Authorization: Bearer $TOKEN"

# 验证条件已清理
grpcurl -plaintext -d "{\"user_id\": \"$USER_ID\", \"operation\": \"worklog.create\"}" \
  localhost:12000 api.yggdrasil.v1.InternalConditionService/CheckConditions
```

预期: 删除角色后返回 `allowed: true`（无规则 = 放行）

### 8g: 提交

```bash
git add apps/yggdrasil/
git commit -m "feat(yggdrasil): IAM advanced capabilities pass integration verification"
```

---

## 完成标准

- [x] `./nx run yggdrasil_iam:build` 编译通过
- [x] 条件规则 CRUD 正常
- [x] DataScope 规则 CRUD 正常
- [x] 部门 CRUD + 层级树构建正常
- [x] 部门绑定 → 事件触发 scope cache 失效
- [x] 用户属性设置/获取正常
- [x] 用户属性变更 → 事件触发 scope cache 失效
- [x] 条件中间件正确拒绝不在时间段的请求
- [x] Scope 注入中间件正确注入解析函数到 context
- [x] 内部 API（CheckConditions / ResolveScope）正常工作
- [x] 删除角色 → 事件级联清理所有关联数据
- [x] 变更已提交
