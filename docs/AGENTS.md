# AI Agent 强约束指南

本文档是给 AI Agent 的强约束工作规范。目标不是“介绍仓库”，而是约束 Agent 在本仓库里应该如何行动、哪些行为绝对禁止、哪些行为只是允许、哪些行为是默认推荐路径。

如果本文档与 Agent 自身习惯冲突，以本文档为准。

## 1. 仓库事实

- 编排工具：`Nx`
- 后端主栈：`Go 1.25`、`Kratos v2.9.2`、`Ent`、`Wire`、`Buf`
- 前端实验项目：`clients/admin`，基于 `Next.js 16`
- 后端实现基线：`examples/template1`、`examples/template2`
- 公共契约目录：`contracts/proto`
- 共享 Go 组件目录：`shared-go`
- 工具链与基础设施 target：`infra`
- Nx 生成器目录：`tools/generators`

## 2. 规则优先级

按以下顺序理解规则：

1. `禁止`
2. `必做`
3. `推荐`
4. `允许`

只要命中更高优先级规则，低优先级规则自动失效。

## 3. 禁止

以下行为一律禁止：

### 3.1 禁止臆造仓库能力

- 禁止假设某个 `Nx target` 存在而不检查 `project.json`
- 禁止把路线图文档当成已实现能力
- 禁止把历史上可能存在过的行为写成当前事实
- 禁止在没看代码前就声称“项目就是这样实现的”

反例：

```text
错误示范：
- 直接说 ./nx run tools:docker:up 可用
- 直接说响应头里已经注入 X-Trace-Id
- 直接说仓库支持 i18n / 熔断 / 分布式事务
```

### 3.2 禁止绕过仓库入口

- 禁止把仓库级操作默认写成零散裸命令
- 禁止新增文档时只给底层命令，不给对应 `nx` 用法
- 禁止跳过 `project.json` 直接发明新的命令名

反例：

```bash
# 禁止作为默认建议
cd examples/template1 && go generate ./...
cd infra/docker && docker compose up -d
buf generate --template buf.gen.api.yaml
```

### 3.3 禁止破坏当前分层

- 禁止在 Service 层直接写 Ent 查询
- 禁止在 Service 层编排复杂事务
- 禁止在 Biz 层直接依赖具体 Ent 类型
- 禁止把 Proto 定义散落到 `contracts/proto` 之外的公共目录中冒充公共契约

### 3.4 禁止输出“看起来完整但无法执行”的指导

- 禁止给不存在的文件路径
- 禁止给不符合当前样例结构的伪代码模板
- 禁止把生成器、模板、实际代码三套写法混在一起不做区分

### 3.5 禁止把示例当成生产承诺

- 禁止宣称 `clients/admin` 代表成熟前端规范
- 禁止把 `examples/template1`、`examples/template2` 描述为业务生产模板
- 禁止把当前占位实现描述为完整平台能力

## 4. 必做

以下动作是 Agent 每次进入具体任务时必须执行的：

### 4.1 修改前必须核对真实入口

至少检查以下之一：

- 目标项目的 `project.json`
- 对应实现文件
- 已存在的示例代码

### 4.2 涉及命令时必须优先给出 `nx` 入口

正确示例：

```bash
./nx run infra:go:init
./nx run infra:docker:up
./nx run contracts:proto
./nx run examples-template1:generate
./nx run shared-go:test
```

### 4.3 涉及新服务时必须优先检查生成器

默认先看：

- `tools/project.json`
- `tools/generators/kratos-base/generator.ts`
- `tools/generators/kratos-base/schema.json`

### 4.4 涉及“当前行为”描述时必须以代码为准

例如判断响应格式，应看：

- `internal/server/http.go`
- `shared-go/kratos/encoder/*.go`

判断错误处理，应看：

- `internal/data/errors.go`
- `shared-go/orm/ent/entutil/error.go`

## 5. 推荐

