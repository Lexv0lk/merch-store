package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	"github.com/Lexv0lk/merch-store/internal/store/domain"
	"github.com/jackc/pgx/v5"
)

type BalanceLocker struct {
}

func NewBalanceLocker() *BalanceLocker {
	return &BalanceLocker{}
}

func (bl *BalanceLocker) LockAndGetUserBalance(ctx context.Context, querier database.Querier, userId int) (int, error) {
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
