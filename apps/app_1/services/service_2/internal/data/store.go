package data

import (
	"context"

	"cyber-ecosystem/shared-go/cache"
	"cyber-ecosystem/shared-go/orm/ent/entutil"

	app1V1 "cyber-ecosystem/apps/app_1/gen/go/v1"
	app1V1connect "cyber-ecosystem/apps/app_1/gen/go/v1/app1V1connect"
	"cyber-ecosystem/apps/app_1/services/service_2/internal/data/ent"
)

type Store struct {
	cache                 *cache.Cache
	db                    *ent.Client
	grpcClientService1    app1V1.BlogServiceClient
	connectClientService1 app1V1connect.BlogServiceClient
}

func NewStore(
	cache *cache.Cache,
	db *ent.Client,
	grpcClientService1 app1V1.BlogServiceClient,
	connectClientService1 app1V1connect.BlogServiceClient,
) (*Store, func(), error) {
	store := &Store{
		cache:                 cache,
		db:                    db,
		grpcClientService1:    grpcClientService1,
		connectClientService1: connectClientService1,
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
