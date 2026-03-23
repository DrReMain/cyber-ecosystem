# 开发指南

本文档面向当前仓库的维护者与开发者，目标是用最短路径说明"现在怎么开发"，并提供可以直接照着写的示例。

## 1. 当前技术基线

- Monorepo：`Nx 22`
- Go：`1.25`
- 微服务框架：`Kratos v2.9.2`
- ORM：`Ent`
- 契约与生成：`Buf + Protobuf`
- 依赖注入：`Wire`
- 前端实验项目：`Next.js 16 + React 19`

## 2. 目录职责

```text
contracts/    公共 Proto 契约
examples/     Kratos 示例服务
gen/          生成产物
infra/        工具链、Buf、Docker 基础设施
shared-go/    共享 Go 组件
tools/        Nx 生成器
clients/      前端客户端
docs/         文档
```

## 3. 先用这些命令

仓库级操作优先通过 `nx` 执行：

```bash
# 基础设施
./nx run infra:go:init          # 安装 Go 工具链
./nx run infra:docker:up        # 启动所有基础设施
./nx run infra:docker:down      # 停止所有基础设施

# Proto 生成
./nx run contracts:proto        # 生成公共契约
./nx run examples-template1:proto    # 生成 template1 Proto
./nx run examples-template1:generate # 完整生成（Proto + Ent + Wire）

# 开发运行
./nx run examples-template1:dev      # 开发模式运行
./nx run examples-template1:build    # 构建二进制

# 测试
./nx run shared-go:test         # 运行共享库测试
```

## 4. 当前后端分层

每个 Kratos 服务默认按下面分层理解：

1. `api/v1/*.proto` - 契约定义
2. `internal/service` - 协议层转换
3. `internal/biz` - 业务语义、事务边界
4. `internal/data` - Ent 持久化、远程依赖
5. `internal/server` - 中间件、transport、编解码

### 4.1 分层约束（强制）

| 层级 | 允许 | 禁止 |
|------|------|------|
| Service | 调用 Biz、Proto 类型转换 | 直接使用 Ent、开启事务 |
| Biz | 定义 Entity/接口、事务编排 | 依赖 Ent 具体类型 |
| Data | Ent 查询、错误映射、远程调用 | 业务逻辑判断 |

### 4.2 检查分层合规

```bash
# 检查 Service 是否直接使用 Ent
grep -r "\.Blog\." examples/template1/internal/service/

# 检查 Biz 是否依赖 Ent 类型
grep -r "ent\." examples/template1/internal/biz/
```

以上命令应该返回空，否则说明分层违规。

## 5. 序列化规范

### 5.1 显式配置（推荐）

遵循 Kratos 官方文档建议，在 `main.go` 中显式配置 `json.MarshalOptions`：

```go
import (
    kratosjson "github.com/go-kratos/kratos/v2/encoding/json"
    "google.golang.org/protobuf/encoding/protojson"
)

func init() {
    flag.StringVar(&flagConf, "conf", "../../configs", "config path")

    // 显式配置 JSON 序列化选项（遵循 Kratos 官方文档）
    // 配置可见，开发者可随时替换
    kratosjson.MarshalOptions = protojson.MarshalOptions{
        EmitUnpopulated: true,  // 零值字段也输出
        UseProtoNames:   false, // 使用 camelCase（createdAt 而非 created_at）
    }
    kratosjson.UnmarshalOptions = protojson.UnmarshalOptions{
        DiscardUnknown: true, // 忽略未知字段
    }
}
```

**设计优势**：
- 配置在业务项目中显式可见，不隐藏在共享库中
- 开发者可以随时替换 codec 实现
- 所有组件通过 `encoding.GetCodec("json")` 获取配置后的 codec

### 5.2 统一消费

所有需要序列化的组件通过 `encoding.GetCodec("json")` 获取 codec：

| 组件 | 文件 | 说明 |
|------|------|------|
| HTTP Encoder | `shared-go/kratos/encoder/response_encoder.go` | 响应序列化 |
| Error Encoder | `shared-go/kratos/encoder/error_encoder.go` | 错误序列化 |
| Connect Codec | `shared-go/kratos/transport/connect/codec.go` | Connect RPC 序列化 |
| Loki Writer | `shared-go/kratos/logging/zap/loki.go` | 日志序列化 |
| Health Check | `shared-go/kratos/transport/connect/health.go` | 健康检查序列化 |

### 5.3 协议感知编码

HTTP Encoder 实现协议感知编码，区分强 Schema 协议和普通 HTTP：

```go
// IsContractProtocol 检测是否为强 Schema 协议（gRPC、Connect）
func IsContractProtocol(r *http.Request) bool {
    // Connect 协议标识
    if r.Header.Get("Connect-Protocol-Version") != "" {
        return true
    }
    // gRPC-Web 协议标识
    if strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
        return true
    }
    return false
}
```

