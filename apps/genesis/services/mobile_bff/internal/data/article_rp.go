package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/utils"

	genesisV1 "cyber-ecosystem/apps/genesis/gen/go/v1"
	"cyber-ecosystem/apps/genesis/services/mobile_bff/internal/biz"
	"cyber-ecosystem/apps/genesis/services/mobile_bff/internal/platform"
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

// region[rgba(0,188,212,0.12)] 🩵 Repo

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

// region[rgba(144,164,174,0.10)] ⚪ Private

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
