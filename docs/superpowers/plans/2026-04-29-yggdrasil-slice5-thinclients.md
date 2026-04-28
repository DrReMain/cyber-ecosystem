# Slice 5: Thin Client Modules

> **闭环标准**: `shared-go/capabilities/` 编译通过 + auth middleware 集成测试。

## 目标

在 `shared-go/capabilities/` 下构建 thin client 模块，让业务服务（worklog 等）通过 gRPC 调用 IAM/Audit/Storage 服务的能力，而不需要本地实现业务逻辑。

每个模块由三部分组成：
1. **接口定义** — Go interface
2. **gRPC 客户端实现** — 调用对应服务的内部 API
3. **本地应用逻辑** — 中间件、拦截器、缓存

## 前置条件

- Slice 4 完成（IAM 服务全部内部 API 可用）
- Audit 服务的 `AuditSinkService` gRPC 端点可用

---

## Step 1: auth 模块 — JWT 校验 + session 验证

**文件**: `shared-go/capabilities/auth/auth.go`

```go
package auth

import (
	"context"
)

// Authenticator validates tokens and returns user identity.
type Authenticator interface {
	ValidateToken(ctx context.Context, token string) (*Identity, error)
}

type Identity struct {
	UserID    string
	SessionID string
}

// ProviderSet wires the gRPC client implementation.
```

**文件**: `shared-go/capabilities/auth/client.go`

```go
package auth

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	yggdrasilV1 "cyber-ecosystem/apps/yggdrasil/gen/go/v1"
)

type grpcClient struct {
	conn   *grpc.ClientConn
	client yggdrasilV1.InternalAuthServiceClient
}

func NewGRPCClient(iamAddr string) (*grpcClient, error) {
	conn, err := grpc.NewClient(iamAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to iam: %w", err)
	}
	return &grpcClient{
		conn:   conn,
		client: yggdrasilV1.NewInternalAuthServiceClient(conn),
	}, nil
}

func (c *grpcClient) ValidateToken(ctx context.Context, token string) (*Identity, error) {
	resp, err := c.client.ValidateToken(ctx, &yggdrasilV1.ValidateTokenRequest{Token: token})
	if err != nil {
		return nil, err
	}
	if !resp.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return &Identity{
		UserID:    resp.UserId,
		SessionID: resp.SessionId,
	}, nil
}

func (c *grpcClient) Close() error {
	return c.conn.Close()
}
```

**文件**: `shared-go/capabilities/auth/middleware.go`

```go
package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

type contextKey string

const identityKey contextKey = "auth_identity"

func IdentityFromContext(ctx context.Context) (*Identity, error) {
	v := ctx.Value(identityKey)
	if v == nil {
		return nil, fmt.Errorf("no identity in context")
	}
	id, ok := v.(*Identity)
	if !ok {
		return nil, fmt.Errorf("invalid identity type")
	}
	return id, nil
}

func WithIdentity(ctx context.Context, id *Identity) context.Context {
	return context.WithValue(ctx, identityKey, id)
}

// Middleware validates the Bearer token via IAM internal API.
func Middleware(authenticator Authenticator) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			tr, _ := transport.FromServerContext(ctx)
			if tr == nil {
				return handler(ctx, req)
			}
			// Extract token from header
			token := extractToken(tr.RequestHeader())
			if token == "" {
				return handler(ctx, req) // Let selector handle public paths
			}
			identity, err := authenticator.ValidateToken(ctx, token)
			if err != nil {
				return nil, fmt.Errorf("authentication failed: %w", err)
			}
			ctx = WithIdentity(ctx, identity)
			return handler(ctx, req)
		}
	}
}

func extractToken(header transport.Header) string {
	auth := header.Get("Authorization")
	if auth == "" {
		return ""
	}
	return strings.TrimPrefix(auth, "Bearer ")
}
```

---

## Step 2: rbac 模块 — 鉴权决策客户端

**文件**: `shared-go/capabilities/rbac/rbac.go`

```go
package rbac

import "context"

// Authorizer checks user permissions via IAM.
type Authorizer interface {
	CheckPermission(ctx context.Context, userID, resource string) (bool, error)
	GetRolesForUser(ctx context.Context, userID string) ([]string, error)
}
```

