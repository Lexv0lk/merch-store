package postgres

import (
	"context"
	"fmt"

	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	"github.com/Lexv0lk/merch-store/internal/store/domain"
)

type TransactionProceeder struct{}

func NewTransactionProceeder() *TransactionProceeder {
	return &TransactionProceeder{}
}

func (tp *TransactionProceeder) ProceedTransaction(ctx context.Context, executor database.Executor, amount uint32, fromUser, toUser *domain.UserInfo) error {
	updateBalanceSQL := `UPDATE balances SET balance = balance - $1 WHERE user_id = $2 AND balance >= $1`
	tag, err := executor.Exec(ctx, updateBalanceSQL, amount, fromUser.Id)
	if err != nil {
		return fmt.Errorf("failed to update balance for fromUser: %w", err)
	} else if tag.RowsAffected() == 0 {
		return &domain.InsufficientBalanceError{}
	}

	updateBalanceSQL = `UPDATE balances SET balance = balance + $1 WHERE user_id = $2`
	_, err = executor.Exec(ctx, updateBalanceSQL, amount, toUser.Id)
	if err != nil {
		return fmt.Errorf("failed to update balance for toUser: %w", err)
	}

	insertTransactionSQL := `INSERT INTO transactions (from_user_id, to_user_id, amount) VALUES ($1, $2, $3)`
	_, err = executor.Exec(ctx, insertTransactionSQL, fromUser.Id, toUser.Id, amount)
	if err != nil {
		return fmt.Errorf("failed to insert transaction record: %w", err)
	}

	return nil
}
