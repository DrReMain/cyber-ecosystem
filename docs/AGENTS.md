# AI Agent 开发指南

> 本文档专为 AI Agent（如 Claude、GPT 等）优化，遵循 [Anthropic AGENTS.md 最佳实践](https://docs.anthropic.com/en/docs/agents-and-tools/agent-development-guide)。

## 元信息

| 属性 | 值 |
|------|-----|
| 项目类型 | Go Microservices (Kratos) |
| 构建系统 | Nx Monorepo |
| 框架版本 | Kratos v2.9.2 |
| Go 版本 | 1.21+ |
| 文档版本 | 1.0.0 |

---

## ⚠️ 关键约束

### 命令执行规范（最高优先级）

```
【强制】所有子项目命令必须通过 nx 执行，严禁执行裸命令。
```

#### 正确的命令格式

```bash
# macOS / Linux
./nx run <project-name>:<target>

# Windows
nx.bat run <project-name>:<target>
```

#### 禁止的命令格式

```bash
# ❌ 禁止直接执行
cd examples/template1 && go generate ./...
cd examples/template1 && kratos run
cd examples/template1 && buf generate
go generate ./...
kratos run
buf generate
```

#### 项目名称映射

| 目录 | 项目名称 |
|------|----------|
| `contracts` | `contracts` |
| `examples/template1` | `examples-template1` |
| `examples/template2` | `examples-template2` |
| `gen` | `gen` |
| `shared-go` | `shared-go` |
| `tools` | `tools` |

---

## 项目结构

```
cyber-ecosystem/
├── contracts/                  # 契约定义
│   └── proto/                 # 通用 Proto
│       ├── common/            # 公共类型
│       └── errors/            # 错误模型
│
├── examples/                   # 示例项目
│   ├── template1/             # Blog CRUD 服务示例
│   └── template2/             # Reading 服务示例
│
├── gen/                        # 生成代码
│   ├── go/                     # 生成的 Go 代码
│   │   ├── common/             # 公共类型
│   │   ├── template1/v1/       # template1 服务生成代码
│   │   └── template2/v1/       # template2 服务生成代码
│   └── oas/                    # OpenAPI 规范
│
├── shared-go/                  # 共享 Go 库
│   ├── kratos/
│   │   ├── encoder/            # 响应编码器
│   │   ├── masks/              # fields_mask 处理
│   │   ├── middleware/         # 中间件
│   │   ├── order_by/           # 排序工具
│   │   └── util/               # 工具函数
│   └── orm/ent/entutil/        # Ent 工具函数
│
├── clients/                    # 客户端项目
│   └── admin/                  # Next.js 管理后台
│
└── tools/                      # 开发工具
    └── docker/                 # Docker Compose
```

---

## 模板对齐基线

以下约定以 `examples/template1` 和 `examples/template2` 的真实实现为准，新增服务时优先对齐：

- Proto 包名使用 `package api.<service>.v1;`。
- `go_package` 使用 `github.com/DrReMain/cyber-ecosystem/gen/go/<service>/v1;<service>V1`。
- 公共 Proto 导入使用 `common/common.proto`。
- 排序工具导入路径固定为 `github.com/DrReMain/cyber-ecosystem/shared-go/kratos/order_by`。
- Service 查询分页统一先调用 `util.GetOrBuildPage(in.Page)`。
- Create 接口默认可返回空响应；只有契约明确要求时才返回 `id` 等新增结果。
- Biz 只在需要原子编排时显式开启事务；Repo 内可通过 `data.InTx` 按需开启事务，并自动复用上层事务。
- 跨服务调用落在 Data 层：远程 gRPC Client 挂在 `Data` 结构体，由 `NewData`/`NewXxxService` 注入。
- Service 必须实现 `RegisterGRPC` / `RegisterHTTP`，并由 `NewRegistrarList` 聚合注册。

---

## 可用命令

### 服务项目命令

| 命令 | 说明 | 示例 |
|------|------|------|
| `proto:api` | 生成 API Proto 代码 | `./nx run examples-template1:proto:api` |
| `proto:conf` | 生成配置 Proto 代码 | `./nx run examples-template1:proto:conf` |
| `proto` | 生成所有 Proto 代码 | `./nx run examples-template1:proto` |
| `generate` | 完整代码生成 | `./nx run examples-template1:generate` |
| `build` | 构建项目 | `./nx run examples-template1:build` |
| `dev` | 开发模式运行 | `./nx run examples-template1:dev` |
| `ent:new` | 新增 Ent Schema | `./nx run examples-template1:ent:new --args="Entity=Blog"` |

### 工具命令

| 命令 | 说明 |
|------|------|
| `tools:go:init` | 安装 Go 工具链 |
| `tools:docker:up` | 启动所有基础设施 |
| `tools:docker:down` | 停止所有基础设施 |
| `tools:docker:postgres` | 启动 PostgreSQL |
| `tools:buf:format` | Buf 格式化 |

---

## 架构决策树

### 新增 API 接口

```
用户请求：新增 XXX 接口
    │
    ├─→ 1. 检查是否需要新 Proto 文件？
    │       │
    │       ├─ 是 → 创建 api/v1/xxx.proto
    │       │       运行 ./nx run <project>:proto:api
    │       │
    │       └─ 否 → 在现有 proto 中添加 rpc
    │
    ├─→ 2. 实现 Service 层
    │       文件：internal/service/xxx.go
    │       职责：协议转换（Proto ↔ Biz Entity）
    │
    ├─→ 3. 实现 Biz 层
    │       文件：internal/biz/xxx.go
    │       职责：业务逻辑、事务边界
    │       定义：Entity、QueryIn/Out、Repo 接口
    │
    ├─→ 4. 实现 Data 层
    │       文件：internal/data/xxx.go
    │       职责：数据访问、Ent 操作
    │
    └─→ 5. 更新 Wire 注册
            文件：各层 service.go / biz.go / data.go
            添加：ProviderSet、NewRegistrarList
```

### 新增数据实体

```
用户请求：新增 XXX 数据表
    │
    ├─→ 1. 创建 Ent Schema
    │       命令：./nx run <project>:ent:new --args="Entity=XXX"
    │       文件：internal/data/ent/schema/xxx.go
    │
    ├─→ 2. 生成 Ent 代码
    │       命令：./nx run <project>:generate
    │
    ├─→ 3. 定义 Biz 层
    │       文件：internal/biz/xxx.go
    │       定义：XXXEntity、XXXQueryIn/Out、XXXRP 接口
    │
    └─→ 4. 实现 Data 层
            文件：internal/data/xxx.go
            实现：NewXXXRP、CRUD 方法
```

### 新增微服务

```
用户请求：新增 XXX 微服务
    │
    ├─→ 1. 复制模板服务
    │       cp -r examples/template1 examples/xxx
    │
    ├─→ 2. 修改 project.json
    │       name: "examples-xxx"
    │
    ├─→ 3. 修改 Proto / 包名 / go_package
    │       文件：api/v1/xxx.proto
    │       检查：package / go_package / service 名称
    │
    ├─→ 4. 修改配置与装配
    │       文件：configs/config.yaml、internal/conf/conf.proto、
    │             cmd/app/wire.go、internal/*/ProviderSet
    │       修改：端口、数据库名、远程服务依赖
    │
    └─→ 5. 生成代码
            命令：./nx run examples-xxx:generate
```

---

## 代码模板

### Proto 定义模板

```protobuf
syntax = "proto3";

package api.{{entity}}.v1;

import "google/api/annotations.proto";
import "google/protobuf/timestamp.proto";
import "google/protobuf/wrappers.proto";
import "buf/validate/validate.proto";
import "common/common.proto";

option go_package = "github.com/DrReMain/cyber-ecosystem/gen/go/{{entity}}/v1;{{entity}}V1";

// ========== Request ==========

message Create{{Entity}}Request {
  // 必填字段
  string name = 1 [(buf.validate.field).string.min_len = 1];
  // 可选字段
  optional string description = 2;
}

message Update{{Entity}}Request {
  string id = 1 [(buf.validate.field).string.len = 20];
  
  optional string name = 2;
  optional string description = 3;
  
  repeated string fields_mask = 100 [(buf.validate.field).repeated = {
    min_items: 1,
    unique: true
  }];
}

message Delete{{Entity}}Request {
  string id = 1 [(buf.validate.field).string.len = 20];
}

message Get{{Entity}}Request {
  string id = 1 [(buf.validate.field).string.len = 20];
}

message Query{{Entity}}Request {
  common.PageRequest page = 1;
  
  optional string id = 2;
  optional string name = 3;
  
  repeated string order_by = 100 [(buf.validate.field).cel = {
    id: "Query{{Entity}}Request.order_by",
    message: "排序格式错误",
    expression: "this.all(item, size(item) == 0 || item.matches('^(created_at|updated_at):(asc|desc)$'))"
  }];
}

// ========== Response ==========

message Create{{Entity}}Response {
  string id = 1;
}

message Update{{Entity}}Response {}

message Delete{{Entity}}Response {}

message Get{{Entity}}Response {
  string id = 1;
  google.protobuf.Timestamp created_at = 2;
  google.protobuf.Timestamp updated_at = 3;
  google.protobuf.StringValue name = 4;
  google.protobuf.StringValue description = 5;
}

message Query{{Entity}}Response {
  common.PageResponse page = 1;
  repeated Get{{Entity}}Response list = 2;
}

// ========== Service ==========

service {{Entity}}Service {
  rpc Create{{Entity}}(Create{{Entity}}Request) returns (Create{{Entity}}Response) {
    option (google.api.http) = {
      post: "/api/v1/{{entity}}"
      body: "*"
    };
  }
  
  rpc Update{{Entity}}(Update{{Entity}}Request) returns (Update{{Entity}}Response) {
    option (google.api.http) = {
      put: "/api/v1/{{entity}}/{id}"
      body: "*"
    };
  }
  
  rpc Delete{{Entity}}(Delete{{Entity}}Request) returns (Delete{{Entity}}Response) {
    option (google.api.http) = { delete: "/api/v1/{{entity}}/{id}" };
  }
  
  rpc Get{{Entity}}(Get{{Entity}}Request) returns (Get{{Entity}}Response) {
    option (google.api.http) = { get: "/api/v1/{{entity}}/{id}" };
  }
  
  rpc Query{{Entity}}(Query{{Entity}}Request) returns (Query{{Entity}}Response) {
    option (google.api.http) = { get: "/api/v1/{{entity}}" };
  }
}
```

### Service 层模板

```go
// internal/service/{{entity}}.go
package service

import (
    "context"

    v1 "github.com/DrReMain/cyber-ecosystem/gen/go/{{entity}}/v1"
    "github.com/DrReMain/cyber-ecosystem/shared-go/kratos/util"
    "github.com/DrReMain/cyber-ecosystem/shared-go/kratos/order_by"
    "github.com/DrReMain/cyber-ecosystem/examples/{{entity}}/internal/biz"

    "github.com/go-kratos/kratos/v2/transport/grpc"
    "github.com/go-kratos/kratos/v2/transport/http"
    "google.golang.org/protobuf/types/known/wrapperspb"
)

type {{Entity}}Service struct {
    v1.Unimplemented{{Entity}}ServiceServer
    {{entity}}UC *biz.{{Entity}}UC
}

func New{{Entity}}Service({{entity}}UC *biz.{{Entity}}UC) *{{Entity}}Service {
    return &{{Entity}}Service{{entity}}UC: {{entity}}UC}
}

func (s *{{Entity}}Service) RegisterGRPC(srv *grpc.Server) {
    v1.Register{{Entity}}ServiceServer(srv, s)
}

func (s *{{Entity}}Service) RegisterHTTP(srv *http.Server) {
    v1.Register{{Entity}}ServiceHTTPServer(srv, s)
}

// Create{{Entity}} 新增
func (s *{{Entity}}Service) Create{{Entity}}(ctx context.Context, in *v1.Create{{Entity}}Request) (*v1.Create{{Entity}}Response, error) {
    entity := &biz.{{Entity}}Entity{
        Name:        in.Name,
        Description: in.Description,
    }
    if err := s.{{entity}}UC.Create{{Entity}}(ctx, entity); err != nil {
        return nil, err
    }
    return &v1.Create{{Entity}}Response{Id: entity.ID}, nil
}

// Get{{Entity}} 详情
func (s *{{Entity}}Service) Get{{Entity}}(ctx context.Context, in *v1.Get{{Entity}}Request) (*v1.Get{{Entity}}Response, error) {
    entity, err := s.{{entity}}UC.Get{{Entity}}(ctx, in.Id)
    if err != nil {
        return nil, err
    }
    return &v1.Get{{Entity}}Response{
        Id:          entity.ID,
        Name:        util.ToPtrWrapper(entity.Name, wrapperspb.String),
        Description: util.ToPtrWrapper(entity.Description, wrapperspb.String),
        CreatedAt:   util.GetPPbTimeFromPTime(entity.CreatedAt),
        UpdatedAt:   util.GetPPbTimeFromPTime(entity.UpdatedAt),
    }, nil
}

// Query{{Entity}} 分页查询
func (s *{{Entity}}Service) Query{{Entity}}(ctx context.Context, in *v1.Query{{Entity}}Request) (*v1.Query{{Entity}}Response, error) {
    out, err := s.{{entity}}UC.Query{{Entity}}(ctx, &biz.{{Entity}}QueryIn{
        PageRequest: util.GetOrBuildPage(in.Page),
        OrderBy:     order_by.ParseOrderBy(in.OrderBy),
        ID:          in.Id,
        Name:        in.Name,
    })
    if err != nil {
        return nil, err
    }
    return &v1.Query{{Entity}}Response{
        Page: out.PageResponse,
        List: func() []*v1.Get{{Entity}}Response {
            result := make([]*v1.Get{{Entity}}Response, len(out.List))
            for i, entity := range out.List {
                result[i] = &v1.Get{{Entity}}Response{
                    Id:          entity.ID,
                    Name:        util.ToPtrWrapper(entity.Name, wrapperspb.String),
                    Description: util.ToPtrWrapper(entity.Description, wrapperspb.String),
                    CreatedAt:   util.GetPPbTimeFromPTime(entity.CreatedAt),
                    UpdatedAt:   util.GetPPbTimeFromPTime(entity.UpdatedAt),
                }
            }
            return result
        }(),
    }, nil
}
```

### Biz 层模板

```go
// internal/biz/{{entity}}.go
package biz

import (
    "context"
    "time"

    "github.com/DrReMain/cyber-ecosystem/gen/go/common"
    "github.com/DrReMain/cyber-ecosystem/shared-go/kratos/order_by"
)

// {{Entity}}Entity 业务实体
type {{Entity}}Entity struct {
    ID          string
    Name        *string
    Description *string
    CreatedAt   *time.Time
    UpdatedAt   *time.Time
}

// {{Entity}}QueryIn 查询入参
type {{Entity}}QueryIn struct {
    *common.PageRequest
    OrderBy []*order_by.OrderBy
    ID      *string
    Name    *string
}

// {{Entity}}QueryOut 查询出参
type {{Entity}}QueryOut struct {
    *common.PageResponse
    List []*{{Entity}}Entity
}

// {{Entity}}RP 仓储接口
type {{Entity}}RP interface {
    Create(ctx context.Context, entity *{{Entity}}Entity) error
    Update(ctx context.Context, fieldsMask []string, entity *{{Entity}}Entity) error
    Delete(ctx context.Context, id string) error
    Get(ctx context.Context, id string) (*{{Entity}}Entity, error)
    Query(ctx context.Context, bo *{{Entity}}QueryIn) (*{{Entity}}QueryOut, error)
}

type {{Entity}}UC struct {
    tm       Transaction
    {{entity}}RP {{Entity}}RP
}

func New{{Entity}}UC(tm Transaction, {{entity}}RP {{Entity}}RP) *{{Entity}}UC {
    return &{{Entity}}UC{tm: tm, {{entity}}RP: {{entity}}RP}
}

func (uc *{{Entity}}UC) Create{{Entity}}(ctx context.Context, entity *{{Entity}}Entity) error {
    return uc.{{entity}}RP.Create(ctx, entity)
}

func (uc *{{Entity}}UC) Update{{Entity}}(ctx context.Context, fieldsMask []string, entity *{{Entity}}Entity) error {
    return uc.tm.InTx(ctx, func(ctx context.Context) error {
        return uc.{{entity}}RP.Update(ctx, fieldsMask, entity)
    })
}

func (uc *{{Entity}}UC) Delete{{Entity}}(ctx context.Context, id string) error {
    return uc.{{entity}}RP.Delete(ctx, id)
}

func (uc *{{Entity}}UC) Get{{Entity}}(ctx context.Context, id string) (*{{Entity}}Entity, error) {
    return uc.{{entity}}RP.Get(ctx, id)
}

func (uc *{{Entity}}UC) Query{{Entity}}(ctx context.Context, bo *{{Entity}}QueryIn) (*{{Entity}}QueryOut, error) {
    return uc.{{entity}}RP.Query(ctx, bo)
}
```

### Data 层模板

#### Data 结构体（含 gRPC 客户端）

> **重要**：gRPC 客户端必须挂载在 `Data` 结构体上，由 `NewData` 统一创建和管理生命周期。这遵循 Kratos 官方 beer-shop 示例模式。

```go
// internal/data/data.go
package data

import (
    "context"
    "os"

    "github.com/DrReMain/cyber-ecosystem/examples/template2/internal/biz"
    "github.com/DrReMain/cyber-ecosystem/examples/template2/internal/conf"
    "github.com/DrReMain/cyber-ecosystem/examples/template2/internal/data/ent"
    "github.com/DrReMain/cyber-ecosystem/examples/template2/internal/data/ent/migrate"
    _ "github.com/DrReMain/cyber-ecosystem/examples/template2/internal/data/ent/runtime"

    template1V1 "github.com/DrReMain/cyber-ecosystem/gen/go/template1/v1"

    "github.com/DrReMain/cyber-ecosystem/shared-go/orm/ent/client"
    "github.com/DrReMain/cyber-ecosystem/shared-go/orm/ent/entutil"

    "github.com/go-kratos/kratos/v2/log"
    "github.com/go-kratos/kratos/v2/middleware/logging"
    "github.com/go-kratos/kratos/v2/middleware/recovery"
    "github.com/go-kratos/kratos/v2/middleware/tracing"
    "github.com/go-kratos/kratos/v2/transport/grpc"

    "github.com/google/wire"
    tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

// Data 结构体包含数据库客户端和 gRPC 客户端
type Data struct {
    db              *ent.Client
    template1Client template1V1.BlogServiceClient  // gRPC 客户端
}

// NewData 创建 Data 实例（包含数据库和 gRPC 连接）
func NewData(c *conf.Bootstrap, logger log.Logger, tp *tracesdk.TracerProvider) (*Data, func(), error) {
    // 1. 创建数据库连接
    drv, err := client.NewEntClient(client.DBConfig{
        Driver:   c.Data.Database.Driver,
        Host:     c.Data.Database.Host,
        Port:     int(c.Data.Database.Port),
        User:     c.Data.Database.User,
        Password: c.Data.Database.Password,
        DBName:   c.Data.Database.DbName,
    })
    if err != nil {
        log.Fatalf("failed opening connection to database: %v", err)
    }
    db := ent.NewClient(ent.Driver(drv))

    // 2. 创建 gRPC 客户端连接
    conn, err := grpc.DialInsecure(
        context.Background(),
        grpc.WithEndpoint(c.Data.ServiceTemplate1.Addr),
        grpc.WithMiddleware(
            tracing.Client(tracing.WithTracerProvider(tp)),
            recovery.Recovery(),
            logging.Client(logger),
        ),
        grpc.WithTimeout(c.Data.ServiceTemplate1.Timeout.AsDuration()),
    )
    if err != nil {
        db.Close()
        return nil, nil, err
    }

    // 3. 统一清理函数
    cleanup := func() {
        db.Close()
        conn.Close()
    }

    // 4. 返回 Data 实例
    return &Data{
        db:              db,
        template1Client: template1V1.NewBlogServiceClient(conn),
    }, cleanup, nil
}

// Wire 注册
var ProviderSet = wire.NewSet(
    NewData,
    wire.Bind(new(biz.Transaction), new(*Data)),
    NewReadingRP,
    NewTemplateClient,  // 注册 TemplateClient
)
```

#### TemplateClient 包装器

```go
// internal/data/reading.go

type templateClient struct {
    data *Data
}

func NewTemplateClient(data *Data) biz.TemplateClient {
    return &templateClient{data: data}
}

func (tc *templateClient) GetBlog(ctx context.Context, in *template1V1.GetBlogRequest) (*template1V1.GetBlogResponse, error) {
    return tc.data.template1Client.GetBlog(ctx, in)
}

func (tc *templateClient) QueryBlog(ctx context.Context, in *template1V1.QueryBlogRequest) (*template1V1.QueryBlogResponse, error) {
    return tc.data.template1Client.QueryBlog(ctx, in)
}
```

#### Repo 实现模板

```go
// internal/data/{{entity}}.go
package data

import (
    "context"

    "github.com/DrReMain/cyber-ecosystem/examples/template1/internal/biz"
    "github.com/DrReMain/cyber-ecosystem/examples/template1/internal/data/ent"
    "github.com/DrReMain/cyber-ecosystem/examples/template1/internal/data/ent/{{entity}}"
    "github.com/DrReMain/cyber-ecosystem/examples/template1/internal/data/ent/schema"
    "github.com/DrReMain/cyber-ecosystem/shared-go/orm/ent/entutil"

    "entgo.io/ent/dialect/sql"
)

type {{entity}}RP struct {
    data *Data
}

func New{{Entity}}RP(data *Data) biz.{{Entity}}RP {
    return &{{entity}}RP{data: data}
}

func (rp *{{entity}}RP) Create(ctx context.Context, entity *biz.{{Entity}}Entity) error {
    if err := rp.data.InTx(ctx, func(ctx context.Context) error {
        client := rp.data.getClient(ctx)
        if err := client.{{Entity}}.Create().
            SetNillableName(entity.Name).
            SetNillableDescription(entity.Description).
            Exec(ctx); err != nil {
            return entutil.HandleError(err)
        }
        return nil
    }); err != nil {
        return err
    }
    return nil
}

func (rp *{{entity}}RP) Update(ctx context.Context, fieldsMask []string, entity *biz.{{Entity}}Entity) error {
    builder := rp.data.getClient(ctx).{{Entity}}.UpdateOneID(entity.ID)
    entutil.MasksHandler{
        "name": {
            entity.Name == nil,
            func() { builder.SetName(schema.{{Entity}}DefaultName()) },
            func() { builder.SetName(*entity.Name) },
        },
        "description": {
            entity.Description == nil,
            func() { builder.SetDescription(schema.{{Entity}}DefaultDescription()) },
            func() { builder.SetDescription(*entity.Description) },
        },
    }.Emit(fieldsMask)
    if err := builder.Exec(ctx); err != nil {
        return entutil.HandleError(err)
    }
    return nil
}

func (rp *{{entity}}RP) Delete(ctx context.Context, id string) error {
    if err := rp.data.getClient(ctx).{{Entity}}.DeleteOneID(id).Exec(ctx); err != nil {
        return entutil.HandleError(err)
    }
    return nil
}

func (rp *{{entity}}RP) Get(ctx context.Context, id string) (*biz.{{Entity}}Entity, error) {
    po, err := rp.data.getClient(ctx).{{Entity}}.Query().
        Where({{entity}}.IDEQ(id)).
        Only(ctx)
    if err != nil {
        return nil, entutil.HandleError(err)
    }
    return &biz.{{Entity}}Entity{
        ID:          po.ID,
        Name:        &po.Name,
        Description: &po.Description,
        CreatedAt:   &po.CreatedAt,
        UpdatedAt:   &po.UpdatedAt,
    }, nil
}

func (rp *{{entity}}RP) Query(ctx context.Context, bo *biz.{{Entity}}QueryIn) (*biz.{{Entity}}QueryOut, error) {
    query := rp.data.getClient(ctx).{{Entity}}.Query()
    
    entutil.WherePtr(query, bo.ID, {{entity}}.IDEQ)
    entutil.WherePtr(query, bo.Name, {{entity}}.NameContainsFold)
    
    entutil.ApplyOrderBy(bo.OrderBy, entutil.FOMapping{
        "created_at": func(sel entutil.SQLSelector) { query.Order(sel({{entity}}.FieldCreatedAt)) },
        "updated_at": func(sel entutil.SQLSelector) { query.Order(sel({{entity}}.FieldUpdatedAt)) },
    })

    total, offset, limit, err := entutil.ApplyPagination(ctx, query, bo.PageRequest,
        entutil.NewPageConfig(entutil.DefaultPageSize, entutil.DefaultPageSizeUnlimit))
    if err != nil {
        return nil, entutil.HandleError(err)
    }

    pos, err := query.All(ctx)
    if err != nil {
        return nil, entutil.HandleError(err)
    }
    
    return &biz.{{Entity}}QueryOut{
        PageResponse: entutil.BuildPageResponse(total, offset, limit),
        List: func() []*biz.{{Entity}}Entity {
            result := make([]*biz.{{Entity}}Entity, len(pos))
            for i, v := range pos {
                result[i] = &biz.{{Entity}}Entity{
                    ID:          v.ID,
                    Name:        &v.Name,
                    Description: &v.Description,
                    CreatedAt:   &v.CreatedAt,
                    UpdatedAt:   &v.UpdatedAt,
                }
            }
            return result
        }(),
    }, nil
}
```

### Ent Schema 模板

```go
// internal/data/ent/schema/{{entity}}.go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
)

type {{Entity}} struct {
    ent.Schema
}

func ({{Entity}}) Mixin() []ent.Mixin {
    return []ent.Mixin{
        // 可选：添加 IDMixin、SoftDeleteMixin 等
    }
}

func ({{Entity}}) Fields() []ent.Field {
    return []ent.Field{
        field.String("name").NotEmpty().Comment("名称"),
        field.String("description").Optional().Comment("描述"),
    }
}

func ({{Entity}}) Indexes() []ent.Index {
    return []ent.Index{
        // 可选：添加索引
    }
}
```

---

## 工具函数参考

### 类型转换

| 函数 | 签名 | 用途 |
|------|------|------|
| `util.GetPTimeFromPPbTime` | `*timestamppb.Timestamp → *time.Time` | Proto 时间转 Go 时间 |
| `util.GetPPbTimeFromPTime` | `*time.Time → *timestamppb.Timestamp` | Go 时间转 Proto 时间 |
| `util.ToPtrWrapper` | `*T → *wrapperspb.XxxValue` | 指针转 Proto Wrapper |
| `util.GetStringFromWrapper` | `*wrapperspb.StringValue → *string` | Proto Wrapper 转指针 |
| `order_by.ParseOrderBy` | `[]string → []*order_by.OrderBy` | 解析排序参数 |

### Ent 工具

| 函数 | 用途 |
|------|------|
| `entutil.HandleError` | 统一转换 Ent 错误为 Kratos 错误 |
| `entutil.MasksHandler.Emit` | 处理 fields_mask 局部更新 |
| `entutil.WherePtr` | 条件拼装（指针版本） |
| `entutil.Where` | 条件拼装（布尔版本） |
| `entutil.ApplyOrderBy` | 应用排序 |
| `entutil.ApplyPagination` | 应用分页 |
| `entutil.BuildPageResponse` | 构建分页响应 |

---

## 错误处理

### Ent 错误映射

| Ent 错误 | Kratos 错误 | 错误码 |
|----------|-------------|--------|
| `IsNotFound` | `NotFound` | `ENT_NOT_FOUND_ERROR` |
| `IsValidationError` | `BadRequest` | `ENT_VALIDATION_ERROR` |
| `IsNotSingular` | `BadRequest` | `ENT_NOT_SINGULAR_ERROR` |
| `IsNotLoaded` | `InternalServer` | `ENT_NOT_LOADED_ERROR` |
| `IsConstraintError` | `Conflict` | `ENT_CONSTRAINT_ERROR` |

---

## 常见任务清单

### 新增 CRUD 接口

- [ ] 1. 定义 Proto（`api/v1/xxx.proto`）
- [ ] 2. 运行 `./nx run <project>:proto:api`
- [ ] 3. 实现 Service 层（`internal/service/xxx.go`）
- [ ] 4. 定义 Biz 层（`internal/biz/xxx.go`）
- [ ] 5. 实现 Data 层（`internal/data/xxx.go`）
- [ ] 6. 更新 Wire 注册（各层 `ProviderSet`）
- [ ] 7. 运行 `./nx run <project>:generate`
- [ ] 8. 运行 `./nx run <project>:build` 验证

### 新增数据表

- [ ] 1. 运行 `./nx run <project>:ent:new --args="Entity=XXX"`
- [ ] 2. 编辑 Schema（`internal/data/ent/schema/xxx.go`）
- [ ] 3. 运行 `./nx run <project>:generate`
- [ ] 4. 实现 Biz 层
- [ ] 5. 实现 Data 层
- [ ] 6. 更新 Wire 注册

### 新增微服务

- [ ] 1. 复制模板：`cp -r examples/template1 examples/xxx`
- [ ] 2. 修改 `project.json` 中的 `name`
- [ ] 3. 修改 Proto 定义
- [ ] 4. 修改配置文件（端口、数据库名）
- [ ] 5. 运行 `./nx run @cyber-ecosystem/kratos-xxx:generate`
- [ ] 6. 实现业务逻辑

---

## 禁止事项

1. **禁止直接执行裸命令** - 必须通过 `./nx` 执行
2. **禁止手动修改生成代码** - `wire_gen.go`、`ent/` 目录下的文件
3. **禁止在 Service 层开事务** - 事务边界在 Biz 层决定
4. **禁止在 Biz 层替换 ctx** - 不要使用 `context.Background()`
5. **禁止直接访问 Data 层** - Service 层只能调用 Biz 层
6. **禁止在 Biz 层创建 gRPC 连接** - gRPC 客户端必须挂载在 Data 结构体上，由 NewData 统一管理

---

## 验证检查

### 代码生成后验证

```bash
# 1. 检查编译
./nx run <project>:build
```

### 提交前检查

- [ ] 所有命令通过 `./nx` 执行
- [ ] 无手动修改生成代码
- [ ] Wire 注册已更新
- [ ] 编译通过
