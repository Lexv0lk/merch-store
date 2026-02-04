package application

import (
	"context"
	"errors"

	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"github.com/Lexv0lk/merch-store/internal/store/domain"
	"golang.org/x/sync/errgroup"
)

const (
	startBalance = 1000
)

type UserInfoCase struct {
	userRepository domain.UserInfoRepository
	logger         logging.Logger
}

func NewUserInfoCase(userRepository domain.UserInfoRepository, logger logging.Logger) *UserInfoCase {
	return &UserInfoCase{
		logger:         logger,
		userRepository: userRepository,
	}
}

func (uic *UserInfoCase) GetUserInfo(ctx context.Context, userId int) (domain.TotalUserInfo, error) {
	group, groupCtx := errgroup.WithContext(ctx)

	var mainInfo domain.MainUserInfo
	var purchases map[domain.Good]uint32
	var transfers domain.CoinTransferHistory

	group.Go(func() error {
		var err error

		username, err := uic.userRepository.FetchUsername(groupCtx, userId)
		if err != nil {
			return err
		}

		balance, err := uic.userRepository.FetchUserBalance(groupCtx, userId)
		if errors.Is(err, &domain.BalanceNotFoundError{}) {
			err = uic.userRepository.CreateBalance(groupCtx, userId, startBalance)
			if err != nil {
				return err
			}

			balance = startBalance
		} else if err != nil {
			return err
		}

		mainInfo = domain.MainUserInfo{
			Username: username,
			Balance:  balance,
		}

		return nil
	})

	group.Go(func() error {
		var err error
		purchases, err = uic.userRepository.FetchUserPurchases(groupCtx, userId)
		return err
	})

	group.Go(func() error {
		var err error
		transfers, err = uic.userRepository.FetchUserCoinTransfers(groupCtx, userId)
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
