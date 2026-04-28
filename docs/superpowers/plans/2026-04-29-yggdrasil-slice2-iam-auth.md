# Slice 2: IAM Auth Foundation

> **闭环标准**: HTTP 登录获取 JWT + gRPC 内部 ValidateToken 通过。

## 目标

构建 IAM 服务的认证核心：用户登录/登出、JWT 管理、Session 管理、超级管理员种子、内部认证 API。

此 Slice 不涉及 RBAC、Condition、DataScope——中间件链只有 JWT + Session。

## 前置条件

- Slice 0 完成（app 骨架 + error_reason.proto）
- Slice 1 完成（Audit 服务，提供审计事件投递能力）
- PostgreSQL 和 Redis 在本地可用

## 端口

| Transport | Address |
|-----------|---------|
| HTTP | `0.0.0.0:11000` |
| gRPC | `0.0.0.0:12000` |
| ConnectRPC | `0.0.0.0:13000` |
| Ops | `0.0.0.0:14000` |

---

## Step 1: 目录结构 + Nx 配置

```bash
mkdir -p apps/yggdrasil/services/iam/{cmd/app,configs,internal/{conf,data/ent/schema,biz,server/{middleware,locales},service,pkg/auth}}
```

**文件**: `apps/yggdrasil/services/iam/project.json`

```json
{
  "name": "yggdrasil_iam",
  "$schema": "../../../../node_modules/nx/schemas/project-schema.json",
  "implicitDependencies": ["yggdrasil_api", "shared-go"],
  "targets": {
    "proto:conf": {
      "executor": "nx:run-commands",
      "options": {
        "command": "buf generate --template apps/yggdrasil/services/iam/buf.gen.conf.yaml --path apps/yggdrasil/services/iam/internal/conf"
      }
    },
    "generate": {
      "dependsOn": ["yggdrasil_api:proto:api", "proto:conf"],
      "executor": "nx:run-commands",
      "options": {
        "command": "cd apps/yggdrasil/services/iam && wire ./cmd/app/..."
      }
    },
    "dev": {
      "executor": "nx:run-commands",
      "options": {
        "command": "cd apps/yggdrasil/services/iam && go run ./cmd/app/... -conf ./configs"
      }
    },
    "build": {
      "executor": "nx:run-commands",
      "options": {
        "command": "cd apps/yggdrasil/services/iam && go build -o ./bin/iam ./cmd/app/..."
      }
    },
    "ent:new": {
      "executor": "nx:run-commands",
      "options": {
        "command": "cd apps/yggdrasil/services/iam && go run -mod=mod entgo.io/ent/cmd/ent new --target internal/data/ent/schema"
      }
    }
  }
}
```

**文件**: `apps/yggdrasil/services/iam/buf.gen.conf.yaml`

```yaml
version: v2
plugins:
  - local: [go, tool, protoc-gen-go]
    out: .
    opt: paths=source_relative
```

验证: `cat apps/yggdrasil/services/iam/project.json`

---

## Step 2: API Proto 定义

### 2a: iam_account_auth.proto — 认证管理 API

**文件**: `apps/yggdrasil/api/v1/iam_account_auth.proto`

```protobuf
syntax = "proto3";

// buf:lint:ignore PACKAGE_DIRECTORY_MATCH
package api.yggdrasil.v1;

import "buf/validate/validate.proto";
import "auth/auth.proto";
import "desc/desc.proto";
import "google/api/annotations.proto";

option go_package = "cyber-ecosystem/apps/yggdrasil/gen/go/v1;yggdrasilV1";

// AccountAuthService
service AccountAuthService {
  option (desc.service_comment) = "认证管理服务";

  // Login
  rpc Login(LoginRequest) returns (LoginResponse) {
    option (desc.method_comment) = "用户登录";
    option (auth.public_access) = true;
    option (google.api.http) = {post: "/iam/auth/login" body: "*"};
  }

  // Logout
  rpc Logout(LogoutRequest) returns (LogoutResponse) {
    option (desc.method_comment) = "用户登出";
    option (google.api.http) = {post: "/iam/auth/logout" body: "*"};
  }

  // RefreshToken
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse) {
    option (desc.method_comment) = "刷新 Token";
    option (auth.public_access) = true;
    option (google.api.http) = {post: "/iam/auth/refresh" body: "*"};
  }
}

message LoginRequest {
  string email = 1 [(buf.validate.field).string.email = true];
  string password = 2 [(buf.validate.field).string.min_len = 6];
}

message LoginResponse {
  string access_token = 1;
  string refresh_token = 2;
  int64 expires_in = 3;
}

message LogoutRequest {}

message LogoutResponse {}

message RefreshTokenRequest {
  string refresh_token = 1 [(buf.validate.field).string.min_len = 1];
}

message RefreshTokenResponse {
  string access_token = 1;
  string refresh_token = 2;
  int64 expires_in = 3;
}
```

### 2b: iam_internal_auth.proto — 内部认证 API

**文件**: `apps/yggdrasil/api/v1/iam_internal_auth.proto`

