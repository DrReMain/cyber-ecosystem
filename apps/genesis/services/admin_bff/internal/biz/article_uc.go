package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// region[rgba(66,165,245,0.15)] 🔵 Port --------------------------------------------------------------------------------

type ArticleRP interface {
	Create(ctx context.Context, a *Article) (*Article, error)
	Update(ctx context.Context, fieldsMask []string, a *Article) (*Article, error)
	UpdateStatus(ctx context.Context, id string, status string) (*Article, error)
	Delete(ctx context.Context, id string) (string, error)
	Get(ctx context.Context, id string) (*Article, error)
	Query(ctx context.Context, in *ArticleQueryIn) (*ArticleQueryOut, error)
	Sort(ctx context.Context, id string, prevID, nextID *string) (*Article, error)
}

// region[rgba(102,187,106,0.15)] 🟢 UC ----------------------------------------------------------------------------------

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

// region[rgba(186,104,200,0.15)] 🟣 Method ------------------------------------------------------------------------------

func (uc *ArticleUC) Create(ctx context.Context, a *Article) (out *Article, err error) {
	err = uc.tm.InTx(ctx, func(ctx context.Context) (e error) {
		out, e = uc.articleRP.Create(ctx, a)
		return
	})
	return
}

func (uc *ArticleUC) Update(ctx context.Context, fieldsMask []string, a *Article) (out *Article, err error) {
	err = uc.tm.InTx(ctx, func(ctx context.Context) (e error) {
		out, e = uc.articleRP.Update(ctx, fieldsMask, a)
		return
	})
	return
}

func (uc *ArticleUC) Delete(ctx context.Context, id string) (out string, err error) {
	err = uc.tm.InTx(ctx, func(ctx context.Context) (e error) {
		out, e = uc.articleRP.Delete(ctx, id)
		return
	})
	return
}

func (uc *ArticleUC) Get(ctx context.Context, id string) (*Article, error) {
	return uc.articleRP.Get(ctx, id)
}

func (uc *ArticleUC) Query(ctx context.Context, in *ArticleQueryIn) (*ArticleQueryOut, error) {
	return uc.articleRP.Query(ctx, in)
}

func (uc *ArticleUC) UpdateStatus(ctx context.Context, id string, target string) (out *Article, err error) {
	err = uc.tm.InTx(ctx, func(ctx context.Context) (e error) {
		out, e = uc.articleRP.UpdateStatus(ctx, id, target)
		return
	})
	return
}

func (uc *ArticleUC) Sort(ctx context.Context, id string, prevID, nextID *string) (out *Article, err error) {
	err = uc.tm.InTx(ctx, func(ctx context.Context) (e error) {
		out, e = uc.articleRP.Sort(ctx, id, prevID, nextID)
		return
	})
	return
}
