package data

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"roci.dev/fracdex"

	"github.com/go-kratos/kratos/v2/log"

	errorspb "cyber-ecosystem/contracts/go/errors"
	"cyber-ecosystem/shared-go/orm/ent/entutil"
	"cyber-ecosystem/shared-go/utils"

	"cyber-ecosystem/apps/genesis/services/base/internal/biz"
	"cyber-ecosystem/apps/genesis/services/base/internal/ent"
	"cyber-ecosystem/apps/genesis/services/base/internal/ent/article"
	"cyber-ecosystem/apps/genesis/services/base/internal/ent/predicate"
	"cyber-ecosystem/apps/genesis/services/base/internal/platform"
)

type articleRP struct {
	RP
}

func NewArticleRP(logger log.Logger, p *platform.Platform) biz.ArticleRP {
	return &articleRP{
		RP: RP{
			log:      log.NewHelper(log.With(logger, "module", "data/article_rp")),
			platform: p,
		},
	}
}

// region[rgba(0,188,212,0.12)] 🩵 Repo --------------------------------------------------------------------------------

func (rp *articleRP) Create(ctx context.Context, a *biz.Article) (*biz.Article, error) {
	created, err := rp.platform.GetClient(ctx).Article.Create().
		SetTitle(*a.Title).
		SetNillableContent(a.Content).
		Save(ctx)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	return mapArticle(created), nil
}

func (rp *articleRP) Update(ctx context.Context, fieldsMask []string, a *biz.Article) (*biz.Article, error) {
	updater := rp.platform.GetClient(ctx).Article.UpdateOneID(*a.ID)
	utils.Handler{
		"title": {
			Condition: a.Title != nil,
			OnTrue:    func() { updater.SetTitle(*a.Title) },
			OnFalse:   func() {},
		},
		"content": {
			Condition: a.Content != nil,
			OnTrue:    func() { updater.SetContent(*a.Content) },
			OnFalse:   func() { updater.SetContent("") },
		},
		"status": {
			Condition: a.Status != nil,
			OnTrue:    func() { updater.SetStatus(*a.Status) },
			OnFalse:   func() {},
		},
	}.Emit(fieldsMask)

	updated, err := updater.Save(ctx)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	return mapArticle(updated), nil
}

func (rp *articleRP) Delete(ctx context.Context, id string) (string, error) {
	if err := rp.platform.GetClient(ctx).Article.DeleteOneID(id).Exec(ctx); err != nil {
		return "", rp.platform.HandleEntError(err)
	}
	return id, nil
}

func (rp *articleRP) Get(ctx context.Context, id string) (*biz.Article, error) {
	d, err := rp.platform.GetClient(ctx).Article.Get(ctx, id)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	return mapArticle(d), nil
}

func (rp *articleRP) Query(ctx context.Context, in *biz.ArticleQueryIn) (*biz.ArticleQueryOut, error) {
	query := rp.platform.GetClient(ctx).Article.Query()
	entutil.WherePtr(query, utils.FromTimestamp(in.PageRequest.CreatedAtA), article.CreatedAtGTE)
	entutil.WherePtr(query, utils.FromTimestamp(in.PageRequest.CreatedAtZ), article.CreatedAtLTE)
	entutil.WherePtr(query, utils.FromTimestamp(in.PageRequest.UpdatedAtA), article.UpdatedAtGTE)
	entutil.WherePtr(query, utils.FromTimestamp(in.PageRequest.UpdatedAtZ), article.UpdatedAtLTE)
	entutil.Where(query, in.ID != nil, func() predicate.Article { return article.IDEQ(*in.ID) })
	entutil.Where(query, in.Title != nil, func() predicate.Article { return article.TitleContainsFold(*in.Title) })
	entutil.Where(query, in.Status != nil, func() predicate.Article { return article.StatusEQ(*in.Status) })
	entutil.ApplyOrderBy(in.OrderBy, ent.Asc, ent.Desc, entutil.FOMapping{
		"createdAt": func(sel entutil.SQLSelector) { query.Order(sel(article.FieldCreatedAt)) },
		"updatedAt": func(sel entutil.SQLSelector) { query.Order(sel(article.FieldUpdatedAt)) },
		"sort":      func(sel entutil.SQLSelector) { query.Order(sel(article.FieldSort)) },
	})
	query.Order(func(s *sql.Selector) { s.OrderBy(s.C(article.FieldSort)) })

	total, offset, limit, err := entutil.ApplyPagination(ctx, query, in.PageRequest,
		entutil.NewPageConfig(entutil.DefaultPageSize, entutil.DefaultPageSizeUnlimit),
		errorspb.ErrorGeneralErrorPaginationInvalidArgument(""),
	)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	items, err := query.All(ctx)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	return &biz.ArticleQueryOut{
		PageResponse: entutil.BuildPageResponse(total, offset, limit),
		List:         utils.SliceMap(items, mapArticle),
	}, nil
}

func (rp *articleRP) Sort(ctx context.Context, id string, prevID, nextID *string) (*biz.Article, error) {
	var prevSort, nextSort string
	client := rp.platform.GetClient(ctx)

	if prevID != nil {
		d, err := client.Article.Get(ctx, *prevID)
		if err != nil {
			return nil, rp.platform.HandleEntError(err)
		}
		prevSort = d.Sort
	}

	if nextID != nil {
		d, err := client.Article.Get(ctx, *nextID)
		if err != nil {
			return nil, rp.platform.HandleEntError(err)
		}
		nextSort = d.Sort
	}

	newSort, err := fracdex.KeyBetween(prevSort, nextSort)
	if err != nil {
		return nil, err
	}

	updated, err := client.Article.UpdateOneID(id).SetSort(newSort).Save(ctx)
	if err != nil {
		return nil, rp.platform.HandleEntError(err)
	}
	return mapArticle(updated), nil
}

// region[rgba(144,164,174,0.10)] ⚪ Private ---------------------------------------------------------------------------

func mapArticle(d *ent.Article) *biz.Article {
	return &biz.Article{
		ID:        &d.ID,
		CreatedAt: &d.CreatedAt,
		UpdatedAt: &d.UpdatedAt,
		Title:     &d.Title,
		Content:   &d.Content,
		Status:    &d.Status,
	}
}
