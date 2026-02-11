package application

import (
	"testing"

	loggingmocks "github.com/Lexv0lk/merch-store/gen/mocks/logging"
	storemocks "github.com/Lexv0lk/merch-store/gen/mocks/store"
	"github.com/Lexv0lk/merch-store/internal/pkg/logging"
	"github.com/Lexv0lk/merch-store/internal/store/domain"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUserInfoCase_GetUserInfo(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name   string
		userId int

		prepareFn func(t *testing.T, ctrl *gomock.Controller) (domain.UserInfoRepository, domain.UsernameGetter, logging.Logger)

		expectedUserInfo domain.TotalUserInfo
		expectedErr      error
	}

	tests := []testCase{
		{
			name:   "successful fetch all info",
			userId: 1,
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (domain.UserInfoRepository, domain.UsernameGetter, logging.Logger) {
				infoRepository := storemocks.NewMockUserInfoRepository(ctrl)
				usernameGetter := storemocks.NewMockUsernameGetter(ctrl)
				logger := loggingmocks.NewMockLogger(ctrl)

				usernameGetter.EXPECT().GetUsername(gomock.Any(), 1).Return("testuser", nil)
				infoRepository.EXPECT().FetchUserBalance(gomock.Any(), 1).Return(uint32(1000), nil)
				infoRepository.EXPECT().FetchUserPurchases(gomock.Any(), 1).Return(map[domain.Good]uint32{
					{Name: "t-shirt"}: 2,
					{Name: "cup"}:     1,
				}, nil)
				infoRepository.EXPECT().FetchUserCoinTransfers(gomock.Any(), 1).Return(domain.TransferHistory{
					IncomingTransfers: []domain.DirectTransfer{
						{TargetID: 10, Amount: 50},
					},
					OutcomingTransfers: []domain.DirectTransfer{
						{TargetID: 20, Amount: 100},
					},
				}, nil)
				usernameGetter.EXPECT().GetUsernames(gomock.Any(), gomock.Any()).Return(map[int]string{
					10: "sender1",
					20: "receiver1",
				}, nil)

				return infoRepository, usernameGetter, logger
			},
			expectedUserInfo: domain.TotalUserInfo{
				Username: "testuser",
				Balance:  1000,
				Goods: map[domain.Good]uint32{
					{Name: "t-shirt"}: 2,
					{Name: "cup"}:     1,
				},
				CoinTransferHistory: domain.NamedTransferHistory{
					IncomingTransfers: []domain.NamedDirectTransfer{
						{TargetUsername: "sender1", Amount: 50},
					},
					OutcomingTransfers: []domain.NamedDirectTransfer{
						{TargetUsername: "receiver1", Amount: 100},
					},
				},
			},
			expectedErr: nil,
		},
		{
			name:   "user not found",
			userId: 999,
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (domain.UserInfoRepository, domain.UsernameGetter, logging.Logger) {
				infoRepository := storemocks.NewMockUserInfoRepository(ctrl)
				usernameGetter := storemocks.NewMockUsernameGetter(ctrl)
				logger := loggingmocks.NewMockLogger(ctrl)

				usernameGetter.EXPECT().GetUsername(gomock.Any(), 999).Return("", &domain.UserNotFoundError{Msg: "user not found"})
				infoRepository.EXPECT().FetchUserPurchases(gomock.Any(), 999).Return(nil, nil).AnyTimes()
				infoRepository.EXPECT().FetchUserCoinTransfers(gomock.Any(), 999).Return(domain.TransferHistory{}, nil).AnyTimes()
				usernameGetter.EXPECT().GetUsernames(gomock.Any(), gomock.Any()).Return(map[int]string{}, nil).AnyTimes()

				return infoRepository, usernameGetter, logger
			},
			expectedUserInfo: domain.TotalUserInfo{},
			expectedErr:      &domain.UserNotFoundError{},
		},
		{
			name:   "fetch purchases error",
			userId: 1,
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (domain.UserInfoRepository, domain.UsernameGetter, logging.Logger) {
				infoRepository := storemocks.NewMockUserInfoRepository(ctrl)
				usernameGetter := storemocks.NewMockUsernameGetter(ctrl)
				logger := loggingmocks.NewMockLogger(ctrl)

				usernameGetter.EXPECT().GetUsername(gomock.Any(), 1).Return("testuser", nil).AnyTimes()
				infoRepository.EXPECT().FetchUserBalance(gomock.Any(), 1).Return(uint32(1000), nil).AnyTimes()
				infoRepository.EXPECT().FetchUserPurchases(gomock.Any(), 1).Return(nil, assert.AnError)
				infoRepository.EXPECT().FetchUserCoinTransfers(gomock.Any(), 1).Return(domain.TransferHistory{}, nil).AnyTimes()
				usernameGetter.EXPECT().GetUsernames(gomock.Any(), gomock.Any()).Return(map[int]string{}, nil).AnyTimes()

				return infoRepository, usernameGetter, logger
			},
			expectedUserInfo: domain.TotalUserInfo{},
			expectedErr:      assert.AnError,
		},
		{
			name:   "fetch coin transfers error",
			userId: 1,
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (domain.UserInfoRepository, domain.UsernameGetter, logging.Logger) {
				infoRepository := storemocks.NewMockUserInfoRepository(ctrl)
				usernameGetter := storemocks.NewMockUsernameGetter(ctrl)
				logger := loggingmocks.NewMockLogger(ctrl)

				usernameGetter.EXPECT().GetUsername(gomock.Any(), 1).Return("testuser", nil).AnyTimes()
				infoRepository.EXPECT().FetchUserBalance(gomock.Any(), 1).Return(uint32(1000), nil).AnyTimes()
				infoRepository.EXPECT().FetchUserPurchases(gomock.Any(), 1).Return(map[domain.Good]uint32{}, nil).AnyTimes()
				infoRepository.EXPECT().FetchUserCoinTransfers(gomock.Any(), 1).Return(domain.TransferHistory{}, assert.AnError)

				return infoRepository, usernameGetter, logger
			},
			expectedUserInfo: domain.TotalUserInfo{},
			expectedErr:      assert.AnError,
		},
		{
			name:   "empty purchases and transfers",
			userId: 2,
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (domain.UserInfoRepository, domain.UsernameGetter, logging.Logger) {
				infoRepository := storemocks.NewMockUserInfoRepository(ctrl)
				usernameGetter := storemocks.NewMockUsernameGetter(ctrl)
				logger := loggingmocks.NewMockLogger(ctrl)

				usernameGetter.EXPECT().GetUsername(gomock.Any(), 2).Return("newuser", nil)
				infoRepository.EXPECT().FetchUserBalance(gomock.Any(), 2).Return(uint32(500), nil)
				infoRepository.EXPECT().FetchUserPurchases(gomock.Any(), 2).Return(map[domain.Good]uint32{}, nil)
				infoRepository.EXPECT().FetchUserCoinTransfers(gomock.Any(), 2).Return(domain.TransferHistory{
					IncomingTransfers:  []domain.DirectTransfer{},
					OutcomingTransfers: []domain.DirectTransfer{},
				}, nil)
				usernameGetter.EXPECT().GetUsernames(gomock.Any()).Return(map[int]string{}, nil)

				return infoRepository, usernameGetter, logger
			},
			expectedUserInfo: domain.TotalUserInfo{
				Username: "newuser",
				Balance:  500,
				Goods:    map[domain.Good]uint32{},
				CoinTransferHistory: domain.NamedTransferHistory{
					IncomingTransfers:  []domain.NamedDirectTransfer{},
					OutcomingTransfers: []domain.NamedDirectTransfer{},
				},
			},
			expectedErr: nil,
		},
	}

	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			infoFetcher, usernameGetter, logger := tt.prepareFn(t, ctrl)
			userInfoCase := NewUserInfoCase(infoFetcher, usernameGetter, logger)

			userInfo, err := userInfoCase.GetUserInfo(t.Context(), tt.userId)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUserInfo, userInfo)
			}
		})
	}
}
