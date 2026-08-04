package data

import (
	"context"
	"log/slog"
	"time"

	"cyber-ecosystem/app/services/system/internal/biz"
	"cyber-ecosystem/app/services/system/internal/platform"
)

type tokenRP struct {
	RP
}

func NewTokenRP(logger *slog.Logger, p *platform.Platform) biz.TokenRP {
	return &tokenRP{
		RP: RP{
			log:      logger.With("module", "data/token_rp"),
			platform: p,
		},
	}
}

// Repo ----------------------------------------------------------------------------------------------------------------

func (r *tokenRP) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {
	if err := r.platform.GetCache().KV.Set(ctx, key, val, ttl); err != nil {
		return r.platform.HandleCacheError(err)
	}
	return nil
}

func (r *tokenRP) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.platform.GetCache().KV.Get(ctx, key)
	if err != nil {
		return nil, r.platform.HandleCacheError(err)
	}
	return val, nil
}

func (r *tokenRP) Del(ctx context.Context, key string) error {
	if err := r.platform.GetCache().KV.Del(ctx, key); err != nil {
		return r.platform.HandleCacheError(err)
	}
	return nil
}