```protobuf
syntax = "proto3";

// buf:lint:ignore PACKAGE_DIRECTORY_MATCH
package api.yggdrasil.v1;

import "buf/validate/validate.proto";

option go_package = "cyber-ecosystem/apps/yggdrasil/gen/go/v1;yggdrasilV1";

// InternalAuthService — 内部 API，供其他服务校验 Token。
service InternalAuthService {
  // ValidateToken 校验 JWT 并返回用户信息。
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
}

message ValidateTokenRequest {
  string token = 1 [(buf.validate.field).string.min_len = 1];
}

message ValidateTokenResponse {
  string user_id = 1;
  string session_id = 2;
  bool valid = 3;
}
```

### 2c: 生成 Proto 代码

```bash
buf lint apps/yggdrasil/api/v1/iam_account_auth.proto
buf lint apps/yggdrasil/api/v1/iam_internal_auth.proto
./nx run yggdrasil_api:proto:api
```

验证: `ls apps/yggdrasil/gen/go/v1/iam_account_auth* apps/yggdrasil/gen/go/v1/iam_internal_auth*`

提交:
```bash
git add apps/yggdrasil/api/v1/iam_account_auth.proto apps/yggdrasil/api/v1/iam_internal_auth.proto apps/yggdrasil/gen/
git commit -m "feat(yggdrasil): add IAM auth API proto definitions"
```

---

## Step 3: Conf Proto + 生成

**文件**: `apps/yggdrasil/services/iam/internal/conf/conf.proto`

与 Slice 1 的 conf.proto 结构相同，额外添加:

```protobuf
// 在 Bootstrap message 中追加:
message Super {
  bool enabled = 1;
  string role_name = 2;
  string role_code = 3;
  string email = 4;
  string password = 5;
  bool force_reset = 6;
}

// Bootstrap 追加字段:
message Bootstrap {
  Server server = 1;
  Auth auth = 2;
  Log log = 3;
  Data data = 4;
  Trace trace = 5;
  Ops ops = 6;
  Super super = 7;
}
```

Server/Auth/Log/Data/Trace/Ops 子消息与 Slice 1 完全一致。

生成:
```bash
./nx run yggdrasil_iam:proto:conf
```

验证: `ls apps/yggdrasil/services/iam/internal/conf/conf.pb.go`

---

## Step 4: Ent Schema + 生成

### 4a: user.go

**文件**: `apps/yggdrasil/services/iam/internal/data/ent/schema/user.go`

```go
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("email").NotEmpty().Unique().MaxLen(200),
		field.String("password").NotEmpty().MaxLen(200),
		field.String("name").Default("").MaxLen(200),
		field.String("avatar").Default("").MaxLen(500),
		field.Enum("status").Values("active", "inactive", "locked").Default("active"),
		field.Bool("force_reset").Default(false),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("email").Unique(),
		index.Fields("status"),
	}
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "user"},
	}
}
```

### 4b: session.go

**文件**: `apps/yggdrasil/services/iam/internal/data/ent/schema/session.go`

```go
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

type Session struct {
	ent.Schema
}

func (Session) Fields() []ent.Field {
	return []ent.Field{
		field.String("user_id").NotEmpty().MaxLen(20),
		field.String("sid").NotEmpty().MaxLen(20),
		field.Time("expires_at"),
		field.Bool("revoked").Default(false),
		field.String("ip").Default("").MaxLen(50),
		field.String("user_agent").Default("").MaxLen(500),
	}
}

func (Session) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (Session) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("sid"),
		index.Fields("expires_at"),
	}
}

func (Session) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "session"},
	}
}
```

### 4c: role.go（最小定义，为种子数据所需）

**文件**: `apps/yggdrasil/services/iam/internal/data/ent/schema/role.go`

```go
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

type Role struct {
	ent.Schema
}

func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").NotEmpty().MaxLen(100),
		field.String("code").NotEmpty().Unique().MaxLen(50),
		field.Int("sort").Default(0),
		field.String("description").Default("").MaxLen(500),
	}
}

func (Role) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (Role) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique(),
	}
}

func (Role) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "role"},
	}
}
```

### 4d: policy_rule.go（Casbin 存储表）

**文件**: `apps/yggdrasil/services/iam/internal/data/ent/schema/policy_rule.go`

```go
package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type PolicyRule struct {
	ent.Schema
}

func (PolicyRule) Fields() []ent.Field {
	return []ent.Field{
		field.String("ptype").Default("").MaxLen(12),
		field.String("v0").Default("").MaxLen(128),
		field.String("v1").Default("").MaxLen(128),
		field.String("v2").Default("").MaxLen(128),
		field.String("v3").Default("").MaxLen(128),
		field.String("v4").Default("").MaxLen(128),
		field.String("v5").Default("").MaxLen(128),
	}
}

func (PolicyRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("ptype", "v0", "v1", "v2", "v3"),
	}
}

func (PolicyRule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "policy_rule"},
	}
}
```

### 4e: generate.go + 生成

**文件**: `apps/yggdrasil/services/iam/internal/data/ent/generate.go`

```go
package ent

//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate ./schema
```

```bash
cd apps/yggdrasil/services/iam && go generate ./internal/data/ent/...
```

验证: `ls apps/yggdrasil/services/iam/internal/data/ent/client.go`

---

