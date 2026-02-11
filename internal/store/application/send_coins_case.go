package application

import (
	"context"
	"fmt"

	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	"github.com/Lexv0lk/merch-store/internal/store/domain"
)

type SendCoinsCase struct {
	txManager            database.TxManager
	userIDFetcher        domain.UserIDFetcher
	transactionProceeder domain.TransactionProceeder
	userBalanceFetcher   domain.UserBalanceFetcher
	balanceCreator       domain.BalanceEnsurer
}

func NewSendCoinsCase(txManager database.TxManager,
	userIDFetcher domain.UserIDFetcher,
	userBalanceFetcher domain.UserBalanceFetcher,
	balanceCreator domain.BalanceEnsurer,
	transactionProceeder domain.TransactionProceeder) *SendCoinsCase {
	return &SendCoinsCase{
		txManager:            txManager,
		userIDFetcher:        userIDFetcher,
		transactionProceeder: transactionProceeder,
		userBalanceFetcher:   userBalanceFetcher,
		balanceCreator:       balanceCreator,
	}
}

func (sc *SendCoinsCase) SendCoins(ctx context.Context, fromUserID int, toUsername string, amount uint32) error {
	toUserID, err := sc.userIDFetcher.FetchUserID(ctx, toUsername)
	if err != nil {
		return &domain.UserNotFoundError{Msg: fmt.Sprintf("user not found: %s", toUsername)}
	}

	if toUserID == fromUserID {
		return &domain.InvalidArgumentsError{Msg: "from_user must differ from to_user"}
	}

	err = sc.balanceCreator.EnsureBalanceCreated(ctx, toUserID, domain.StartBalance)
	if err != nil {
		return fmt.Errorf("failed to ensure balance for user %d: %w", toUserID, err)
	}

	return sc.txManager.WithinTransaction(ctx, func(ctx context.Context, executor database.QueryExecuter) error {
		fromUserBalance, err := sc.userBalanceFetcher.FetchUserBalance(ctx, fromUserID)

		if fromUserBalance < amount {
			return &domain.InsufficientBalanceError{Msg: fmt.Sprintf("user %d has insufficient balance", fromUserID)}
		}

		err = sc.transactionProceeder.ProceedTransaction(ctx, executor, amount, fromUserID, toUserID)
		if err != nil {
			return fmt.Errorf("failed to proceed transaction: %w", err)
		}

		return nil
	})
}
