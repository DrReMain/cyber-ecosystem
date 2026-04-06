package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/contracts/go/common"
	"cyber-ecosystem/shared-go/utils"

	app1V1 "cyber-ecosystem/apps/app_1/gen/go/v1"
)

// Model ----------------------------------------------------------------------------------------------------------------

type BlogWithReadingEntity struct {
	ID           string
	CreatedAt    *time.Time
	UpdatedAt    *time.Time
	Title        *string
	Content      *string
	PublishedAt  *time.Time
	ReadingCount int64
}

type ReadingQueryIn struct {
	*common.PageRequest
	OrderBy      []*utils.OrderBy
	ID           *string
	Title        *string
	PublishedAtA *time.Time
	PublishedAtZ *time.Time
}

type ReadingQueryOut struct {
	*common.PageResponse
	List []*BlogWithReadingEntity
}

// Port -----------------------------------------------------------------------------------------------------------------

type ReadingRP interface {
	GetReadingCount(ctx context.Context, blogID string) (int64, error)
	GetReadingCounts(ctx context.Context, blogIDs []string) (map[string]int64, error)
	IncrementReading(ctx context.Context, blogID string) error

	QueryBlog(ctx context.Context, in *app1V1.QueryBlogReadingRequest) (*app1V1.QueryBlogReadingResponse, error)
	GetBlog(ctx context.Context, in *app1V1.GetBlogReadingRequest) (*app1V1.GetBlogReadingResponse, error)
}

// UC -------------------------------------------------------------------------------------------------------------------

type ReadingUC struct {
	UC
	readingRP ReadingRP
}

func NewReadingUC(logger log.Logger, tm Transaction, readingRP ReadingRP) *ReadingUC {
	return &ReadingUC{
		UC: UC{
			log: log.NewHelper(log.With(logger, "module", "biz/uc_reading")),
			tm:  tm,
		},
		readingRP: readingRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *ReadingUC) QueryBlog(ctx context.Context, bo *ReadingQueryIn) (*ReadingQueryOut, error) {
	resp, err := uc.readingRP.QueryBlog(ctx, &app1V1.QueryBlogReadingRequest{
		Page:         bo.PageRequest,
		Id:           bo.ID,
		Title:        bo.Title,
		PublishedAtA: utils.ToTimestamp(bo.PublishedAtA),
		PublishedAtZ: utils.ToTimestamp(bo.PublishedAtZ),
		OrderBy:      utils.StringifyOrderBy(bo.OrderBy),
	})
	if err != nil {
		return nil, err
	}

	blogIds := make([]string, len(resp.List))
	for i, blog := range resp.List {
		blogIds[i] = blog.Id
	}

	readingCounts, err := uc.readingRP.GetReadingCounts(ctx, blogIds)
	if err != nil {
		return nil, err
	}

	list := make([]*BlogWithReadingEntity, 0, len(resp.List))
	for _, blog := range resp.List {
		list = append(list, &BlogWithReadingEntity{
			ID:           blog.Id,
			CreatedAt:    utils.FromTimestamp(blog.CreatedAt),
			UpdatedAt:    utils.FromTimestamp(blog.UpdatedAt),
			Title:        utils.Unwrap[string](blog.Title),
			Content:      utils.Unwrap[string](blog.Content),
			PublishedAt:  utils.FromTimestamp(blog.PublishedAt),
			ReadingCount: readingCounts[blog.Id],
		})
	}

	return &ReadingQueryOut{
		PageResponse: resp.Page,
		List:         list,
	}, nil
}

func (uc *ReadingUC) GetBlog(ctx context.Context, id string) (*BlogWithReadingEntity, error) {
	blog, err := uc.readingRP.GetBlog(ctx, &app1V1.GetBlogReadingRequest{Id: id})
	if err != nil {
		return nil, err
	}

	readingCount, err := uc.readingRP.GetReadingCount(ctx, id)
	if err != nil {
		return nil, err
	}

	return &BlogWithReadingEntity{
		ID:           blog.Id,
		CreatedAt:    utils.FromTimestamp(blog.CreatedAt),
		UpdatedAt:    utils.FromTimestamp(blog.UpdatedAt),
		Title:        utils.Unwrap[string](blog.Title),
		Content:      utils.Unwrap[string](blog.Content),
		PublishedAt:  utils.FromTimestamp(blog.PublishedAt),
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