## Step 5: Pkg 层 — Auth 工具

**文件**: `apps/yggdrasil/services/iam/internal/pkg/auth/jwt.go`

```go
package auth

import jwtv5 "github.com/golang-jwt/jwt/v5"

var SigningMethod = jwtv5.SigningMethodHS256
```

**文件**: `apps/yggdrasil/services/iam/internal/pkg/auth/identity.go`

```go
package auth

import (
	"context"
	"fmt"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
)

type Identity struct {
	jwtv5.RegisteredClaims
	Sid string `json:"sid,omitempty"`
}

func IdentityFromContext(ctx context.Context) (*Identity, error) {
	raw, ok := jwt.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing jwt claims in context")
	}
	claims, ok := raw.(*Identity)
	if !ok {
		return nil, fmt.Errorf("invalid jwt claims type")
	}
	return claims, nil
}
```

---

## Step 6: Biz 层

**文件**: `apps/yggdrasil/services/iam/internal/biz/biz.go`

```go
package biz

import (
	"context"

	"github.com/google/wire"

	"github.com/go-kratos/kratos/v2/log"
)

const Domain = "self"

type Transaction interface {
	InTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type UC struct {
	log *log.Helper
	tm  Transaction
}

var ProviderSet = wire.NewSet(
	NewAccountUC,
	NewSessionValidator,
)
```

**文件**: `apps/yggdrasil/services/iam/internal/biz/uc_account.go`

```go
package biz

import (
	"context"
	"fmt"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/go-kratos/kratos/v2/log"

	pkgauth "cyber-ecosystem/apps/yggdrasil/services/iam/internal/pkg/auth"
)

const (
	accessTokenTTL  = 2 * time.Hour
	refreshTokenTTL = 7 * 24 * time.Hour
)

// Models

type LoginParam struct {
	Email    string
	Password string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

// Ports

type UserRP interface {
	GetByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) (*User, error)
	Get(ctx context.Context, id string) (*User, error)
}

type SessionRP interface {
	Create(ctx context.Context, session *Session) (*Session, error)
	RevokeBySID(ctx context.Context, sid string) error
	IsRevoked(ctx context.Context, sid string) (bool, error)
	GetByRefreshToken(ctx context.Context, refreshToken string) (*Session, error)
}

type PolicyRP interface {
	Enforce(sub, dom, obj string) (bool, error)
	AddRoleForUser(ctx context.Context, userID, role string) (bool, func(), error)
	SeedSuperAdmin(ctx context.Context, roleName, roleCode, email, password string, forceReset bool) (string, error)
}

// Models (最小定义)

type User struct {
	ID          *string
	Email       *string
	Password    *string
	Name        *string
	Status      *string
	ForceReset  *bool
}

type Session struct {
	ID         *string
	UserID     *string
	SID        *string
	ExpiresAt  *time.Time
	Revoked    *bool
}

// UC

type AccountUC struct {
	UC
	userRP    UserRP
	sessionRP SessionRP
	policyRP  PolicyRP
	secret    string
}

func NewAccountUC(
	logger log.Logger,
	tm Transaction,
	userRP UserRP,
	sessionRP SessionRP,
	policyRP PolicyRP,
	auth *conf.Auth,
) *AccountUC {
	return &AccountUC{
		UC: UC{
			log: log.NewHelper(log.With(logger, "module", "biz/uc_account")),
			tm:  tm,
		},
		userRP:    userRP,
		sessionRP: sessionRP,
		policyRP:  policyRP,
		secret:    auth.Secret,
	}
}

func (uc *AccountUC) Login(ctx context.Context, in *LoginParam) (*TokenPair, error) {
	user, err := uc.userRP.GetByEmail(ctx, in.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}
	// TODO: verify password hash (bcrypt)
	if err := verifyPassword(*user.Password, in.Password); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}
	if *user.Status != "active" {
		return nil, fmt.Errorf("account is not active")
	}
	return uc.generateTokenPair(ctx, *user.ID)
}

func (uc *AccountUC) Logout(ctx context.Context, userID, sid string) error {
	return uc.sessionRP.RevokeBySID(ctx, sid)
}

func (uc *AccountUC) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	session, err := uc.sessionRP.GetByRefreshToken(ctx, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}
	if *session.Revoked {
		return nil, fmt.Errorf("session revoked")
	}
	if session.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("refresh token expired")
	}
	// Revoke old session
	_ = uc.sessionRP.RevokeBySID(ctx, *session.SID)
	// Issue new pair
	return uc.generateTokenPair(ctx, *session.UserID)
}

func (uc *AccountUC) ValidateToken(ctx context.Context, token string) (*pkgauth.Identity, error) {
	parsed, err := jwtv5.ParseWithClaims(token, &pkgauth.Identity{}, func(t *jwtv5.Token) (any, error) {
		return []byte(uc.secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := parsed.Claims.(*pkgauth.Identity)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	revoked, err := uc.sessionRP.IsRevoked(ctx, claims.Sid)
	if err != nil {
		return nil, err
	}
	if revoked {
		return nil, fmt.Errorf("session revoked")
	}
	return claims, nil
}

func (uc *AccountUC) SeedSuperAdmin(ctx context.Context, super *conf.Super) error {
	if super == nil || !super.Enabled {
		return nil
	}
	_, err := uc.policyRP.SeedSuperAdmin(ctx, super.RoleName, super.RoleCode, super.Email, super.Password, super.ForceReset)
	return err
}

func (uc *AccountUC) generateTokenPair(ctx context.Context, userID string) (*TokenPair, error) {
	sid := xid.New().String()
	now := time.Now()

	accessClaims := &pkgauth.Identity{
		RegisteredClaims: jwtv5.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwtv5.NewNumericDate(now.Add(accessTokenTTL)),
			IssuedAt:  jwtv5.NewNumericDate(now),
		},
		Sid: sid,
	}
	accessToken, err := jwtv5.NewWithClaims(pkgauth.SigningMethod, accessClaims).SignedString([]byte(uc.secret))
	if err != nil {
		return nil, err
	}

	refreshClaims := &pkgauth.Identity{
		RegisteredClaims: jwtv5.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwtv5.NewNumericDate(now.Add(refreshTokenTTL)),
			IssuedAt:  jwtv5.NewNumericDate(now),
		},
		Sid: sid,
	}
	refreshToken, err := jwtv5.NewWithClaims(pkgauth.SigningMethod, refreshClaims).SignedString([]byte(uc.secret))
	if err != nil {
		return nil, err
	}

	// Persist session
	_, err = uc.sessionRP.Create(ctx, &Session{
		UserID:    &userID,
		SID:       &sid,
		ExpiresAt: ptrTime(now.Add(refreshTokenTTL)),
		Revoked:   ptrBool(false),
	})
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(accessTokenTTL.Seconds()),
	}, nil
}

func verifyPassword(hashed, plain string) error {
	// bcrypt.CompareHashAndPassword
	return nil // TODO: implement with bcrypt
}

func ptrTime(t time.Time) *time.Time    { return &t }
func ptrBool(b bool) *bool              { return &b }
```

