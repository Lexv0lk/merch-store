package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type TxManager interface {
	WithinTransaction(ctx context.Context, txFn TxFunc) error
}

type TxFunc func(ctx context.Context, executor QueryExecuter) error

type DelegateTxManager struct {
	txBeginner TxBeginner
}

func NewDelegateTxManager(txBeginner TxBeginner) *DelegateTxManager {
	return &DelegateTxManager{
		txBeginner: txBeginner,
	}
}

func (tm *DelegateTxManager) WithinTransaction(ctx context.Context, txFn TxFunc) error {
	tx, err := tm.txBeginner.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		err := tx.Rollback(ctx)
		if err != nil && err != pgx.ErrTxClosed {
			fmt.Printf("failed to rollback transaction: %v\n", err)
		}
	}()

	err = txFn(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to execute logic within transaction: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
