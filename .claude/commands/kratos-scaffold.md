# Kratos Service Scaffold

Guide for scaffolding and implementing Go microservices using the Kratos framework in this monorepo. Use this when creating a new service, adding a new aggregate/entity, or implementing business logic following the established patterns.

This skill assumes proto files and generation are already set up (see `/proto-design`). Focus here is on the Go service implementation.

---

## 1. Service Directory Structure

```
apps/<app>/services/<service>/
  cmd/app/
    main.go              # Bootstrap: config → logger → metrics → tracing → sentry → wire
    wire.go              # Wire injection (build tag: wireinject)
    wire_gen.go          # Wire generated (DO NOT EDIT)
    logger.go            # Zap logger setup
    metrics.go           # OTel MeterProvider
    tracing.go           # OTel TracerProvider
    sentry.go            # Sentry init
    resource.go          # OTel resource with service metadata
  internal/
    conf/
      conf.proto         # Config proto definition
      conf.pb.go         # Generated config types
    ent/
      schema/            # Ent entity schemas
        article.go
        local_mixins/    # Service-specific mixins
          soft_delete.go
          sort.go
    platform/            # Infrastructure container
      interface.go
      platform.go
      platform_ent.go
      platform_ent_handler.go
      platform_cache.go
      platform_cache_handler.go
    data/                # Repository implementations
      data.go
      article_rp.go
    biz/                 # Domain layer
      biz.go
      article.go
      article_uc.go
      article_fsm.go     # If entity has state machine
    service/             # Proto handler layer
      service.go
      article.go
    server/              # Transport setup
      server.go
      grpc.go
      http.go
      connect.go
      ops.go
    i18n/
      generate.go
      i18n.go
      i18n.protos
      locales/
        v1.zh-CN.yaml
        v1.en-US.yaml
  configs/
    config.yaml
  project.json           # Nx targets
  buf.gen.conf.yaml      # Proto generation config
```

---

## 2. Layer Architecture

Strict layered structure with unidirectional dependencies:

```
server → service → biz ← data → platform
                      ↑         ↑
                   proto      ent
```

**Dependency rule:** arrows point inward. Inner layers MUST NOT import outer layers.

**Exception:** biz layer MAY import proto (generated error codes) for `ErrorErrorReasonXxx("")` calls. This is the only allowed biz → outer dependency.

| Layer | Package | Responsibility |
|-------|---------|----------------|
| server | `internal/server/` | Transport setup, middleware chain |
| service | `internal/service/` | Proto request/response mapping |
| biz | `internal/biz/` | Domain models, use cases, RP port interfaces |
| data | `internal/data/` | Repository implementations (RP) |
| platform | `internal/platform/` | DB, cache, error handling |
| ent | `internal/ent/` | Ent ORM schemas + generated code |

---

## 3. Adding a New Aggregate

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
    "cyber-ecosystem/shared-go/orm/ent/entutil"
    "cyber-ecosystem/shared-go/utils"
    common "cyber-ecosystem/contracts/go/common"
)

type Comment struct {
    ID        *string
    CreatedAt *utils.TimeString
    UpdatedAt *utils.TimeString
    Content   *string
    Status    *string
}

type CommentQueryIn struct {
    entutil.PageRequest
    entutil.OrderBy
    Content *string
    Status  *string
}

