package application

import (
	"context"

	"github.com/Lexv0lk/merch-store/internal/store/domain"
)

type SendCoinsCase struct {
	coinsTransferer domain.CoinsTransferer
}

func NewSendCoinsCase(coinsTransferer domain.CoinsTransferer) *SendCoinsCase {
	return &SendCoinsCase{
		coinsTransferer: coinsTransferer,
	}
}

func (sc *SendCoinsCase) SendCoins(ctx context.Context, fromUsername string, toUsername string, amount int) error {
	err := sc.coinsTransferer.TransferCoins(ctx, fromUsername, toUsername, amount)
	if err != nil {
		return err
	}

	return nil
}
