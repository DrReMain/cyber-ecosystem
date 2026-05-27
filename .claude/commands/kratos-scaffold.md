# Kratos Service Scaffold

Step-by-step guide for scaffolding and implementing Go microservices using the Kratos framework in this monorepo. Use this when creating a new service, adding a new aggregate/entity, or implementing business logic.

This skill assumes proto files and generation are already set up (see `/proto-design`). For architecture reference, middleware chains, utils, and project structure, see `docs/stacks/kratos-go.md`.

---

## 1. Service Directory Structure

```
apps/<app>/services/<service>/    # use underscores (admin_bff, not admin-bff)
  cmd/app/
    main.go, wire.go, wire_gen.go
    logger.go, metrics.go, tracing.go, sentry.go, resource.go
  internal/
    conf/
      conf.proto, conf.pb.go
    ent/schema/                  # [DB capability]
      {aggregate}.go
      local_mixins/
        soft_delete.go, sort.go
    platform/
      interface.go, platform.go
      platform_ent.go            # [DB] Ent client
      platform_ent_handler.go    # [DB] Ent error mapping
      platform_grpc.go           # [gRPC Client] Remote service clients
      platform_cache.go, platform_cache_handler.go
    data/
      data.go
      {aggregate}_rp.go
    biz/
      biz.go
      {aggregate}.go
      {aggregate}_uc.go
      {aggregate}_fsm.go
    service/
      service.go
      {aggregate}.go
    server/
      server.go, grpc.go, http.go, connect.go, ops.go
    i18n/
      generate.go, i18n.go, i18n.protos
      locales/v1.zh-CN.yaml, v1.en-US.yaml
  configs/config.yaml
  project.json
  buf.gen.conf.yaml
```

For the full project structure with annotations, see `docs/stacks/kratos-go.md` Section 14.

---

## 2. Adding a New Aggregate

When adding a new entity (e.g., "Comment"), follow this order:

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

// region[rgba(186,104,200,0.15)] 🟣 Method

func (uc *CommentUC) Create(ctx context.Context, c *Comment) (out *Comment, err error) {
    err = uc.tm.InTx(ctx, func(ctx context.Context) error {
        out, err = uc.commentRP.Create(ctx, c)
        return err
    })
    return
}
// ... Update, Delete follow same pattern with InTx
// ... Get, Query do NOT need InTx (read-only)
```

### Step 4: Biz Layer — FSM (if entity has states)

Create `internal/biz/comment_fsm.go`:

```go
// region[rgba(255,238,88,0.12)] 🟡 FSM States

const (
    CommentStatusDraft     = "draft"
    CommentStatusPublished = "published"
)

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

// region[rgba(186,104,200,0.15)] 🟣 Domain Method

func (c *Comment) TransitionTo(ctx context.Context, target string) error {
    c.Status = utils.Ptr(utils.Deref(c.Status, CommentStatusDraft))
    f := newCommentFSM(*c.Status, c)
    if err := f.Event(ctx, target); err != nil {
        return appV1.ErrorErrorReasonCommentInvalidStatusTransition("").WithCause(err)
    }
    return nil
}
```

### Step 5: Data Layer — Repository

#### With Database (Ent)

Create `internal/data/comment_rp.go`:

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

// region[rgba(0,188,212,0.12)] 🩵 Repo

func (rp *commentRP) Create(ctx context.Context, c *biz.Comment) (*biz.Comment, error) {
    created, err := rp.platform.GetClient(ctx).Comment.Create().
        SetContent(*c.Content).
        Save(ctx)
    if err != nil {
        return nil, rp.platform.HandleEntError(err)
    }
    return mapComment(created), nil
}

// Update uses utils.Handler for field mask pattern:
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

#### With gRPC Client (calling another service)

The data layer calls remote service clients and maps between proto types and biz models:

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

// Proto-to-biz mapping:
func protoToComment(resp *appV1.GetCommentResponse) *biz.Comment {
    return &biz.Comment{
        ID:        utils.Unwrap(resp.Id),
        CreatedAt: utils.FromTimestamp(resp.CreatedAt),
        UpdatedAt: utils.FromTimestamp(resp.UpdatedAt),
        Content:   utils.Unwrap(resp.Content),
        Status:    utils.Unwrap(resp.Status),
    }
}
```