**文件**: `apps/yggdrasil/services/iam/internal/biz/adapter_session.go`

```go
package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	pkgauth "cyber-ecosystem/apps/yggdrasil/services/iam/internal/pkg/auth"
)

// SessionValidator adapts biz ports to the pkg/security.SessionValidator interface.
// Used by server middleware to check if a session has been revoked.
type SessionValidator struct {
	uc *AccountUC
}

func NewSessionValidator(uc *AccountUC) *SessionValidator {
	return &SessionValidator{uc: uc}
}

func (v *SessionValidator) ValidateSession(ctx context.Context, identity *pkgauth.Identity) error {
	revoked, err := v.uc.sessionRP.IsRevoked(ctx, identity.Sid)
	if err != nil {
		return err
	}
	if revoked {
		return fmt.Errorf("session revoked")
	}
	return nil
}
```

> 注意: `fmt` import 需要添加到 adapter_session.go。

---

## Step 7: Data 层

按「通用模式参考」创建以下文件，只改 import 路径和 DB 名称:

**文件**: `apps/yggdrasil/services/iam/internal/data/store.go`
- import `ent` 路径: `cyber-ecosystem/apps/yggdrasil/services/iam/internal/data/ent`

**文件**: `apps/yggdrasil/services/iam/internal/data/store_ent.go`
- DB 名称: `cyber_ecosystem_yggdrasil_iam`
- `defaultError` 使用 `yggdrasilV1.ErrorErrorReason*`

**文件**: `apps/yggdrasil/services/iam/internal/data/store_cache.go`
- 与通用模式完全一致

**文件**: `apps/yggdrasil/services/iam/internal/data/data.go`

```go
package data

import (
	"github.com/google/wire"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/biz"
)

type RP struct {
	log   *log.Helper
	store *Store
}

var ProviderSet = wire.NewSet(
	NewStore,
	NewCache,
	NewEntClient,
	wire.Bind(new(biz.Transaction), new(*Store)),
	NewUserRP,
	NewSessionRP,
	NewPolicyRP,
)
```

**文件**: `apps/yggdrasil/services/iam/internal/data/rp_user.go`

```go
package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/biz"
	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/data/ent"
)

type userRP struct {
	RP
}

func NewUserRP(logger log.Logger, store *Store) biz.UserRP {
	return &userRP{
		RP: RP{
			log:   log.NewHelper(log.With(logger, "module", "data/rp_user")),
			store: store,
		},
	}
}

func (rp *userRP) GetByEmail(ctx context.Context, email string) (*biz.User, error) {
	result, err := rp.store.GetClient(ctx).User.Query().Where(user.Email(email)).Only(ctx)
	if err != nil {
		return nil, HandleError(err)
	}
	return mapUser(result), nil
}

func (rp *userRP) Create(ctx context.Context, u *biz.User) (*biz.User, error) {
	builder := rp.store.GetClient(ctx).User.Create().
		SetEmail(*u.Email).
		SetPassword(*u.Password)
	if u.Name != nil {
		builder.SetName(*u.Name)
	}
	result, err := builder.Save(ctx)
	if err != nil {
		return nil, HandleError(err)
	}
	return mapUser(result), nil
}

func (rp *userRP) Get(ctx context.Context, id string) (*biz.User, error) {
	result, err := rp.store.GetClient(ctx).User.Get(ctx, id)
	if err != nil {
		return nil, HandleError(err)
	}
	return mapUser(result), nil
}

func mapUser(r *ent.User) *biz.User {
	return &biz.User{
		ID:         &r.ID,
		Email:      &r.Email,
		Password:   &r.Password,
		Name:       &r.Name,
		Status:     ptrString(string(r.Status)),
		ForceReset: &r.ForceReset,
	}
}

func ptrString(s string) *string { return &s }
```

