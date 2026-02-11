package application

import (
	"context"

	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"github.com/Lexv0lk/merch-store/internal/store/domain"
	"golang.org/x/sync/errgroup"
)

type UserInfoCase struct {
	userRepository domain.UserInfoRepository
	usernameGetter domain.UsernameGetter
	logger         logging.Logger
}

func NewUserInfoCase(userRepository domain.UserInfoRepository, usernameGetter domain.UsernameGetter, logger logging.Logger) *UserInfoCase {
	return &UserInfoCase{
		logger:         logger,
		userRepository: userRepository,
		usernameGetter: usernameGetter,
	}
}

func (uic *UserInfoCase) GetUserInfo(ctx context.Context, userId int) (domain.TotalUserInfo, error) {
	group, groupCtx := errgroup.WithContext(ctx)

	var mainInfo domain.MainUserInfo
	var purchases map[domain.Good]uint32
	var transfers domain.NamedTransferHistory

	group.Go(func() error {
		var err error

		username, err := uic.usernameGetter.GetUsername(groupCtx, userId)
		if err != nil {
			return err
		}

		balance, err := uic.userRepository.FetchUserBalance(groupCtx, userId)
		if err != nil {
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
		rawTransfers, err := uic.userRepository.FetchUserCoinTransfers(groupCtx, userId)
		if err != nil {
			return err
		}

		transfers, err = convertToNamedTransferHistory(groupCtx, rawTransfers, uic.usernameGetter)
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

func convertToNamedTransferHistory(ctx context.Context, tf domain.TransferHistory, usernameGetter domain.UsernameGetter) (domain.NamedTransferHistory, error) {
	usernames, err := usernameGetter.GetUsernames(ctx, extractUserIDs(tf)...)
	if err != nil {
		return domain.NamedTransferHistory{}, err
	}

	namedTF := domain.NamedTransferHistory{
		IncomingTransfers:  make([]domain.NamedDirectTransfer, 0, len(tf.IncomingTransfers)),
		OutcomingTransfers: make([]domain.NamedDirectTransfer, 0, len(tf.OutcomingTransfers)),
	}

	for _, transfer := range tf.IncomingTransfers {
		namedTF.IncomingTransfers = append(namedTF.IncomingTransfers, domain.NamedDirectTransfer{
			TargetUsername: usernames[transfer.TargetID],
			Amount:         transfer.Amount,
		})
	}

	for _, transfer := range tf.OutcomingTransfers {
		namedTF.OutcomingTransfers = append(namedTF.OutcomingTransfers, domain.NamedDirectTransfer{
			TargetUsername: usernames[transfer.TargetID],
			Amount:         transfer.Amount,
		})
	}

	return namedTF, nil
}

func extractUserIDs(tf domain.TransferHistory) []int {
	userIDSet := make(map[int]struct{})
	for _, transfer := range tf.IncomingTransfers {
		userIDSet[transfer.TargetID] = struct{}{}
	}
	for _, transfer := range tf.OutcomingTransfers {
		userIDSet[transfer.TargetID] = struct{}{}
	}

	userIDs := make([]int, 0, len(userIDSet))
	for id := range userIDSet {
		userIDs = append(userIDs, id)
	}

	return userIDs
}
