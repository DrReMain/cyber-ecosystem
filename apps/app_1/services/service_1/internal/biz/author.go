package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type AuthorEntity struct {
	ID        string
	CreatedAt *time.Time
	UpdatedAt *time.Time
	Name      *string
}

type AuthorRP interface {
	Create(context.Context, *AuthorEntity) error
	Get(context.Context, string) (*AuthorEntity, error)
}

type AuthorUC struct {
	log *log.Helper
	tm  Transaction

	authorRP AuthorRP
}

func NewAuthorUC(logger log.Logger, tm Transaction, authorRP AuthorRP) *AuthorUC {
	return &AuthorUC{
		log:      log.NewHelper(log.With(logger, "module", "biz/author")),
		tm:       tm,
		authorRP: authorRP,
	}
}

// ---------------------------------------------------------------------------------------------------------------------

func (uc *AuthorUC) CreateAuthor(ctx context.Context, entity *AuthorEntity) error {
	return uc.authorRP.Create(ctx, entity)
}

func (uc *AuthorUC) GetAuthor(ctx context.Context, id string) (*AuthorEntity, error) {
	return uc.authorRP.Get(ctx, id)
}
