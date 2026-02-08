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
	Balance  uint32
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

func (pct *CoinsTransferer) TransferCoins(ctx context.Context, fromUsername string, toUsername string, amount uint32) error {
	if fromUsername == toUsername {
		return &domain.InvalidArgumentsError{Msg: "fromUsername must differ from toUsername"}
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

	users, err := getTargetUsers(ctx, tx, fromUsername, toUsername)
	if err != nil {
		return err
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

	err = proceedTransaction(ctx, tx, amount, fromUser, toUser)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func proceedTransaction(ctx context.Context, executor database.Executor, amount uint32, fromUser, toUser *userInfo) error {
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

func getTargetUsers(ctx context.Context, querier database.Querier, fromUsername, toUsername string) ([]userInfo, error) {
	usersSelectSQL := `SELECT u.id, u.username, b.balance
FROM users u
JOIN balances b ON b.user_id = u.id
WHERE u.username = ANY($1)
ORDER BY u.id
FOR UPDATE`
	rows, err := querier.Query(ctx, usersSelectSQL, []string{fromUsername, toUsername})
	if err != nil {
		return nil, fmt.Errorf("failed to select users for update: %w", err)
	}

	users := make([]userInfo, 0, 2)
	for rows.Next() {
		var user userInfo
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
