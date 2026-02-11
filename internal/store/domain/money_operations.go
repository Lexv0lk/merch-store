package domain

import (
	"context"

	"github.com/Lexv0lk/merch-store/internal/pkg/database"
)

type TransactionProceeder interface {
	ProceedTransaction(ctx context.Context, executor database.Executor, amount uint32, fromUserID, toUserID int) error
}

type Purchaser interface {
	ProcessPurchase(ctx context.Context, executor database.Executor, userId int, good GoodInfo) error
}
