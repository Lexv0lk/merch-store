package application

import (
	"context"

	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"github.com/Lexv0lk/merch-store/internal/store/domain"
)

type SendCoinsCase struct {
	coinsTransferer domain.CoinsTransferer
	logger          logging.Logger
}

func NewSendCoinsCase(coinsTransferer domain.CoinsTransferer, logger logging.Logger) *SendCoinsCase {
	return &SendCoinsCase{
		coinsTransferer: coinsTransferer,
		logger:          logger,
	}
}

func (sc *SendCoinsCase) SendCoins(ctx context.Context, fromUsername string, toUsername string, amount int) error {
	err := sc.coinsTransferer.TransferCoins(ctx, fromUsername, toUsername, amount)
	if err != nil {
		return err
	}

	return nil
}
