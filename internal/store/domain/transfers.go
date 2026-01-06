package domain

import "context"

type CoinsTransferer interface {
	TransferCoins(ctx context.Context, fromUsername string, toUsername string, amount int) error
}
