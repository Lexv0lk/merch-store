package postgres

import (
	"context"
	"errors"

	"github.com/Lexv0lk/merch-store/internal/auth/domain"
	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"github.com/jackc/pgx/v5"
)

type UsersRepository struct {
	querier database.Querier
	logger  logging.Logger
}

func NewUsersRepository(querier database.Querier, logger logging.Logger) *UsersRepository {
	return &UsersRepository{
		querier: querier,
		logger:  logger,
	}
}

func (r *UsersRepository) CreateUser(ctx context.Context, username, hashedPassword string) (domain.UserInfo, error) {
	creationSQL := `INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id, username, password_hash`

	var userInfo domain.UserInfo
	row := r.querier.QueryRow(ctx, creationSQL, username, hashedPassword)
	err := row.Scan(&userInfo.ID, &userInfo.Username, &userInfo.PasswordHash)
	if err != nil {
		return domain.UserInfo{}, err
	}

	return userInfo, nil
}

func (r *UsersRepository) TryGetUserInfo(ctx context.Context, username string) (domain.UserInfo, bool, error) {
	var userInfo domain.UserInfo
	querySQL := `SELECT id, username, password_hash FROM users WHERE username = $1`

	row := r.querier.QueryRow(ctx, querySQL, username)
	err := row.Scan(&userInfo.ID, &userInfo.Username, &userInfo.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.UserInfo{}, false, nil
		}

		return domain.UserInfo{}, false, err
	}

	return userInfo, true, nil
}

func (r *UsersRepository) GetUserID(ctx context.Context, username string) (int, error) {
	var userID int
	querySQL := `SELECT id FROM users WHERE username = $1`

	row := r.querier.QueryRow(ctx, querySQL, username)
	err := row.Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, &domain.UserNotFoundError{}
		}

		return 0, err
	}

	return userID, nil
}

func (r *UsersRepository) GetUsernames(ctx context.Context, userIDs []int) (map[int]string, error) {
	querySQL := `SELECT id, username FROM users WHERE id = ANY($1)`

	rows, err := r.querier.Query(ctx, querySQL, userIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	usernames := make(map[int]string)

	for rows.Next() {
		var id int
		var username string

		if err := rows.Scan(&id, &username); err != nil {
			return nil, err
		}

		usernames[id] = username
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return usernames, nil
}
