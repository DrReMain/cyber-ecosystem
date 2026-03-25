# 架构蓝图（当前状态）

本文档描述仓库当前已经形成的结构边界，以及后续演进不建议打破的原则。

## 1. 顶层结构

```text
cyber-ecosystem/
├── contracts/     # 公共 proto 契约
├── examples/      # 示例服务基线（template1/template2）
├── gen/           # 生成产物
├── shared-go/     # 共享 Go 能力
├── infra/         # 工具链与本地基础设施
├── tools/         # 生成器与工程工具
├── clients/       # 客户端工程
└── docs/          # 文档
```

## 2. 后端服务分层

1. `api/v1/*.proto`：契约层
2. `internal/service`：协议转换层
3. `internal/biz`：业务语义与事务边界
4. `internal/data`：持久化与远程依赖
5. `internal/server`：transport、middleware、encoder

关键约束：

- Service 不直接依赖 Ent
- Biz 不依赖 Ent 具体模型
- Data 负责错误映射与基础设施细节

## 3. 协议与返回模型边界

当前设计：

- 成功响应：保持协议原生 body（HTTP/gRPC/Connect）
- 错误语义：`reason` 为核心主键，`message` 可 i18n，`details` 为结构化信息
- 元信息：通过 header 暴露（`X-Trace-Id`、`X-Response-Success`、`X-Error-Reason`）

这意味着“统一语义与元信息”，而不是强制“统一 body 包装”。

## 4. 共享能力边界

`shared-go` 目前沉淀的关键能力包括：

- 编解码器（HTTP/Connect）
- 中间件（trace header、response meta、validate、header forward）
- i18n 抽象与默认 provider
- Connect transport 封装
- Ent 通用辅助（事务、分页、错误映射）

规则：

- 能跨服务复用的能力，优先进入 `shared-go`
- 业务服务只保留服务自治逻辑（如本服务错误枚举、locale）

## 5. 生成器边界

`tools/generators/kratos-base` 产物需与当前基线一致：

- 默认带错误枚举 proto（`<projectName>_error_reason.proto`）
- 默认带 i18n locale 骨架
- 默认带当前 server 中间件链与错误处理方式

模板变更必须通过“生成 + 运行”冒烟验证。

## 6. 当前风险

- 生成器与手工演进代码仍存在持续漂移风险
- 客户端 SDK 统一归一化层尚未落地（connect-web/ts-rest）
- 路线图能力（限流、熔断、分布式事务）仍未进入默认基线
