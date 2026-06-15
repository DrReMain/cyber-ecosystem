# edge_mobile server 层后续工作（TODO）

记录已设计 / 分析但尚未实施的 server 层工作，作为排期 backlog。

- **一、错误体系对外契约（机制已落地）**——**错误 i18n 后端不做，交由客户端翻译**（已决策）。
- **二、k3s 部署适配（待实现）**：健康探针 / 优雅停机 / 监控 / pprof

---

## 一、错误体系对外契约

### 1.1 错误机制（已落地）

- 所有错误通过 proto 枚举定义；`protoc-gen-go-errors` 生成工厂 `ErrorXxx(format, args...) *errors.Error`。
- 每个错误携带：`Reason`（枚举 key，**客户端翻译的查表 key**）、`Message`（空 = 客户端按 Reason 翻译；非空 = 直出透传）、`Metadata`（结构化模板参数）。
- 中间件错误（validator / recovery 等）在 app `init()` 里覆盖为枚举工厂产出（Reason = 枚举 key），与业务错误统一。

### 1.2 i18n 决策：后端不做，客户端翻译（已定）

- **决定：后端不做错误 i18n，最终翻译全部由客户端完成。**
- **理由**：第一方客户端（mobile / admin）是富客户端、自带 i18n；开放 API 偏开发者对接，要的是稳定 error code + 结构化详情，英文 message 够用。后端做 i18n 会变成"第二套翻译系统"且带来每服务一套 bundle 的重复负担（旧仓库的痛点：三服务 `v1.*.yaml` key 100% 重复 + 每服务一套工具链），收益不值。
- **对外契约**：
  - `Reason` = 稳定的机器可读枚举 key，**即契约本身**，客户端按它查本地翻译表；同时是错误类别（供客户端做重试 / 权限等逻辑判断）。
  - `Metadata` = 结构化模板参数（如 `{field, id, count}`），客户端按 Reason 取本地模板 + 套参渲染。
  - `Message` = **翻译 / 直出的开关信号**：空 → 客户端按 Reason 翻译；非空 → 客户端**原样直出、不翻译**（用于动态消息、上游错误透传等 override 场景）。开发态文案进日志，**不进 Message**。
- **客户端规则（一句话）**：`Message != ""` → 直出；否则按 `Reason` + `Metadata` 翻译。比旧仓库的 `x-md-global-i18n-message` 魔法串更简单——信号就是 Message 本身。
- **由此消失的复杂度**：后端 i18n 中间件、bundle、生成器（geni18n）、locale 解析、服务间"不重复转换"那套——全部不需要。

### 1.3 错误返回的 transport 行为（实测）

#### a. 错误内容三协议一致 ✓

edge_mobile 实测（invalid = 校验错误 / valid = 业务错误）：

| 客户端 → 服务器 | invalid | valid |
|---|---|---|
| grpcurl → grpc (:12002) | InvalidArgument + ErrorInfo(GENERAL_ERROR_VALIDATION_FAILED) | InvalidArgument + ErrorInfo(STATUS_INVALID_TRANSITION), "not implemented" |
| grpcurl → connect (:13002, gRPC 协议) | 同上 | 同上 |
| curl → http (:11002) | `{code:400, reason:VALIDATION_FAILED}` | `{code:400, reason:STATUS, message:"not implemented"}` |
| curl → connect (:13002, Connect 协议) | `{code:"invalid_argument", details:[ErrorInfo(reason)]}` | `{code:"invalid_argument", message:"not implemented", details:[ErrorInfo]}` |

- **reason / message 四组合一致**。
- code / 结构因协议而异（各协议原生）：gRPC = code 名、http = 状态码、connect = code 字符串；gRPC / connect 用 `Details(ErrorInfo)`，http 扁平 JSON。
- connect 服务器同时服务 Connect + gRPC 两种协议。

#### b. transport ReplyHeader 错误路径差异（边缘 gap）

