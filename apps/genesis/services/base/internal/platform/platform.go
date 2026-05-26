package platform

import (
	"context"

	"github.com/google/wire"

	"github.com/go-kratos/kratos/v2/log"

	"cyber-ecosystem/shared-go/cache"
	"cyber-ecosystem/shared-go/orm/ent/entutil"

	"cyber-ecosystem/apps/genesis/services/base/internal/ent"
)

type CacheErrorHandler func(error) error
type EntErrorHandler func(error) error

type Platform struct {
	cache            *cache.Cache
	handleCacheError CacheErrorHandler
	db               *ent.Client
	handleEntError   EntErrorHandler
}

func NewPlatform(
	logger log.Logger,
	cache *cache.Cache,
	handleCacheError CacheErrorHandler,
	db *ent.Client,
	handleEntError EntErrorHandler,
) (*Platform, func(), error) {
	helper := log.NewHelper(log.With(logger, "module", "platform/platform"))
	p := &Platform{
		cache:            cache,
		handleCacheError: handleCacheError,
		db:               db,
		handleEntError:   handleEntError,
	}
	return p,
		func() {
			if err := cache.Client.Close(); err != nil {
				helper.Warnf("failed to close cache client: %v", err)
			}
			if err := db.Close(); err != nil {
				helper.Warnf("failed to close database client: %v", err)
			}
		},
		nil
}

func (p *Platform) InTx(ctx context.Context, fn func(context.Context) error) error {
	return entutil.InTx(ctx, ent.TxFromContext, ent.NewTxContext, p.db.Tx, fn)
}

func (p *Platform) GetClient(ctx context.Context) *ent.Client {
	return entutil.GetClientFromTx(ctx, ent.TxFromContext, func(tx *ent.Tx) *ent.Client { return tx.Client() }, p.db)
}

func (p *Platform) HandleEntError(err error) error {
	return p.handleEntError(err)
}

func (p *Platform) GetCache() *cache.Cache {
	return p.cache
}

func (p *Platform) HandleCacheError(err error) error {
	return p.handleCacheError(err)
}

var ProviderSet = wire.NewSet(
	NewPlatform,
	NewCache,
	NewCacheErrorHandler,
	NewEntClient,
	NewEntErrorHandler,
)
