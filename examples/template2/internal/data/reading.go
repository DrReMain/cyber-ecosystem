package data

import (
	"context"

	"github.com/DrReMain/cyber-ecosystem/examples/template2/internal/biz"
	"github.com/DrReMain/cyber-ecosystem/examples/template2/internal/data/ent"
	"github.com/DrReMain/cyber-ecosystem/examples/template2/internal/data/ent/reading"

	template1V1 "github.com/DrReMain/cyber-ecosystem/gen/go/template1/v1"
)

type readingRP struct {
	data *Data
}

func NewReadingRP(data *Data) biz.ReadingRP {
	return &readingRP{data: data}
}

func (rp *readingRP) GetReadingCount(ctx context.Context, blogID string) (int64, error) {
	r, err := rp.data.getClient(ctx).Reading.Query().
		Where(reading.BlogIDEQ(blogID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return 0, nil
		}
		return 0, HandleError(err)
	}
	return r.ReadingCount, nil
}

func (rp *readingRP) GetReadingCounts(ctx context.Context, blogIDs []string) (map[string]int64, error) {
	if len(blogIDs) == 0 {
		return make(map[string]int64), nil
	}

	readings, err := rp.data.getClient(ctx).Reading.Query().
		Where(reading.BlogIDIn(blogIDs...)).
		All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}

	result := make(map[string]int64, len(readings))
	for _, r := range readings {
		result[r.BlogID] = r.ReadingCount
	}
	return result, nil
}

func (rp *readingRP) IncrementReading(ctx context.Context, blogID string) error {
	client := rp.data.getClient(ctx)

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
		SetReadingCount(r.ReadingCount + 1).
		Exec(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

// ---------------------------------------------------------------------------------------------------------------------

func (rp *readingRP) QueryBlog(ctx context.Context, in *template1V1.QueryBlogRequest) (*template1V1.QueryBlogResponse, error) {
	return rp.data.template1BlogService.QueryBlog(ctx, in)
}

func (rp *readingRP) GetBlog(ctx context.Context, in *template1V1.GetBlogRequest) (*template1V1.GetBlogResponse, error) {
	return rp.data.template1BlogService.GetBlog(ctx, in)
}
