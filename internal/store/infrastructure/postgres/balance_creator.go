package postgres

import (
	"context"

	"github.com/Lexv0lk/merch-store/internal/pkg/database"
)

type BalanceCreator struct {
	executor database.Executor
}

func NewBalanceCreator(executor database.Executor) *BalanceCreator {
	return &BalanceCreator{
		executor: executor,
	}
}

func (bc *BalanceCreator) EnsureBalanceCreated(ctx context.Context, userId int, startValue uint32) error {
	sql := `INSERT INTO balances (user_id, balance) VALUES ($1, $2) ON CONFLICT (user_id) DO NOTHING`

	_, err := bc.executor.Exec(ctx, sql, userId, startValue)
	return err
}
