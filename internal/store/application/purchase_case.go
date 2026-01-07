package application

import (
	"context"

	"github.com/Lexv0lk/merch-store/internal/store/domain"
)

type PurchaseCase struct {
	purchaseHandler domain.PurchaseHandler
}

func NewPurchaseCase(purchaseHandler domain.PurchaseHandler) *PurchaseCase {
	return &PurchaseCase{
		purchaseHandler: purchaseHandler,
	}
}

func (pc *PurchaseCase) PurchaseGood(ctx context.Context, userId int, goodName string) error {
	err := pc.purchaseHandler.HandlePurchase(ctx, userId, goodName)
	if err != nil {
		return err
	}

	return nil
}
