package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"

	app1V1 "cyber-ecosystem/apps/app_1/gen/go/v1"
	"cyber-ecosystem/apps/app_1/services/service_2/internal/biz"
	"cyber-ecosystem/apps/app_1/services/service_2/internal/data/ent"
	"cyber-ecosystem/apps/app_1/services/service_2/internal/data/ent/reading"
)

type ReadingRP struct {
	RP
}

func NewReadingRP(logger log.Logger, store *Store) biz.ReadingRP {
	return &ReadingRP{
		RP: RP{
			log:   log.NewHelper(log.With(logger, "module", "data/rp_reading")),
			store: store,
		},
	}
}

// Repo ----------------------------------------------------------------------------------------------------------------

func (rp *ReadingRP) GetReadingCount(ctx context.Context, blogID string) (int64, error) {
	r, err := rp.store.GetClient(ctx).Reading.Query().Where(reading.BlogIDEQ(blogID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return 0, nil
		}
		return 0, HandleError(err)
	}
	return r.ReadingCount, nil
}

func (rp *ReadingRP) GetReadingCounts(ctx context.Context, blogIDs []string) (map[string]int64, error) {
	result := make(map[string]int64, len(blogIDs))
	for _, id := range blogIDs {
		result[id] = 0
	}
	readings, err := rp.store.GetClient(ctx).Reading.Query().
		Where(reading.BlogIDIn(blogIDs...)).
		All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}
	for _, r := range readings {
		result[r.BlogID] = r.ReadingCount
	}
	return result, nil
}

func (rp *ReadingRP) IncrementReading(ctx context.Context, blogID string) error {
	client := rp.store.GetClient(ctx)

	r, err := client.Reading.Query().
		Where(reading.BlogIDEQ(blogID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			if err := client.Reading.Create().
				SetBlogID(blogID).
				SetReadingCount(1).
				Exec(ctx); err != nil {
				return HandleError(err)
			}
			return nil
		}
		return HandleError(err)
	}
	if err := client.Reading.UpdateOneID(r.ID).
		AddReadingCount(1).
		Exec(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

// Repo ----------------------------------------------------------------------------------------------------------------

func (rp *ReadingRP) QueryBlog(ctx context.Context, in *app1V1.QueryBlogReadingRequest) (*app1V1.QueryBlogReadingResponse, error) {
	resp, err := rp.store.connectClientService1.QueryBlog(ctx, &app1V1.QueryBlogRequest{
		Page:         in.Page,
		Id:           in.Id,
		Title:        in.Title,
		PublishedAtA: in.PublishedAtA,
		PublishedAtZ: in.PublishedAtZ,
		OrderBy:      in.OrderBy,
	})
	if err != nil {
		return nil, errors.FromError(err)
	}
	list := make([]*app1V1.BlogWithReading, 0, len(resp.List))
	for _, blog := range resp.List {
		list = append(list, &app1V1.BlogWithReading{
			Id:          blog.Id,
			CreatedAt:   blog.CreatedAt,
			UpdatedAt:   blog.UpdatedAt,
			Title:       blog.Title,
			Content:     blog.Content,
			PublishedAt: blog.PublishedAt,
		})
	}
	return &app1V1.QueryBlogReadingResponse{
		Page: resp.Page,
		List: list,
	}, nil
}

func (rp *ReadingRP) GetBlog(ctx context.Context, in *app1V1.GetBlogReadingRequest) (*app1V1.GetBlogReadingResponse, error) {
	blog, err := rp.store.grpcClientService1.GetBlog(ctx, &app1V1.GetBlogRequest{Id: in.Id})
	if err != nil {
		return nil, errors.FromError(err)
	}
	return &app1V1.GetBlogReadingResponse{
		Id:          blog.Id,
		CreatedAt:   blog.CreatedAt,
		UpdatedAt:   blog.UpdatedAt,
		Title:       blog.Title,
		Content:     blog.Content,
		PublishedAt: blog.PublishedAt,
	}, nil
}
