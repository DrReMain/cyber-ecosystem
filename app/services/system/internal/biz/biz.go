package biz

import (
	"context"
	"log/slog"

	"github.com/google/wire"
)

type Transaction interface {
	InTx(ctx context.Context, fn func(context.Context) error) error
}

type UC struct {
	log *slog.Logger
	tm  Transaction
}

var ProviderSet = wire.NewSet(
	NewAuthUC,
	NewItemUC,
	NewResourceUC,
	NewUserUC,
)