对于强 Schema 协议，响应直接返回不包装；对于普通 HTTP，响应使用 `common.Reply` 包装。
```

### 5.3 协议感知编码

HTTP Encoder 实现协议感知，区分强 Schema 协议和纯 HTTP：

| 协议 | 处理方式 | 原因 |
|------|----------|------|
| gRPC/Connect | 直接返回 Proto | 自带 Schema，无需包装 |
| 纯 HTTP | 包装 Response | 前端需要统一格式 |

**实现**（[`shared-go/kratos/encoder/response_encoder.go`](shared-go/kratos/encoder/response_encoder.go)）：

```go
// 强 Schema 协议检测
func IsContractProtocol(r *http.Request) bool {
    return r.Header.Get("Connect-Protocol-Version") != "" ||
           strings.Contains(r.Header.Get("Content-Type"), "application/grpc")
}
```

### 5.4 字段命名对照

| Proto 定义 | JSON 输出 | 说明 |
|------------|-----------|------|
| `created_at` | `createdAt` | camelCase |
| `page_no` | `pageNo` | camelCase |
| `page_size` | `pageSize` | camelCase |

### 5.5 统一消费方

所有需要 JSON 序列化的组件都使用全局 Codec：

| 组件 | 文件 | 使用方式 |
|------|------|----------|
| HTTP Encoder | `encoder/response_encoder.go` | `encoding.GetCodec("json")` |
| Error Encoder | `encoder/error_encoder.go` | `encoding.GetCodec("json")` |
| Connect Codec | `transport/connect/codec.go` | `encoding.GetCodec("json")` |
| Loki Writer | `logging/zap/loki.go` | `encoding.GetCodec("json")` |
| Health Check | `transport/connect/health.go` | `encoding.GetCodec("json")` |

## 6. 跨服务调用规范

### 6.1 两种调用方式

template2 展示了调用 template1 的两种方式：

| 方式 | 配置项 | 适用场景 |
|------|--------|----------|
| gRPC | `service_template1` | 内部服务、高性能 |
| Connect | `service_template1connect` | Web 友好、跨语言 |

### 6.2 客户端创建示例

```go
// gRPC 客户端
func NewTemplate1BlogService(c *conf.Data, logger log.Logger, tp *tracesdk.TracerProvider) template1V1.BlogServiceClient {
    conn, err := grpc.DialInsecure(
        context.Background(),
        grpc.WithEndpoint(c.ServiceTemplate1.Addr),
        grpc.WithTimeout(c.ServiceTemplate1.Timeout.AsDuration()),
        grpc.WithMiddleware(middlewares...),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to dial template1: %w", err)  // 返回错误，不要 panic
    }
    return template1V1.NewBlogServiceClient(conn)
}

// Connect 客户端
func NewTemplate1ConnectBlogService(c *conf.Data, logger log.Logger, tp *tracesdk.TracerProvider) template1V1connect.BlogServiceClient {
    conn, err := connect.DialInsecure(
        context.Background(),
        connect.WithEndpoint(c.ServiceTemplate1Connect.Addr),
        connect.WithTimeout(c.ServiceTemplate1Connect.Timeout.AsDuration()),
        connect.WithMiddleware(middlewares...),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to dial template1: %w", err)  // 返回错误，不要 panic
    }
    return template1V1connect.NewBlogServiceClient(conn.HTTPClient(), conn.Endpoint(), conn.ClientOptions()...)
}
```

### 6.3 错误处理要求

**禁止使用 `panic` 或 `log.Fatalf`**，应该返回 error 让调用方决定处理方式：

```go
// ❌ 错误示范
if err != nil {
    panic(err)
}

// ✅ 正确做法
if err != nil {
    return nil, fmt.Errorf("failed to connect: %w", err)
}
```

## 7. 最常见任务与写法

### 7.1 新增或修改 Proto

示例来自 `template1`：

```protobuf
syntax = "proto3";

package api.template1.v1;

import "buf/validate/validate.proto";
import "common/common.proto";
import "google/api/annotations.proto";
import "google/protobuf/timestamp.proto";
import "google/protobuf/wrappers.proto";

option go_package = "github.com/DrReMain/cyber-ecosystem/gen/go/template1/v1;template1V1";

service BlogService {
  rpc CreateBlog(CreateBlogRequest) returns (CreateBlogResponse) {
    option (google.api.http) = {
      post: "/api/v1/blog"
      body: "*"
    };
  }
}

message CreateBlogRequest {
  optional string title = 1;
  optional string content = 2;
  google.protobuf.Timestamp published_at = 3;
}

