package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/contracts/go/common"
	"cyber-ecosystem/shared-go/utils"
)

// Model ----------------------------------------------------------------------------------------------------------------

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
	OrderBy      []*utils.OrderBy
	ID           *string
	Title        *string
	PublishedAtA *time.Time
	PublishedAtZ *time.Time
}

type BlogQueryOut struct {
	*common.PageResponse
	List []*BlogEntity
}

// Port -----------------------------------------------------------------------------------------------------------------

type BlogRP interface {
	Create(ctx context.Context, entity *BlogEntity) error
	Update(ctx context.Context, fieldsMask []string, entity *BlogEntity) error
	Delete(ctx context.Context, id string) error
	DeleteBatch(ctx context.Context, ids []string) (int, error)
	Get(ctx context.Context, id string) (*BlogEntity, error)
	Query(ctx context.Context, in *BlogQueryIn) (*BlogQueryOut, error)
}

// UC -------------------------------------------------------------------------------------------------------------------

type BlogUC struct {
	UC
	blogRP BlogRP
}

func NewBlogUC(logger log.Logger, tm Transaction, blogRP BlogRP) *BlogUC {
	return &BlogUC{
		UC: UC{
			log: log.NewHelper(log.With(logger, "module", "biz/uc_blog")),
			tm:  tm,
		},
		blogRP: blogRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

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