**Type conversion rule of thumb:**
- `utils.Ptr(val)` — `string` → `*string` (for request fields: `optional string` in proto)
- `utils.Unwrap(wrapper)` — `*wrapperspb.StringValue` → `*string` (for response fields)
- `utils.Wrap(ptr, utils.StringW)` — `*string` → `*wrapperspb.StringValue` (for building proto responses in service layer)
- `utils.FromTimestamp(ts)` — `*timestamppb.Timestamp` → `*time.Time`

### Step 6: Service Layer

Create `internal/service/comment.go`:

```go
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

func (s *CommentService) RegisterGRPC(srv *grpc.Server) {
    <app>V1.RegisterCommentServiceServer(srv, s)
}

// For services without HTTP annotations — no-ops:
func (s *CommentService) RegisterHTTP(_ *http.Server)       {}
func (s *CommentService) RegisterConnect(_ *connect.Server) {}

// For services WITH HTTP annotations — register handlers:
func (s *CommentService) RegisterHTTP(srv *http.Server) {
    <app>V1.RegisterCommentServiceHTTPServer(srv, s)
}
func (s *CommentService) RegisterConnect(srv *connect.Server) {
    srv.Register(<app>V1connect.NewCommentServiceHandler(s, srv.HandlerOptions()...))
}
```

### Step 7: Wire Everything

Update each layer's ProviderSet:

- `data/data.go`: add `NewCommentRP` to ProviderSet
- `biz/biz.go`: add `NewCommentUC` to ProviderSet
- `service/service.go`: add `NewCommentService` to ProviderSet and registrar list

For services with gRPC client capability, also add `NewGRPCXxxClient` to `platform/` ProviderSet.

Then regenerate: `./nx run <project>:generate:wire`

---

## 3. Key Patterns

### BFF UC Layer Mirrors Base Patterns

BFF services that call a base service via gRPC should still follow the same UC layer patterns as the base service:

- Wrap all write operations (Create, Update, Delete, Sort, UpdateStatus) in `uc.tm.InTx()`. Even though BFF's `InTx()` is a no-op passthrough (no database), keeping the pattern consistent means the UC layer looks identical across base and BFF services.
- Include FSM for status transitions (`TransitionTo` method). The BFF's `UpdateStatus` should do Get → TransitionTo → Update, same as the base service.
- The `UpdateStatus` method should NOT be in the RP port interface — it's composed from Get + TransitionTo + Update at the UC level.

**No distributed transaction concern:** When BFFs are stateless proxies (no own database) calling a single base service, the transaction boundary is always within the base service. BFF's `InTx()` is a passthrough — there's no cross-service consistency issue. This only changes if future services each have their own database and an operation must update data across them.

### Transaction Management

All write operations (Create, Update, Delete, Sort, UpdateStatus) wrap in `uc.tm.InTx()`:

```go
func (uc *ArticleUC) Create(ctx context.Context, a *Article) (out *Article, err error) {
    err = uc.tm.InTx(ctx, func(ctx context.Context) error {
        out, err = uc.articleRP.Create(ctx, a)
        return err
    })
    return
}
```

Read operations (Get, Query) skip transactions — they're read-only.

### UpdateStatus Pattern

Load entity → transition state → save with field mask:

```go
func (uc *ArticleUC) UpdateStatus(ctx context.Context, id string, target string) (out *Article, err error) {
    err = uc.tm.InTx(ctx, func(ctx context.Context) error {
        a, e := uc.articleRP.Get(ctx, id)
        if e != nil { return e }
        if e = a.TransitionTo(ctx, target); e != nil { return e }
        out, e = uc.articleRP.Update(ctx, []string{"status"}, a)
        return e
    })
    return
}
```

### Sort Pattern (Fractional Indexing)

Uses `fracdex.KeyBetween(prevSort, nextSort)` to generate a sort key between two positions:

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

### Query with Pagination, Filtering, Sorting