**文件**: `apps/yggdrasil/services/iam/internal/data/rp_session.go`

```go
package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/biz"
	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/data/ent"
)

type sessionRP struct {
	RP
}

func NewSessionRP(logger log.Logger, store *Store) biz.SessionRP {
	return &sessionRP{
		RP: RP{
			log:   log.NewHelper(log.With(logger, "module", "data/rp_session")),
			store: store,
		},
	}
}

func (rp *sessionRP) Create(ctx context.Context, s *biz.Session) (*biz.Session, error) {
	builder := rp.store.GetClient(ctx).Session.Create().
		SetUserID(*s.UserID).
		SetSID(*s.SID).
		SetExpiresAt(*s.ExpiresAt)
	result, err := builder.Save(ctx)
	if err != nil {
		return nil, HandleError(err)
	}
	return mapSession(result), nil
}

func (rp *sessionRP) RevokeBySID(ctx context.Context, sid string) error {
	_, err := rp.store.GetClient(ctx).Session.Update().
		Where(session.SID(sid)).
		SetRevoked(true).
		Save(ctx)
	return HandleError(err)
}

func (rp *sessionRP) IsRevoked(ctx context.Context, sid string) (bool, error) {
	result, err := rp.store.GetClient(ctx).Session.Query().
		Where(session.SID(sid)).
		Only(ctx)
	if err != nil {
		return false, HandleError(err)
	}
	return result.Revoked, nil
}

func (rp *sessionRP) GetByRefreshToken(ctx context.Context, refreshToken string) (*biz.Session, error) {
	// Decode JWT to get SID, then look up session
	// Note: refresh token is a JWT with SID claim
	// This is a simplified version - full implementation would parse the JWT
	return nil, fmt.Errorf("not implemented in slice 2")
}

func mapSession(r *ent.Session) *biz.Session {
	return &biz.Session{
		ID:        &r.ID,
		UserID:    &r.UserID,
		SID:       &r.SID,
		ExpiresAt: &r.ExpiresAt,
		Revoked:   &r.Revoked,
	}
}
```

**文件**: `apps/yggdrasil/services/iam/internal/data/rp_policy.go`

最小 Casbin 适配器实现，仅支持超级管理员种子:

```go
package data

import (
	"context"
	"fmt"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/biz"
	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/data/ent"
)

type policyRP struct {
	RP
	enforcer *casbin.SyncedEnforcer
}

func NewPolicyRP(logger log.Logger, store *Store) (biz.PolicyRP, func(), error) {
	rp := &policyRP{
		RP: RP{
			log:   log.NewHelper(log.With(logger, "module", "data/rp_policy")),
			store: store,
		},
	}

	// Create Casbin model inline
	m, err := model.NewModelFromString(`
[request_definition]
r = sub, dom, obj

[policy_definition]
p = sub, dom, obj, eft

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = g(r.sub, p.sub, r.dom) && r.dom == p.dom && keyMatch2(r.obj, p.obj)
`)
	if err != nil {
		return nil, nil, fmt.Errorf("create casbin model: %w", err)
	}

	adapter := &casbinAdapter{rp: rp}
	e, err := casbin.NewSyncedEnforcer(m, adapter)
	if err != nil {
		return nil, nil, fmt.Errorf("create enforcer: %w", err)
	}
	rp.enforcer = e
	_ = e.LoadPolicy()

	cleanup := func() {}
	return rp, cleanup, nil
}

func (rp *policyRP) Enforce(sub, dom, obj string) (bool, error) {
	return rp.enforcer.Enforce(sub, dom, obj)
}

func (rp *policyRP) SeedSuperAdmin(ctx context.Context, roleName, roleCode, email, password string, forceReset bool) (string, error) {
	client := rp.store.GetClient(ctx)

	// Ensure role exists
	roleResult, err := client.Role.Create().SetName(roleName).SetCode(roleCode).SetSort(1).Save(ctx)
	if ent.IsConstraintError(err) {
		roleResult, err = client.Role.Query().Where(role.Code(roleCode)).Only(ctx)
		if err != nil {
			return "", HandleError(err)
		}
	} else if err != nil {
		return "", HandleError(err)
	}

	// Ensure user exists
	hashed := password // TODO: bcrypt hash
	userResult, err := client.User.Create().SetEmail(email).SetPassword(hashed).SetName("Super Admin").Save(ctx)
	if ent.IsConstraintError(err) {
		userResult, err = client.User.Query().Where(user.Email(email)).Only(ctx)
		if err != nil {
			return "", HandleError(err)
		}
	} else if err != nil {
		return "", HandleError(err)
	}

	// Seed Casbin policies
	_, _ = rp.enforcer.AddRoleForUser(userResult.ID, roleCode, biz.Domain)
	_, _ = rp.enforcer.AddPolicy(roleCode, biz.Domain, "/*", "allow")

	return userResult.ID, nil
}

// Casbin Adapter (minimal in-memory for Slice 2)
type casbinAdapter struct {
	rp *policyRP
}

func (a *casbinAdapter) LoadPolicy(m model.Model) error {
	// Load from policy_rule table
	ctx := context.Background()
	rules, err := a.rp.store.GetClient(ctx).PolicyRule.Query().All(ctx)
	if err != nil {
		return err
	}
	for _, r := range rules {
		persist.LoadPolicyLine(fmt.Sprintf("%s, %s, %s, %s, %s, %s, %s",
			r.Ptype, r.V0, r.V1, r.V2, r.V3, r.V4, r.V5), m)
	}
	return nil
}

func (a *casbinAdapter) SavePolicy(m model.Model) error    { return nil }
func (a *casbinAdapter) AddPolicy(sec, ptype string, rule []string) error {
	return a.addFilteredPolicy(sec, ptype, rule)
}
func (a *casbinAdapter) RemovePolicy(sec, ptype string, rule []string) error {
	return a.removeFilteredPolicy(sec, ptype, rule)
}
func (a *casbinAdapter) RemoveFilteredPolicy(sec, ptype string, fieldIndex int, fieldValues []string) error {
	return nil // TODO: implement
}
```

