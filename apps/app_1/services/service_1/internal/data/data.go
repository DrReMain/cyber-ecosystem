package data

import (
	"context"

	"github.com/google/wire"

	"cyber-ecosystem/shared-go/orm/ent/entutil"

	"cyber-ecosystem/apps/app_1/services/service_1/internal/biz"
	"cyber-ecosystem/apps/app_1/services/service_1/internal/conf"
	"cyber-ecosystem/apps/app_1/services/service_1/internal/data/ent"
)

type Data struct {
	db *ent.Client
}

func NewData(c *conf.Data, db *ent.Client) (*Data, func(), error) {
	data := &Data{
		db: db,
	}
	close := func() {
		db.Close()
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
	NewEntClient,
	wire.Bind(new(biz.Transaction), new(*Data)),
	NewBlogRP,
	NewAuthorRP,
)
