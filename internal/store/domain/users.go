package domain

import "context"

type UserInfoRepository interface {
	FetchUsername(ctx context.Context, userId int) (string, error)
	FetchUserBalance(ctx context.Context, userId int) (uint32, error)
	FetchUserPurchases(ctx context.Context, userId int) (map[Good]uint32, error)
	FetchUserCoinTransfers(ctx context.Context, userId int) (CoinTransferHistory, error)
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