```go
func (rp *articleRP) Query(ctx context.Context, in *biz.ArticleQueryIn) (*biz.ArticleQueryOut, error) {
    query := rp.platform.GetClient(ctx).Article.Query()

    // Time range filters from PageRequest
    entutil.WherePtr(query, utils.FromTimestamp(in.PageRequest.CreatedAtA), article.CreatedAtGTE)
    entutil.WherePtr(query, utils.FromTimestamp(in.PageRequest.CreatedAtZ), article.CreatedAtLTE)

    // Conditional filters
    entutil.Where(query, in.Title != nil, func() predicate.Article { return article.TitleContainsFold(*in.Title) })
    entutil.Where(query, in.Status != nil, func() predicate.Article { return article.StatusEQ(*in.Status) })

    // Order by (user-specified)
    entutil.ApplyOrderBy(in.OrderBy, ent.Asc, ent.Desc, entutil.FOMapping{
        "createdAt": func(sel entutil.SQLSelector) { query.Order(sel(article.FieldCreatedAt)) },
        "updatedAt": func(sel entutil.SQLSelector) { query.Order(sel(article.FieldUpdatedAt)) },
        "sort":      func(sel entutil.SQLSelector) { query.Order(sel(article.FieldSort)) },
    })

    // Default sort order (always applied last)
    query.Order(func(s *sql.Selector) { s.OrderBy(s.C(article.FieldSort)) })

    // Pagination
    total, offset, limit, err := entutil.ApplyPagination(ctx, query, in.PageRequest,
        entutil.NewPageConfig(entutil.DefaultPageSize, entutil.DefaultPageSizeUnlimit),
        errorspb.ErrorGeneralErrorPaginationInvalidArgument(""),
    )
    // ...
    return &biz.ArticleQueryOut{
        PageResponse: entutil.BuildPageResponse(total, offset, limit),
        List:         utils.SliceMap(items, mapArticle),
    }, nil
}
```

### Proto-to-Biz Mapping (Service Layer)

```go
func (s *ArticleService) CreateArticle(ctx context.Context, in *genesisV1.CreateArticleRequest) (*genesisV1.CreateArticleResponse, error) {
    a := &biz.Article{
        Title:   in.Title,
        Content: in.Content,
    }
    created, err := s.articleUC.Create(ctx, a)
    if err != nil { return nil, err }
    return &genesisV1.CreateArticleResponse{
        Id: utils.Wrap(created.ID, utils.StringW),
    }, nil
}

func (s *ArticleService) articleToProto(a *biz.Article) *genesisV1.GetArticleResponse {
    return &genesisV1.GetArticleResponse{
        Id:        utils.Wrap(a.ID, utils.StringW),
        CreatedAt: utils.ToTimestamp(a.CreatedAt),
        UpdatedAt: utils.ToTimestamp(a.UpdatedAt),
        Title:     utils.Wrap(a.Title, utils.StringW),
        Content:   utils.Wrap(a.Content, utils.StringW),
        Status:    utils.Wrap(a.Status, utils.StringW),
    }
}
```

---

## 4. Platform Variants by Capability

The Platform struct and ProviderSet change based on which capabilities are enabled.

### With Database (Ent)

```go
// platform.go
type Platform struct {
    cache            *cache.Cache
    handleCacheError CacheErrorHandler
    db               *ent.Client
    handleEntError   EntErrorHandler
}

var ProviderSet = wire.NewSet(
    NewPlatform,
    NewCache, NewCacheErrorHandler,
    NewEntClient, NewEntErrorHandler,
)
```

### With gRPC Client (no database)

```go
// platform.go
type Platform struct {
    cache            *cache.Cache
    handleCacheError CacheErrorHandler
    articleClient    appV1.ArticleServiceClient
    resourceClient   appV1.ResourceServiceClient
}

var ProviderSet = wire.NewSet(
    NewPlatform,
    NewCache, NewCacheErrorHandler,
    NewGRPCArticleClient, NewGRPCResourceClient,
)
```

See `docs/stacks/kratos-go.md` Section 9 (gRPC Client) for the `platform_grpc.go` middleware setup. The client middleware chain MUST end with `grpc_status.Client()` as the innermost middleware — it maps gRPC transport errors to Kratos errors with reason codes so that i18n, logging, and error reporting all work correctly. Import: `"cyber-ecosystem/shared-go/kratos/middleware/grpc_status"`.

