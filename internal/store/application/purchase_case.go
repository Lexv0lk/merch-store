package application

import (
	"context"
	"fmt"

	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	"github.com/Lexv0lk/merch-store/internal/store/domain"
)

type PurchaseCase struct {
	goodsRepository domain.GoodsRepository
	balanceLocker   domain.UserBalanceLocker
	purchaser       domain.Purchaser
	txManager       database.TxManager
}

func NewPurchaseCase(goodsRepository domain.GoodsRepository, balanceLocker domain.UserBalanceLocker,
	purchaser domain.Purchaser, txManager database.TxManager) *PurchaseCase {
	return &PurchaseCase{
		goodsRepository: goodsRepository,
		balanceLocker:   balanceLocker,
		purchaser:       purchaser,
		txManager:       txManager,
	}
}

func (pc *PurchaseCase) BuyItem(ctx context.Context, userId int, goodName string) error {
	goodInfo, err := pc.goodsRepository.GetGoodInfo(ctx, goodName)
	if err != nil {
		return fmt.Errorf("failed to get good info: %w", err)
	}

	return pc.txManager.WithinTransaction(ctx, func(ctx context.Context, executor database.QueryExecuter) error {
		balance, err := pc.balanceLocker.LockAndGetUserBalance(ctx, executor, userId)
		if err != nil {
			return fmt.Errorf("failed to lock and get user balance: %w", err)
		}

		if balance < goodInfo.Price {
			return &domain.InsufficientBalanceError{Msg: "insufficient balance"}
		}

		err = pc.purchaser.ProcessPurchase(ctx, executor, userId, goodInfo)
		if err != nil {
			return fmt.Errorf("failed to process purchase: %w", err)
		}

		return nil
	})
}
