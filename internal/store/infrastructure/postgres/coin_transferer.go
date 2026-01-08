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

type userInfo struct {
	Id       int
	Username string
	Balance  int
}

type CoinsTransferer struct {
	txBeginner database.TxBeginner
	logger     logging.Logger
}

func NewCoinsTransferer(txBeginner database.TxBeginner, logger logging.Logger) *CoinsTransferer {
	return &CoinsTransferer{
		txBeginner: txBeginner,
		logger:     logger,
	}
}

func (pct *CoinsTransferer) TransferCoins(ctx context.Context, fromUsername string, toUsername string, amount int) error {
	if fromUsername == toUsername {
		return &domain.InvalidArgumentsError{Msg: "fromUsername must differ from toUsername"}
	} else if amount <= 0 {
		return &domain.InvalidArgumentsError{Msg: "amount must be positive"}
	}

	tx, err := pct.txBeginner.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		err = tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			pct.logger.Error("failed to rollback transaction", "error", err)
		}
	}()

	usersSelectSQL := `SELECT id, username, balance FROM users WHERE username = ANY($1) ORDER BY id FOR UPDATE`
	rows, err := tx.Query(ctx, usersSelectSQL, []string{fromUsername, toUsername})
	if err != nil {
		return fmt.Errorf("failed to select users for update: %w", err)
	}

	users := make([]userInfo, 0, 2)
	for rows.Next() {
		var user userInfo
		err = rows.Scan(&user.Id, &user.Username, &user.Balance)
		if err != nil {
			return fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, user)
	}
	rows.Close()

	if len(users) != 2 {
		return &domain.UserNotFoundError{}
	}

	var fromUser, toUser *userInfo
	if users[0].Username == fromUsername {
		fromUser = &users[0]
		toUser = &users[1]
	} else {
		fromUser = &users[1]
		toUser = &users[0]
	}

	if fromUser.Balance < amount {
		return &domain.InsufficientBalanceError{}
	}

	updateBalanceSQL := `UPDATE users SET balance = balance - $1 WHERE id = $2 AND balance >= $1`
	tag, err := tx.Exec(ctx, updateBalanceSQL, amount, fromUser.Id)
	if err != nil {
		return fmt.Errorf("failed to update balance for fromUser: %w", err)
	} else if tag.RowsAffected() == 0 {
		return &domain.InsufficientBalanceError{}
	}

	updateBalanceSQL = `UPDATE users SET balance = balance + $1 WHERE id = $2`
	_, err = tx.Exec(ctx, updateBalanceSQL, amount, toUser.Id)
	if err != nil {
		return fmt.Errorf("failed to update balance for toUser: %w", err)
	}

	insertTransactionSQL := `INSERT INTO transactions (from_user_id, to_user_id, amount) VALUES ($1, $2, $3)`
	_, err = tx.Exec(ctx, insertTransactionSQL, fromUser.Id, toUser.Id, amount)
	if err != nil {
		return fmt.Errorf("failed to insert transaction record: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
