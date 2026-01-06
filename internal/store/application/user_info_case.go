package application

import (
	"context"

	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"github.com/Lexv0lk/merch-store/internal/store/domain"
	"golang.org/x/sync/errgroup"
)

type UserInfoCase struct {
	infoFetcher domain.UserInfoFetcher
	logger      logging.Logger
}

func NewUserInfoCase(infoFetcher domain.UserInfoFetcher, logger logging.Logger) *UserInfoCase {
	return &UserInfoCase{
		logger:      logger,
		infoFetcher: infoFetcher,
	}
}

func (uic *UserInfoCase) GetUserInfo(ctx context.Context, userId string) (domain.TotalUserInfo, error) {
	group, groupCtx := errgroup.WithContext(ctx)

	var mainInfo domain.MainUserInfo
	var purchases map[domain.Good]int
	var transfers domain.CoinTransferHistory

	group.Go(func() error {
		var err error
		mainInfo, err = uic.infoFetcher.FetchMainUserInfo(groupCtx, userId)
		return err
	})

	group.Go(func() error {
		var err error
		purchases, err = uic.infoFetcher.FetchUserPurchases(groupCtx, mainInfo.Username)
		return err
	})

	group.Go(func() error {
		var err error
		transfers, err = uic.infoFetcher.FetchUserCoinTransfers(groupCtx, mainInfo.Username)
		return err
	})

	err := group.Wait()
	if err != nil {
		return domain.TotalUserInfo{}, err
	}

	return domain.TotalUserInfo{
		Username:            mainInfo.Username,
		Balance:             mainInfo.Balance,
		Goods:               purchases,
		CoinTransferHistory: transfers,
	}, nil
}