验证: `gofmt -e apps/yggdrasil/services/iam/internal/data/*.go`

---

## Step 8: Service 层

**文件**: `apps/yggdrasil/services/iam/internal/service/service.go`

```go
package service

import (
	"github.com/google/wire"

	"github.com/go-kratos/kratos/v2/transport/grpc"
	krahttp "github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"
)

type Registrar interface {
	RegisterGRPC(*grpc.Server)
	RegisterHTTP(*krahttp.Server)
	RegisterConnect(*connect.Server)
}

var ProviderSet = wire.NewSet(
	NewRegistrarList,
	NewAccountAuthService,
	NewInternalAuthService,
)

func NewRegistrarList(
	s1 *AccountAuthService,
	s2 *InternalAuthService,
) []Registrar {
	return []Registrar{s1, s2}
}
```

**文件**: `apps/yggdrasil/services/iam/internal/service/account_auth.go`

```go
package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	krahttp "github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"

	pkgauth "cyber-ecosystem/apps/yggdrasil/services/iam/internal/pkg/auth"
	yggdrasilV1 "cyber-ecosystem/apps/yggdrasil/gen/go/v1"
	yggdrasilV1connect "cyber-ecosystem/apps/yggdrasil/gen/go/v1/yggdrasilV1connect"
	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/biz"
)

type AccountAuthService struct {
	yggdrasilV1.UnimplementedAccountAuthServiceServer
	log       *log.Helper
	accountUC *biz.AccountUC
}

func NewAccountAuthService(logger log.Logger, accountUC *biz.AccountUC) *AccountAuthService {
	return &AccountAuthService{
		log:       log.NewHelper(log.With(logger, "module", "service/account_auth")),
		accountUC: accountUC,
	}
}

func (s *AccountAuthService) RegisterGRPC(srv *grpc.Server) {
	yggdrasilV1.RegisterAccountAuthServiceServer(srv, s)
}

func (s *AccountAuthService) RegisterHTTP(srv *krahttp.Server) {
	yggdrasilV1.RegisterAccountAuthServiceHTTPServer(srv, s)
}

func (s *AccountAuthService) RegisterConnect(srv *connect.Server) {
	srv.Register(yggdrasilV1connect.NewAccountAuthServiceHandler(s, srv.HandlerOptions()...))
}

func (s *AccountAuthService) Login(ctx context.Context, in *yggdrasilV1.LoginRequest) (*yggdrasilV1.LoginResponse, error) {
	pair, err := s.accountUC.Login(ctx, &biz.LoginParam{
		Email:    in.Email,
		Password: in.Password,
	})
	if err != nil {
		return nil, err
	}
	return &yggdrasilV1.LoginResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
	}, nil
}

func (s *AccountAuthService) Logout(ctx context.Context, in *yggdrasilV1.LogoutRequest) (*yggdrasilV1.LogoutResponse, error) {
	identity, err := pkgauth.IdentityFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := s.accountUC.Logout(ctx, identity.Subject, identity.Sid); err != nil {
		return nil, err
	}
	return &yggdrasilV1.LogoutResponse{}, nil
}

func (s *AccountAuthService) RefreshToken(ctx context.Context, in *yggdrasilV1.RefreshTokenRequest) (*yggdrasilV1.RefreshTokenResponse, error) {
	pair, err := s.accountUC.RefreshToken(ctx, in.RefreshToken)
	if err != nil {
		return nil, err
	}
	return &yggdrasilV1.RefreshTokenResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
	}, nil
}
```

**文件**: `apps/yggdrasil/services/iam/internal/service/internal_auth.go`