message CreateBlogResponse {}
```

修改完后执行：

```bash
./nx run examples-template1:proto
```

如果还涉及 Ent / Wire / `go mod tidy`，直接执行：

```bash
./nx run examples-template1:generate
```

### 7.2 写 Service 层

Service 层的目标很单一：把 Proto request 转成 Biz entity，再把 Biz output 转回 Proto response。

```go
func (s *BlogService) CreateBlog(ctx context.Context, in *template1V1.CreateBlogRequest) (*template1V1.CreateBlogResponse, error) {
    entity := &biz.BlogEntity{
        Title:       in.Title,
        Content:     in.Content,
        PublishedAt: util.GetPTimeFromPPbTime(in.PublishedAt),
    }
    if err := s.blogUC.CreateBlog(ctx, entity); err != nil {
        return nil, err
    }
    return &template1V1.CreateBlogResponse{}, nil
}
```

### 7.3 写 Biz 层

Biz 层负责实体、接口、事务边界。

```go
type BlogEntity struct {
    ID          string
    Title       *string
    Content     *string
    PublishedAt *time.Time
    CreatedAt   *time.Time
    UpdatedAt   *time.Time
}

type BlogRP interface {
    Create(context.Context, *BlogEntity) error
    Update(context.Context, []string, *BlogEntity) error
    Get(context.Context, string) (*BlogEntity, error)
}

func (uc *BlogUC) UpdateBlog(ctx context.Context, fieldsMask []string, entity *BlogEntity) error {
    return uc.tm.InTx(ctx, func(ctx context.Context) error {
        return uc.blogRP.Update(ctx, fieldsMask, entity)
    })
}
```

### 7.4 写 Data 层

Data 层负责 Ent 查询、错误映射、事务复用。

```go
func (rp *blogRP) Create(ctx context.Context, entity *biz.BlogEntity) error {
    return rp.data.InTx(ctx, func(ctx context.Context) error {
        client := rp.data.getClient(ctx)
        if err := client.Blog.Create().
            SetNillableTitle(entity.Title).
            SetNillableContent(entity.Content).
            SetNillablePublishedAt(entity.PublishedAt).
            Exec(ctx); err != nil {
            return HandleError(err)
        }
        return nil
    })
}
```

### 7.5 注册 Service

```go
type Registrar interface {
    RegisterGRPC(*grpc.Server)
    RegisterHTTP(*http.Server)
    RegisterConnect(*connect.Server)
}

var ProviderSet = wire.NewSet(
    NewRegistrarList,
    NewBlogService,
)

func NewRegistrarList(s1 *BlogService) []Registrar {
    return []Registrar{s1}
}
```

## 8. 新增服务的推荐路径

优先用生成器：

```bash
./nx run tools:g:kratos-base
```

生成后要检查三件事：

1. `buf.yaml` 是否追加了模块路径
2. 新项目 `project.json` 是否可执行
3. 生成路径是否符合你的目录策略

当前注意点：

- 生成器默认生成到 `services/<name>`
- 当前已有样例主要在 `examples/`
- 所以新增服务时要主动确认目录策略，不要默默让仓库继续分叉

## 9. 当前默认约束

### 9.1 命令约束

- 仓库级说明优先写 `nx`
- 只有在解释底层原理时才补充裸命令

### 9.2 文档约束

- 只写已实现事实
- 规划能力必须明确标注为未落地

### 9.3 代码约束

- Service 不直接碰 Ent
- Biz 不直接依赖 Ent 具体实现
- Data 统一做错误映射
- HTTP 响应统一包装为 `common.Reply`
- 初始化失败返回 error，不要 `panic` 或 `log.Fatalf`

### 9.4 序列化约束

- 必须在 `main.go` 的 `init()` 调用 `sharedjson.Init()`
- JSON 输出使用 camelCase 字段名
- 零值字段默认输出

## 10. 开发前后检查

### 10.1 开发前

- [ ] 确认目标项目名
- [ ] 确认对应 `project.json` target
- [ ] 确认要对齐哪份样例
- [ ] 阅读相关文档

### 10.2 开发后

- [ ] 是否执行了 `proto` 或 `generate`
- [ ] ProviderSet / Wire / registrar 是否补齐
- [ ] 分层约束是否遵守（检查 Service/Biz）
- [ ] 序列化初始化是否调用
- [ ] 错误处理是否正确（无 panic/log.Fatalf）
- [ ] 文档是否与实现一致
- [ ] 是否误把"规划能力"写成"当前能力"

### 10.3 提交前

- [ ] 运行 `./nx run shared-go:test` 确保测试通过
- [ ] 运行 `./nx run infra:buf:lint` 确保 Proto 规范
- [ ] 检查是否有重复代码需要提取

## 11. 常见问题

### 11.1 JSON 字段名是 snake_case 而不是 camelCase

检查是否调用了 `sharedjson.Init()`。

### 11.2 跨服务调用失败

1. 检查目标服务是否启动
2. 检查配置中的地址是否正确
3. 检查中间件是否正确配置

### 11.3 测试失败

运行详细测试输出：

```bash
go test ./shared-go/... -v
```

## 12. 相关文档

- [AGENTS.md](./AGENTS.md) - AI Agent 工作约束
- [ARCHITECTURE_BLUEPRINT.md](./ARCHITECTURE_BLUEPRINT.md) - 架构边界
- [ROADMAP.md](./ROADMAP.md) - 未来路线图
