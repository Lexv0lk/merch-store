package postgres

import (
	"context"
	"fmt"

	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	"github.com/jackc/pgx/v5"
)

type TxFunc func(ctx context.Context, executor database.QueryExecuter) error

type TxManager struct {
	txBeginner database.TxBeginner
}

func (tm *TxManager) WithinTransaction(ctx context.Context, txFn TxFunc) error {
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
