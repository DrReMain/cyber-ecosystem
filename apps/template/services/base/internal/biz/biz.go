package biz

import (
	"context"

	"github.com/google/wire"

	"github.com/go-kratos/kratos/v2/log"
)

type Transaction interface {
	InTx(ctx context.Context, fn func(context.Context) error) error
}

type UC struct {
	log *log.Helper
	tm  Transaction
}

var ProviderSet = wire.NewSet(
	NewResourceUC,
	NewMessageUC,
)
