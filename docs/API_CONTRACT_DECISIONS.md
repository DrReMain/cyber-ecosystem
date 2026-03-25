# API 契约决策（当前版）

本文档记录“契约优先”相关核心决策，并标注当前落地状态。  
更新时间：2026-03-26。

## 1. 总体决策

### 1.1 契约真源

- `proto` 是后端接口真源（SSOT）
- OpenAPI 为兼容层产物，不反向驱动后端契约

### 1.2 协议分层

- gRPC：内部高性能调用主路径
- Connect：同时支持内部和外部调用
- HTTP + OpenAPI：兼容小程序/鸿蒙/外部 REST 场景

### 1.3 返回策略

- 成功 body 保持协议原生（不做全局统一包装）
- 错误语义统一为 `reason/message/details`
- 元信息通过 header 对齐（`X-Trace-Id` 等）

## 2. 当前落地状态

### 2.1 已落地

- `shared-go/kratos/transport/connect` 已落地并在示例服务启用
- `gen/oas/*` 已生成 OpenAPI 兼容输出
- 错误处理已统一到 `common.ErrorBody` 语义层
- i18n 抽象与默认 go-i18n provider 已落地

### 2.2 未落地

- 前端统一 SDK（connect-web / ts-rest 归一化层）尚未进入仓库默认能力
- 多端统一“协议无感消费”的客户端抽象仍待补齐

## 3. 前端协议建议（决策）

| 客户端类型 | 主路径 | 兼容路径 |
|------------|--------|----------|
| Web/App（支持 Connect） | Connect | HTTP+OAS |
| 小程序/鸿蒙 | HTTP+OAS（可配 ts-rest） | 无 |
| 外部系统 | HTTP+OAS | gRPC/Connect（按集成能力） |

## 4. 工具选型结论

- Connect：后端到现代客户端的首选 RPC 协议
- ts-rest：HTTP+OAS 的 TypeScript 兼容消费层候选
- Kubb：按需使用，不作为默认路径

说明：当前仓库尚未内置 ts-rest/Kubb 客户端工程，仍处于后端先行阶段。

## 5. 后续推进顺序

1. 先完成客户端归一化 SDK（统一错误与元信息消费）
2. 再完善多端脚手架（connect-web 与 ts-rest 双模板）
3. 最后补齐跨协议契约一致性回归测试
