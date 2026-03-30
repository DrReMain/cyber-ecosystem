# 非 Unary 接口实施计划

## 1. 背景与目标

### 1.1 问题背景

当前仓库是通用开发平台 monorepo，采用 Proto SSOT 设计。对于 Unary 接口，gRPC、ConnectRPC、HTTP 三协议完全一致。但非 Unary 接口（Streaming、文件传输等）无法通过 Proto SSOT 实现。

### 1.2 设计决策

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           设计决策                                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   决策一：Proto SSOT 仅覆盖 Unary 接口                                     │
│           三协议 (gRPC/ConnectRPC/HTTP) 的 Unary 接口完全一致              │
│                                                                             │
│   决策二：非 Unary 接口统一使用 HTTP 协议                                  │
│           通过独立 Handler 实现，不走 Kratos Transport                       │
│           不使用 WebSocket/MQTT，简化架构                                  │
│                                                                             │
│   决策三：作为通用平台，优先建立规范                                       │
│           代码组织、命名规范、设计模式、约束通过代码体现                    │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 1.3 Proto SSOT 边界

| 场景 | gRPC | ConnectRPC | HTTP | 实现方式 |
|------|------|------------|------|----------|
| **Unary CRUD** | ✅ Proto SSOT | ✅ Proto SSOT | ✅ Proto SSOT | Kratos Transport |
| **Server Streaming** | ✅ Proto SSOT | ✅ Proto SSOT | ❌ | 不实现 |
| **事件推送** | ❌ | ❌ | ✅ SSE | 独立 Handler |
| **流式拉取** | ❌ | ❌ | ✅ NDJSON | 独立 Handler |
| **文件传输** | ❌ | ❌ | ✅ Range | 独立 Handler |
| **流媒体** | ❌ | ❌ | ✅ HLS/DASH | Media Server |

---

## 2. Monorepo 代码组织

### 2.1 目录结构

```
cyber-ecosystem/
├── apps/                              # 应用层
│   └── app_1/
│       ├── api/                       # Proto 定义 (SSOT)
│       │   ├── v1/
│       │   │   ├── blog.proto        # Unary API (Proto SSOT)
│       │   │   └── stream.proto      # 消息结构定义 (非接口)
│       │   └── buf.gen.api.yaml
│       │
│       ├── gen/                      # 生成的代码
│       │   └── go/v1/
│       │
│       ├── services/
│       │   └── service_1/
│       │       └── internal/
│       │           ├── stream/        # 流式处理模块
│       │           │   ├── hub/      # 订阅管理
│       │           │   ├── handler/  # HTTP Handler
│       │           │   └── provider.go
│       │           │
│       │           ├── file/         # 文件处理模块
│       │           │   ├── handler/
│       │           │   └── provider.go
│       │           │
│       │           ├── biz/           # 业务逻辑层
│       │           │   ├── blog.go
│       │           │   └── file.go
│       │           │
│       │           └── service/       # 服务层
│       │               └── blog.go
│       │
│       └── clients/                   # 客户端 SDK (可选)
│           └── ts/
│               ├── sse-client.ts
│               ├── ndjson-client.ts
│               └── file-client.ts
│
├── shared-go/                         # Backend 共享库
│   └── kratos/
│       └── transport/
│           └── connect/
│
├── packages/                          # 共享配置包
│   ├── ts-config/
│   └── eslint-config/
│
└── contracts/                        # 共享契约定义
    └── go/common/
```

### 2.2 模块职责划分

| 目录 | 职责 | 归属 |
|------|------|------|
| `api/` | Proto 定义、消息结构 | 团队共同维护 |
| `stream/hub/` | 订阅管理、消息广播 | Backend 实现 |
| `stream/handler/` | HTTP Handler 实现 | Backend 实现 |
| `file/handler/` | 文件上传/下载 Handler | Backend 实现 |
| `biz/` | 业务逻辑 | 业务开发者 |
| `service/` | 服务层 (调用 biz) | 业务开发者 |
| `clients/ts/` | 客户端 SDK | 前端/移动端开发者 |

