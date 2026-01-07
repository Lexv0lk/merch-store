package domain

import "context"

type UserInfoFetcher interface {
	FetchMainUserInfo(ctx context.Context, userId int) (MainUserInfo, error)
	FetchUserPurchases(ctx context.Context, username string) (map[Good]int, error)
	FetchUserCoinTransfers(ctx context.Context, username string) (CoinTransferHistory, error)
}

type MainUserInfo struct {
	Username string
	Balance  uint32
}

type TotalUserInfo struct {
	Username            string
	Balance             uint32
	Goods               map[Good]int
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
	Amount     int
}
