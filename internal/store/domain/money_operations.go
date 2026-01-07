package domain

import "context"

type CoinsTransferer interface {
	TransferCoins(ctx context.Context, fromUsername string, toUsername string, amount int) error
}

type PurchaseHandler interface {
	HandlePurchase(ctx context.Context, userId int, goodName string) error
}
