package postgres

import (
	"context"
	"errors"

	"github.com/Lexv0lk/merch-store/internal/auth/domain"
	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	"github.com/jackc/pgx/v5"
)

type UsersRepository struct {
	queryExecuter database.QueryExecuter
}

func NewUsersRepository(queryExecuter database.QueryExecuter) *UsersRepository {
	return &UsersRepository{
		queryExecuter: queryExecuter,
	}
}

func (r *UsersRepository) CreateUser(ctx context.Context, username, hashedPassword string) (domain.UserInfo, error) {
	creationSQL := `INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id, username, password_hash`

	var userInfo domain.UserInfo
	row := r.queryExecuter.QueryRow(ctx, creationSQL, username, hashedPassword)
	err := row.Scan(&userInfo.ID, &userInfo.Username, &userInfo.PasswordHash)
	if err != nil {
		return domain.UserInfo{}, err
	}

	return userInfo, nil
}

func (r *UsersRepository) TryGetUserInfo(ctx context.Context, username string) (domain.UserInfo, bool, error) {
	var userInfo domain.UserInfo
	querySQL := `SELECT id, username, password_hash FROM users WHERE username = $1`

	row := r.queryExecuter.QueryRow(ctx, querySQL, username)
	err := row.Scan(&userInfo.ID, &userInfo.Username, &userInfo.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.UserInfo{}, false, nil
		}

		return domain.UserInfo{}, false, err
	}

	return userInfo, true, nil
}