**文件**: `shared-go/capabilities/rbac/client.go`

```go
package rbac

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	yggdrasilV1 "cyber-ecosystem/apps/yggdrasil/gen/go/v1"
)

type grpcClient struct {
	conn   *grpc.ClientConn
	client yggdrasilV1.InternalAuthorizationServiceClient
}

func NewGRPCClient(iamAddr string) (*grpcClient, error) {
	conn, err := grpc.NewClient(iamAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return &grpcClient{conn: conn, client: yggdrasilV1.NewInternalAuthorizationServiceClient(conn)}, nil
}

func (c *grpcClient) CheckPermission(ctx context.Context, userID, resource string) (bool, error) {
	resp, err := c.client.CheckPermission(ctx, &yggdrasilV1.CheckPermissionRequest{
		UserId:   userID,
		Resource: resource,
	})
	if err != nil {
		return false, err
	}
	return resp.Allowed, nil
}

func (c *grpcClient) GetRolesForUser(ctx context.Context, userID string) ([]string, error) {
	resp, err := c.client.GetRolesForUser(ctx, &yggdrasilV1.GetRolesForUserRequest{UserId: userID})
	if err != nil {
		return nil, err
	}
	return resp.Roles, nil
}

func (c *grpcClient) Close() error { return c.conn.Close() }
```

**文件**: `shared-go/capabilities/rbac/middleware.go`

```go
package rbac

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	"cyber-ecosystem/shared-go/capabilities/auth"
)

func Middleware(authorizer Authorizer) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			identity, err := auth.IdentityFromContext(ctx)
			if err != nil {
				return handler(ctx, req)
			}
			tr, _ := transport.FromServerContext(ctx)
			if tr == nil {
				return handler(ctx, req)
			}
			allowed, err := authorizer.CheckPermission(ctx, identity.UserID, tr.Operation())
			if err != nil {
				return nil, err
			}
			if !allowed {
				return nil, fmt.Errorf("access denied")
			}
			return handler(ctx, req)
		}
	}
}
```

---

## Step 3: audit 模块 — 审计事件发送

**文件**: `shared-go/capabilities/audit/audit.go`

```go
package audit

import "context"

type Event struct {
	ActorID      string
	ActorName    string
	DepartmentID string
	Action       string
	ResourceType string
	ResourceID   string
	ServiceName  string
	Result       string
	IP           string
	UserAgent    string
	Detail       string
}

type Sender interface {
	Send(ctx context.Context, events []*Event) error
}
```

**文件**: `shared-go/capabilities/audit/client.go`

```go
package audit

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	yggdrasilV1 "cyber-ecosystem/apps/yggdrasil/gen/go/v1"
)

type grpcClient struct {
	conn   *grpc.ClientConn
	client yggdrasilV1.AuditSinkServiceClient
}

func NewGRPCClient(auditAddr string) (*grpcClient, error) {
	conn, err := grpc.NewClient(auditAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return &grpcClient{conn: conn, client: yggdrasilV1.NewAuditSinkServiceClient(conn)}, nil
}

func (c *grpcClient) Send(ctx context.Context, events []*Event) error {
	pbEvents := make([]*yggdrasilV1.AuditEvent, len(events))
	for i, e := range events {
		pbEvents[i] = &yggdrasilV1.AuditEvent{
			ActorId:      e.ActorID,
			ActorName:    e.ActorName,
			DepartmentId: e.DepartmentID,
			Action:       e.Action,
			ResourceType: e.ResourceType,
			ResourceId:   e.ResourceID,
			ServiceName:  e.ServiceName,
			Result:       e.Result,
			Ip:           e.IP,
			UserAgent:    e.UserAgent,
			Detail:       e.Detail,
		}
	}
	_, err := c.client.SubmitEvents(ctx, &yggdrasilV1.SubmitEventsRequest{Events: pbEvents})
	return err
}

func (c *grpcClient) Close() error { return c.conn.Close() }
```

**文件**: `shared-go/capabilities/audit/buffered.go`

