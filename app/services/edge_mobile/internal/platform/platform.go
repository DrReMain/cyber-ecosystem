package platform

import (
	"context"
	"log/slog"

	"github.com/google/wire"
)

type Platform struct {
	// TODO gRPC clients, cache, error handlers
}

func NewPlatform(logger *slog.Logger) (*Platform, func(), error) {
	p := &Platform{}
	return p, func() {}, nil
}

func (p *Platform) InTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

var ProviderSet = wire.NewSet(
	NewPlatform,
)