type CommentQueryOut struct {
    *entutil.PageResponse
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
func (s *CommentService) RegisterHTTP(_ *http.Server)       {}
func (s *CommentService) RegisterConnect(_ *connect.Server) {}
```

### Step 7: Wire Everything

Update each layer's ProviderSet:

- `data/data.go`: add `NewCommentRP` to ProviderSet
- `biz/biz.go`: add `NewCommentUC` to ProviderSet
- `service/service.go`: add `NewCommentService` to ProviderSet and registrar list

Then regenerate: `./nx run <project>:generate:wire`

---

## 4. Key Patterns

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

Key utils:
- `utils.Wrap(ptr, utils.StringW)` — `*string` → `*wrapperspb.StringValue`
- `utils.StringW(value)` — `string` → `*wrapperspb.StringValue`
- `utils.ToTimestamp(timePtr)` — `*time.Time` → `*timestamppb.Timestamp`
- `utils.SliceMap(slice, fn)` — `[]T` → `[]R`
- `utils.EnsurePageRequest(req)` — nil-safe page request
- `utils.ParseOrderBy([]string)` — `["field:asc"]` → `[]*OrderBy`

---

## 5. Server Setup

### Middleware Chain

All transport servers share the same middleware in this order:

```
i18n → recovery → ratelimit → metrics → tracing → connect_span* → metadata → logging → validate → error_report
```

`connect_span` only appears on the Connect server.

### Registrar Pattern

Services register themselves to transports via the `Registrar` interface:

```go
type Registrar interface {
    RegisterGRPC(*grpc.Server)
    RegisterHTTP(*http.Server)
    RegisterConnect(*connect.Server)
}
```

For gRPC-only services (base service), `RegisterHTTP` and `RegisterConnect` are no-ops.

### Error Override in server.go

```go
func init() {
    recovery.ErrUnknownRequest = errorspb.ErrorGeneralErrorUnspecified("").WithCause(recovery.ErrUnknownRequest)
    ratelimit.ErrLimitExceed = errorspb.ErrorFlowErrorRateLimited("").WithCause(ratelimit.ErrLimitExceed)
    validate.ErrVALIDATOR = errorspb.ErrorGeneralErrorValidationFailed("").WithCause(validate.ErrVALIDATOR)
}
```

---

## 6. Platform Layer

The Platform struct is the infrastructure container. It holds cache and DB clients and provides error handlers.

```go
type Platform struct {
    cache            *cache.Cache
    client           *ent.Client
    HandleEntError   func(error) error
    HandleCacheError func(error) error
}

func (p *Platform) InTx(ctx context.Context, fn func(context.Context) error) error { ... }
func (p *Platform) GetClient(ctx context.Context) *ent.Client { ... }
func (p *Platform) GetCache() *cache.Cache { ... }
```

Error handlers map infrastructure errors to standardized `errorspb` errors. The data layer calls them via `rp.platform.HandleEntError(err)` and `rp.platform.HandleCacheError(err)`.

---

## 7. Wire DI

### Composition Root (`cmd/app/wire.go`)

```go
func wireApp(...) (*kratos.App, func(), error) {
    panic(wire.Build(
        server.ProviderSet,
        service.ProviderSet,
        biz.ProviderSet,
        data.ProviderSet,
        i18n.ProviderSet,
        platform.ProviderSet,
        wire.Bind(new(biz.Transaction), new(*platform.Platform)),
        newApp,
    ))
}
```

Each package's `ProviderSet` registers its constructors. Interface bindings (`wire.Bind`) live exclusively in `wire.go`.

### main.go Bootstrap Order

```
config → logger → metrics → tracing → sentry → wireApp() → app.Run()
```

`newApp` accepts all server instances but may not start all of them. For base (gRPC-only) services, only gRPC and Ops servers are started.

---

## 8. Ent Three-Stage Bootstrap

Local mixins (`soft_delete.go`, `sort.go`) import generated Ent packages that don't exist on first generation. Work around this chicken-and-egg problem:

1. **Stage 1**: Comment out local_mixins from schema, comment out `intercept` feature in ent/generate.go. Run `generate:ent`.
2. **Stage 2**: Uncomment `intercept` feature. Run `generate:ent` again.
3. **Stage 3**: Uncomment local_mixins from schema. Run `generate:ent` again.

After the first successful generation, subsequent runs work normally without staging.

---

## 9. Nx Targets

Standard targets in `project.json`:

```json
{
  "targets": {
    "proto:conf":       "Generate config proto",
    "generate:i18n":    "Generate i18n stubs",
    "generate:ent":     "Generate Ent ORM",
    "generate:wire":    "Generate Wire DI",
    "generate":         "Full chain: proto + i18n + ent + wire + go mod tidy",
    "dev":              "Run with kratos",
    "build":            "Build binary with version",
    "ent:new":          "Create new Ent schema"
  }
}
```

Always use `./nx run <project>:<target>` from workspace root. Never cd into subdirectories or run Go tools directly.

---

## 10. Common Pitfalls

### GCI Import Formatting

After creating Go files, imports may need formatting. Run `gci write` or the project's format target if available. Imports follow this grouping:

1. Standard library
2. Third-party (github.com, etc.)
3. This project (`cyber-ecosystem/`)

### HTTP/Connect Registration

If a proto service has NO `google.api.http` annotations, the generated code will NOT include `RegisterXxxHTTPServer`. The service's `RegisterHTTP` and `RegisterConnect` must be no-ops. Do NOT try to call non-existent registration functions.

### newApp Server Selection

For gRPC-only services, `newApp` only appends gRPC and Ops servers to `kratos.Server()`. HTTP and Connect server instances exist (for Wire compatibility) but are not started.

### Database Name Convention

Database names follow the pattern: `cyber_ecosystem_<app>_service_<service>`. The database must exist before starting the service.

### i18n.protos Relative Paths

Paths in `i18n.protos` are relative from `internal/i18n/` to the proto file. Count directory levels carefully:
- To contracts: `../../../../../../contracts/errors/codes_general.proto`
- To app errors: `../../../../api/v1/error/error_reason.proto`