以下是 Agent 的默认推荐行为。

### 5.1 推荐的工作顺序

1. 看 `README.md`
2. 看 `docs/DEVELOPMENT.md`
3. 看目标目录 `project.json`
4. 看样例实现
5. 再输出建议或直接改代码

### 5.2 推荐的新增后端服务方式

优先使用生成器：

```bash
./nx run tools:g:kratos-base
```

不要默认建议直接复制 `examples/template1`。

### 5.3 推荐的后端对齐基线

- Proto 包名：`api.<service>.v1`
- `go_package`：`github.com/DrReMain/cyber-ecosystem/gen/go/<service>/v1;<service>V1`
- Service 只做协议转换
- Biz 定义实体、接口、事务边界
- Data 实现 Ent 与远程依赖
- 列表查询先走 `util.GetOrBuildPage`
- 排序字符串走 `order_by.ParseOrderBy`

## 6. 允许

以下行为是允许的，但不是默认首选：

- 允许解释底层真实命令，但要同时说明对应 `nx target`
- 允许参考 `examples/template1` 手工补代码，但前提是用户任务不适合生成器
- 允许指出当前架构局限，但必须区分“现状问题”和“未来建议”

## 7. 可直接复用的正确示例

### 7.1 Proto 示例

```protobuf
syntax = "proto3";

package api.template1.v1;

import "buf/validate/validate.proto";
import "common/common.proto";
import "google/api/annotations.proto";
import "google/protobuf/timestamp.proto";

option go_package = "github.com/DrReMain/cyber-ecosystem/gen/go/template1/v1;template1V1";

service BlogService {
  rpc QueryBlog(QueryBlogRequest) returns (QueryBlogResponse) {
    option (google.api.http) = {get: "/api/v1/blog"};
  }
}

message QueryBlogRequest {
  common.PageRequest page = 1;
  optional string title = 2;
  repeated string order_by = 100;
}
```

### 7.2 Service 示例

```go
func (s *BlogService) QueryBlog(ctx context.Context, in *template1V1.QueryBlogRequest) (*template1V1.QueryBlogResponse, error) {
	out, err := s.blogUC.QueryBlog(ctx, &biz.BlogQueryIn{
		PageRequest: util.GetOrBuildPage(in.Page),
		OrderBy:     order_by.ParseOrderBy(in.OrderBy),
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

### 7.3 Biz 示例

```go
func (uc *BlogUC) UpdateBlog(ctx context.Context, fieldsMask []string, entity *BlogEntity) error {
	return uc.tm.InTx(ctx, func(ctx context.Context) error {
		return uc.blogRP.Update(ctx, fieldsMask, entity)
	})
}
```

### 7.4 Data 示例

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
		}.Emit(fieldsMask)
		return HandleError(builder.Exec(ctx))
	})
}
```

## 8. 任务类型到默认动作

### 8.1 用户说“新增接口”

默认动作：

1. 看目标服务 `api/v1/*.proto`
2. 修改 proto
3. 运行对应 `proto` / `generate`
4. 实现 `service` / `biz` / `data`

### 8.2 用户说“新增服务”

默认动作：

1. 检查是否应使用 `tools:g:kratos-base`
2. 生成后核对 `buf.yaml`
3. 校验新项目 `project.json`
4. 再补业务逻辑

### 8.3 用户说“检查项目是否合理”

默认动作：

1. 看 `README.md`、`docs`
2. 看根目录和各 `project.json`
3. 看模板实现
4. 只输出已验证的问题，不输出想象中的问题

## 9. 最终自检清单

Agent 在输出结论前，至少确认以下问题：

- 我引用的命令在真实 `project.json` 里存在吗？
- 我描述的行为在真实代码里存在吗？
- 我是否把“规划”说成了“现状”？
- 我是否给了最短可执行路径？
- 我是否指出了风险边界，而不是泛泛而谈？
