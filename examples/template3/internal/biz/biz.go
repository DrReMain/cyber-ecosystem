package biz

import (
	"context"

	"github.com/google/wire"
)

type Transaction interface {
	InTx(context.Context, func(context.Context) error) error
}

var ProviderSet = wire.NewSet(
	NewHelloUC,
)
