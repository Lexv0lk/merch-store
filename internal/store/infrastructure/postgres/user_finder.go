package postgres

import (
	"context"
	"fmt"

	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	"github.com/Lexv0lk/merch-store/internal/store/domain"
)

type UserFinder struct{}

func NewUserFinder() *UserFinder {
	return &UserFinder{}
}

func (uf *UserFinder) GetTargetUsers(ctx context.Context, querier database.Querier, fromUsername, toUsername string) ([]domain.UserInfo, error) {
	usersSelectSQL := `SELECT u.id, u.username, b.balance
FROM balances b
WHERE u.username = ANY($1)
ORDER BY u.id
FOR UPDATE`
	rows, err := querier.Query(ctx, usersSelectSQL, []string{fromUsername, toUsername})
	if err != nil {
		return nil, fmt.Errorf("failed to select users for update: %w", err)
	}

	users := make([]domain.UserInfo, 0, 2)
	for rows.Next() {
		var user domain.UserInfo
		err = rows.Scan(&user.Id, &user.Username, &user.Balance)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, user)
	}
	rows.Close()

	if len(users) != 2 {
		return nil, &domain.UserNotFoundError{}
	}

	return users, nil
}
