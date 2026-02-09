package postgres

import (
	"context"
	"fmt"

	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	"github.com/Lexv0lk/merch-store/internal/store/domain"
)

type PurchaseHandler struct{}

func NewPurchaseHandler() *PurchaseHandler {
	return &PurchaseHandler{}
}

func (ph *PurchaseHandler) ProcessPurchase(ctx context.Context, executor database.Executor, userId int, good domain.GoodInfo) error {
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