### With Both Database and gRPC Client (monolith)

Combine both sets of fields and constructors.

---

## 5. Ent Three-Stage Bootstrap

Local mixins (`soft_delete.go`, `sort.go`) import generated Ent packages that don't exist on first generation. Work around this chicken-and-egg problem:

1. **Stage 1**: Comment out local_mixins from schema, comment out `intercept` feature in ent/generate.go. Run `generate:ent`.
2. **Stage 2**: Uncomment `intercept` feature. Run `generate:ent` again.
3. **Stage 3**: Uncomment local_mixins from schema. Run `generate:ent` again.

After the first successful generation, subsequent runs work normally without staging.

---

## 6. Common Pitfalls

### Single Go Module (No go mod init)

This monorepo uses a single Go module at the repository root. Do NOT run `go mod init` in any service subdirectory. New services are just packages under the existing module — no separate go.mod needed.

### Platform ProviderSet Must Include gRPC Client Constructors

The platform layer's `ProviderSet` must list ALL constructors that Wire needs to resolve, including `NewGRPCXxxClient` functions. Omitting them causes "no provider found" errors:

```go
var ProviderSet = wire.NewSet(
    NewPlatform,
    NewGRPCArticleClient,  // MUST include — Wire needs this to resolve XxxServiceClient
)
```

### OpsServer Must Be a Concrete Struct (Not Type Alias)

Do NOT use `type OpsServer = ops.Server` — Wire cannot resolve type aliases. Define a concrete struct with `Start(ctx) error` and `Stop(ctx) error` methods. See `admin_bff/internal/server/ops.go` for the full implementation.

### GCI Import Formatting

After creating Go files, imports may need formatting. Run `gci write` or the project's format target if available. Imports follow this grouping:

1. Standard library
2. Third-party (github.com, etc.)
3. This project (`cyber-ecosystem/`)

### HTTP/Connect Registration

If a proto service has NO `google.api.http` annotations, the generated code will NOT include `RegisterXxxHTTPServer`. The service's `RegisterHTTP` and `RegisterConnect` must be no-ops. Do NOT try to call non-existent registration functions.

### newApp Server Selection

All services create gRPC, HTTP, Connect, and Ops servers (for structural consistency). `newApp` accepts all four but selectively starts them — see `docs/stacks/kratos-go.md` Section 9 (`newApp` Server Selection).

### Database Name Convention

Database names follow the pattern: `cyber_ecosystem_<app>_service_<service>`. The database must exist before starting the service.

### i18n.protos Relative Paths

Paths in `i18n.protos` are relative from `internal/i18n/` to the proto file. Count directory levels carefully:
- To contracts: `../../../../../../contracts/errors/codes_general.proto`
- To app errors: `../../../../api/v1/error/error_reason.proto`

### Services Without Database

Services without Ent skip the `generate:ent` target and have no `ent/` directory. Their `project.json` omits `generate:ent` and `ent:new` targets. The `Data` config section only has `Cache` and remote service addresses — no `Database` or `Storage`.

### CORS Configuration

HTTP and Connect servers MUST include CORS filter (using `github.com/gorilla/handlers`):

```go
krahttp.Filter(handlers.CORS(
    handlers.AllowedOrigins([]string{"*"}),
    handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}),
    handlers.AllowedHeaders([]string{"Content-Type", i18n.DefaultHeaderLang, "Authorization"}),
    handlers.MaxAge(86400),
)),
```

### i18n Bundle

All services use the same minimal `NewI18nBundle()` — no logger parameter:

```go
func NewI18nBundle() (*i18n.Bundle, error) {
    return i18n.NewBundleFS(locales, "locales", "v1", language.Make("zh-CN"))
}
```

### wireApp Signature

`wireApp` has the same parameters regardless of capabilities — `*conf.Client` is NOT needed because remote service config is inside `*conf.Data`:

```go
func wireApp(
    *conf.Server, *conf.Log, *conf.Data, *conf.Ops,
    log.Logger, *tracesdk.TracerProvider, *metricsdk.MeterProvider,
    metric.Int64Counter, metric.Float64Histogram,
) (*kratos.App, func(), error)
```
