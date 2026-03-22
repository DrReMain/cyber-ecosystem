# 架构蓝图

本文档描述当前仓库已经形成的结构边界，以及后续演进时不建议打破的基本原则。

## 1. 当前结构

```text
cyber-ecosystem/
├── contracts/     # 公共契约中心
├── examples/      # 示例服务与实现基线
├── gen/           # 生成产物
├── shared-go/     # 共享 Go 库
├── infra/         # 工具链和本地基础设施
├── tools/         # Nx 生成器
├── clients/       # 客户端项目
└── docs/          # 文档
```

## 2. 当前形成的架构分层

后端服务以 `Kratos + Ent` 为主，按以下层次组织：

1. `api/v1/*.proto`
2. `internal/service`
3. `internal/biz`
4. `internal/data`
5. `internal/server`

其中：

- 契约定义是变更入口
- Service 做协议转换
- Biz 管业务语义和事务边界
- Data 管持久化和远程依赖
- Server 负责 transport、中间件、编解码

## 3. 当前已经成立的边界

- 公共 Proto 要进入 `contracts/proto`，不要散落复制
- 共享 Go 能力优先沉淀到 `shared-go`
- 仓库级操作通过 `Nx target` 暴露，而不是靠口口相传的脚本
- 示例服务用于说明模式，不等于最终业务域划分

## 4. 当前架构优势

- 契约、共享库、示例服务已经分层
- `template2` 展示了跨服务聚合的基本模式
- `infra` 提供了工具链安装和 Docker 编排入口
- `tools/generators/kratos-base` 已具备继续平台化的基础

## 5. 当前架构风险

- 新服务生成器默认产物目录与现有 `examples/` 基线不一致
- 跨服务 client 直接放在 `Data` 中，短期实用，长期可能让 Data 层过宽
- `gen/` 只有产物没有 target，生成链路分散在各项目内，可发现性一般
- 文档和设计草案容易超前于实现，导致认知偏差
- 前端、治理、国际化、分布式事务等能力尚未形成统一抽象

## 6. 演进建议

- 先统一“当前有效规范”，再继续扩展平台愿景
- 优先补齐生成器、基础设施、文档索引之间的一致性
- 将“已实现能力”和“规划能力”持续分仓式表达，避免互相污染
