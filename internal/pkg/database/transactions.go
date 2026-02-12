package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"github.com/jackc/pgx/v5"
)

type TxManager interface {
	WithinTransaction(ctx context.Context, txFn TxFunc) error
}

type TxFunc func(ctx context.Context, executor QueryExecuter) error

type DelegateTxManager struct {
	txBeginner TxBeginner
	logger     logging.Logger
}

func NewDelegateTxManager(txBeginner TxBeginner, logger logging.Logger) *DelegateTxManager {
	return &DelegateTxManager{
		txBeginner: txBeginner,
		logger:     logger,
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
		rollErr := tx.Rollback(ctx)
		if rollErr != nil && !errors.Is(rollErr, pgx.ErrTxClosed) {
			tm.logger.Error("failed to rollback transaction", "error", rollErr)
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
