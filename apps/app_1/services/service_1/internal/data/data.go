package data

import (
	"context"

	"github.com/google/wire"

	"cyber-ecosystem/shared-go/cache"
	"cyber-ecosystem/shared-go/orm/ent/entutil"

	"cyber-ecosystem/apps/app_1/services/service_1/internal/biz"
	"cyber-ecosystem/apps/app_1/services/service_1/internal/data/ent"
)

type Data struct {
	cache *cache.Cache
	db    *ent.Client
}

func NewData(db *ent.Client, cache *cache.Cache) (*Data, func(), error) {
	data := &Data{
		cache: cache,
		db:    db,
	}
	close := func() {
		if cache.Client != nil {
			_ = cache.Client.Close()
		}
		if db != nil {
			_ = db.Close()
		}
	}
	return data, close, nil
}

func (d *Data) getClient(ctx context.Context) *ent.Client {
	return entutil.GetClientFromTx(ctx, ent.TxFromContext, func(tx *ent.Tx) *ent.Client { return tx.Client() }, d.db)
}

func (d *Data) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return entutil.InTx(ctx, ent.TxFromContext, ent.NewTxContext, d.db.Tx, fn)
}

var ProviderSet = wire.NewSet(
	NewData,
	NewCache,
	NewEntClient,
	wire.Bind(new(biz.Transaction), new(*Data)),
	NewBlogRP,
	NewAuthorRP,
)
