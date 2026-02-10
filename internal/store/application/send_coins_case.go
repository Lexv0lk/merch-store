package application

import (
	"context"
	"fmt"

	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	"github.com/Lexv0lk/merch-store/internal/store/domain"
)

type SendCoinsCase struct {
	txManager            database.TxManager
	userFinder           domain.UserFinder
	transactionProceeder domain.TransactionProceeder
}

func NewSendCoinsCase(txManager database.TxManager, userFinder domain.UserFinder, transactionProceeder domain.TransactionProceeder) *SendCoinsCase {
	return &SendCoinsCase{
		txManager:            txManager,
		userFinder:           userFinder,
		transactionProceeder: transactionProceeder,
	}
}

func (sc *SendCoinsCase) SendCoins(ctx context.Context, fromUsername string, toUsername string, amount uint32) error {
	if fromUsername == toUsername {
		return &domain.InvalidArgumentsError{Msg: "fromUsername must differ from toUsername"}
	}

	return sc.txManager.WithinTransaction(ctx, func(ctx context.Context, executor database.QueryExecuter) error {
		users, err := sc.userFinder.GetTargetUsers(ctx, executor, fromUsername, toUsername)
		if err != nil {
			return fmt.Errorf("failed to find target users: %w", err)
		}

		var fromUser, toUser *domain.UserInfo
		if users[0].Username == fromUsername {
			fromUser = &users[0]
			toUser = &users[1]
		} else {
			fromUser = &users[1]
			toUser = &users[0]
		}

		if fromUser.Balance < amount {
			return &domain.InsufficientBalanceError{Msg: fmt.Sprintf("user %s has insufficient balance", fromUsername)}
		}

		err = sc.transactionProceeder.ProceedTransaction(ctx, executor, amount, fromUser, toUser)
		if err != nil {
			return fmt.Errorf("failed to proceed transaction: %w", err)
		}

		return nil
	})
}
