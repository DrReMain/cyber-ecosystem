package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/orm/ent/entutil"
	"cyber-ecosystem/shared-go/utils"

	app1V1 "cyber-ecosystem/apps/app_1/gen/go/v1"
	"cyber-ecosystem/apps/app_1/services/service_1/internal/biz"
	"cyber-ecosystem/apps/app_1/services/service_1/internal/data/ent"
	"cyber-ecosystem/apps/app_1/services/service_1/internal/data/ent/blog"
	"cyber-ecosystem/apps/app_1/services/service_1/internal/data/ent/predicate"
	"cyber-ecosystem/apps/app_1/services/service_1/internal/data/ent/schema"
)

type BlogRP struct {
	RP
}

func NewBlogRP(logger log.Logger, store *Store) biz.BlogRP {
	return &BlogRP{
		RP: RP{
			log:   log.NewHelper(log.With(logger, "module", "data/rp_blog")),
			store: store,
		},
	}
}

// Repo ----------------------------------------------------------------------------------------------------------------

func (rp *BlogRP) Create(ctx context.Context, entity *biz.BlogEntity) error {
	if err := rp.store.InTx(ctx, func(ctx context.Context) error {
		client := rp.store.GetClient(ctx)
		if err := client.Blog.Create().
			SetNillableTitle(entity.Title).
			SetNillableContent(entity.Content).
			SetNillablePublishedAt(entity.PublishedAt).
			Exec(ctx); err != nil {
			return HandleError(err)
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (rp *BlogRP) Update(ctx context.Context, fieldsMask []string, entity *biz.BlogEntity) error {
	if err := rp.store.InTx(ctx, func(ctx context.Context) error {
		builder := rp.store.GetClient(ctx).Blog.UpdateOneID(entity.ID)
		utils.Handler{
			"title": utils.MaskAction{
				Condition: entity.Title == nil,
				OnTrue:    func() { builder.SetTitle(schema.BlogDefaultTitle()) },
				OnFalse:   func() { builder.SetTitle(*entity.Title) },
			},
			"content": utils.MaskAction{
				Condition: entity.Content == nil,
				OnTrue:    func() { builder.SetContent(schema.BlogDefaultContent()) },
				OnFalse:   func() { builder.SetContent(*entity.Content) },
			},
			"publishedAt": utils.MaskAction{
				Condition: entity.PublishedAt == nil,
				OnTrue:    func() { builder.ClearPublishedAt() },
				OnFalse:   func() { builder.SetPublishedAt(*entity.PublishedAt) },
			},
		}.Emit(fieldsMask)
		if err := builder.Exec(ctx); err != nil {
			return HandleError(err)
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (rp *BlogRP) Delete(ctx context.Context, id string) error {
	if err := rp.store.GetClient(ctx).Blog.DeleteOneID(id).Exec(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *BlogRP) DeleteBatch(ctx context.Context, ids []string) (int, error) {
	count, err := rp.store.GetClient(ctx).Blog.Delete().Where(blog.IDIn(ids...)).Exec(ctx)
	if err != nil {
		return 0, HandleError(err)
	}
	return count, err
}

func (rp *BlogRP) Get(ctx context.Context, id string) (*biz.BlogEntity, error) {
	e, err := rp.store.GetClient(ctx).Blog.Get(ctx, id)
	if err != nil {
		return nil, HandleError(err)
	}
	return mapBlog(e), nil
}

func (rp *BlogRP) Query(ctx context.Context, bo *biz.BlogQueryIn) (*biz.BlogQueryOut, error) {
	query := rp.store.GetClient(ctx).Blog.Query()
	entutil.WherePtr(query, utils.FromTimestamp(bo.PageRequest.CreatedAtA), blog.CreatedAtGTE)
	entutil.WherePtr(query, utils.FromTimestamp(bo.PageRequest.CreatedAtZ), blog.CreatedAtLTE)
	entutil.WherePtr(query, utils.FromTimestamp(bo.PageRequest.UpdatedAtA), blog.UpdatedAtGTE)
	entutil.WherePtr(query, utils.FromTimestamp(bo.PageRequest.UpdatedAtZ), blog.UpdatedAtLTE)
	entutil.Where(query, bo.ID != nil, func() predicate.Blog { return blog.IDEQ(*bo.ID) })
	entutil.WherePtr(query, bo.Title, blog.TitleContainsFold)
	entutil.WherePtr(query, bo.PublishedAtA, blog.PublishedAtGTE)
	entutil.WherePtr(query, bo.PublishedAtZ, blog.PublishedAtLTE)
	entutil.ApplyOrderBy(bo.OrderBy, ent.Asc, ent.Desc, entutil.FOMapping{
		"createdAt": func(sel entutil.SQLSelector) { query.Order(sel(blog.FieldCreatedAt)) },
		"updatedAt": func(sel entutil.SQLSelector) { query.Order(sel(blog.FieldUpdatedAt)) },
	})

	total, offset, limit, err := entutil.ApplyPagination(ctx, query, bo.PageRequest,
		entutil.NewPageConfig(entutil.DefaultPageSize, entutil.DefaultPageSizeMax),
		app1V1.ErrorErrorReasonPaginationInvalidArgument(""),
	)
	if err != nil {
		return nil, HandleError(err)
	}

	pos, err := query.All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}

	return &biz.BlogQueryOut{
		PageResponse: entutil.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(pos, mapBlog),
	}, nil
}

func mapBlog(v *ent.Blog) *biz.BlogEntity {
	return &biz.BlogEntity{
		ID:          v.ID,
		CreatedAt:   &v.CreatedAt,
		UpdatedAt:   &v.UpdatedAt,
		Title:       &v.Title,
		Content:     &v.Content,
		PublishedAt: v.PublishedAt,
	}
}