kratos `Transporter` 只暴露 `ReplyHeader()`（无 trailer API）。handler 经 `tr.ReplyHeader().Set(k,v)` 设的响应头，错误路径去向不一致：

| 协议 | 错误时 ReplyHeader 去向 |
|---|---|
| grpc | response header（`grpc.SetHeader` 无条件） |
| http | response header（`w.Header()`） |
| connect（unary） | **`*connect.Error.Meta()`**（`attachReplyHeadersToConnectError`） |
| connect（stream） | response header + error.Meta（`copyReplyHeadersToConn` 在错误检查前执行） |

- connect unary 错误：实测确认（`integration/TestReplyHeaderOnErrorConnectUnary`）——reply header 在 `error.Meta`，不在 response header。
- 根因：connect 错误自包含（走 end-stream 帧，无独立 response header）；grpc header/trailer 分离。
- **影响**：只波及 transport 层 ReplyHeader（handler 手动设的响应头）；错误自身的 `Metadata`（客户端取参用的结构化参数）经错误编码器序列化，三协议一致，**不受影响**。

### 1.4 TODO — 错误契约的落地要点

- [ ] 错误工厂**结构化参数进 `Metadata`、`Message` 默认留空**（= 翻译）；只在需要直出透传时才填 `Message`。
- [ ] validator：把多条 violation（`protovalidate.ValidationError.Violations` → `{field, rule}`）作为**结构化参数**放进 Metadata，客户端据此渲染"字段 X 不合法"等本地文案。
- [ ] verbatim 透传纪律：`Message` 非空 = 故意直出（用户态文案）；**禁止把开发态英文塞进 `Message`**（会被客户端原样展示给用户）。可选 lint 或显式工厂（如 `ErrorXxxDirect(...)`）防误用。
- [ ]（可选 / 延后）出口 sanitize：unknown reason（裸 DB error、`fmt.Errorf` 经 kratos `FromError` 兜底会外泄 `err.Error()`）→ 抹成通用安全文案，内部细节只进日志。**与 i18n 无关，是安全收敛**；等有真实 handler / DB 再补即可。
- [ ]（低优先）transport ReplyHeader 错误路径三协议对齐。

---

## 二、k3s 部署适配（待实现）

适配 k3s 是 server 层的**系统性工作**（非注册一两个端点），含健康探针、优雅停机、监控、调试四块。metrics / pprof 上生产需要启用（用户更正了"只做核心"的初判）。

### 2.1 现状 gap

| 能力 | http (:11002) | grpc (:12002) | connect (:13002) |
|---|---|---|---|
| 中间件链（recovery/metadata/logging/validator） | ✓ | ✓ | ✓ |
| `grpc.health.v1` | ✗ | ✓（kratos 自动） | ✓（我们 `health.Register`） |
| HTTP 探针 `/healthz` `/readyz` | ✗ | n/a | `/healthz` |
| drain 窗口（`StopTimeout`） | ✗（默认 0 = 无 drain） | — | — |
| `/metrics`（Prometheus） | ✗ | ✗ | ✗ |
| `/debug/pprof` | ✗ | ✗ | ✗ |

> k8s Deployment 清单目前不存在（`deploy/k8s` 下只有 db/observability/storage 基础设施）。下述探针 endpoint 是"server 层先备好、清单写时直接用"。

### 2.2 关键发现（设计依据，均已核对源码）

- kratos App 默认 `stopTimeout=0`（`app.go:106` `if a.opts.stopTimeout > 0`）→ 收到 k3s SIGTERM 后**直接关 server，不 drain 在途请求**，滚动更新会丢请求。须显式设 `StopTimeout`。
- kratos App 有 `AfterStart` / `BeforeStop` 钩子（`options.go:36/37`）；`App.Stop()` **先跑 beforeStop 再 cancel**（`app.go:Stop`）→ readiness 可在 drain 前翻 503，让 endpoint controller 先摘流量。
- 各 server 的 `Stop(stopCtx)` **并发**执行（各自 goroutine，共享 `stopTimeout` 上下文）——不能靠"注册顺序"保证某个 server 先停。
- kratos grpc transport **默认开** health + reflection + admin（`admin.Register` 搭便车带来 `channelz.v1` + CSDS；`reflection.Register` 一次注册 v1 + v1alpha）；http 与我们的 connect 是 **opt-in**。这是**理念差异**（connect-go 极简、core 不强塞运维服务）**非能力差**——我们 connect 故意 opt-in：对齐 connect-go 习惯 + 让 core 可提取给上游 Kratos PR（spec §3.2）。