```go
package audit

import (
	"context"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// BufferedSender batches events and sends them periodically.
type BufferedSender struct {
	sender  Sender
	buffer  []*Event
	mu      sync.Mutex
	flush   chan struct{}
	stop    chan struct{}
	log     *log.Helper
	maxSize int
}

func NewBufferedSender(sender Sender, logger log.Logger, maxSize int, interval time.Duration) *BufferedSender {
	bs := &BufferedSender{
		sender:  sender,
		buffer:  make([]*Event, 0, maxSize),
		flush:   make(chan struct{}, 1),
		stop:    make(chan struct{}),
		log:     log.NewHelper(log.With(logger, "module", "audit/buffered")),
		maxSize: maxSize,
	}
	go bs.run(interval)
	return bs
}

func (bs *BufferedSender) Send(ctx context.Context, events []*Event) error {
	bs.mu.Lock()
	bs.buffer = append(bs.buffer, events...)
	shouldFlush := len(bs.buffer) >= bs.maxSize
	bs.mu.Unlock()

	if shouldFlush {
		select {
		case bs.flush <- struct{}{}:
		default:
		}
	}
	return nil // Best-effort buffering
}

func (bs *BufferedSender) Close() {
	close(bs.stop)
}

func (bs *BufferedSender) run(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			bs.doFlush()
		case <-bs.flush:
			bs.doFlush()
		case <-bs.stop:
			bs.doFlush()
			return
		}
	}
}

func (bs *BufferedSender) doFlush() {
	bs.mu.Lock()
	events := bs.buffer
	bs.buffer = make([]*Event, 0, bs.maxSize)
	bs.mu.Unlock()

	if len(events) == 0 {
		return
	}
	if err := bs.sender.Send(context.Background(), events); err != nil {
		bs.log.Errorf("flush audit events failed: %v", err)
		// Re-queue on failure? Or drop? For simplicity, we drop.
	}
}
```

---

## Step 4: datascope 模块 — 数据权限客户端

**文件**: `shared-go/capabilities/datascope/datascope.go`

```go
package datascope

import "context"

type ScopeRule struct {
	Field string
	Op    string
	Value string
}

type EffectiveScope struct {
	IsAll           bool
	SelfFilter      bool
	DeptFilter      bool
	AttributeFilter bool
	DeptIDs         []string
	Rules           []ScopeRule
	Logic           string
}

type Resolver interface {
	ResolveScope(ctx context.Context, userID, resource string) (*EffectiveScope, error)
}
```

**文件**: `shared-go/capabilities/datascope/client.go`

```go
package datascope

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	yggdrasilV1 "cyber-ecosystem/apps/yggdrasil/gen/go/v1"
)

type grpcClient struct {
	conn   *grpc.ClientConn
	client yggdrasilV1.InternalScopeServiceClient
}

func NewGRPCClient(iamAddr string) (*grpcClient, error) {
	conn, err := grpc.NewClient(iamAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return &grpcClient{conn: conn, client: yggdrasilV1.NewInternalScopeServiceClient(conn)}, nil
}

func (c *grpcClient) ResolveScope(ctx context.Context, userID, resource string) (*EffectiveScope, error) {
	resp, err := c.client.ResolveScope(ctx, &yggdrasilV1.ResolveScopeRequest{
		UserId:   userID,
		Resource: resource,
	})
	if err != nil {
		return nil, err
	}
	scope := &EffectiveScope{
		IsAll:           resp.IsAll,
		SelfFilter:      resp.SelfFilter,
		DeptFilter:      resp.DeptFilter,
		DeptIDs:         resp.DeptIds,
		Logic:           resp.Logic,
	}
	for _, r := range resp.Rules {
		scope.Rules = append(scope.Rules, ScopeRule{Field: r.Field, Op: r.Op, Value: r.Value})
	}
	return scope, nil
}

func (c *grpcClient) Close() error { return c.conn.Close() }
```

---

## Step 5: condition 模块 — ABAC 条件客户端

**文件**: `shared-go/capabilities/condition/condition.go`

