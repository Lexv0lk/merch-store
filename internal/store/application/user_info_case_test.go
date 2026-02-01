//go:generate mockgen
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

		prepareFn func(t *testing.T, ctrl *gomock.Controller) (domain.UserInfoFetcher, logging.Logger)

		expectedUserInfo domain.TotalUserInfo
		expectedErr      error
	}

	tests := []testCase{
		{
			name:   "successful fetch all info",
			userId: 1,
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (domain.UserInfoFetcher, logging.Logger) {
				infoFetcher := storemocks.NewMockUserInfoFetcher(ctrl)
				logger := loggingmocks.NewMockLogger(ctrl)

				infoFetcher.EXPECT().FetchMainUserInfo(gomock.Any(), 1).Return(domain.MainUserInfo{
					Username: "testuser",
					Balance:  1000,
				}, nil)
				infoFetcher.EXPECT().FetchUserPurchases(gomock.Any(), 1).Return(map[domain.Good]uint32{
					{Name: "t-shirt"}: 2,
					{Name: "cup"}:     1,
				}, nil)
				infoFetcher.EXPECT().FetchUserCoinTransfers(gomock.Any(), 1).Return(domain.CoinTransferHistory{
					IncomingTransfers: []domain.DirectTransfer{
						{TargetName: "sender1", Amount: 50},
					},
					OutcomingTransfers: []domain.DirectTransfer{
						{TargetName: "receiver1", Amount: 100},
					},
				}, nil)

				return infoFetcher, logger
			},
			expectedUserInfo: domain.TotalUserInfo{
				Username: "testuser",
				Balance:  1000,
				Goods: map[domain.Good]uint32{
					{Name: "t-shirt"}: 2,
					{Name: "cup"}:     1,
				},
				CoinTransferHistory: domain.CoinTransferHistory{
					IncomingTransfers: []domain.DirectTransfer{
						{TargetName: "sender1", Amount: 50},
					},
					OutcomingTransfers: []domain.DirectTransfer{
						{TargetName: "receiver1", Amount: 100},
					},
				},
			},
			expectedErr: nil,
		},
		{
			name:   "user not found",
			userId: 999,
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (domain.UserInfoFetcher, logging.Logger) {
				infoFetcher := storemocks.NewMockUserInfoFetcher(ctrl)
				logger := loggingmocks.NewMockLogger(ctrl)

				infoFetcher.EXPECT().FetchMainUserInfo(gomock.Any(), 999).Return(domain.MainUserInfo{}, &domain.UserNotFoundError{Msg: "user not found"})
				infoFetcher.EXPECT().FetchUserPurchases(gomock.Any(), 999).Return(nil, nil).AnyTimes()
				infoFetcher.EXPECT().FetchUserCoinTransfers(gomock.Any(), 999).Return(domain.CoinTransferHistory{}, nil).AnyTimes()

				return infoFetcher, logger
			},
			expectedUserInfo: domain.TotalUserInfo{},
			expectedErr:      &domain.UserNotFoundError{},
		},
		{
			name:   "fetch purchases error",
			userId: 1,
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (domain.UserInfoFetcher, logging.Logger) {
				infoFetcher := storemocks.NewMockUserInfoFetcher(ctrl)
				logger := loggingmocks.NewMockLogger(ctrl)

				infoFetcher.EXPECT().FetchMainUserInfo(gomock.Any(), 1).Return(domain.MainUserInfo{
					Username: "testuser",
					Balance:  1000,
				}, nil).AnyTimes()
				infoFetcher.EXPECT().FetchUserPurchases(gomock.Any(), 1).Return(nil, assert.AnError)
				infoFetcher.EXPECT().FetchUserCoinTransfers(gomock.Any(), 1).Return(domain.CoinTransferHistory{}, nil).AnyTimes()

				return infoFetcher, logger
			},
			expectedUserInfo: domain.TotalUserInfo{},
			expectedErr:      assert.AnError,
		},
		{
			name:   "fetch coin transfers error",
			userId: 1,
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (domain.UserInfoFetcher, logging.Logger) {
				infoFetcher := storemocks.NewMockUserInfoFetcher(ctrl)
				logger := loggingmocks.NewMockLogger(ctrl)

				infoFetcher.EXPECT().FetchMainUserInfo(gomock.Any(), 1).Return(domain.MainUserInfo{
					Username: "testuser",
					Balance:  1000,
				}, nil).AnyTimes()
				infoFetcher.EXPECT().FetchUserPurchases(gomock.Any(), 1).Return(map[domain.Good]uint32{}, nil).AnyTimes()
				infoFetcher.EXPECT().FetchUserCoinTransfers(gomock.Any(), 1).Return(domain.CoinTransferHistory{}, assert.AnError)

				return infoFetcher, logger
			},
			expectedUserInfo: domain.TotalUserInfo{},
			expectedErr:      assert.AnError,
		},
		{
			name:   "empty purchases and transfers",
			userId: 2,
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) (domain.UserInfoFetcher, logging.Logger) {
				infoFetcher := storemocks.NewMockUserInfoFetcher(ctrl)
				logger := loggingmocks.NewMockLogger(ctrl)

				infoFetcher.EXPECT().FetchMainUserInfo(gomock.Any(), 2).Return(domain.MainUserInfo{
					Username: "newuser",
					Balance:  500,
				}, nil)
				infoFetcher.EXPECT().FetchUserPurchases(gomock.Any(), 2).Return(map[domain.Good]uint32{}, nil)
				infoFetcher.EXPECT().FetchUserCoinTransfers(gomock.Any(), 2).Return(domain.CoinTransferHistory{
					IncomingTransfers:  []domain.DirectTransfer{},
					OutcomingTransfers: []domain.DirectTransfer{},
				}, nil)

				return infoFetcher, logger
			},
			expectedUserInfo: domain.TotalUserInfo{
				Username: "newuser",
				Balance:  500,
				Goods:    map[domain.Good]uint32{},
				CoinTransferHistory: domain.CoinTransferHistory{
					IncomingTransfers:  []domain.DirectTransfer{},
					OutcomingTransfers: []domain.DirectTransfer{},
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

			infoFetcher, logger := tt.prepareFn(t, ctrl)
			userInfoCase := NewUserInfoCase(infoFetcher, logger)

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
