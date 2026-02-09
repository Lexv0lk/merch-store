package domain

import (
	"context"

	"github.com/Lexv0lk/merch-store/internal/pkg/database"
)

type CoinsTransferer interface {
	TransferCoins(ctx context.Context, fromUsername string, toUsername string, amount uint32) error
}

type Purchaser interface {
	ProcessPurchase(ctx context.Context, executor database.Executor, userId int, good GoodInfo) error
}