```go
package condition

import "context"

type Checker interface {
	CheckConditions(ctx context.Context, userID, operation string) (bool, error)
}
```

**文件**: `shared-go/capabilities/condition/client.go`

```go
package condition

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	yggdrasilV1 "cyber-ecosystem/apps/yggdrasil/gen/go/v1"
)

type grpcClient struct {
	conn   *grpc.ClientConn
	client yggdrasilV1.InternalConditionServiceClient
}

func NewGRPCClient(iamAddr string) (*grpcClient, error) {
	conn, err := grpc.NewClient(iamAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return &grpcClient{conn: conn, client: yggdrasilV1.NewInternalConditionServiceClient(conn)}, nil
}

func (c *grpcClient) CheckConditions(ctx context.Context, userID, operation string) (bool, error) {
	resp, err := c.client.CheckConditions(ctx, &yggdrasilV1.CheckConditionsRequest{
		UserId:    userID,
		Operation: operation,
	})
	if err != nil {
		return false, err
	}
	return resp.Allowed, nil
}

func (c *grpcClient) Close() error { return c.conn.Close() }
```

**文件**: `shared-go/capabilities/condition/middleware.go`

```go
package condition

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	"cyber-ecosystem/shared-go/capabilities/auth"
)

func Middleware(checker Checker) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			identity, err := auth.IdentityFromContext(ctx)
			if err != nil {
				return handler(ctx, req)
			}
			tr, _ := transport.FromServerContext(ctx)
			if tr == nil {
				return handler(ctx, req)
			}
			allowed, err := checker.CheckConditions(ctx, identity.UserID, tr.Operation())
			if err != nil {
				return nil, err
			}
			if !allowed {
				return nil, fmt.Errorf("condition denied")
			}
			return handler(ctx, req)
		}
	}
}
```

---

## Step 6: security 模块 — 中间件编排

**文件**: `shared-go/capabilities/security/security.go`

```go
package security

import (
	"github.com/go-kratos/kratos/v2/middleware"

	"cyber-ecosystem/shared-go/capabilities/auth"
	"cyber-ecosystem/shared-go/capabilities/rbac"
	"cyber-ecosystem/shared-go/capabilities/condition"
)

// Options configures which capabilities to enable.
type Options struct {
	Auth      bool
	RBAC      bool
	Condition bool
	DataScope bool
}

// BuildAuthMiddlewares constructs the auth-gated middleware chain.
// Called by business service's buildMiddlewares() function.
func BuildAuthMiddlewares(
	opts Options,
	authenticator auth.Authenticator,
	authorizer rbac.Authorizer,
	conditionChecker condition.Checker,
	scopeInjector middleware.Middleware,
) []middleware.Middleware {
	var mws []middleware.Middleware
	if opts.Auth {
		mws = append(mws, auth.Middleware(authenticator))
	}
	if opts.RBAC {
		mws = append(mws, rbac.Middleware(authorizer))
	}
	if opts.Condition {
		mws = append(mws, condition.Middleware(conditionChecker))
	}
	if opts.DataScope && scopeInjector != nil {
		mws = append(mws, scopeInjector)
	}
	return mws
}
```

---

## Step 7: 编译验证

```bash
# 确保所有 generated proto 代码存在
./nx run yggdrasil_api:proto:api

# 编译 shared-go
cd shared-go && go build ./...
```

验证: 无编译错误

提交:
```bash
git add shared-go/capabilities/
git commit -m "feat(yggdrasil): add thin client capability modules"
```

---

## 完成标准

- [x] `shared-go/capabilities/auth/` — JWT 校验客户端 + 中间件
- [x] `shared-go/capabilities/rbac/` — 鉴权决策客户端 + 中间件
- [x] `shared-go/capabilities/audit/` — 审计事件客户端 + 缓冲发送
- [x] `shared-go/capabilities/datascope/` — DataScope 规则客户端
- [x] `shared-go/capabilities/condition/` — ABAC 条件客户端 + 中间件
- [x] `shared-go/capabilities/security/` — 中间件编排
- [x] `go build ./...` 在 `shared-go/` 编译通过
- [x] 变更已提交
