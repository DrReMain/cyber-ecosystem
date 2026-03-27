package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/contracts/go/common"
	"cyber-ecosystem/shared-go/kratos/order_by"
)

type BlogEntity struct {
	ID          string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
	Title       *string
	Content     *string
	PublishedAt *time.Time
}

type BlogQueryIn struct {
	*common.PageRequest
	OrderBy      []*order_by.OrderBy
	ID           *string
	Title        *string
	PublishedAtA *time.Time
	PublishedAtZ *time.Time
}

type BlogQueryOut struct {
	*common.PageResponse
	List []*BlogEntity
}

type BlogRP interface {
	Create(context.Context, *BlogEntity) error
	Update(context.Context, []string, *BlogEntity) error
	Delete(context.Context, string) error
	DeleteBatch(context.Context, []string) (int, error)
	Get(context.Context, string) (*BlogEntity, error)
	Query(context.Context, *BlogQueryIn) (*BlogQueryOut, error)
}

type BlogUC struct {
	log *log.Helper
	tm  Transaction

	blogRP BlogRP
}

func NewBlogUC(logger log.Logger, tm Transaction, blogRP BlogRP) *BlogUC {
	return &BlogUC{
		log:    log.NewHelper(log.With(logger, "module", "biz/blog")),
		tm:     tm,
		blogRP: blogRP,
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func (uc *BlogUC) CreateBlog(ctx context.Context, entity *BlogEntity) error {
	return uc.blogRP.Create(ctx, entity)
}

func (uc *BlogUC) UpdateBlog(ctx context.Context, fieldsMask []string, entity *BlogEntity) error {
	return uc.tm.InTx(ctx, func(ctx context.Context) error {
		return uc.blogRP.Update(ctx, fieldsMask, entity)
	})
}

func (uc *BlogUC) DeleteBlog(ctx context.Context, id string) error {
	return uc.blogRP.Delete(ctx, id)
}

func (uc *BlogUC) DeleteBatchBlog(ctx context.Context, ids []string) (int, error) {
	return uc.blogRP.DeleteBatch(ctx, ids)
}

func (uc *BlogUC) GetBlog(ctx context.Context, id string) (*BlogEntity, error) {
	return uc.blogRP.Get(ctx, id)
}

func (uc *BlogUC) QueryBlog(ctx context.Context, bo *BlogQueryIn) (*BlogQueryOut, error) {
	return uc.blogRP.Query(ctx, bo)
}