---

## 3. 代码规范 (Backend)

### 3.1 Handler 规范

所有独立 Handler 必须遵循以下结构：

```go
// internal/stream/handler/sse_handler.go

// Handler 命名: {功能}_handler.go
// 结构体命名: {功能}Handler

type SSEHandler struct {
    biz     *biz.StreamUC      // 依赖业务逻辑
    conf    *conf.SSE         // 依赖配置
    logger  *log.Helper       // 日志
}

// 构造函数命名: New{功能}Handler
// 参数顺序: Logger, Config, UC, 其他依赖
func NewSSEHandler(
    logger *log.Helper,
    conf *conf.SSE,
    biz *biz.StreamUC,
) *SSEHandler {
    return &SSEHandler{
        biz:    biz,
        conf:   conf,
        logger: log.NewHelper(log.With(logger, "module", "handler/sse")),
    }
}

// HTTP Handler 方法签名统一为 Handle(w http.ResponseWriter, r *http.Request)
func (h *SSEHandler) Handle(w http.ResponseWriter, r *http.Request) {
    // 1. 请求解析
    topics := parseTopics(r.URL.Query())
    
    // 2. 参数验证
    if err := h.validateTopics(topics); err != nil {
        h.writeError(w, err)
        return
    }
    
    // 3. 调用业务逻辑
    client, err := h.biz.Subscribe(r.Context(), topics)
    if err != nil {
        h.writeError(w, err)
        return
    }
    defer h.biz.Unsubscribe(client)
    
    // 4. 处理响应
    h.writeSSE(w, client)
}

// 私有方法统一以小写字母开头
func (h *SSEHandler) validateTopics(topics []string) error { ... }
func (h *SSEHandler) writeError(w http.ResponseWriter, err error) { ... }
func (h *SSEHandler) writeSSE(w http.ResponseWriter, client *SSEClient) { ... }
```

### 3.2 错误处理规范

```go
// internal/pkg/errors/errors.go

// 错误码定义
const (
    ErrCodeInvalidArgument = "INVALID_ARGUMENT"
    ErrCodeNotFound        = "NOT_FOUND"
    ErrCodeUnauthorized    = "UNAUTHORIZED"
    ErrCodeInternal        = "INTERNAL"
)

// 错误响应格式
type ErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details string `json:"details,omitempty"`
}

// NewError 创建错误
func NewError(code string, message string) error {
    return &handlerError{code: code, message: message}
}

// Handler 中错误处理
func (h *Handler) writeError(w http.ResponseWriter, err error) {
    var he *handlerError
    if errors.As(err, &he) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(httpStatusFromCode(he.code))
        json.NewEncoder(w).Encode(ErrorResponse{
            Code:    he.code,
            Message: he.message,
        })
        return
    }
    // 未知错误
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusInternalServerError)
    json.NewEncoder(w).Encode(ErrorResponse{
        Code:    ErrCodeInternal,
        Message: "internal server error",
    })
}

func httpStatusFromCode(code string) int {
    switch code {
    case ErrCodeInvalidArgument: return http.StatusBadRequest
    case ErrCodeNotFound:       return http.StatusNotFound
    case ErrCodeUnauthorized:   return http.StatusUnauthorized
    case ErrCodeInternal:       return http.StatusInternalServerError
    default:                    return http.StatusInternalServerError
    }
}
```

### 3.3 Biz 层规范

```go
// internal/stream/biz/stream.go

// UC 命名: {功能}UC
// 接口命名: {功能}Repo / {功能}Port

type StreamUC struct {
    log     *log.Helper
    broker  EventBroker   // 接口，用于依赖注入
}

type EventBroker interface {
    Subscribe(topics []string) (*Subscriber, error)
    Unsubscribe(*Subscriber) error
    Publish(topic string, event *StreamEvent) error
}

// 事件发布方法命名: On{业务动作}{实体}
type StreamEvent struct {
    Type    string
    Payload []byte
}

func (uc *StreamUC) OnBlogCreated(ctx context.Context, blog *BlogEntity) error {
    payload, _ := json.Marshal(blog)
    return uc.broker.Publish("blog.created", &StreamEvent{
        Type:    "BLOG_CREATED",
        Payload: payload,
    })
}
```

