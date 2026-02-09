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
	txBeginner database.QueryTxBeginner
	logger     logging.Logger
}

func NewUsersRepository(txBeginner database.QueryTxBeginner, logger logging.Logger) *UsersRepository {
	return &UsersRepository{
		txBeginner: txBeginner,
		logger:     logger,
	}
}

func (r *UsersRepository) CreateUser(ctx context.Context, username, hashedPassword string, startBalance uint32) (domain.UserInfo, error) {
	tx, err := r.txBeginner.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	defer func() {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			r.logger.Error("failed to rollback user creation transaction", "error", err)
		}
	}()

	userInfo, err := createNewUser(ctx, tx, username, hashedPassword)
	if err != nil {
		return domain.UserInfo{}, err
	}

	err = createUserBalance(ctx, tx, userInfo.ID, startBalance)
	if err != nil {
		return domain.UserInfo{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return domain.UserInfo{}, err
	}

	return userInfo, nil
}

func (r *UsersRepository) TryGetUserInfo(ctx context.Context, username string) (domain.UserInfo, bool, error) {
	var userInfo domain.UserInfo
	querySQL := `SELECT id, username, password_hash FROM users WHERE username = $1`

	row := r.txBeginner.QueryRow(ctx, querySQL, username)
	err := row.Scan(&userInfo.ID, &userInfo.Username, &userInfo.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.UserInfo{}, false, nil
		}

		return domain.UserInfo{}, false, err
	}

	return userInfo, true, nil
}

func createNewUser(ctx context.Context, querier database.Querier, username, hashedPassword string) (domain.UserInfo, error) {
	creationSQL := `INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id, username, password_hash`

	var userInfo domain.UserInfo
	row := querier.QueryRow(ctx, creationSQL, username, hashedPassword)
	err := row.Scan(&userInfo.ID, &userInfo.Username, &userInfo.PasswordHash)
	if err != nil {
		return domain.UserInfo{}, err
	}

	return userInfo, nil
}

func createUserBalance(ctx context.Context, querier database.QueryExecuter, userID int, balance uint32) error {
	creationSQL := `INSERT INTO balances (user_id, balance) VALUES ($1, $2)`

	_, err := querier.Exec(ctx, creationSQL, userID, balance)
	return err
}
