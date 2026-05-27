package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/utils"

	genesisV1 "cyber-ecosystem/apps/genesis/gen/go/v1"
	"cyber-ecosystem/apps/genesis/services/admin_bff/internal/biz"
	"cyber-ecosystem/apps/genesis/services/admin_bff/internal/platform"
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
	resp, err := rp.platform.GetArticleClient().CreateArticle(ctx, &genesisV1.CreateArticleRequest{
		Title:   a.Title,
		Content: a.Content,
	})
	if err != nil {
		return nil, err
	}
	return &biz.Article{ID: utils.Unwrap(resp.Id)}, nil
}

func (rp *articleRP) Update(ctx context.Context, fieldsMask []string, a *biz.Article) (*biz.Article, error) {
	resp, err := rp.platform.GetArticleClient().UpdateArticle(ctx, &genesisV1.UpdateArticleRequest{
		Id:         a.ID,
		Title:      a.Title,
		Content:    a.Content,
		FieldsMask: fieldsMask,
	})
	if err != nil {
		return nil, err
	}
	return &biz.Article{ID: utils.Unwrap(resp.Id)}, nil
}

func (rp *articleRP) Delete(ctx context.Context, id string) (string, error) {
	resp, err := rp.platform.GetArticleClient().DeleteArticle(ctx, &genesisV1.DeleteArticleRequest{
		Id: utils.Ptr(id),
	})
	if err != nil {
		return "", err
	}
	return utils.Deref(utils.Unwrap(resp.Id), ""), nil
}

func (rp *articleRP) Get(ctx context.Context, id string) (*biz.Article, error) {
	resp, err := rp.platform.GetArticleClient().GetArticle(ctx, &genesisV1.GetArticleRequest{
		Id: utils.Ptr(id),
	})
	if err != nil {
		return nil, err
	}
	return protoToArticle(resp), nil
}

func (rp *articleRP) Query(ctx context.Context, in *biz.ArticleQueryIn) (*biz.ArticleQueryOut, error) {
	resp, err := rp.platform.GetArticleClient().QueryArticle(ctx, &genesisV1.QueryArticleRequest{
		Page:    utils.EnsurePageRequest(in.PageRequest),
		OrderBy: utils.StringifyOrderBy(in.OrderBy),
		Title:   in.Title,
		Status:  in.Status,
	})
	if err != nil {
		return nil, err
	}
	return &biz.ArticleQueryOut{
		PageResponse: resp.Page,
		List:         utils.SliceMap(resp.List, protoToArticle),
	}, nil
}

func (rp *articleRP) Sort(ctx context.Context, id string, prevID, nextID *string) (*biz.Article, error) {
	resp, err := rp.platform.GetArticleClient().SortArticle(ctx, &genesisV1.SortArticleRequest{
		Id:     utils.Ptr(id),
		PrevId: prevID,
		NextId: nextID,
	})
	if err != nil {
		return nil, err
	}
	return &biz.Article{ID: utils.Unwrap(resp.Id)}, nil
}

func (rp *articleRP) UpdateStatus(ctx context.Context, id string, target string) (*biz.Article, error) {
	resp, err := rp.platform.GetArticleClient().UpdateArticleStatus(ctx, &genesisV1.UpdateArticleStatusRequest{
		Id:     utils.Ptr(id),
		Status: utils.Ptr(target),
	})
	if err != nil {
		return nil, err
	}
	return &biz.Article{ID: utils.Unwrap(resp.Id)}, nil
}

// region[rgba(144,164,174,0.10)] ⚪ Private ---------------------------------------------------------------------------

func protoToArticle(resp *genesisV1.GetArticleResponse) *biz.Article {
	return &biz.Article{
		ID:        utils.Unwrap(resp.Id),
		CreatedAt: utils.FromTimestamp(resp.CreatedAt),
		UpdatedAt: utils.FromTimestamp(resp.UpdatedAt),
		Title:     utils.Unwrap(resp.Title),
		Content:   utils.Unwrap(resp.Content),
		Status:    utils.Unwrap(resp.Status),
	}
}
