# 文档索引

`docs/` 分为四类内容：当前有效规范、技术决策记录、现状评审、未来路线图。

## 当前有效

- [AGENTS.md](./AGENTS.md): 面向 AI Agent 的仓库工作约束
- [DEVELOPMENT.md](./DEVELOPMENT.md): 面向开发者的日常开发指南
- [ARCHITECTURE_BLUEPRINT.md](./ARCHITECTURE_BLUEPRINT.md): 当前仓库架构蓝图与边界

## 技术决策记录

- [API_CONTRACT_DECISIONS.md](./API_CONTRACT_DECISIONS.md): API 契约技术选型决策，涵盖 tRPC、oRPC、ts-rest、Kubb、Connect-RPC 的对比分析与架构设计

## 现状评审

- [PROJECT_REVIEW.md](./PROJECT_REVIEW.md): 对结构、扩展性、稳定性、文档一致性的审视

## 未来路线

- [ROADMAP.md](./ROADMAP.md): 尚未落地但值得继续推进的主题，包括错误治理、i18n、限流熔断、分布式事务、链路上下文增强、契约架构升级

## 使用建议

- 要开始编码：先读 `DEVELOPMENT.md`
- 要让 Agent 接手：先读 `AGENTS.md`
- 要理解仓库边界：读 `ARCHITECTURE_BLUEPRINT.md`
- 要决定下一步补什么：读 `PROJECT_REVIEW.md` 和 `ROADMAP.md`
