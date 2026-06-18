package platform

import (
	"context"

	"github.com/google/wire"

	"cyber-ecosystem/shared-go/cache"
	"cyber-ecosystem/shared-go/orm/ent/entutil"

	"cyber-ecosystem/app/services/edge_mobile/internal/ent"
)

type EntErrorHandler func(error) error
type CacheErrorHandler func(error) error

type Platform struct {
	cache            *cache.Cache
	handleCacheError CacheErrorHandler
	db               *ent.Client
	handleEntError   EntErrorHandler
}

// NewPlatform assembles the platform resources behind a single facade. It does
// not own their lifecycle: redis and DB each return their own cleanup from
// NewCache / NewEntClient, and wire chains those cleanups for both graceful
// shutdown and partial injection failure.
func NewPlatform(
	cache *cache.Cache,
	handleCacheError CacheErrorHandler,
	db *ent.Client,
	handleEntError EntErrorHandler,
) (*Platform, error) {
	return &Platform{
		cache:            cache,
		handleCacheError: handleCacheError,
		db:               db,
		handleEntError:   handleEntError,
	}, nil
}

func (p *Platform) InTx(ctx context.Context, fn func(context.Context) error) error {
	return entutil.InTx(ctx, ent.TxFromContext, ent.NewTxContext, p.db.Tx, fn)
}

func (p *Platform) GetClient(ctx context.Context) *ent.Client {
	return entutil.GetClientFromTx(ctx, ent.TxFromContext, func(tx *ent.Tx) *ent.Client { return tx.Client() }, p.db)
}
func (p *Platform) HandleEntError(err error) error { return p.handleEntError(err) }

func (p *Platform) GetCache() *cache.Cache           { return p.cache }
func (p *Platform) HandleCacheError(err error) error { return p.handleCacheError(err) }

var ProviderSet = wire.NewSet(
	NewPlatform,
	NewCache,
	NewCacheErrorHandler,
	NewEntClient,
	NewEntErrorHandler,
)
