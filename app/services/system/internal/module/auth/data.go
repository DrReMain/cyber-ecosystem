package auth

import (
	"context"
	"log/slog"
	"time"

	"cyber-ecosystem/app/services/system/internal/platform"
	"cyber-ecosystem/app/services/system/internal/shared"
)

type tokenRP struct {
	shared.RP
}

func NewTokenRP(logger *slog.Logger, p *platform.Platform) TokenRP {
	return &tokenRP{
		RP: shared.NewRP(logger.With("module", "module/token_rp"), p),
	}
}

// Repo ----------------------------------------------------------------------------------------------------------------

func (r *tokenRP) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	if err := r.Platform.GetCache().KV.Set(ctx, key, val, ttl); err != nil {
		return r.Platform.HandleCacheError(err)
	}
	return nil
}

func (r *tokenRP) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.Platform.GetCache().KV.Get(ctx, key)
	if err != nil {
		return nil, r.Platform.HandleCacheError(err)
	}
	return val, nil
}

func (r *tokenRP) Del(ctx context.Context, key string) error {
	if err := r.Platform.GetCache().KV.Del(ctx, key); err != nil {
		return r.Platform.HandleCacheError(err)
	}
	return nil
}