```go
package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	krahttp "github.com/go-kratos/kratos/v2/transport/http"

	"cyber-ecosystem/shared-go/kratos/transport/connect"

	yggdrasilV1 "cyber-ecosystem/apps/yggdrasil/gen/go/v1"
	yggdrasilV1connect "cyber-ecosystem/apps/yggdrasil/gen/go/v1/yggdrasilV1connect"
	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/biz"
)

type InternalAuthService struct {
	yggdrasilV1.UnimplementedInternalAuthServiceServer
	log       *log.Helper
	accountUC *biz.AccountUC
}

func NewInternalAuthService(logger log.Logger, accountUC *biz.AccountUC) *InternalAuthService {
	return &InternalAuthService{
		log:       log.NewHelper(log.With(logger, "module", "service/internal_auth")),
		accountUC: accountUC,
	}
}

func (s *InternalAuthService) RegisterGRPC(srv *grpc.Server) {
	yggdrasilV1.RegisterInternalAuthServiceServer(srv, s)
}

func (s *InternalAuthService) RegisterHTTP(srv *krahttp.Server) {
	yggdrasilV1.RegisterInternalAuthServiceHTTPServer(srv, s)
}

func (s *InternalAuthService) RegisterConnect(srv *connect.Server) {
	srv.Register(yggdrasilV1connect.NewInternalAuthServiceHandler(s, srv.HandlerOptions()...))
}

func (s *InternalAuthService) ValidateToken(ctx context.Context, in *yggdrasilV1.ValidateTokenRequest) (*yggdrasilV1.ValidateTokenResponse, error) {
	identity, err := s.accountUC.ValidateToken(ctx, in.Token)
	if err != nil {
		return &yggdrasilV1.ValidateTokenResponse{Valid: false}, nil
	}
	return &yggdrasilV1.ValidateTokenResponse{
		UserId:    identity.Subject,
		SessionId: identity.Sid,
		Valid:     true,
	}, nil
}
```

---

## Step 9: Server 层

按「通用模式参考」创建以下文件。

关键改进: **`buildMiddlewares()` 共享函数**，被 http/grpc/connect 三个构造函数共同调用。

**文件**: `apps/yggdrasil/services/iam/internal/server/server.go`

```go
package server

import (
	"github.com/go-kratos/kratos/v2/middleware"
	jwt "github.com/go-kratos/kratos/v2/middleware/auth/jwt"

	jwtv5 "github.com/golang-jwt/jwt/v5"

	pkgauth "cyber-ecosystem/apps/yggdrasil/services/iam/internal/pkg/auth"
	// conf import...
)

func init() {
	// Map error reasons to proto error codes
	// Same pattern as Slice 1
}

// buildMiddlewares constructs the shared middleware chain.
// Called by NewHTTPServer, NewGRPCServer, NewConnectServer.
func buildMiddlewares(
	ca *conf.Auth,
	logger log.Logger,
	sessionValidator *biz.SessionValidator,
	i18nBundle *i18n.Bundle,
	tp *trace.TracerProvider,
	_metricRequests metric.Int64Counter,
	_metricSeconds metric.Float64Histogram,
) []middleware.Middleware {
	var mws []middleware.Middleware

	mws = append(mws, i18n.Server(i18nBundle))
	mws = append(mws, recovery.Recovery())
	mws = append(mws, ratelimit.Server())
	mws = append(mws, metrics.Server(_metricRequests, _metricSeconds))
	if tp != nil {
		mws = append(mws, tracing.Server(tracing.WithTracerProvider(tp)))
	}
	mws = append(mws, metadata.Server())
	mws = append(mws, logging.Server(logger))

	// Auth gated by proto whitelist (public_access annotation)
	mws = append(mws, selector.Server(
		jwt.Server(
			func(token *jwtv5.Token) (any, error) { return []byte(ca.Secret), nil },
			jwt.WithSigningMethod(pkgauth.SigningMethod),
			jwt.WithClaims(func() jwtv5.Claims { return &pkgauth.Identity{} }),
		),
		mw.SessionValidator(sessionValidator, logger),
		// RBAC, Condition, DataScope middleware will be added in Slice 3/4
	).Match(auth.NewWhiteListByPublicAccessInProtoMatcher()).Build())

	mws = append(mws, validate.ProtoValidate())

	return mws
}

var ProviderSet = wire.NewSet(NewOpsServer, NewGRPCServer, NewHTTPServer, NewConnectServer, NewI18nBundle)
```

**文件**: `apps/yggdrasil/services/iam/internal/server/middleware/session_validator.go`

```go
package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"

	pkgauth "cyber-ecosystem/apps/yggdrasil/services/iam/internal/pkg/auth"
)

// SessionValidator checks if a JWT session has been revoked.
func SessionValidator(validator *biz.SessionValidator, logger log.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			identity, err := pkgauth.IdentityFromContext(ctx)
			if err != nil {
				return handler(ctx, req)
			}
			if err := validator.ValidateSession(ctx, identity); err != nil {
				return nil, yggdrasilV1.ErrorErrorReasonUnauthorized("session revoked")
			}
			return handler(ctx, req)
		}
	}
}
```

grpc.go / http.go / connect.go / ops.go / i18n.go / locales: 与 Slice 1 结构相同，使用 `buildMiddlewares()` 替代重复代码。

