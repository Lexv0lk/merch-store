package domain

import "context"

type CoinsTransferer interface {
	TransferCoins(ctx context.Context, fromUsername string, toUsername string, amount uint32) error
}

type PurchaseHandler interface {
	HandlePurchase(ctx context.Context, userId int, goodName string) error
}
