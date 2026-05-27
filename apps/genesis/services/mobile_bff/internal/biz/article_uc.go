package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// region[rgba(66,165,245,0.15)] 🔵 Port

type ArticleRP interface {
	Create(ctx context.Context, a *Article) (*Article, error)
	Query(ctx context.Context, in *ArticleQueryIn) (*ArticleQueryOut, error)
}

// region[rgba(102,187,106,0.15)] 🟢 UC

type ArticleUC struct {
	UC
	articleRP ArticleRP
}

func NewArticleUC(logger log.Logger, tm Transaction, articleRP ArticleRP) *ArticleUC {
	return &ArticleUC{
		UC: UC{
			log: log.NewHelper(log.With(logger, "module", "biz/article_uc")),
			tm:  tm,
		},
		articleRP: articleRP,
	}
}

// region[rgba(186,104,200,0.15)] 🟣 Method

func (uc *ArticleUC) Create(ctx context.Context, a *Article) (out *Article, err error) {
	err = uc.tm.InTx(ctx, func(ctx context.Context) (e error) {
		out, e = uc.articleRP.Create(ctx, a)
		return
	})
	return
}

func (uc *ArticleUC) Query(ctx context.Context, in *ArticleQueryIn) (*ArticleQueryOut, error) {
	return uc.articleRP.Query(ctx, in)
}
