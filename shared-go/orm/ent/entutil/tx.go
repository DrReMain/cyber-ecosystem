package entutil

import (
	"context"
	"fmt"
)

type TxCommitter interface {
	Commit() error
	Rollback() error
}

func InTx[Tx interface {
	TxCommitter
	comparable
}](
	ctx context.Context,
	txFromCtx func(ctx context.Context) Tx,
	newTxCtx func(ctx context.Context, tx Tx) context.Context,
	startTx func(ctx context.Context) (Tx, error),
	fn func(context.Context) error,
) error {
	var zero Tx
	if tx := txFromCtx(ctx); tx != zero {
		return fn(ctx)
	}

	tx, err := startTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start a transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()

	txCtx := newTxCtx(ctx, tx)

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx failed: %w, rollback also failed: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}

func GetClientFromTx[Tx comparable, Client any](
	ctx context.Context,
	txFromCtx func(ctx context.Context) Tx,
	getClientFromTx func(tx Tx) Client,
	baseClient Client,
) Client {
	var zero Tx
	if tx := txFromCtx(ctx); tx != zero {
		return getClientFromTx(tx)
	}
	return baseClient
}
