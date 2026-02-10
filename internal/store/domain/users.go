package domain

import (
	"context"

	"github.com/Lexv0lk/merch-store/internal/pkg/database"
)

const StartBalance uint32 = 1000

type BalanceEnsurer interface {
	EnsureBalanceCreated(ctx context.Context, userId int, startValue uint32) error
}

type UserFinder interface {
	GetTargetUsers(ctx context.Context, querier database.Querier, fromUsername, toUsername string) ([]UserInfo, error)
}

type UserBalanceLocker interface {
	LockAndGetUserBalance(ctx context.Context, querier database.Querier, userId int) (int, error)
}

type UserInfoRepository interface {
	FetchUsername(ctx context.Context, userId int) (string, error)
	FetchUserBalance(ctx context.Context, userId int) (uint32, error)
	FetchUserPurchases(ctx context.Context, userId int) (map[Good]uint32, error)
	FetchUserCoinTransfers(ctx context.Context, userId int) (CoinTransferHistory, error)
}

type UserInfo struct {
	Id       int
	Username string
	Balance  uint32
}

type MainUserInfo struct {
	Username string
	Balance  uint32
}

type TotalUserInfo struct {
	Username            string
	Balance             uint32
	Goods               map[Good]uint32
	CoinTransferHistory CoinTransferHistory
}

type Good struct {
	Name string
}

type CoinTransferHistory struct {
	IncomingTransfers  []DirectTransfer
	OutcomingTransfers []DirectTransfer
}

type DirectTransfer struct {
	TargetName string
	Amount     uint32
}
