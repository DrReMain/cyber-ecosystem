package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// region[rgba(66,165,245,0.15)] 🔵 Port --------------------------------------------------------------------------------

type ArticleRP interface {
	Create(ctx context.Context, a *Article) (*Article, error)
	Update(ctx context.Context, fieldsMask []string, a *Article) (*Article, error)
	Delete(ctx context.Context, id string) (string, error)
	Get(ctx context.Context, id string) (*Article, error)
	Query(ctx context.Context, in *ArticleQueryIn) (*ArticleQueryOut, error)
	Sort(ctx context.Context, id string, prevID, nextID *string) (*Article, error)
	UpdateStatus(ctx context.Context, id string, target string) (*Article, error)
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

func (uc *ArticleUC) Create(ctx context.Context, a *Article) (*Article, error) {
	return uc.articleRP.Create(ctx, a)
}

func (uc *ArticleUC) Update(ctx context.Context, fieldsMask []string, a *Article) (*Article, error) {
	return uc.articleRP.Update(ctx, fieldsMask, a)
}

func (uc *ArticleUC) Delete(ctx context.Context, id string) (string, error) {
	return uc.articleRP.Delete(ctx, id)
}

func (uc *ArticleUC) Get(ctx context.Context, id string) (*Article, error) {
	return uc.articleRP.Get(ctx, id)
}

func (uc *ArticleUC) Query(ctx context.Context, in *ArticleQueryIn) (*ArticleQueryOut, error) {
	return uc.articleRP.Query(ctx, in)
}

func (uc *ArticleUC) Sort(ctx context.Context, id string, prevID, nextID *string) (*Article, error) {
	return uc.articleRP.Sort(ctx, id, prevID, nextID)
}

func (uc *ArticleUC) UpdateStatus(ctx context.Context, id string, target string) (*Article, error) {
	return uc.articleRP.UpdateStatus(ctx, id, target)
}
