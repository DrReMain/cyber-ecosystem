# 开发指南（当前有效）

本文档描述仓库当前可执行的开发路径，以 `examples/template1`、`examples/template2` 和 `shared-go` 现状为准。

## 1. 技术基线

- Monorepo：`Nx`
- Go：`1.25`
- 后端：`Kratos + Ent + Wire`
- 契约：`Buf + Protobuf`
- 共享能力：`shared-go`

## 2. 目录职责

```text
contracts/    公共 proto 契约
examples/     示例服务（当前后端实现基线）
gen/          代码生成产物
infra/        工具链与本地基础设施
shared-go/    共享 Go 组件
tools/        生成器与工程工具
clients/      客户端工程
docs/         文档
```

## 3. 常用命令（优先 Nx）

```bash
# 基础设施
./nx run infra:go:init
./nx run infra:docker:up
./nx run infra:docker:down

# 契约生成
./nx run contracts:proto
./nx run examples-template1:proto
./nx run examples-template2:proto

# 服务生成（proto + ent + wire + tidy）
./nx run examples-template1:generate
./nx run examples-template2:generate

# 运行
./nx run examples-template1:dev
./nx run examples-template2:dev

# 测试
./nx run shared-go:test
```

## 4. 分层与边界

后端服务默认分层：

1. `api/v1/*.proto`
2. `internal/service`
3. `internal/biz`
4. `internal/data`
5. `internal/server`

约束：

- `service` 只做协议转换，不做持久化与事务编排
- `biz` 负责业务语义与事务边界，不依赖 Ent 具体类型
- `data` 负责 Ent/远程依赖实现与错误映射
- `server` 负责 middleware、transport、encoder

## 5. 返回与错误处理规范

### 5.1 成功响应

- HTTP/gRPC/Connect 成功响应均返回原生 proto 消息
- 不再使用 `common.Reply` 包装

### 5.2 错误响应

- HTTP：`common.ErrorBody {reason, message, details}`
- gRPC：status + details
- Connect：Connect error，`common.ErrorBody` 放在 error detail

### 5.3 元信息

默认通过 middleware 注入：

- `X-Trace-Id`
- `X-Response-Success`
- `X-Error-Reason`（失败时）

参考实现：

- `shared-go/kratos/middleware/traceheader`
- `shared-go/kratos/middleware/responsemeta`

## 6. i18n 规范

### 6.1 当前实现

- 默认 i18n 库：`go-i18n`
- 抽象接口位于：`shared-go/kratos/i18n`
- 语言从 `Accept-Language` 解析，支持 `q` 权重

### 6.2 消息优先级

1. 业务显式设置的 `message`（最高优先级）
2. 根据 `reason` + 语言翻译
3. 未命中时回落到 `reason`

### 6.3 locale 组织

- 每个服务自治维护：`internal/locales/active.<lang>.json`
- reason 必须来源于 proto 错误枚举，不允许业务硬编码任意字符串

## 7. 校验与错误码

### 7.1 proto 校验

使用 `buf.validate` 注解字段规则，服务端通过 middleware 执行：

```go
validate.ProtoValidate(
    templateXV1.ErrorReason_ERROR_REASON_VALIDATOR.String(),
    validate.UseDefaultError,
)
```

### 7.2 错误枚举

- 每个服务应定义 `*_error_reason.proto`
- `ERROR_REASON_UNSPECIFIED = 0` 必须存在
- `(errors.code)` 作为 transport code 映射来源

## 8. 数据层错误映射

Data 层统一通过 `entutil.HandleEntError(...)` 映射 Ent 错误：

- `NotFound -> 404`
- `Validation -> 400`
- `Constraint -> 409`

并绑定到服务自身的错误枚举 reason。

## 9. 生成器开发与验收

生成器目录：`tools/generators/kratos-base`

涉及模板改动时，建议最小验收流程：

1. 生成临时服务
2. 跑 `proto`/`generate:ent`/`generate`
3. 启动服务做一次真实运行冒烟
4. 清理临时服务与 `buf.yaml` 临时模块项

## 10. 提交前检查清单

- 受影响 target 可执行
- `shared-go` 测试通过
- 文档描述与实现一致
- 无新增 `panic` / `log.Fatalf` 初始化硬退出
- 无过时返回模型（`common.Reply`）残留
