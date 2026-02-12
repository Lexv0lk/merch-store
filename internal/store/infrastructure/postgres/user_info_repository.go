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
	fromUserID int
	toUserID   int
	amount     int
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

func (uif *UserInfoRepository) FetchUserCoinTransfers(ctx context.Context, userId int) (domain.TransferHistory, error) {
	fromUserSQL := `SELECT from_user_id, to_user_id, amount FROM transactions WHERE from_user_id = $1`
	outcomingRows, err := uif.queryExecuter.Query(ctx, fromUserSQL, userId)
	if err != nil {
		return domain.TransferHistory{}, err
	}
	defer outcomingRows.Close()

	toUserSQL := `SELECT from_user_id, to_user_id, amount FROM transactions WHERE to_user_id = $1`
	incomingRows, err := uif.queryExecuter.Query(ctx, toUserSQL, userId)
	if err != nil {
		return domain.TransferHistory{}, err
	}
	defer incomingRows.Close()

	transferHistory := domain.TransferHistory{}

	transferHistory.OutcomingTransfers, err = processRows(outcomingRows, func(tr transaction) int { return tr.toUserID })
	if err != nil {
		return domain.TransferHistory{}, err
	}

	transferHistory.IncomingTransfers, err = processRows(incomingRows, func(tr transaction) int { return tr.fromUserID })
	if err != nil {
		return domain.TransferHistory{}, err
	}

	return transferHistory, nil
}

func processRows(rows pgx.Rows, getTargetIDFn func(tr transaction) int) ([]domain.DirectTransfer, error) {
	result := make([]domain.DirectTransfer, 0)

	for rows.Next() {
		var transfer transaction
		if err := rows.Scan(&transfer.fromUserID, &transfer.toUserID, &transfer.amount); err != nil {
			return nil, err
		}

		result = append(result, domain.DirectTransfer{
			TargetID: getTargetIDFn(transfer),
			Amount:   uint32(transfer.amount),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
