# Kratos 微服务开发指南

本文档面向人类开发者，详细介绍基于 Kratos 框架的微服务开发规范与最佳实践。

## 目录

- [项目概述](#项目概述)
- [快速开始](#快速开始)
- [项目结构](#项目结构)
- [命令执行规范](#命令执行规范)
- [架构分层](#架构分层)
- [API 层开发规范](#api-层开发规范)
- [Service 层开发规范](#service-层开发规范)
- [Biz 层开发规范](#biz-层开发规范)
- [Data 层开发规范](#data-层开发规范)
- [Server 层开发规范](#server-层开发规范)
- [代码生成流程](#代码生成流程)
- [实战案例](#实战案例)
- [常见问题](#常见问题)

---

## 项目概述

本项目是一个基于 [Kratos](https://go-kratos.dev/) 框架的微服务生态系统，采用以下技术栈：

| 技术 | 用途 |
|------|------|
| **Kratos** | 微服务框架 |
| **Ent** | ORM 框架 |
| **Wire** | 依赖注入 |
| **Buf** | Proto 管理 |
| **Nx** | Monorepo 构建系统 |
| **PostgreSQL** | 主数据库 |

### 核心设计原则

1. **分层架构**：Service → Biz → Data，职责清晰
2. **契约优先**：Proto 定义即契约，代码自动生成
3. **事务边界**：Biz 层决定事务边界，Data 层自动复用
4. **局部更新**：通过 `fields_mask` 实现精确字段更新

---

## 与模板对齐

`template1` 和 `template2` 是当前仓库的实现基线。新增服务时默认遵循以下约定：

- Proto 包名格式固定为 `api.<service>.v1`。
- `go_package` 格式固定为 `github.com/DrReMain/cyber-ecosystem/gen/go/<service>/v1;<service>V1`。
- 公共 Proto 导入统一使用 `common/common.proto`。
- 排序工具导入路径统一使用 `github.com/DrReMain/cyber-ecosystem/shared-go/kratos/order_by`。
- Service 查询分页统一使用 `util.GetOrBuildPage(in.Page)`。
- Create 接口默认可以返回空响应；除非契约明确要求，否则不要强制返回新建记录 ID。
- Biz 层只在需要原子编排时显式开启事务；简单单仓储操作可以直接调用 Repo，由 Data 层按需开启并复用事务。
- 远程服务 gRPC Client 统一挂在 `internal/data.Data` 上，并通过 `NewData`/`NewXxxService` 注入。
- 新增 Service 后，需要同时维护 `service.ProviderSet`、`NewRegistrarList` 和 `cmd/app/wire.go` 的装配关系。

---

## 快速开始

### 环境准备

```bash
# 安装 Go 工具链
./nx run tools:go:init

# 启动基础设施（PostgreSQL、Redis 等）
./nx run tools:docker:up
```

### 开发流程

```bash
# 1. 定义 Proto（api/v1/xxx.proto）
# 2. 生成代码
./nx run examples-template1:generate

# 3. 实现业务逻辑（service/biz/data）
# 4. 本地运行
./nx run examples-template1:dev
```

---

## 项目结构

```
cyber-ecosystem/
├── nx                          # Nx 启动脚本（macOS/Linux）
├── nx.bat                      # Nx 启动脚本（Windows）
├── nx.json                     # Nx 配置
├── go.mod                      # Go 模块定义
├── buf.yaml                    # Buf 配置
│
├── contracts/                  # 契约定义
│   ├── project.json
│   └── proto/                  # 公共 Proto 定义
│       ├── common/             # 公共类型
│       └── errors/             # 错误定义
│
├── gen/                        # 生成代码（公共）
│   ├── project.json
│   ├── go/                     # 生成的 Go 代码
│   │   ├── common/             # 公共类型
│   │   ├── template1/v1/       # template1 服务生成代码
│   │   └── template2/v1/       # template2 服务生成代码
│   └── oas/                    # OpenAPI 规范
│
├── shared-go/                  # 共享 Go 库
│   ├── project.json
│   ├── kratos/                 # Kratos 相关工具
│   │   ├── encoder/            # 响应编码器
│   │   ├── masks/              # fields_mask 处理
│   │   ├── middleware/         # 中间件
│   │   ├── order_by/           # 排序工具
│   │   └── util/               # 工具函数
│   └── orm/                    # ORM 相关
│       └── ent/entutil/        # Ent 工具函数
│
├── examples/                   # 示例服务
│   ├── template1/              # 示例服务1（Blog CRUD）
│   │   ├── project.json
│   │   ├── api/v1/             # Proto 定义
│   │   ├── cmd/app/            # 应用入口
│   │   ├── configs/            # 配置文件
│   │   └── internal/
│   │       ├── biz/            # 业务逻辑层
│   │       ├── conf/           # 配置定义
│   │       ├── data/           # 数据访问层
│   │       │   └── ent/        # Ent Schema 和生成代码
│   │       ├── server/         # HTTP/gRPC 服务器
│   │       └── service/        # 服务层
│   └── template2/              # 示例服务2（Reading）
│
├── clients/                    # 前端客户端
│   └── admin/                  # 管理后台（Next.js）
│
└── tools/                      # 开发工具
    ├── project.json
    └── docker/                 # Docker Compose 配置
```

---

## 命令执行规范

### ⚠️ 重要：必须通过 Nx 执行命令

本项目使用 Nx 进行 Monorepo 管理，**所有子项目命令必须通过 `./nx`（或 `nx.bat`）执行**，严禁直接执行裸命令。

#### 正确示例

```bash
# ✅ 正确：通过 nx 执行
./nx run examples-template1:generate
./nx run examples-template1:dev
./nx run examples-template1:build
```

#### 错误示例

```bash
# ❌ 错误：直接执行裸命令
cd examples/template1 && go generate ./...
cd examples/template1 && kratos run
cd examples/template1 && buf generate
```

### 常用命令速查

| 命令 | 说明 |
|------|------|
| `./nx run <project>:proto` | 生成 Proto 代码 |
| `./nx run <project>:proto:api` | 仅生成 API Proto 代码 |
| `./nx run <project>:proto:conf` | 仅生成配置 Proto 代码 |
| `./nx run <project>:generate` | 完整代码生成（Proto + Ent + Wire） |
| `./nx run <project>:build` | 构建项目 |
| `./nx run <project>:dev` | 开发模式运行 |
| `./nx run <project>:ent:new --args="Entity=<name>"` | 新增 Ent Schema |
| `./nx run tools:docker:up` | 启动所有基础设施 |
| `./nx run tools:docker:down` | 停止所有基础设施 |
| `./nx run tools:go:init` | 安装 Go 工具链 |

### 项目名称对照

| 目录 | 项目名称 |
|------|----------|
| `examples/template1` | `examples-template1` |
| `examples/template2` | `examples-template2` |
| `gen` | `gen` |
| `shared-go` | `shared-go` |
| `tools` | `tools` |

---

## 架构分层

```
┌─────────────────────────────────────────────────────────────┐
│                        Proto Layer                           │
│  api/v1/*.proto - 定义 API 契约                              │
└─────────────────────────────────────────────────────────────┘
                              ↓ 代码生成
┌─────────────────────────────────────────────────────────────┐
│                       Service Layer                          │
│  internal/service/*.go - 协议转换（Proto ↔ Biz Entity）      │
│  职责：类型转换、参数校验（已由中间件完成）                    │
│  禁止：事务编排、直接访问 Data 层                             │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                         Biz Layer                            │
│  internal/biz/*.go - 业务逻辑、用例编排                       │
│  职责：定义实体、仓储接口、事务边界                            │
│  关键：一个业务动作需要原子性时，在此层包住                    │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                        Data Layer                            │
│  internal/data/*.go - 数据访问、Ent 操作                      │
│  职责：实现仓储接口、事务管理、错误转换                        │
│  关键：使用 getClient(ctx) 自动复用事务                       │
└─────────────────────────────────────────────────────────────┘
```

---

## API 层开发规范

### Request 定义规范

#### 可选字段使用 `optional`

```protobuf
message UpdateBlogRequest {
  string id = 1 [(buf.validate.field).string.len = 20];
  
  // 可选标量字段使用 optional，配合 fields_mask 实现局部更新
  optional string title = 2;
  optional string content = 3;
  google.protobuf.Timestamp published_at = 4;

  // fields_mask 指定要更新的字段
  repeated string fields_mask = 100 [(buf.validate.field).repeated = {
    min_items: 1
    unique: true
  }];
}
```

#### `fields_mask` 语义

| 场景 | 行为 |
|------|------|
| 字段不在 `fields_mask` 中 | 不修改 |
| 字段在 `fields_mask` 中，但值为空 | 修改为零值/默认值/清空 |
| 字段在 `fields_mask` 中，且有值 | 修改为指定值 |

### Response 定义规范

#### 可选字段使用 Wrapper 类型

```protobuf
message GetBlogResponse {
  string id = 1;
  google.protobuf.Timestamp created_at = 2;
  google.protobuf.Timestamp updated_at = 3;
  
  // 可选字段使用 wrapper，稳定输出 "有值/null"
  google.protobuf.StringValue title = 4;
  google.protobuf.StringValue content = 5;
  google.protobuf.Timestamp published_at = 6;
}
```

### 分页查询规范

```protobuf
message QueryBlogRequest {
  // 复用公共分页请求
  common.PageRequest page = 1;
  
  // 查询条件
  optional string id = 2 [(buf.validate.field).string.len = 20];
  optional string title = 3;
  google.protobuf.Timestamp published_at_a = 4;  // 开始时间
  google.protobuf.Timestamp published_at_z = 5;  // 结束时间

  // 排序：格式为 "field:(asc|desc)"
  repeated string order_by = 100 [(buf.validate.field).cel = {
    id: "QueryBlogRequest.order_by"
    message: "排序格式错误"
    expression: "this.all(item, size(item) == 0 || item.matches('^(created_at|updated_at):(asc|desc)$'))"
  }];
}
```

#### `PageRequest` 内置字段

| 字段 | 说明 |
|------|------|
| `page_no` | 页码（从 1 开始） |
| `page_size` | 每页数量 |
| `all` | 是否查询全部 |
| `created_at_a / created_at_z` | 创建时间范围 |
| `updated_at_a / updated_at_z` | 更新时间范围 |

### HTTP 路由注解

```protobuf
service BlogService {
  rpc CreateBlog(CreateBlogRequest) returns (CreateBlogResponse) {
    option (google.api.http) = {
      post: "/api/v1/blog"
      body: "*"
    };
  }
  
  rpc UpdateBlog(UpdateBlogRequest) returns (UpdateBlogResponse) {
    option (google.api.http) = {
      put: "/api/v1/blog/{id}"
      body: "*"
    };
  }
  
  rpc DeleteBlog(DeleteBlogRequest) returns (DeleteBlogResponse) {
    option (google.api.http) = { delete: "/api/v1/blog/{id}" };
  }
  
  rpc GetBlog(GetBlogRequest) returns (GetBlogResponse) {
    option (google.api.http) = { get: "/api/v1/blog/{id}" };
  }
  
  rpc QueryBlog(QueryBlogRequest) returns (QueryBlogResponse) {
    option (google.api.http) = { get: "/api/v1/blog" };
  }
}
```

### 新增模块开发顺序

1. 编写 `api/v1/xxx.proto`
2. 定义完整的 CRUD 接口
3. Request 可选字段使用 `optional`
4. Response 可选字段使用 wrapper
5. 更新接口添加 `fields_mask`
6. 查询接口复用 `common.PageRequest`
7. 需要排序时添加 `repeated string order_by`
8. 补充 `buf.validate` 和 `google.api.http` 注解
9. 运行代码生成：`./nx run <project>:api`
10. 实现 `internal/service` / `internal/biz` / `internal/data`

---

## Service 层开发规范

### 职责定位

**Service 层只做协议转换**：
- `proto request → biz entity / biz query`
- `biz result → proto response`

### 禁止事项

- ❌ 不要在 Service 层写事务编排
- ❌ 不要直接访问 Data 层

### 典型实现

```go
package service

import (
    "context"

    templateV1 "github.com/DrReMain/cyber-ecosystem/gen/go/template1/v1"
    "github.com/DrReMain/cyber-ecosystem/shared-go/kratos/util"
    "github.com/DrReMain/cyber-ecosystem/shared-go/kratos/order_by"
    "github.com/DrReMain/cyber-ecosystem/examples/template1/internal/biz"

    "github.com/go-kratos/kratos/v2/transport/grpc"
    "github.com/go-kratos/kratos/v2/transport/http"
    "google.golang.org/protobuf/types/known/wrapperspb"
)

type BlogService struct {
    templateV1.UnimplementedBlogServiceServer
    blogUC *biz.BlogUC
}

func NewBlogService(blogUC *biz.BlogUC) *BlogService {
    return &BlogService{blogUC: blogUC}
}

func (s *BlogService) RegisterGRPC(srv *grpc.Server) {
    templateV1.RegisterBlogServiceServer(srv, s)
}

func (s *BlogService) RegisterHTTP(srv *http.Server) {
    templateV1.RegisterBlogServiceHTTPServer(srv, s)
}

// CreateBlog 新增
func (s *BlogService) CreateBlog(ctx context.Context, in *templateV1.CreateBlogRequest) (*templateV1.CreateBlogResponse, error) {
    // Proto → Biz Entity
    entity := &biz.BlogEntity{
        Title:       in.Title,
        Content:     in.Content,
        PublishedAt: util.GetPTimeFromPPbTime(in.PublishedAt),
    }
    if err := s.blogUC.CreateBlog(ctx, entity); err != nil {
        return nil, err
    }
    return &templateV1.CreateBlogResponse{}, nil
}

// GetBlog 详情
func (s *BlogService) GetBlog(ctx context.Context, in *templateV1.GetBlogRequest) (*templateV1.GetBlogResponse, error) {
    entity, err := s.blogUC.GetBlog(ctx, in.Id)
    if err != nil {
        return nil, err
    }
    // Biz Entity → Proto
    return &templateV1.GetBlogResponse{
        Id:          entity.ID,
        Title:       util.ToPtrWrapper(entity.Title, wrapperspb.String),
        Content:     util.ToPtrWrapper(entity.Content, wrapperspb.String),
        PublishedAt: util.GetPPbTimeFromPTime(entity.PublishedAt),
    }, nil
}

// QueryBlog 分页查询
func (s *BlogService) QueryBlog(ctx context.Context, in *templateV1.QueryBlogRequest) (*templateV1.QueryBlogResponse, error) {
    out, err := s.blogUC.QueryBlog(ctx, &biz.BlogQueryIn{
        PageRequest:  util.GetOrBuildPage(in.Page),
        OrderBy:      order_by.ParseOrderBy(in.OrderBy),
        ID:           in.Id,
        Title:        in.Title,
        PublishedAtA: util.GetPTimeFromPPbTime(in.PublishedAtA),
        PublishedAtZ: util.GetPTimeFromPPbTime(in.PublishedAtZ),
    })
    if err != nil {
        return nil, err
    }
    // 转换响应列表
    return &templateV1.QueryBlogResponse{
        Page: out.PageResponse,
        List: func() []*templateV1.GetBlogResponse {
            result := make([]*templateV1.GetBlogResponse, len(out.List))
            for i, entity := range out.List {
                result[i] = &templateV1.GetBlogResponse{
                    Id:          entity.ID,
                    Title:       util.ToPtrWrapper(entity.Title, wrapperspb.String),
                    Content:     util.ToPtrWrapper(entity.Content, wrapperspb.String),
                    PublishedAt: util.GetPPbTimeFromPTime(entity.PublishedAt),
                    CreatedAt:   util.GetPPbTimeFromPTime(entity.CreatedAt),
                    UpdatedAt:   util.GetPPbTimeFromPTime(entity.UpdatedAt),
                }
            }
            return result
        }(),
    }, nil
}
```

### 常用转换函数

| 函数 | 用途 |
|------|------|
| `util.GetPTimeFromPPbTime(ts)` | `*timestamppb.Timestamp → *time.Time` |
| `util.GetPPbTimeFromPTime(t)` | `*time.Time → *timestamppb.Timestamp` |
| `util.ToPtrWrapper(ptr, wrapperspb.String)` | `*string → *wrapperspb.StringValue` |
| `util.GetStringFromWrapper(v)` | `*wrapperspb.StringValue → *string` |
| `order_by.ParseOrderBy(in.OrderBy)` | `[]string → []*order_by.OrderBy` |
| `util.GetOrBuildPage(in.Page)` | 标准化分页参数，补默认值 |

### Wire 注册

```go
// service/service.go
var ProviderSet = wire.NewSet(
    NewRegistrarList,
    NewBlogService,
)

func NewRegistrarList(s1 *BlogService) []Registrar {
    return []Registrar{s1}
}
```

---

## Biz 层开发规范

### 职责定位

- 定义业务实体、查询对象、仓储接口
- 定义用例编排，以及事务边界

### 实体定义

```go
// biz/blog.go

// BlogEntity 业务实体
type BlogEntity struct {
    ID          string
    Title       *string
    Content     *string
    PublishedAt *time.Time
    CreatedAt   *time.Time
    UpdatedAt   *time.Time
}

// BlogQueryIn 查询入参
type BlogQueryIn struct {
    *common.PageRequest
    OrderBy      []*order_by.OrderBy
    ID           *string
    Title        *string
    PublishedAtA *time.Time
    PublishedAtZ *time.Time
}

// BlogQueryOut 查询出参
type BlogQueryOut struct {
    *common.PageResponse
    List []*BlogEntity
}
```

### 仓储接口定义

```go
// BlogRP 仓储接口
type BlogRP interface {
    Create(ctx context.Context, entity *BlogEntity) error
    Update(ctx context.Context, fieldsMask []string, entity *BlogEntity) error
    Delete(ctx context.Context, id string) error
    DeleteBatch(ctx context.Context, ids []string) (int, error)
    Get(ctx context.Context, id string) (*BlogEntity, error)
    Query(ctx context.Context, bo *BlogQueryIn) (*BlogQueryOut, error)
}
```

### UseCase 实现

```go
type BlogUC struct {
    tm     Transaction
    blogRP BlogRP
}

func NewBlogUC(tm Transaction, blogRP BlogRP) *BlogUC {
    return &BlogUC{tm: tm, blogRP: blogRP}
}

// UpdateBlog 更新（biz 决定事务边界）
func (uc *BlogUC) UpdateBlog(ctx context.Context, fieldsMask []string, entity *BlogEntity) error {
    return uc.tm.InTx(ctx, func(ctx context.Context) error {
        return uc.blogRP.Update(ctx, fieldsMask, entity)
    })
}
```

### 事务使用约定

| 规则 | 说明 |
|------|------|
| ✅ 在 Biz 层决定事务边界 | 一个业务动作需要原子性时，在此层包住 |
| ❌ 不要在 Service 层开事务 | Service 层只做协议转换 |
| ❌ 不要替换 ctx | 不要把 `ctx` 换成 `context.Background()` |
| ✅ Data 层自动复用事务 | Repo 内调用 `InTx` 会自动复用当前事务 |

### Wire 注册

```go
// biz/biz.go
var ProviderSet = wire.NewSet(
    NewBlogUC,
)
```

---

## Data 层开发规范

### 职责定位

- 统一管理 Ent Client、gRPC 客户端和事务
- 实现 Biz 定义的仓储接口
- 使用 `entutil` 做局部更新、分页、排序、条件拼装
- 统一转换 Ent 错误

### gRPC 客户端管理

> **重要**：gRPC 客户端应该挂载在 `Data` 结构体上，由 `NewData` 统一创建和管理生命周期。这遵循 Kratos 官方 beer-shop 示例的模式。

#### Data 结构体定义

```go
// data/data.go

type Data struct {
    db              *ent.Client
    template1Client template1V1.BlogServiceClient  // gRPC 客户端
}
```

#### NewData 创建 gRPC 连接

```go
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
```

#### TemplateClient 包装器

```go
// data/reading.go

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

#### 配置文件结构

```yaml
# configs/config.yaml
data:
  database:
    driver: postgres
    host: 127.0.0.1
    port: 5432
    user: postgres
    password: postgres
    db_name: cyber_ecosystem_template2
  service_template1:        # 远程服务配置
    addr: 127.0.0.1:9000
    timeout: 1s
```

#### Wire 注册

```go
// data/data.go
var ProviderSet = wire.NewSet(
    NewData,
    wire.Bind(new(biz.Transaction), new(*Data)),
    NewReadingRP,
    NewTemplateClient,  // 注册 TemplateClient
)
```

### 事务管理

```go
// data/tx.go

// getClient 获取数据库客户端（自动复用事务）
func (d *Data) getClient(ctx context.Context) *ent.Client {
    if tx := ent.TxFromContext(ctx); tx != nil {
        return tx.Client()
    }
    return d.db
}

// InTx 在事务中执行（支持嵌套复用）
func (d *Data) InTx(ctx context.Context, fn func(context.Context) error) error {
    // 如果已在事务中，直接复用
    if tx := ent.TxFromContext(ctx); tx != nil {
        return fn(ctx)
    }

    // 开启新事务
    tx, err := d.db.Tx(ctx)
    if err != nil {
        return fmt.Errorf("failed to start a transaction: %w", err)
    }
    defer func() {
        if v := recover(); v != nil {
            _ = tx.Rollback()
            panic(v)
        }
    }()

    // 将事务注入 context
    txCtx := ent.NewTxContext(ctx, tx)

    if err := fn(txCtx); err != nil {
        if rbErr := tx.Rollback(); rbErr != nil {
            return fmt.Errorf("tx failed: %w, rollback also failed: %v", err, rbErr)
        }
        return err
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("committing transaction: %w", err)
    }
    return nil
}
```

### Repo 实现

```go
// data/blog.go

type blogRP struct {
    data *Data
}

func NewBlogRP(data *Data) biz.BlogRP {
    return &blogRP{data: data}
}

// Create 创建
func (rp *blogRP) Create(ctx context.Context, entity *biz.BlogEntity) error {
    if err := rp.data.InTx(ctx, func(ctx context.Context) error {
        client := rp.data.getClient(ctx)
        if err := client.Blog.Create().
            SetNillableTitle(entity.Title).
            SetNillableContent(entity.Content).
            SetNillablePublishedAt(entity.PublishedAt).
            Exec(ctx); err != nil {
            return entutil.HandleError(err)
        }
        return nil
    }); err != nil {
        return err
    }
    return nil
}

// Update 局部更新（fields_mask）
func (rp *blogRP) Update(ctx context.Context, fieldsMask []string, entity *biz.BlogEntity) error {
    builder := rp.data.getClient(ctx).Blog.UpdateOneID(entity.ID)
    
    // 使用 MasksHandler 处理 fields_mask
    entutil.MasksHandler{
        "title": {
            entity.Title == nil,
            func() { builder.SetTitle(schema.BlogDefaultTitle()) },
            func() { builder.SetTitle(*entity.Title) },
        },
        "content": {
            entity.Content == nil,
            func() { builder.SetContent(schema.BlogDefaultContent()) },
            func() { builder.SetContent(*entity.Content) },
        },
        "published_at": {
            entity.PublishedAt == nil,
            func() { builder.ClearPublishedAt() },
            func() { builder.SetPublishedAt(*entity.PublishedAt) },
        },
    }.Emit(fieldsMask)
    
    if err := builder.Exec(ctx); err != nil {
        return entutil.HandleError(err)
    }
    return nil
}

// Query 分页查询
func (rp *blogRP) Query(ctx context.Context, bo *biz.BlogQueryIn) (*biz.BlogQueryOut, error) {
    query := rp.data.getClient(ctx).Blog.Query()
    
    // 条件拼装
    entutil.WherePtr(query, bo.ID, blog.IDEQ)
    entutil.WherePtr(query, bo.Title, blog.TitleContainsFold)
    entutil.WherePtr(query, bo.PublishedAtA, blog.PublishedAtGTE)
    entutil.WherePtr(query, bo.PublishedAtZ, blog.PublishedAtLTE)
    
    // 排序
    entutil.ApplyOrderBy(bo.OrderBy, entutil.FOMapping{
        "created_at": func(sel entutil.SQLSelector) { query.Order(sel(blog.FieldCreatedAt)) },
        "updated_at": func(sel entutil.SQLSelector) { query.Order(sel(blog.FieldUpdatedAt)) },
    })

    // 分页
    total, offset, limit, err := entutil.ApplyPagination(ctx, query, bo.PageRequest,
        entutil.NewPageConfig(entutil.DefaultPageSize, entutil.DefaultPageSizeUnlimit))
    if err != nil {
        return nil, entutil.HandleError(err)
    }

    pos, err := query.All(ctx)
    if err != nil {
        return nil, entutil.HandleError(err)
    }
    
    return &biz.BlogQueryOut{
        PageResponse: entutil.BuildPageResponse(total, offset, limit),
        List: func() []*biz.BlogEntity {
            result := make([]*biz.BlogEntity, len(pos))
            for i, v := range pos {
                result[i] = &biz.BlogEntity{
                    ID:          v.ID,
                    Title:       &v.Title,
                    Content:     &v.Content,
                    PublishedAt: v.PublishedAt,
                    CreatedAt:   &v.CreatedAt,
                    UpdatedAt:   &v.UpdatedAt,
                }
            }
            return result
        }(),
    }, nil
}
```

### Ent 错误处理

```go
// entutil.HandleError 统一转换 Ent 错误
func HandleError(err error) error {
    switch {
    case ent.IsNotFound(err):
        return errors.NotFound("ENT_NOT_FOUND_ERROR", "未找到相关数据").WithCause(err)
    case ent.IsValidationError(err):
        return errors.BadRequest("ENT_VALIDATION_ERROR", "数据校验失败").WithCause(err)
    case ent.IsNotSingular(err):
        return errors.BadRequest("ENT_NOT_SINGULAR_ERROR", "数据不唯一").WithCause(err)
    case ent.IsNotLoaded(err):
        return errors.InternalServer("ENT_NOT_LOADED_ERROR", "数据未加载").WithCause(err)
    case ent.IsConstraintError(err):
        return errors.Conflict("ENT_CONSTRAINT_ERROR", "内容已存在，请勿重复提交").WithCause(err)
    default:
        return err
    }
}
```

### Wire 注册

```go
// data/data.go
var ProviderSet = wire.NewSet(
    NewData,
    wire.Bind(new(biz.Transaction), new(*Data)),
    NewBlogRP,
)
```

---

## Server 层开发规范

### 职责定位

- 组织 HTTP/gRPC 共用的 middleware
- 接入 proto validate 校验
- 统一成功响应/错误响应封装

### HTTP Server

```go
// server/http.go

func NewHTTPServer(c *conf.Server, services []service.Registrar, logger log.Logger) *http.Server {
    var opts = []http.ServerOption{
        http.Middleware(
            recovery.Recovery(),
            logging.Server(logger),
            validate.ProtoValidate(validate.UseProtoMessage),
        ),
        http.ResponseEncoder(encoder.NewResponseEncoder(ResponseBuildBody)),
        http.ErrorEncoder(encoder.NewErrorEncoder(ErrorBuildBody)),
    }
    
    if c.Http.Network != "" {
        opts = append(opts, http.Network(c.Http.Network))
    }
    if c.Http.Addr != "" {
        opts = append(opts, http.Address(c.Http.Addr))
    }
    if c.Http.Timeout != nil {
        opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
    }
    
    srv := http.NewServer(opts...)
    for _, svc := range services {
        svc.RegisterHTTP(srv)
    }
    return srv
}

// ResponseBuildBody 构建成功响应
func ResponseBuildBody(v any) (any, error) {
    reply := &common.Reply{
        T:       time.Now().UnixMilli(),
        Success: true,
        Msg:     "OK",
        Result:  nil,
    }
    if m, ok := v.(proto.Message); ok {
        if anyVal, err := anypb.New(m); err != nil {
            return nil, err
        } else {
            reply.Result = anyVal
        }
    }
    return reply, nil
}

// ErrorBuildBody 构建错误响应
func ErrorBuildBody(err *errors.Error) any {
    return &common.Reply{
        T:       time.Now().UnixMilli(),
        Success: false,
        Msg:     err.Message,
        Result:  nil,
    }
}
```

### gRPC Server

```go
// server/grpc.go

func NewGRPCServer(c *conf.Server, services []service.Registrar, logger log.Logger) *grpc.Server {
    var opts = []grpc.ServerOption{
        grpc.Middleware(
            recovery.Recovery(),
            logging.Server(logger),
            validate.ProtoValidate(validate.UseProtoMessage),
        ),
    }
    
    if c.Grpc.Network != "" {
        opts = append(opts, grpc.Network(c.Grpc.Network))
    }
    if c.Grpc.Addr != "" {
        opts = append(opts, grpc.Address(c.Grpc.Addr))
    }
    if c.Grpc.Timeout != nil {
        opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
    }
    
    srv := grpc.NewServer(opts...)
    for _, svc := range services {
        svc.RegisterGRPC(srv)
    }
    return srv
}
```

### Wire 注册

```go
// server/server.go
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer)
```

---

## 代码生成流程

### 1. Proto 生成

```bash
# 生成 API 相关代码
./nx run <project>:proto:api

# 生成配置相关代码
./nx run <project>:proto:conf

# 同时生成 API 和配置
./nx run <project>:proto
```

### 2. 完整生成流程

```bash
./nx run <project>:generate
```

该命令会依次执行：
1. `proto:api` - 生成 Proto API 代码
2. `proto:conf` - 生成 Proto 配置代码
3. `go generate ./...` - 生成 Ent 和 Wire 代码
4. `go mod tidy` - 整理依赖

### 3. 新增 Ent Schema

```bash
./nx run <project>:ent:new --args="Entity=Blog"
```

---

## 实战案例

### 业务需求

`template2` 服务提供博客展示和阅读量记录功能：
- 查询博客列表（带阅读量）
- 获取博客详情（不记录阅读）
- 记录阅读行为（阅读量+1）

Blog 的基础数据管理（CRUD）在 `template1` 服务中完成，`template2` 服务需调用 `template1` 服务获取博客数据。

### RESTful API 设计原则

> **重要**：GET 请求应该是安全（safe）和幂等（idempotent）的，不应该修改服务器状态。

```
GET  /api/v1/reading/blog/{id}      # 获取博客详情（不修改数据）
POST /api/v1/reading/blog/{id}/read # 记录阅读行为（阅读量+1）
```

### 开发步骤

#### 1. 定义 Proto

```protobuf
// api/v1/reading.proto
service ReadingService {
  rpc QueryBlog(QueryBlogRequest) returns (QueryBlogResponse) {
    option (google.api.http) = {get: "/api/v1/reading/blog"};
  }
  rpc GetBlog(GetBlogRequest) returns (GetBlogResponse) {
    option (google.api.http) = {get: "/api/v1/reading/blog/{id}"};
  }
  rpc RecordReading(RecordReadingRequest) returns (RecordReadingResponse) {
    option (google.api.http) = {
      post: "/api/v1/reading/blog/{id}/read"
      body: "*"
    };
  }
}
```

#### 2. 定义 Ent Schema

```go
// internal/data/ent/schema/reading.go
func (Reading) Fields() []ent.Field {
    return []ent.Field{
        field.String("blog_id").NotEmpty().Comment("文章ID"),
        field.Int64("reading_count").Comment("阅读量"),
    }
}

func (Reading) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("blog_id").Unique(),
    }
}
```

#### 3. 实现 Biz 层

```go
// internal/biz/reading.go
type BlogWithReadingEntity struct {
    ID           string
    Title        *string
    Content      *string
    PublishedAt  *time.Time
    ReadingCount int64
}

type ReadingRP interface {
    GetReadingCount(ctx context.Context, blogID string) (int64, error)
    IncrementReading(ctx context.Context, blogID string) error
}

type TemplateClient interface {
    GetBlog(ctx context.Context, in *templateV1.GetBlogRequest) (*templateV1.GetBlogResponse, error)
}

// RecordReading 记录阅读行为（阅读量+1）
func (uc *ReadingUC) RecordReading(ctx context.Context, id string) (int64, error) {
    var readingCount int64
    if err := uc.tm.InTx(ctx, func(ctx context.Context) error {
        if err := uc.readingRP.IncrementReading(ctx, id); err != nil {
            return err
        }
        count, err := uc.readingRP.GetReadingCount(ctx, id)
        if err != nil {
            return err
        }
        readingCount = count
        return nil
    }); err != nil {
        return 0, err
    }
    return readingCount, nil
}
```

#### 4. 配置文件

```yaml
# configs/config.yaml
server:
  http:
    addr: 0.0.0.0:8001
  grpc:
    addr: 0.0.0.0:9001
data:
  database:
    driver: postgres
    host: 127.0.0.1
    port: 5432
    user: postgres
    password: postgres
    db_name: cyber_ecosystem_template2
client:
  template:
    addr: 127.0.0.1:9000
    timeout: 1s
```

#### 5. 生成代码

```bash
./nx run examples-template2:generate
```

#### 6. 构建验证

```bash
./nx run examples-template2:build
```

---

## 常见问题

### Q: 为什么必须通过 Nx 执行命令？

A: 本项目采用 Monorepo 架构，Nx 提供了：
- 统一的命令入口
- 依赖关系管理
- 缓存机制
- 跨平台兼容性

### Q: 如何调试生成的代码？

A: 生成的代码位于 `gen/` 目录和 `internal/data/ent/` 目录，可以查看但**不要手动修改**。如需调整，应修改 Proto 或 Ent Schema 后重新生成。

### Q: 事务嵌套会怎样？

A: Data 层的 `InTx` 方法支持嵌套调用。如果已在事务中，会自动复用当前事务，不会重复开启。

### Q: 如何添加新的微服务？

A: 
1. 复制 `examples/template1` 目录
2. 修改 `project.json` 中的 `name`
3. 修改 Proto 定义
4. 运行 `./nx run <new-project>:generate`

### Q: fields_mask 为空会怎样？

A: `fields_mask` 为空时，`MasksHandler.Emit` 不会执行任何更新操作。建议在 Proto 层添加 `min_items: 1` 校验。

---

## 注意事项

1. **不要手动修改生成的代码**（如 `wire_gen.go`、`ent/` 目录下的文件）
2. **所有事务边界在 Biz 层决定**
3. **Service 层只做协议转换**
4. **Data 层统一使用 `getClient(ctx)` 获取数据库客户端**
5. **错误统一通过 `entutil.HandleError` 转换**
6. **新增 Service 后需同步 `ProviderSet` 和 `NewRegistrarList`**
7. **所有命令必须通过 `./nx` 执行**
8. **gRPC 客户端必须挂载在 Data 结构体上**，由 `NewData` 统一创建和管理生命周期（遵循 Kratos beer-shop 模式）