### 3.4 Provider 规范

```go
// internal/stream/provider.go

// Provider 文件命名: provider.go
// ProviderSet 变量命名: {模块}ProviderSet

var StreamProviderSet = wire.NewSet(
    hub.NewBroadcaster,
    NewStreamUC,
    handler.NewSSEHandler,
    handler.NewNDJSONHandler,
    wire.Bind(new(EventBroker), new(*hub.Broadcaster)),
)
```

### 3.5 配置规范

```go
// internal/conf/conf.proto

message SSE {
    int32 heartbeat_interval = 1;  // 心跳间隔 (秒)
    int32 max_connections = 2;     // 最大连接数
}

message Stream {
    int32 max_message_size = 1;   // 最大消息大小 (字节)
    int32 send_buffer_size = 2;    // 发送缓冲区大小
}

message File {
    string storage_path = 1;       // 存储路径
    int64 chunk_size = 2;          // 分片大小
    int64 max_file_size = 3;      // 最大文件大小
    repeated string allowed_types = 4; // 允许的文件类型
}
```

---

## 4. Proto 定义

### 4.1 stream.proto (消息结构，非接口)

```protobuf
syntax = "proto3";

package api.app_1.v1;

import "google/protobuf/timestamp.proto";
import "google/protobuf/any.proto";

// ============================================================================
// 事件类型枚举
// ============================================================================

enum StreamEventType {
    STREAM_EVENT_TYPE_UNSPECIFIED = 0;
    STREAM_EVENT_TYPE_BLOG_CREATED = 1;
    STREAM_EVENT_TYPE_BLOG_UPDATED = 2;
    STREAM_EVENT_TYPE_BLOG_DELETED = 3;
}

// ============================================================================
// SSE 事件格式
// ============================================================================

message SSEEvent {
    string id = 1;
    StreamEventType event_type = 2;
    string topic = 3;
    google.protobuf.Timestamp timestamp = 4;
    google.protobuf.Any data = 5;
}

// ============================================================================
// NDJSON Stream 帧格式
// ============================================================================

message NDJSONFrame {
    oneof frame {
        NDJSONMetadata metadata = 1;
        NDJSONMessage message = 2;
        NDJSONError error = 3;
        NDJSONTrailers trailers = 4;
    }
}

message NDJSONMetadata {
    string trace_id = 1;
    string request_id = 2;
}

message NDJSONMessage {
    int32 sequence = 1;
    string message_type = 2;
    bytes payload = 3;
}

message NDJSONError {
    string code = 1;
    string message = 2;
}

message NDJSONTrailers {
    string code = 1;
    string message = 2;
}

// ============================================================================
// 文件上传元数据
// ============================================================================

message FileMetadata {
    string upload_id = 1;
    string filename = 2;
    int64 total_size = 3;
    int64 chunk_size = 4;
    int32 chunk_count = 5;
}

message ChunkStatus {
    int32 chunk_index = 1;
    bool uploaded = 2;
    string checksum = 3;
}
```

---

## 5. 端点定义

### 5.1 HTTP 端点

| 方法 | 路径 | 协议 | 用途 |
|------|------|------|------|
| GET | `/sse/events` | SSE | 事件订阅推送 |
| GET | `/stream/blogs` | NDJSON | 博客流式拉取 |
| POST | `/files/init` | HTTP | 初始化上传 |
| POST | `/files/upload` | HTTP | 分片上传 |
| GET | `/files/{id}` | HTTP Range | 文件下载 |
| GET | `/files/{id}/chunks` | HTTP | 分片状态 |

### 5.2 响应格式

**SSE 事件格式**:
```json
data: {"id":"xxx","event":"BLOG_CREATED","topic":"blog","data":{"id":"1","title":"Hello"}}
```

