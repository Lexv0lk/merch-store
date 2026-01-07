package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"github.com/Lexv0lk/merch-store/internal/store/domain"
	"github.com/jackc/pgx/v5"
)

type goodInfo struct {
	id    int
	name  string
	price int
}

type PurchaseHandler struct {
	queryTxBeginner domain.QueryTxBeginner
	logger          logging.Logger
}

func NewPurchaseHandler(queryTxBeginner domain.QueryTxBeginner, logger logging.Logger) *PurchaseHandler {
	return &PurchaseHandler{
		queryTxBeginner: queryTxBeginner,
		logger:          logger,
	}
}

func (ph *PurchaseHandler) HandlePurchase(ctx context.Context, userId int, goodName string) error {
	findGoodSQL := `SELECT id, name, price FROM goods WHERE name = $1`
	var good goodInfo
	err := ph.queryTxBeginner.QueryRow(ctx, findGoodSQL, goodName).Scan(&good.id, &good.name, &good.price)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &domain.GoodNotFoundError{Msg: fmt.Sprintf("good %s not found", goodName)}
		}

		return fmt.Errorf("failed to find good: %w", err)
	}

	tx, err := ph.queryTxBeginner.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return err
	}

	defer func() {
		err := tx.Rollback(ctx)
		if err != nil && err != pgx.ErrTxClosed {
			ph.logger.Error("failed to rollback transaction", "error", err)
		}
	}()

	lockUserSQL := `SELECT balance FROM users WHERE id = $1 FOR UPDATE`
	var balance int
	err = tx.QueryRow(ctx, lockUserSQL, userId).Scan(&balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &domain.UserNotFoundError{Msg: fmt.Sprintf("user with id %d not found", userId)}
		}

		return fmt.Errorf("failed to lock user row: %w", err)
	}

	if balance < good.price {
		return &domain.InsufficientBalanceError{Msg: "insufficient balance"}
	}

	updateBalanceSQL := `UPDATE users SET balance = balance - $1 WHERE id = $2`
	_, err = tx.Exec(ctx, updateBalanceSQL, good.price, userId)
	if err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	insertPurchaseSQL := `INSERT INTO purchases (user_id, good_id) VALUES ($1, $2)`
	_, err = tx.Exec(ctx, insertPurchaseSQL, userId, good.id)
	if err != nil {
		return fmt.Errorf("failed to insert purchase record: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
