# 错误体系：机制、transport 行为与 i18n 对接

整合错误机制设计、transport 层错误返回行为（实测）、以及 i18n 对接的后续工作。

## 一、分析

### 1.1 错误机制

- 所有错误通过 proto 枚举定义；`protoc-gen-go-errors` 生成工厂 `ErrorXxx(format, args...) *errors.Error`。
- 每个错误携带：`Reason`（枚举 key，i18n 翻译 key）、`Message`（文案）、`Metadata`（模板变量）。
- 中间件错误（validator / recovery 等）在 app `init()` 里覆盖为枚举工厂产出（Reason = 枚举 key），与业务错误统一。

### 1.2 i18n vs override：用 Message 是否为空当信号

- `ErrorXxx("")` → Message 空 → **i18n 模式**：Reason 当 key 翻译，Metadata 当模板变量。
- `ErrorXxx("文案")` → Message 非空 → **override 模式**：原样直出，不翻译。
- 相比旧仓库的魔法 metadata 字符串开关：无魔法串、复用 `format` 参数、语义清晰、生成函数不改。去掉 fallback（key 缺失即暴露）。

### 1.3 错误返回的 transport 行为（实测）

#### a. 错误内容三协议一致 ✓

mobile_bff 实测（invalid = 校验错误 / valid = 业务错误）：

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

kratos `Transporter` 只暴露 `ReplyHeader()`（无 trailer API）。handler 经 `tr.ReplyHeader().Set(k,v)`
设的响应头，错误路径去向不一致：

| 协议 | 错误时 ReplyHeader 去向 |
|---|---|
| grpc | response header（`grpc.SetHeader` 无条件） |
| http | response header（`w.Header()`） |
| connect（unary） | **`*connect.Error.Meta()`**（`attachReplyHeadersToConnectError`） |
| connect（stream） | response header + error.Meta（`copyReplyHeadersToConn` 在错误检查前执行） |

- connect unary 错误：实测确认（`integration/TestReplyHeaderOnErrorConnectUnary`）——reply header 在
  `error.Meta`，不在 response header。
- 根因：connect 错误自包含（走 end-stream 帧，无独立 response header）；grpc header/trailer 分离。
- **影响**：只波及 transport 层 ReplyHeader（handler 手动设的响应头）；错误自身的 `Metadata`
  （i18n 模板变量等）经错误编码器序列化，三协议一致，**不受影响**。

## 二、TODO（i18n 阶段）

- [ ] `i18n.Server(bundle)` 中间件：拦截响应错误；Message 空 → `Localize(Reason, Metadata, lang)`；
      非空 → 原样（override）。构建新错误对象（不改共享 / 包级错误）。
- [ ] validator：提取多条 violation（`protovalidate.ValidationError.Violations` → `{field, message}`），
      随错误携带；默认留服务端，直暴模式可选（信号机制复用 Message 非空 / 结构化 metadata，i18n 阶段定）。
- [ ] 错误 Reason → 翻译 bundle 的生成 / 维护（key 来自枚举）。
- [ ] 语言解析（transport header / context）。
- [ ]（低优先）transport ReplyHeader 错误路径对齐：让 connect 错误也填 response header，使三协议一致。

## 三、总结

- **错误内容（reason / message）三协议一致**：grpc / connect(gRPC) / http / connect(Connect) 四种方式
  实测通过；code / 结构各协议原生，合理。
- **transport ReplyHeader 错误路径有差异**（connect unary → error.Meta），但属边缘场景（handler 在错误路径
  设 transport 响应头），不影响错误内容。
- **i18n 机制**：用 Message 是否为空当 i18n / override 信号（去掉旧仓库魔法 metadata 字符串）；当前错误
  机制（枚举 + ErrorXxx）已天然支持，i18n 中间件待实现。
