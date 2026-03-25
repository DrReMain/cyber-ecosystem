# AI Agent 工作约束（当前有效）

本文档定义 AI Agent 在本仓库的执行约束。  
目标：让 Agent 输出可执行、可验证、与当前代码一致的结果。

## 1. 仓库事实（必须按此理解）

- Monorepo 编排：`Nx`
- 后端主栈：`Go 1.25`、`Kratos`、`Ent`、`Wire`、`Buf`
- 示例服务基线：`examples/template1`、`examples/template2`
- 共享能力目录：`shared-go`
- 生成器目录：`tools/generators/kratos-base`
- 契约中心：`contracts/proto`

## 2. 规则优先级

1. 禁止
2. 必做
3. 推荐
4. 允许

只要命中高优先级，低优先级规则自动失效。

## 3. 禁止

### 3.1 禁止臆造现状

- 禁止把路线图当成已实现能力
- 禁止不看代码就声明“仓库已支持某功能”
- 禁止假设 Nx target 存在而不核对 `project.json`

### 3.2 禁止过时返回模型

- 禁止继续使用或建议 `common.Reply`
- 禁止把“HTTP 包装一层统一结构”描述为当前默认
- 禁止把 `validate.ProtoValidate(validate.UseProtoMessage)` 作为当前标准用法

当前默认是：
- 成功 body 保持协议原生
- 错误 body 使用 `common.ErrorBody`（HTTP）或错误详情承载（gRPC/Connect）
- 校验中间件使用 `validate.ProtoValidate(reason, formatter)`

### 3.3 禁止破坏分层

- 禁止在 `service` 层直接写 Ent 查询
- 禁止在 `service` 层开启/编排事务
- 禁止在 `biz` 层依赖 Ent 具体类型
- 禁止在共享约束场景硬编码 `reason` 字符串（必须来自 proto 枚举）

### 3.4 禁止不可执行指导

- 禁止给不存在的路径、target、命令
- 禁止混淆“生成器模板代码”和“业务服务代码”
- 禁止只给底层命令、不提供 Nx 入口

### 3.5 禁止初始化阶段硬退出

- 禁止新增 `panic` / `log.Fatalf` 处理初始化失败
- 必须返回 `error` 给上层处理

## 4. 必做

### 4.1 修改前先核实真实入口

至少核对：
- 目标项目 `project.json`
- 目标实现文件
- 对应模板（若任务涉及生成器）

### 4.2 命令优先给 Nx 入口

优先示例：
- `./nx run infra:docker:up`
- `./nx run contracts:proto`
- `./nx run examples-template1:generate`
- `./nx run shared-go:test`

### 4.3 涉及错误/返回/i18n 时必须核对这些文件

- `shared-go/kratos/encoder/*.go`
- `shared-go/kratos/middleware/responsemeta/*.go`
- `shared-go/kratos/middleware/traceheader/*.go`
- `shared-go/kratos/i18n/*.go`
- `shared-go/kratos/transport/connect/*.go`
- `examples/*/internal/server/*.go`

### 4.4 修改后必须做最小验证

- 至少运行受影响范围测试
- 默认优先：`./nx run shared-go:test`
- 涉及模板变更时，必须做一次生成器冒烟（生成 + generate + 启动验证）

### 4.5 新增服务默认约束

- 必须生成 `api/v1/*_error_reason.proto`
- `reason` 必须来自错误枚举，不能业务层硬编码任意字符串
- 默认创建 `internal/locales/active.zh-Hans.json` 与 `active.en.json`
- 默认接入 i18n 错误本地化中间件

## 5. 推荐

### 5.1 默认工作顺序

1. 读 `docs/README.md`
2. 读 `docs/DEVELOPMENT.md`
3. 读本文档
4. 看目标 `project.json`
5. 看示例实现与 shared-go
6. 再给方案/改代码

### 5.2 返回与错误推荐口径

- 成功响应：原生 proto body
- 错误响应：
  - HTTP：`common.ErrorBody`
  - gRPC：status + details
  - Connect：Connect error + `common.ErrorBody` detail
- 元信息优先放 header：`X-Trace-Id`、`X-Response-Success`、`X-Error-Reason`

### 5.3 i18n 推荐口径

- `message` 优先使用业务显式设置值
- 若无 message，再按 `reason` + `Accept-Language` 翻译
- `Accept-Language` 解析需支持 `q` 权重（当前实现基于 `x/text/language`）

## 6. 允许

- 允许解释底层命令，但必须同时给 Nx 入口
- 允许指出架构不足，但必须明确“现状”与“建议”边界
- 允许临时手工改示例服务，但涉及新服务默认建议走生成器