### 2.3 TODO — 健康探针

- [ ] http `/healthz`（liveness，handler 可达即 `200 "ok"`）+ `/readyz`（readiness，`ready` flag 驱动 `200/503`），用 kratos `srv.HandleFunc`（`transport/http/server.go:253`）。
- [ ] readiness 生命周期：包级 `var ready atomic.Bool` + `MarkReady()` / `MarkNotReady()`；`kratos.AfterStart` → `MarkReady`，`kratos.BeforeStop` → `MarkNotReady`（drain 前翻 503）。
- [ ] liveness / readiness 分离：停机时 **liveness 保持 200**（避免被 SIGKILL 打断 drain）、**readiness 翻 503**（摘流量）。
- [ ] k8s 探针指 `:11002`；`terminationGracePeriodSeconds`（默认 30s）> `graceful_stop_timeout`。
- [ ]（Phase A polish）readiness 接入依赖检查（DB/redis ready），把单一 serving flag 升级为 readiness 检查器；可选把 grpc/connect health 也接到同一个 readiness 控制器。

### 2.4 TODO — 优雅停机

- [ ] `kratos.StopTimeout(drain 窗口)`，默认 15s；`conf.proto` 的 `Server` 加 `google.protobuf.Duration graceful_stop_timeout = 4;`，`config.yaml` 给默认值。
- [ ] `newApp` 加 `*conf.Server` 参数（wire 注入，已在图里），读 `graceful_stop_timeout`；nil / ≤0 时默认 15s。
- [ ] 停机时序（已核对）：SIGTERM → `BeforeStop`（ready=false）→ `a.cancel()` → 各 server `Stop(stopCtx)` 并发 drain → main `defer cleanup()`。

### 2.5 TODO — 监控 `/metrics`（上生产启用；归 observability block）

- [ ] `/metrics`（Prometheus）：请求计数 / 延迟 / 错误率中间件埋点 + `promhttp` endpoint。
- [ ] 归 observability block：与 SigNoz 接入（trace / meter exporter）一起做。**注意**：昂贵的不是 endpoint 本身，而是事后补 ctx 传播 + 依赖注入——埋点（context.Context 透传 + 注入式 logger 而非全局 logger + otel span 包一层）现在就该做对，exporter 后接。

### 2.6 TODO — pprof（上生产启用）

- [ ] `/debug/pprof`，配置 / env 门控（默认关），in-cluster port-forward profile。
- [ ] 单独 debug port 或 http path；注意 prod 暴露风险（profile 开销 + goroutine / 内存状态泄漏）。

---

## 三、总结

- **错误体系机制已落地**（枚举 + `ErrorXxx`），三协议错误内容（reason / message）实测一致；code / 结构各协议原生，合理。
- **transport ReplyHeader 错误路径有差异**（connect unary → error.Meta），属边缘场景，不影响错误内容。
- **i18n 决策：后端不做，客户端翻译**。后端职责 = 返回稳定的 `Reason`（契约/翻译 key）+ 结构化 `Metadata` 参数 + `Message`（空=翻译 / 非空=直出透传）；不建任何后端 i18n 设施。
- **待办**：
  - 错误契约落地要点（结构化参数进 Metadata / validator violation / 可选出口 sanitize）。
  - k3s 部署适配——健康探针 / 优雅停机 / metrics / pprof 四块；metrics 归 observability block。
