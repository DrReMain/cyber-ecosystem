# 开发指南

本文档面向当前仓库的维护者与开发者，目标是用最短路径说明“现在怎么开发”，并提供可以直接照着写的示例。

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
./nx run infra:go:init
./nx run infra:docker:up
./nx run infra:docker:down
./nx run infra:buf:format
./nx run infra:buf:lint
./nx run contracts:proto
./nx run examples-template1:proto
./nx run examples-template1:generate
./nx run examples-template1:dev
./nx run examples-template1:build
./nx run examples-template1:ent:new --args="Entity=Blog"
./nx run shared-go:test
./nx run tools:g:kratos-base
```

## 4. 当前后端分层

每个 Kratos 服务默认按下面分层理解：

1. `api/v1/*.proto`
2. `internal/service`
3. `internal/biz`
4. `internal/data`
5. `internal/server`

职责约束：

- Proto：定义契约
- Service：协议层转换
- Biz：业务语义、事务边界
- Data：Ent 持久化、远程依赖
- Server：中间件、transport、编解码

## 5. 最常见任务与写法

### 5.1 新增或修改 Proto

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

### 5.2 写 Service 层

Service 层的目标很单一：把 Proto request 转成 Biz entity，再把 Biz output 转回 Proto response。

示例：

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

列表查询示例：

```go
func (s *BlogService) QueryBlog(ctx context.Context, in *template1V1.QueryBlogRequest) (*template1V1.QueryBlogResponse, error) {
	out, err := s.blogUC.QueryBlog(ctx, &biz.BlogQueryIn{
		PageRequest: util.GetOrBuildPage(in.Page),
		OrderBy:     order_by.ParseOrderBy(in.OrderBy),
		ID:          in.Id,
		Title:       in.Title,
	})
	if err != nil {
		return nil, err
	}
	return &template1V1.QueryBlogResponse{
		Page: out.PageResponse,
	}, nil
}
```

Service 层不要做这些事：

- 不直接写 Ent 查询
- 不自己开启复杂事务
- 不在这里拼远程 gRPC client

### 5.3 写 Biz 层

Biz 层负责实体、接口、事务边界。

结构示例：

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
```

事务边界示例：

```go
func (uc *BlogUC) UpdateBlog(ctx context.Context, fieldsMask []string, entity *BlogEntity) error {
	return uc.tm.InTx(ctx, func(ctx context.Context) error {
		return uc.blogRP.Update(ctx, fieldsMask, entity)
	})
}
```

简单单仓储动作可以直接透传：

```go
func (uc *BlogUC) GetBlog(ctx context.Context, id string) (*BlogEntity, error) {
	return uc.blogRP.Get(ctx, id)
}
```

### 5.4 写 Data 层

Data 层负责 Ent 查询、错误映射、事务复用，以及当前仓库里已有的远程依赖注入。

创建示例：

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

`fields_mask` 更新示例：

```go
func (rp *blogRP) Update(ctx context.Context, fieldsMask []string, entity *biz.BlogEntity) error {
	return rp.data.InTx(ctx, func(ctx context.Context) error {
		builder := rp.data.getClient(ctx).Blog.UpdateOneID(entity.ID)
		masks.Handler{
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
		}.Emit(fieldsMask)
		return HandleError(builder.Exec(ctx))
	})
}
```

分页查询示例：

```go
func (rp *blogRP) Query(ctx context.Context, bo *biz.BlogQueryIn) (*biz.BlogQueryOut, error) {
	query := rp.data.getClient(ctx).Blog.Query()
	entutil.WherePtr(query, bo.Title, blog.TitleContainsFold)
	entutil.ApplyOrderBy(bo.OrderBy, ent.Asc, ent.Desc, entutil.FOMapping{
		"created_at": func(sel entutil.SQLSelector) { query.Order(sel(blog.FieldCreatedAt)) },
		"updated_at": func(sel entutil.SQLSelector) { query.Order(sel(blog.FieldUpdatedAt)) },
	})
	total, offset, limit, err := entutil.ApplyPagination(ctx, query, bo.PageRequest,
		entutil.NewPageConfig(entutil.DefaultPageSize, entutil.DefaultPageSizeUnlimit))
	if err != nil {
		return nil, HandleError(err)
	}
	_ = total
	_ = offset
	_ = limit
	return nil, nil
}
```

### 5.5 注册 Service

新 Service 需要进入聚合注册链路。

示例：

```go
type Registrar interface {
	RegisterGRPC(*grpc.Server)
	RegisterHTTP(*http.Server)
}

var ProviderSet = wire.NewSet(
	NewRegistrarList,
	NewBlogService,
)

func NewRegistrarList(s1 *BlogService) []Registrar {
	return []Registrar{s1}
}
```

## 6. 新增服务的推荐路径

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

## 7. 当前默认约束

### 7.1 命令约束

- 仓库级说明优先写 `nx`
- 只有在解释底层原理时才补充裸命令

### 7.2 文档约束

- 只写已实现事实
- 规划能力必须明确标注为未落地

### 7.3 代码约束

- Service 不直接碰 Ent
- Biz 不直接依赖 Ent 具体实现
- Data 统一做错误映射
- HTTP 响应统一包装为 `common.Reply`

## 8. 开发前后检查

开发前：

- 确认目标项目名
- 确认对应 `project.json` target
- 确认要对齐哪份样例

开发后：

- 是否执行了 `proto` 或 `generate`
- ProviderSet / Wire / registrar 是否补齐
- 文档是否与实现一致
- 是否误把“规划能力”写成“当前能力”
