package platform

import (
	"context"

	"github.com/google/wire"

	"cyber-ecosystem/shared-go/cache"
	"cyber-ecosystem/shared-go/mq"
	"cyber-ecosystem/shared-go/orm/ent/entutil"
	"cyber-ecosystem/shared-go/storage"

	"cyber-ecosystem/app/services/system/internal/ent"
)

type EntErrorHandler func(error) error
type CacheErrorHandler func(error) error
type StorageErrorHandler func(error) error
type MQErrorHandler func(error) error

type Platform struct {
	cache              *cache.Cache
	handleCacheError   CacheErrorHandler
	db                 *ent.Client
	handleEntError     EntErrorHandler
	storage            *storage.Storage
	handleStorageError StorageErrorHandler
	mq                 *mq.MQ
	handleMQError      MQErrorHandler
}

// NewPlatform assembles the platform resources behind a single facade. It does
// not own their lifecycle: each resource provider returns its own cleanup, and
// wire chains those cleanups for both graceful shutdown and partial injection
// failure.
func NewPlatform(
	cache *cache.Cache,
	handleCacheError CacheErrorHandler,
	db *ent.Client,
	handleEntError EntErrorHandler,
	storage *storage.Storage,
	handleStorageError StorageErrorHandler,
	mq *mq.MQ,
	handleMQError MQErrorHandler,
) (*Platform, error) {
	return &Platform{
		cache:              cache,
		handleCacheError:   handleCacheError,
		db:                 db,
		handleEntError:     handleEntError,
		storage:            storage,
		handleStorageError: handleStorageError,
		mq:                 mq,
		handleMQError:      handleMQError,
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

func (p *Platform) GetStorage() *storage.Storage       { return p.storage }
func (p *Platform) HandleStorageError(err error) error { return p.handleStorageError(err) }

func (p *Platform) GetMQ() *mq.MQ                 { return p.mq }
func (p *Platform) HandleMQError(err error) error { return p.handleMQError(err) }

var ProviderSet = wire.NewSet(
	NewPlatform,
	NewCache,
	NewCacheErrorHandler,
	NewEntClient,
	NewEntErrorHandler,
	NewStorage,
	NewStorageErrorHandler,
	NewMQ,
	NewMQErrorHandler,
)
