package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// Model ----------------------------------------------------------------------------------------------------------------

type AuthorEntity struct {
	ID        string
	CreatedAt *time.Time
	UpdatedAt *time.Time
	Name      *string
}

// Port -----------------------------------------------------------------------------------------------------------------

type AuthorRP interface {
	Create(context.Context, *AuthorEntity) error
	Get(context.Context, string) (*AuthorEntity, error)
}

// UC -------------------------------------------------------------------------------------------------------------------

type AuthorUC struct {
	UC
	authorRP AuthorRP
}

func NewAuthorUC(logger log.Logger, tm Transaction, authorRP AuthorRP) *AuthorUC {
	return &AuthorUC{
		UC: UC{
			log: log.NewHelper(log.With(logger, "module", "biz/uc_author")),
			tm:  tm,
		},
		authorRP: authorRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *AuthorUC) CreateAuthor(ctx context.Context, entity *AuthorEntity) error {
	return uc.authorRP.Create(ctx, entity)
}

func (uc *AuthorUC) GetAuthor(ctx context.Context, id string) (*AuthorEntity, error) {
	return uc.authorRP.Get(ctx, id)
}
