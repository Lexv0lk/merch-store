package domain

import (
	"context"

	"github.com/Lexv0lk/merch-store/internal/pkg/database"
)

const StartBalance uint32 = 1000

type BalanceEnsurer interface {
	EnsureBalanceCreated(ctx context.Context, userId int, startValue uint32) error
}

type UserBalanceLocker interface {
	LockAndGetUserBalance(ctx context.Context, querier database.Querier, userId int) (int, error)
}

type UserInfoRepository interface {
	UserBalanceFetcher
	FetchUserPurchases(ctx context.Context, userId int) (map[Good]uint32, error)
	FetchUserCoinTransfers(ctx context.Context, userId int) (TransferHistory, error)
}

type UsernameGetter interface {
	GetUsername(ctx context.Context, userId int) (string, error)
	GetUsernames(ctx context.Context, userId ...int) (map[int]string, error)
}

type UserIDFetcher interface {
	FetchUserID(ctx context.Context, username string) (int, error)
}

type UserBalanceFetcher interface {
	FetchUserBalance(ctx context.Context, userId int) (uint32, error)
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
	CoinTransferHistory NamedTransferHistory
}

type Good struct {
	Name string
}

type TransferHistory struct {
	IncomingTransfers  []DirectTransfer
	OutcomingTransfers []DirectTransfer
}

type NamedTransferHistory struct {
	IncomingTransfers  []NamedDirectTransfer
	OutcomingTransfers []NamedDirectTransfer
}

type DirectTransfer struct {
	TargetID int
	Amount   uint32
}

type NamedDirectTransfer struct {
	TargetUsername string
	Amount         uint32
}
