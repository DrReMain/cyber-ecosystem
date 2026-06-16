package platform

import (
	"context"
	"log/slog"

	"github.com/google/wire"

	"cyber-ecosystem/shared-go/orm/ent/entutil"

	"cyber-ecosystem/app/services/edge_mobile/internal/ent"
)

type EntErrorHandler func(error) error

type Platform struct {
	db             *ent.Client
	handleEntError EntErrorHandler
}

func NewPlatform(logger *slog.Logger, db *ent.Client, handleEntError EntErrorHandler) (*Platform, func(), error) {
	p := &Platform{
		db:             db,
		handleEntError: handleEntError,
	}
	return p, func() {
		if err := db.Close(); err != nil {
			logger.Warn("failed to close database client", "error", err)
		}
	}, nil
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

var ProviderSet = wire.NewSet(
	NewPlatform,
	NewEntClient,
	NewEntErrorHandler,
)
