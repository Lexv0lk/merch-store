package postgres

import (
	"context"

	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"github.com/Lexv0lk/merch-store/internal/store/domain"
)

type transaction struct {
	targetUsername string
	fromUsername   string
	toUsername     string
	amount         int
}

type UserInfoFetcher struct {
	querier database.Querier
	logger  logging.Logger
}

func NewUserInfoFetcher(querier database.Querier, logger logging.Logger) *UserInfoFetcher {
	return &UserInfoFetcher{
		querier: querier,
		logger:  logger,
	}
}

func (uif *UserInfoFetcher) FetchMainUserInfo(ctx context.Context, userId int) (domain.MainUserInfo, error) {
	sql := `SELECT username, balance FROM users WHERE id = $1`
	var userInfo domain.MainUserInfo
	err := uif.querier.QueryRow(ctx, sql, userId).Scan(&userInfo.Username, &userInfo.Balance)
	if err != nil {
		return domain.MainUserInfo{}, err
	}

	return userInfo, nil
}

func (uif *UserInfoFetcher) FetchUserPurchases(ctx context.Context, userId int) (map[domain.Good]int, error) {
	sql := `SELECT g.name, COUNT(*) FROM purchases p
			JOIN goods g ON p.good_id = g.id
			WHERE p.user_id = $1
			GROUP BY g.name`
	rows, err := uif.querier.Query(ctx, sql, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	goods := make(map[domain.Good]int)
	for rows.Next() {
		var good domain.Good
		var count int
		if err := rows.Scan(&good.Name, &count); err != nil {
			return nil, err
		}

		goods[good] = count
	}

	return goods, nil
}

func (uif *UserInfoFetcher) FetchUserCoinTransfers(ctx context.Context, userId int) (domain.CoinTransferHistory, error) {
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
	rows, err := uif.querier.Query(ctx, sql, userId)
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
				Amount:     transfer.amount,
			})
		} else {
			transferHistory.OutcomingTransfers = append(transferHistory.OutcomingTransfers, domain.DirectTransfer{
				TargetName: transfer.toUsername,
				Amount:     transfer.amount,
			})
		}
	}

	return transferHistory, nil
}