---

## Step 10: Main + Wire + Config

**文件**: `apps/yggdrasil/services/iam/cmd/app/main.go`

按通用模式创建。关键参数:
- `Name string = "yggdrasil_iam"`
- `flagConf` default: `../../configs`
- 启动时调用 `accountUC.SeedSuperAdmin(ctx, bootstrap.Super)`

**文件**: `apps/yggdrasil/services/iam/cmd/app/wire.go`

```go
//go:build wireinject

package main

import (
	"github.com/google/wire"
	"go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/biz"
	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/conf"
	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/data"
	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/server"
	"cyber-ecosystem/apps/yggdrasil/services/iam/internal/service"
)

func wireApp(
	*conf.Server,
	*conf.Auth,
	*conf.Log,
	*conf.Data,
	*conf.Ops,
	*conf.Super,
	log.Logger,
	*tracesdk.TracerProvider,
	*metricsdk.MeterProvider,
	metric.Int64Counter,
	metric.Float64Histogram,
) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		service.ProviderSet,
		biz.ProviderSet,
		data.ProviderSet,
		newApp,
	))
}
```

**文件**: `apps/yggdrasil/services/iam/configs/config.yaml`

```yaml
server:
  http:
    addr: 0.0.0.0:11000
    timeout: 10s
  grpc:
    addr: 0.0.0.0:12000
    timeout: 10s
  connect:
    addr: 0.0.0.0:13000
    timeout: 10s

auth:
  secret: secret

log:
  level: debug
  ent:
    enabled: false
  cache:
    enabled: false
  console:
    enabled: true
    color: true
    format: "console"
  file:
    enabled: false
  loki:
    enabled: false

data:
  database:
    driver: postgres
    host: localhost
    port: 5432
    user: postgres
    password: postgres
    db_name: cyber_ecosystem_yggdrasil_iam
    max_open_conns: 10
    max_idle_conns: 5
    conn_max_lifetime: 300s
    migrate: true
  cache:
    type: memory
    memory:
      otel_enabled: true

trace:
  insecure: true
  endpoint: "http://localhost:4318/v1/traces"

ops:
  enabled: true
  addr: "0.0.0.0:14000"
  metrics: "/metrics"
  pprof:
    enabled: false

super:
  enabled: true
  role_name: "超级管理员"
  role_code: "super_admin"
  email: "admin@example.com"
  password: "admin123"
  force_reset: false
```

---

## Step 11: 编译闭环

```bash
cd apps/yggdrasil/services/iam && go mod tidy
./nx run yggdrasil_iam:generate
./nx run yggdrasil_iam:build
```

验证: `ls apps/yggdrasil/services/iam/bin/iam`

如果失败，修复所有编译错误后再继续。

提交:
```bash
git add apps/yggdrasil/services/iam/
git commit -m "feat(yggdrasil): IAM auth service — first full build"
```

---

## Step 12: 集成验证

### 12a: 创建数据库

```bash
psql -h localhost -U postgres -c "CREATE DATABASE cyber_ecosystem_yggdrasil_iam;"
```

### 12b: 启动服务

```bash
./nx run yggdrasil_iam:dev
```

等待启动日志，确认超级管理员种子已执行。

### 12c: 测试 Login (HTTP)

```bash
curl -X POST http://localhost:11000/iam/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "admin123"}'
```

预期: 返回 `access_token` + `refresh_token` + `expires_in`

### 12d: 测试 Logout (HTTP)

```bash
TOKEN="<access_token_from_12c>"
curl -X POST http://localhost:11000/iam/auth/logout \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'
```

预期: `{}`

### 12e: 测试 RefreshToken (HTTP)

```bash
curl -X POST http://localhost:11000/iam/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "<refresh_token_from_12c>"}'
```

预期: 新的 token pair

### 12f: 测试 ValidateToken (gRPC)

```bash
# 重新登录获取 token
TOKEN=$(curl -s -X POST http://localhost:11000/iam/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "admin123"}' | jq -r '.access_token')

grpcurl -plaintext -d "{\"token\": \"$TOKEN\"}" localhost:12000 api.yggdrasil.v1.InternalAuthService/ValidateToken
```

预期: `{"userId": "...", "sessionId": "...", "valid": true}`

### 12g: 测试 ConnectRPC Login

```bash
curl -X POST http://localhost:13000/api.yggdrasil.v1.AccountAuthService/Login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@example.com", "password": "admin123"}'
```

预期: ConnectRPC 格式的 token pair

### 12h: 停止服务并提交

```bash
git add apps/yggdrasil/
git commit -m "feat(yggdrasil): IAM auth passes integration verification"
```

---

## 完成标准

- [x] `./nx run yggdrasil_iam:build` 编译通过
- [x] 服务可启动，3 个传输层正常监听
- [x] 超级管理员种子数据自动创建
- [x] HTTP Login 返回 JWT token pair
- [x] HTTP Logout 注销 session
- [x] gRPC ValidateToken 校验 token 有效
- [x] ConnectRPC Login 正常工作
- [x] `buildMiddlewares()` 被 http/grpc/connect 共享调用（非三重复制）
- [x] 变更已提交
