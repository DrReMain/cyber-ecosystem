package data

import (
	"context"

	"cyber-ecosystem/shared-go/cache"
	"cyber-ecosystem/shared-go/orm/ent/entutil"

	"cyber-ecosystem/apps/app_1/services/service_1/internal/data/ent"
)

type Store struct {
	cache *cache.Cache
	db    *ent.Client
}

func NewStore(cache *cache.Cache, db *ent.Client) (*Store, func(), error) {
	store := &Store{
		cache: cache,
		db:    db,
	}
	return store,
		func() {
			if cache != nil && cache.Client != nil {
				_ = cache.Client.Close()
			}
			if db != nil {
				_ = db.Close()
			}
		},
		nil
}

func (s *Store) InTx(ctx context.Context, fn func(context.Context) error) error {
	return entutil.InTx(ctx, ent.TxFromContext, ent.NewTxContext, s.db.Tx, fn)
}

func (s *Store) GetClient(ctx context.Context) *ent.Client {
	return entutil.GetClientFromTx(ctx, ent.TxFromContext, func(tx *ent.Tx) *ent.Client { return tx.Client() }, s.db)
}

func (s *Store) GetCache() *cache.Cache {
	return s.cache
}
