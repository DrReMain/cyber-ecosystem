package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/apps/app_1/services/service_1/internal/biz"
)

type AuthorRP struct {
	RP
}

func NewAuthorRP(logger log.Logger, store *Store) biz.AuthorRP {
	return &AuthorRP{
		RP: RP{
			log:   log.NewHelper(log.With(logger, "module", "data/rp_author")),
			store: store,
		},
	}
}

// Repo ----------------------------------------------------------------------------------------------------------------

func (rp *AuthorRP) Create(ctx context.Context, entity *biz.AuthorEntity) error {
	client := rp.store.GetClient(ctx)
	if err := client.Author.Create().
		SetNillableName(entity.Name).
		Exec(ctx); err != nil {
		return HandleError(err)
	}
	return nil
}

func (rp *AuthorRP) Get(ctx context.Context, id string) (*biz.AuthorEntity, error) {
	e, err := rp.store.GetClient(ctx).Author.Get(ctx, id)
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
