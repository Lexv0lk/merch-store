package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"github.com/Lexv0lk/merch-store/internal/store/domain"
	"github.com/jackc/pgx/v5"
)

type PurchaseHandler struct {
	queryTxBeginner database.QueryTxBeginner
	logger          logging.Logger
}

func NewPurchaseHandler(queryTxBeginner database.QueryTxBeginner, logger logging.Logger) *PurchaseHandler {
	return &PurchaseHandler{
		queryTxBeginner: queryTxBeginner,
		logger:          logger,
	}
}

func (ph *PurchaseHandler) HandlePurchase(ctx context.Context, userId int, goodName string) error {
	good, err := GetGoodInfo(ctx, ph.queryTxBeginner, goodName)
	if err != nil {
		return err
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

	balance, err := GetAndLockUserBalance(ctx, tx, userId)
	if err != nil {
		return err
	}

	if balance < good.Price {
		return &domain.InsufficientBalanceError{Msg: "insufficient balance"}
	}

	err = ProcessPurchase(ctx, tx, userId, good)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func ProcessPurchase(ctx context.Context, executor database.Executor, userId int, good domain.GoodInfo) error {
	updateBalanceSQL := `UPDATE balances SET balance = balance - $1 WHERE user_id = $2`
	_, err := executor.Exec(ctx, updateBalanceSQL, good.Price, userId)
	if err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	insertPurchaseSQL := `INSERT INTO purchases (user_id, good_id) VALUES ($1, $2)`
	_, err = executor.Exec(ctx, insertPurchaseSQL, userId, good.Id)
	if err != nil {
		return fmt.Errorf("failed to insert purchase record: %w", err)
	}

	return nil
}

func GetAndLockUserBalance(ctx context.Context, querier database.Querier, userId int) (int, error) {
	lockUserSQL := `SELECT balance FROM balances WHERE user_id = $1 FOR UPDATE`

	var balance int
	err := querier.QueryRow(ctx, lockUserSQL, userId).Scan(&balance)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, &domain.UserNotFoundError{Msg: fmt.Sprintf("user with id %d not found", userId)}
		}

		return 0, fmt.Errorf("failed to lock user row: %w", err)
	}

	return balance, nil
}

func GetGoodInfo(ctx context.Context, querier database.Querier, name string) (domain.GoodInfo, error) {
	findGoodSQL := `SELECT id, name, price FROM goods WHERE name = $1`

	var good domain.GoodInfo
	err := querier.QueryRow(ctx, findGoodSQL, name).Scan(&good.Id, &good.Name, &good.Price)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.GoodInfo{}, &domain.GoodNotFoundError{Msg: fmt.Sprintf("good %s not found", name)}
		}

		return domain.GoodInfo{}, fmt.Errorf("failed to find good: %w", err)
	}

	return good, nil
}
