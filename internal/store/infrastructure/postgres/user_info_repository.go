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

type transaction struct {
	targetUsername string
	fromUsername   string
	toUsername     string
	amount         int
}

type UserInfoRepository struct {
	queryExecuter database.QueryExecuter
	logger        logging.Logger
}

func NewUserInfoRepository(queryExecuter database.QueryExecuter, logger logging.Logger) *UserInfoRepository {
	return &UserInfoRepository{
		queryExecuter: queryExecuter,
		logger:        logger,
	}
}

func (uif *UserInfoRepository) FetchUsername(ctx context.Context, userId int) (string, error) {
	sql := `SELECT username FROM users WHERE id = $1`
	var username string
	err := uif.queryExecuter.QueryRow(ctx, sql, userId).Scan(&username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", &domain.UserNotFoundError{Msg: fmt.Sprintf("user with id %d not found", userId)}
		}

		return "", err
	}

	return username, nil
}

func (uif *UserInfoRepository) FetchUserBalance(ctx context.Context, userId int) (uint32, error) {
	sql := `SELECT balance FROM balances WHERE user_id = $1`
	var balance uint32

	err := uif.queryExecuter.QueryRow(ctx, sql, userId).Scan(&balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, &domain.BalanceNotFoundError{Msg: fmt.Sprintf("balance for user with id %d not found", userId)}
		}

		return 0, err
	}

	return balance, nil
}

func (uif *UserInfoRepository) FetchUserPurchases(ctx context.Context, userId int) (map[domain.Good]uint32, error) {
	defer func() {
		if r := recover(); r != nil {
			uif.logger.Error("panic recovered in FetchUserPurchases", "error", r)
		}
	}()

	sql := `SELECT g.name, COUNT(*) FROM purchases p
			JOIN goods g ON p.good_id = g.id
			WHERE p.user_id = $1
			GROUP BY g.name`
	rows, err := uif.queryExecuter.Query(ctx, sql, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	goods := make(map[domain.Good]uint32)
	for rows.Next() {
		var good domain.Good
		var count int
		if err := rows.Scan(&good.Name, &count); err != nil {
			return nil, err
		}

		goods[good] = uint32(count)
	}

	return goods, nil
}

func (uif *UserInfoRepository) FetchUserCoinTransfers(ctx context.Context, userId int) (domain.CoinTransferHistory, error) {
	defer func() {
		if r := recover(); r != nil {
			uif.logger.Error("panic recovered in FetchUserCoinTransfers", "error", r)
		}
	}()

	sql := `
SELECT u_from.username, u_from.username, u_to.username, amount FROM transactions
JOIN users u_from ON transactions.from_user_id = u_from.id
JOIN users u_to  ON transactions.to_user_id = u_to.id
WHERE from_user_id = $1

UNION ALL

SELECT u_to.username, u_from.username, u_to.username, amount FROM transactions
JOIN users u_from ON transactions.from_user_id = u_from.id
JOIN users u_to  ON transactions.to_user_id = u_to.id
WHERE to_user_id = $1;
`
	rows, err := uif.queryExecuter.Query(ctx, sql, userId)
	if err != nil {
		return domain.CoinTransferHistory{}, err
	}
	defer rows.Close()

	transferHistory := domain.CoinTransferHistory{
		IncomingTransfers:  make([]domain.DirectTransfer, 0),
		OutcomingTransfers: make([]domain.DirectTransfer, 0),
	}

	for rows.Next() {
		var transfer transaction
		if err := rows.Scan(&transfer.targetUsername, &transfer.fromUsername, &transfer.toUsername, &transfer.amount); err != nil {
			return domain.CoinTransferHistory{}, err
		}

		if transfer.toUsername == transfer.targetUsername {
			transferHistory.IncomingTransfers = append(transferHistory.IncomingTransfers, domain.DirectTransfer{
				TargetName: transfer.fromUsername,
				Amount:     uint32(transfer.amount),
			})
		} else {
			transferHistory.OutcomingTransfers = append(transferHistory.OutcomingTransfers, domain.DirectTransfer{
				TargetName: transfer.toUsername,
				Amount:     uint32(transfer.amount),
			})
		}
	}

	return transferHistory, nil
}