**NDJSON 格式**:
```json
{"_type":"metadata","trace_id":"xxx"}
{"_type":"message","seq":1,"message_type":"Blog","payload":{"id":"1","title":"Hello"}}
{"_type":"trailers","code":"OK"}
```

---

## 6. 实施阶段

### Phase 1: Proto 定义与规范建立

**目标**: 定义消息结构，建立 Handler 开发规范

**任务**:
- [ ] 创建 `api/v1/stream.proto`
- [ ] 建立 `internal/pkg/errors/` 错误处理包
- [ ] 编写 Handler 模板注释

**产出**:
- Proto 定义文件
- 错误处理包
- Handler 开发规范文档

### Phase 2: SSE Handler 实现

**目标**: 实现 SSE 事件订阅推送

**任务**:
- [ ] 创建 `stream/hub/broadcaster.go` (订阅管理)
- [ ] 创建 `stream/handler/sse_handler.go`
- [ ] 创建 `stream/biz/stream.go`
- [ ] 创建 `stream/provider.go`
- [ ] 集成到 HTTP Server

**产出**:
- 可运行的 SSE Handler
- 单元测试
- 使用示例

### Phase 3: NDJSON Stream Handler 实现

**目标**: 实现 HTTP NDJSON 流式响应

**任务**:
- [ ] 创建 `stream/handler/ndjson_handler.go`
- [ ] 实现帧编码/解码
- [ ] 实现流式拉取接口
- [ ] 单元测试

**产出**:
- 可运行的 NDJSON Handler
- 单元测试

### Phase 4: File Handler 实现

**目标**: 实现分片上传/断点下载

**任务**:
- [ ] 创建 `file/handler/file_handler.go`
- [ ] 创建 `file/biz/file.go`
- [ ] 创建 `file/provider.go`
- [ ] 实现上传初始化
- [ ] 实现分片上传
- [ ] 实现断点下载

**产出**:
- 可运行的 File Handler
- 单元测试

### Phase 5: Biz 层集成

**目标**: 将事件发布集成到业务逻辑

**任务**:
- [ ] 修改 `service/blog.go` 调用事件发布
- [ ] 端到端测试

### Phase 6: Client SDK (可选)

**目标**: 提供 TypeScript 客户端 SDK

**任务**:
- [ ] `clients/ts/sse-client.ts`
- [ ] `clients/ts/ndjson-client.ts`
- [ ] `clients/ts/file-client.ts`
- [ ] 使用文档

---

## 7. 配置定义

### 7.1 conf.proto 扩展

```protobuf
message SSE {
    int32 heartbeat_interval = 1;
    int32 max_connections = 2;
}

message Stream {
    int32 max_message_size = 1;
    int32 send_buffer_size = 2;
}

message File {
    string storage_path = 1;
    int64 chunk_size = 2;
    int64 max_file_size = 3;
    repeated string allowed_types = 4;
}
```

### 7.2 config.yaml

```yaml
server:
  sse:
    heartbeat_interval: 30
    max_connections: 10000
  stream:
    max_message_size: 1048576
    send_buffer_size: 262144
  file:
    storage_path: /data/files
    chunk_size: 1048576
    max_file_size: 10737418240
    allowed_types:
      - image/*
      - video/*
      - application/pdf
```

---

## 8. 质量保障

### 8.1 代码规范检查

```bash
# lint
./nx run tools:golangci-lint

# format
go fmt ./...

# vet
go vet ./...
```

### 8.2 测试要求

| 组件 | 测试类型 | 覆盖率目标 |
|------|----------|------------|
| Handler | 单元测试 | 80% |
| Biz | 单元测试 | 80% |
| Hub/Broadcaster | 单元测试 | 90% |

### 8.3 集成测试

- SSE 连接和事件推送测试
- NDJSON 流式响应测试
- 文件上传/下载测试

---

## 9. 后续扩展

- [ ] 消息持久化
- [ ] 在线用户列表
- [ ] 直播管理 API
- [ ] TypeScript Client SDK
