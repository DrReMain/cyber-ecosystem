package platform

import (
	"context"

	"github.com/google/wire"

	"cyber-ecosystem/shared-go/capability/payment"
	"cyber-ecosystem/shared-go/orm/ent/entutil"

	"cyber-ecosystem/app/services/esports/internal/ent"
)

type EntErrorHandler func(error) error
type PaymentErrorHandler func(error) error

type Platform struct {
	db             *ent.Client
	handleEntError EntErrorHandler
	payments       *Payments
	handlePayError PaymentErrorHandler
}

func NewPlatform(
	db *ent.Client,
	handleEntError EntErrorHandler,
	payments *Payments,
	handlePayError PaymentErrorHandler,
) (*Platform, error) {
	return &Platform{
		db:             db,
		handleEntError: handleEntError,
		payments:       payments,
		handlePayError: handlePayError,
	}, nil
}

func (p *Platform) InTx(ctx context.Context, fn func(context.Context) error) error {
	return entutil.InTx(ctx, ent.TxFromContext, ent.NewTxContext, p.db.Tx, fn)
}

func (p *Platform) GetClient(ctx context.Context) *ent.Client {
	return entutil.GetClientFromTx(ctx, ent.TxFromContext, func(tx *ent.Tx) *ent.Client { return tx.Client() }, p.db)
}

func (p *Platform) HandleEntError(err error) error { return p.handleEntError(err) }

func (p *Platform) PaymentFor(provider string) (*payment.Payment, error) {
	return p.payments.backend(provider)
}

func (p *Platform) Payments() *Payments { return p.payments }

func (p *Platform) HandlePaymentError(err error) error { return p.handlePayError(err) }

var ProviderSet = wire.NewSet(
	NewPlatform,
	NewEntClient,
	NewEntErrorHandler,
	NewPayments,
	NewPaymentErrorHandler,
)
