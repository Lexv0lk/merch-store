package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	"github.com/Lexv0lk/merch-store/internal/store/domain"
	"github.com/jackc/pgx/v5"
)

type BalancesRepository struct {
	executor database.Executor
}

func NewBalancesRepository(executor database.Executor) *BalancesRepository {
	return &BalancesRepository{
		executor: executor,
	}
}

func (br *BalancesRepository) EnsureBalanceCreated(ctx context.Context, userId int, startValue uint32) error {
	sql := `INSERT INTO balances (user_id, balance) VALUES ($1, $2) ON CONFLICT (user_id) DO NOTHING`

	_, err := br.executor.Exec(ctx, sql, userId, startValue)
	return err
}

func (br *BalancesRepository) LockAndGetUserBalance(ctx context.Context, querier database.Querier, userId int) (uint32, error) {
	lockUserSQL := `SELECT balance FROM balances WHERE user_id = $1 FOR UPDATE`

	var balance uint32
	err := querier.QueryRow(ctx, lockUserSQL, userId).Scan(&balance)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, &domain.UserNotFoundError{Msg: fmt.Sprintf("user with id %d not found", userId)}
		}

		return 0, fmt.Errorf("failed to lock user row: %w", err)
	}

	return balance, nil
}
