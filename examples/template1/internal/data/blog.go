package data

import (
	"context"

	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/biz"
	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/data/ent"
	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/data/ent/blog"
	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/data/ent/predicate"
	"github.com/DrReMain/cyber-ecosystem/examples/template1/internal/data/ent/schema"

	template1V1 "github.com/DrReMain/cyber-ecosystem/gen/go/template1/v1"
	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/masks"
	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/util"
	"github.com/DrReMain/cyber-ecosystem/shared-go/orm/ent/entutil"

	"github.com/go-kratos/kratos/v2/log"
	//"entgo.io/ent/dialect/sql"
)

type blogRP struct {
	log *log.Helper

	data *Data
}

func NewBlogRP(logger log.Logger, data *Data) biz.BlogRP {
	return &blogRP{
		log:  log.NewHelper(log.With(logger, "module", "data/blog")),
		data: data,
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func (rp *blogRP) Create(ctx context.Context, entity *biz.BlogEntity) error {
	// 演示 repo 内判断是否有事务透传，没有则开启事务
	if err := rp.data.InTx(ctx, func(ctx context.Context) error {
		client := rp.data.getClient(ctx)
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

func (rp *blogRP) Update(ctx context.Context, fieldsMask []string, entity *biz.BlogEntity) error {
	// 演示事务透传
	if err := rp.data.InTx(ctx, func(ctx context.Context) error {
		// 演示根据fields_mask来更新字段
		builder := rp.data.getClient(ctx).Blog.UpdateOneID(entity.ID)
		masks.Handler{
			"title": {
				entity.Title == nil,
				func() { builder.SetTitle(schema.BlogDefaultTitle()) },
				func() { builder.SetTitle(*entity.Title) },
			},
			"content": {
				entity.Content == nil,
				func() { builder.SetContent(schema.BlogDefaultContent()) },
				func() { builder.SetContent(*entity.Content) },
			},
			"publishedAt": {
				entity.PublishedAt == nil,
				func() { builder.ClearPublishedAt() },
				func() { builder.SetPublishedAt(*entity.PublishedAt) },
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

func (rp *blogRP) Delete(ctx context.Context, id string) error {
	if err := rp.data.getClient(ctx).Blog.DeleteOneID(id).Exec(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *blogRP) DeleteBatch(ctx context.Context, ids []string) (int, error) {
	count, err := rp.data.getClient(ctx).Blog.Delete().Where(blog.IDIn(ids...)).Exec(ctx)
	if err != nil {
		return 0, HandleError(err)
	}
	return count, err
}

func (rp *blogRP) Get(ctx context.Context, id string) (*biz.BlogEntity, error) {
	e, err := rp.data.getClient(ctx).Blog.Get(ctx, id)
	if err != nil {
		return nil, HandleError(err)
	}
	res := &biz.BlogEntity{
		ID:          e.ID,
		Title:       &e.Title,
		Content:     &e.Content,
		PublishedAt: e.PublishedAt,
		CreatedAt:   &e.CreatedAt,
		UpdatedAt:   &e.UpdatedAt,
	}
	return res, nil
}

func (rp *blogRP) Query(ctx context.Context, bo *biz.BlogQueryIn) (*biz.BlogQueryOut, error) {
	query := rp.data.getClient(ctx).Blog.Query()
	entutil.WherePtr(query, util.GetPTimeFromPPbTime(bo.PageRequest.CreatedAtA), blog.CreatedAtGTE)
	entutil.WherePtr(query, util.GetPTimeFromPPbTime(bo.PageRequest.CreatedAtZ), blog.CreatedAtLTE)
	entutil.WherePtr(query, util.GetPTimeFromPPbTime(bo.PageRequest.UpdatedAtA), blog.UpdatedAtGTE)
	entutil.WherePtr(query, util.GetPTimeFromPPbTime(bo.PageRequest.UpdatedAtZ), blog.UpdatedAtLTE)
	entutil.Where(query, bo.ID != nil, func() predicate.Blog { return blog.IDEQ(*bo.ID) })
	entutil.WherePtr(query, bo.Title, blog.TitleContainsFold)
	entutil.WherePtr(query, bo.PublishedAtA, blog.PublishedAtGTE)
	entutil.WherePtr(query, bo.PublishedAtZ, blog.PublishedAtLTE)
	//query.Order(func(selector *sql.Selector) {
	//	selector.OrderExpr(sql.IsNull(selector.C(blog.FieldPublishedAt)))
	//})
	entutil.ApplyOrderBy(bo.OrderBy, ent.Asc, ent.Desc, entutil.FOMapping{
		"createdAt": func(sel entutil.SQLSelector) { query.Order(sel(blog.FieldCreatedAt)) },
		"updatedAt": func(sel entutil.SQLSelector) { query.Order(sel(blog.FieldUpdatedAt)) },
	})

	total, offset, limit, err := entutil.ApplyPagination(ctx, query, bo.PageRequest,
		entutil.NewPageConfig(entutil.DefaultPageSize, entutil.DefaultPageSizeMax),
		template1V1.ErrorReason_ERROR_REASON_PAGINATION_INVALID_ARGUMENT.String(),
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
		List: func() []*biz.BlogEntity {
			result := make([]*biz.BlogEntity, len(pos))
			for i, v := range pos {
				result[i] = &biz.BlogEntity{
					ID:          v.ID,
					Title:       &v.Title,
					Content:     &v.Content,
					PublishedAt: v.PublishedAt,
					CreatedAt:   &v.CreatedAt,
					UpdatedAt:   &v.UpdatedAt,
				}
			}
			return result
		}(),
	}, nil
}
