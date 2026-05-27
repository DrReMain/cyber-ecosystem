---
name: kratos-scaffold
description: Use when creating a new Go microservice, adding an aggregate/entity, implementing Kratos business logic, or scaffolding service layers (ent, biz, data, service, wire)
---

# Kratos Service Scaffold

Step-by-step scaffold guide for Go microservices using Kratos in this monorepo.

**Prerequisites:** Proto files and generation already set up (`/proto-design`).

---

## When to Use

- Creating a new Go service (base service, BFF)
- Adding a new aggregate/entity to an existing service
- Implementing biz, data, or service layers
- Wiring new providers with Wire

## Key Rules

- Single Go module — no `go mod init` in service subdirectories
- All write operations wrap in `uc.tm.InTx()`
- Read operations skip transactions
- BFF UC layer mirrors base service patterns (even though BFF's `InTx()` is a no-op passthrough)
- Platform ProviderSet must include ALL gRPC client constructors
- OpsServer must be a concrete struct, not a type alias
- HTTP/Connect registration: no-ops if proto has no `google.api.http` annotations
- CORS filter required on HTTP and Connect servers

---

## Adding a New Aggregate — Step by Step

### Step 1: Ent Schema

Create `internal/ent/schema/comment.go`:

```go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"

    "cyber-ecosystem/shared-go/orm/ent/mixins"
    "cyber-ecosystem/apps/<app>/services/<service>/internal/ent/schema/local_mixins"
)

type Comment struct {
    ent.Schema
}

func (Comment) Fields() []ent.Field {
    return []ent.Field{
        field.String("content").MaxLen(1000),
        field.String("status").Default("draft").MaxLen(10),
    }
}

func (Comment) Mixin() []ent.Mixin {
    return []ent.Mixin{
        mixins.IDStringMixin{},
        mixins.CreatedUpdatedMixin{},
        local_mixins.SortMixin{SoftDelete: true},
        local_mixins.SoftDeleteMixin{},
    }
}
```

Generate: `./nx run <project>:generate:ent`

### Step 2: Biz Layer — Model

Create `internal/biz/comment.go`:

```go
package biz

import (
    "time"

    "cyber-ecosystem/contracts/go/common"
    "cyber-ecosystem/shared-go/utils"
)

// region[rgba(239,83,80,0.15)] 🔴 Model

type Comment struct {
    ID        *string
    CreatedAt *time.Time
    UpdatedAt *time.Time
    Content   *string
    Status    *string
}

type CommentQueryIn struct {
    *common.PageRequest
    OrderBy []*utils.OrderBy
    Content *string
    Status  *string
}

type CommentQueryOut struct {
    *common.PageResponse
    List []*Comment
}

// endregion
```

### Step 3: Biz Layer — RP Port + UC

Create `internal/biz/comment_uc.go`:

```go
// region[rgba(66,165,245,0.15)] 🔵 Port

type CommentRP interface {
    Create(ctx context.Context, c *Comment) (*Comment, error)
    Update(ctx context.Context, fieldsMask []string, c *Comment) (*Comment, error)
    Delete(ctx context.Context, id string) (string, error)
    Get(ctx context.Context, id string) (*Comment, error)
    Query(ctx context.Context, in *CommentQueryIn) (*CommentQueryOut, error)
    Sort(ctx context.Context, id string, prevID, nextID *string) (*Comment, error)
}

// endregion

// region[rgba(102,187,106,0.15)] 🟢 UC

type CommentUC struct {
    UC
    commentRP CommentRP
}

func NewCommentUC(logger log.Logger, tm Transaction, commentRP CommentRP) *CommentUC {
    return &CommentUC{
        UC: UC{
            log: log.NewHelper(log.With(logger, "module", "biz/comment_uc")),
            tm:  tm,
        },
        commentRP: commentRP,
    }
}

// endregion

// region[rgba(186,104,200,0.15)] 🟣 Method

func (uc *CommentUC) Create(ctx context.Context, c *Comment) (out *Comment, err error) {
    err = uc.tm.InTx(ctx, func(ctx context.Context) error {
        out, err = uc.commentRP.Create(ctx, c)
        return err
    })
    return
}

// Update, Delete follow same pattern with InTx
// Get, Query do NOT need InTx (read-only)

// endregion
```

### Step 4: Biz Layer — FSM (if entity has states)

Create `internal/biz/comment_fsm.go`:

```go
// region[rgba(255,238,88,0.12)] 🟡 FSM States

const (
    CommentStatusDraft     = "draft"
    CommentStatusPublished = "published"
)

// endregion

// region[rgba(255,167,38,0.15)] 🟠 FSM

func newCommentFSM(current string, c *Comment) *fsm.FSM {
    return fsm.NewFSM(current,
        []fsm.EventDesc{
            {Name: CommentStatusPublished, Src: []string{CommentStatusDraft}, Dst: CommentStatusPublished},
        },
        map[string]fsm.Callback{
            "after_" + CommentStatusPublished: func(_ context.Context, _ *fsm.Event) { *c.Status = CommentStatusPublished },
        },
    )
}

// endregion

// region[rgba(186,104,200,0.15)] 🟣 Domain Method

func (c *Comment) TransitionTo(ctx context.Context, target string) error {
    c.Status = utils.Ptr(utils.Deref(c.Status, CommentStatusDraft))
    f := newCommentFSM(*c.Status, c)
    if err := f.Event(ctx, target); err != nil {
        return appV1.ErrorErrorReasonCommentInvalidStatusTransition("").WithCause(err)
    }
    return nil
}

// endregion
```

### Step 5: Data Layer — Repository

**With Database (Ent):** Create `internal/data/comment_rp.go`:

```go
type commentRP struct {
    RP
}

func NewCommentRP(logger log.Logger, p *platform.Platform) biz.CommentRP {
    return &commentRP{
        RP: RP{
            log:      log.NewHelper(log.With(logger, "module", "data/comment_rp")),
            platform: p,
        },
    }
}

func (rp *commentRP) Create(ctx context.Context, c *biz.Comment) (*biz.Comment, error) {
    created, err := rp.platform.GetClient(ctx).Comment.Create().
        SetContent(*c.Content).
        Save(ctx)
    if err != nil {
        return nil, rp.platform.HandleEntError(err)
    }
    return mapComment(created), nil
}

func (rp *commentRP) Update(ctx context.Context, fieldsMask []string, c *biz.Comment) (*biz.Comment, error) {
    updater := rp.platform.GetClient(ctx).Comment.UpdateOneID(*c.ID)
    utils.Handler{
        "content": {
            Condition: c.Content != nil,
            OnTrue:    func() { updater.SetContent(*c.Content) },
            OnFalse:   func() { updater.SetContent("") },
        },
    }.Emit(fieldsMask)
    updated, err := updater.Save(ctx)
    if err != nil {
        return nil, rp.platform.HandleEntError(err)
    }
    return mapComment(updated), nil
}
```

**With gRPC Client (calling another service):**

```go
func (rp *commentRP) Get(ctx context.Context, id string) (*biz.Comment, error) {
    resp, err := rp.platform.GetCommentClient().GetComment(ctx,
        &appV1.GetCommentRequest{Id: utils.Ptr(id)},
    )
    if err != nil {
        return nil, err
    }
    return protoToComment(resp), nil
}
```

Type conversion: `utils.Ptr(val)`, `utils.Unwrap(wrapper)`, `utils.Wrap(ptr, utils.StringW)`, `utils.FromTimestamp(ts)`.

### Step 6: Service Layer

Create `internal/service/comment.go`:

```go
// region[rgba(236,64,122,0.15)] 🩷 Struct

type CommentService struct {
    <app>V1.UnimplementedCommentServiceServer
    log       *log.Helper
    commentUC *biz.CommentUC
}

func NewCommentService(logger log.Logger, commentUC *biz.CommentUC) *CommentService {
    return &CommentService{
        log:       log.NewHelper(log.With(logger, "module", "service/comment")),
        commentUC: commentUC,
    }
}

// endregion

// region[rgba(255,167,38,0.15)] 🟠 Handler

func (s *CommentService) RegisterGRPC(srv *grpc.Server) {
    <app>V1.RegisterCommentServiceServer(srv, s)
}

// No HTTP annotations → no-ops:
func (s *CommentService) RegisterHTTP(_ *http.Server)       {}
func (s *CommentService) RegisterConnect(_ *connect.Server) {}

// With HTTP annotations:
func (s *CommentService) RegisterHTTP(srv *http.Server) {
    <app>V1.RegisterCommentServiceHTTPServer(srv, s)
}
func (s *CommentService) RegisterConnect(srv *connect.Server) {
    srv.Register(<app>V1connect.NewCommentServiceHandler(s, srv.HandlerOptions()...))
}

// endregion

// region[rgba(186,104,200,0.15)] 🟣 Method

func (s *CommentService) CreateComment(ctx context.Context, in *<app>V1.CreateCommentRequest) (*<app>V1.CreateCommentResponse, error) {
    c := &biz.Comment{Content: in.Content}
    created, err := s.commentUC.Create(ctx, c)
    if err != nil { return nil, err }
    return &<app>V1.CreateCommentResponse{Id: utils.Wrap(created.ID, utils.StringW)}, nil
}

// endregion
```

### Step 7: Wire Everything

Update each layer's ProviderSet:

- `data/data.go`: add `NewCommentRP` to ProviderSet
- `biz/biz.go`: add `NewCommentUC` to ProviderSet
- `service/service.go`: add `NewCommentService` to ProviderSet and registrar list
- `platform/` (gRPC client): add `NewGRPCXxxClient` to ProviderSet

Regenerate: `./nx run <project>:generate:wire`

### Step 8: Observability Setup

Use `/observability` to configure error reporting and metrics scraping.

---

## Platform Variants

### With Database (Ent)

```go
type Platform struct {
    cache            *cache.Cache
    handleCacheError CacheErrorHandler
    db               *ent.Client
    handleEntError   EntErrorHandler
}
```

`InTx` opens a real database transaction.

### With gRPC Client (no database)

```go
type Platform struct {
    cache            *cache.Cache
    handleCacheError CacheErrorHandler
    articleClient    appV1.ArticleServiceClient
}
```

`InTx` calls `fn(ctx)` directly (passthrough — no cross-service consistency issue for stateless BFF proxies).

---

## Key Patterns

### BFF UC Layer Mirrors Base Patterns

BFF services that call a base service via gRPC should still follow the same UC patterns:
- Wrap writes in `uc.tm.InTx()` (no-op passthrough, but consistent pattern)
- Include FSM for status transitions
- `UpdateStatus` = Get → TransitionTo → Update at UC level

### Sort Pattern (Fractional Indexing)

```go
func (rp *articleRP) Sort(ctx context.Context, id string, prevID, nextID *string) (*biz.Article, error) {
    var prevSort, nextSort string
    client := rp.platform.GetClient(ctx)
    if prevID != nil {
        d, err := client.Article.Get(ctx, *prevID)
        if err != nil { return nil, rp.platform.HandleEntError(err) }
        prevSort = d.Sort
    }
    if nextID != nil {
        d, err := client.Article.Get(ctx, *nextID)
        if err != nil { return nil, rp.platform.HandleEntError(err) }
        nextSort = d.Sort
    }
    newSort, err := fracdex.KeyBetween(prevSort, nextSort)
    // ... update entity sort field
}
```

---

## Ent Three-Stage Bootstrap

Local mixins import generated packages that don't exist on first generation:

1. **Stage 1**: Comment out local_mixins from schema, comment out `intercept` feature in `ent/generate.go`. Run `generate:ent`.
2. **Stage 2**: Uncomment `intercept` feature. Run `generate:ent` again.
3. **Stage 3**: Uncomment local_mixins from schema. Run `generate:ent` again.

After the first successful generation, subsequent runs work normally.

---

## Common Pitfalls

### Single Go Module (No go mod init)

This monorepo uses a single Go module at the repository root. Do NOT run `go mod init` in any service subdirectory.

### Platform ProviderSet Must Include gRPC Client Constructors

Omitting `NewGRPCXxxClient` from the ProviderSet causes "no provider found" wire errors.

### OpsServer Must Be a Concrete Struct

Do NOT use `type OpsServer = ops.Server` — Wire cannot resolve type aliases. Define a concrete struct with `Start(ctx) error` and `Stop(ctx) error` methods.

### GCI Import Formatting

After creating Go files, imports may need formatting. Imports follow: stdlib → third-party → `cyber-ecosystem/`.

### HTTP/Connect Registration

No `google.api.http` annotations → no generated `RegisterXxxHTTPServer` → must be no-ops.

### newApp Server Selection

All services create gRPC, HTTP, Connect, and Ops servers. `newApp` selectively starts them by commenting/uncommenting `srv = append(...)` lines.

### Database Name Convention

Database names: `cyber_ecosystem_<app>_service_<service>`. The database must exist before starting.

### i18n.protos Relative Paths

From `internal/i18n/` to contracts: `../../../../../../contracts/errors/codes_general.proto`. To app errors: `../../../../api/v1/error/error_reason.proto`.

### Services Without Database

No Ent → skip `generate:ent`, no `ent/` directory, omit ent targets from `project.json`.

### CORS Configuration

HTTP and Connect servers MUST include CORS filter:

```go
krahttp.Filter(handlers.CORS(
    handlers.AllowedOrigins([]string{"*"}),
    handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}),
    handlers.AllowedHeaders([]string{"Content-Type", i18n.DefaultHeaderLang, "Authorization"}),
    handlers.MaxAge(86400),
)),
```

### wireApp Signature

Same parameters regardless of capabilities — `*conf.Client` is NOT needed (remote service config is inside `*conf.Data`):

```go
func wireApp(
    *conf.Server, *conf.Log, *conf.Data, *conf.Ops,
    log.Logger, *tracesdk.TracerProvider, *metricsdk.MeterProvider,
    metric.Int64Counter, metric.Float64Histogram,
) (*kratos.App, func(), error)
```

---

## Nx Targets

```bash
./nx run <project>:generate        # Run all generation targets
./nx run <project>:generate:ent    # Generate Ent ORM
./nx run <project>:generate:wire   # Generate Wire DI
./nx run <project>:generate:i18n   # Generate i18n stubs
./nx run <project>:build           # Compile binary
./nx run <project>:dev             # Run with kratos
```

---

For deep architecture details, see `docs/stacks/kratos-go.md`.
