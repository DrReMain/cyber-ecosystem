package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/app_1/services/service_1/internal/biz"
)

type authorRP struct {
	log *log.Helper

	data *Data
}

func NewAuthorRP(logger log.Logger, data *Data) biz.AuthorRP {
	return &authorRP{
		log:  log.NewHelper(log.With(logger, "module", "data/author")),
		data: data,
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func (rp *authorRP) Create(ctx context.Context, entity *biz.AuthorEntity) error {
	client := rp.data.getClient(ctx)
	if err := client.Author.Create().
		SetNillableName(entity.Name).
		Exec(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *authorRP) Get(ctx context.Context, id string) (*biz.AuthorEntity, error) {
	e, err := rp.data.getClient(ctx).Author.Get(ctx, id)
	if err != nil {
		return nil, HandleError(err)
	}
	res := &biz.AuthorEntity{
		ID:        e.ID,
		CreatedAt: &e.CreatedAt,
		UpdatedAt: &e.UpdatedAt,
		Name:      &e.Name,
	}
	return res, nil
}
