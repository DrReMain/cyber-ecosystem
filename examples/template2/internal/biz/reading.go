package biz

import (
	"context"
	"time"

	"github.com/DrReMain/cyber-ecosystem/gen/go/common"
	templateV1 "github.com/DrReMain/cyber-ecosystem/gen/go/template1/v1"
	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/order_by"
	"github.com/DrReMain/cyber-ecosystem/shared-go/kratos/util"
)

type BlogWithReadingEntity struct {
	ID           string
	Title        *string
	Content      *string
	PublishedAt  *time.Time
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
	ReadingCount int64
}

type ReadingQueryIn struct {
	*common.PageRequest
	OrderBy      []*order_by.OrderBy
	ID           *string
	Title        *string
	PublishedAtA *time.Time
	PublishedAtZ *time.Time
}

type ReadingQueryOut struct {
	*common.PageResponse
	List []*BlogWithReadingEntity
}

type ReadingRP interface {
	GetReadingCount(ctx context.Context, blogID string) (int64, error)
	GetReadingCounts(ctx context.Context, blogIDs []string) (map[string]int64, error)
	IncrementReading(ctx context.Context, blogID string) error

	QueryBlog(ctx context.Context, in *templateV1.QueryBlogRequest) (*templateV1.QueryBlogResponse, error)
	GetBlog(ctx context.Context, in *templateV1.GetBlogRequest) (*templateV1.GetBlogResponse, error)
}

type ReadingUC struct {
	tm Transaction

	readingRP ReadingRP
}

func NewReadingUC(tm Transaction, readingRP ReadingRP) *ReadingUC {
	return &ReadingUC{tm: tm, readingRP: readingRP}
}

// ---------------------------------------------------------------------------------------------------------------------

func (uc *ReadingUC) QueryBlog(ctx context.Context, bo *ReadingQueryIn) (*ReadingQueryOut, error) {
	resp, err := uc.readingRP.QueryBlog(ctx, &templateV1.QueryBlogRequest{
		Page:         bo.PageRequest,
		Id:           bo.ID,
		Title:        bo.Title,
		PublishedAtA: util.GetPPbTimeFromPTime(bo.PublishedAtA),
		PublishedAtZ: util.GetPPbTimeFromPTime(bo.PublishedAtZ),
		OrderBy:      order_by.StringifyOrderBy(bo.OrderBy),
	})
	if err != nil {
		return nil, err
	}

	blogIDs := make([]string, len(resp.List))
	for i, blog := range resp.List {
		blogIDs[i] = blog.Id
	}

	readingCounts, err := uc.readingRP.GetReadingCounts(ctx, blogIDs)
	if err != nil {
		return nil, err
	}

	list := make([]*BlogWithReadingEntity, len(resp.List))
	for i, blog := range resp.List {
		list[i] = &BlogWithReadingEntity{
			ID:           blog.Id,
			Title:        util.GetStringFromWrapper(blog.Title),
			Content:      util.GetStringFromWrapper(blog.Content),
			PublishedAt:  util.GetPTimeFromPPbTime(blog.PublishedAt),
			CreatedAt:    util.GetPTimeFromPPbTime(blog.CreatedAt),
			UpdatedAt:    util.GetPTimeFromPPbTime(blog.UpdatedAt),
			ReadingCount: readingCounts[blog.Id],
		}
	}

	return &ReadingQueryOut{
		PageResponse: resp.Page,
		List:         list,
	}, nil
}

func (uc *ReadingUC) GetBlog(ctx context.Context, id string) (*BlogWithReadingEntity, error) {
	blog, err := uc.readingRP.GetBlog(ctx, &templateV1.GetBlogRequest{Id: id})
	if err != nil {
		return nil, err
	}

	readingCount, err := uc.readingRP.GetReadingCount(ctx, id)
	if err != nil {
		return nil, err
	}

	return &BlogWithReadingEntity{
		ID:           blog.Id,
		Title:        util.GetStringFromWrapper(blog.Title),
		Content:      util.GetStringFromWrapper(blog.Content),
		PublishedAt:  util.GetPTimeFromPPbTime(blog.PublishedAt),
		CreatedAt:    util.GetPTimeFromPPbTime(blog.CreatedAt),
		UpdatedAt:    util.GetPTimeFromPPbTime(blog.UpdatedAt),
		ReadingCount: readingCount,
	}, nil
}

func (uc *ReadingUC) RecordReading(ctx context.Context, id string) (int64, error) {
	var readingCount int64
	if err := uc.tm.InTx(ctx, func(ctx context.Context) error {
		if err := uc.readingRP.IncrementReading(ctx, id); err != nil {
			return err
		}
		count, err := uc.readingRP.GetReadingCount(ctx, id)
		if err != nil {
			return err
		}
		readingCount = count
		return nil
	}); err != nil {
		return 0, err
	}
	return readingCount, nil
}
