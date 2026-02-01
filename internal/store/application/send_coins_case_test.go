//go:generate mockgen
package application

import (
	"testing"

	storemocks "github.com/Lexv0lk/merch-store/gen/mocks/store"
	"github.com/Lexv0lk/merch-store/internal/store/domain"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestSendCoinsCase_SendCoins(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name         string
		fromUsername string
		toUsername   string
		amount       uint32

		prepareFn func(t *testing.T, ctrl *gomock.Controller) domain.CoinsTransferer

		expectedErr error
	}

	tests := []testCase{
		{
			name:         "successful transfer",
			fromUsername: "sender",
			toUsername:   "receiver",
			amount:       100,
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.CoinsTransferer {
				coinsTransferer := storemocks.NewMockCoinsTransferer(ctrl)
				coinsTransferer.EXPECT().TransferCoins(gomock.Any(), "sender", "receiver", uint32(100)).Return(nil)
				return coinsTransferer
			},
			expectedErr: nil,
		},
		{
			name:         "same username error",
			fromUsername: "sameuser",
			toUsername:   "sameuser",
			amount:       100,
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.CoinsTransferer {
				coinsTransferer := storemocks.NewMockCoinsTransferer(ctrl)
				coinsTransferer.EXPECT().TransferCoins(gomock.Any(), "sameuser", "sameuser", uint32(100)).Return(&domain.InvalidArgumentsError{Msg: "fromUsername must differ from toUsername"})
				return coinsTransferer
			},
			expectedErr: &domain.InvalidArgumentsError{},
		},
		{
			name:         "insufficient balance",
			fromUsername: "pooruser",
			toUsername:   "receiver",
			amount:       1000000,
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.CoinsTransferer {
				coinsTransferer := storemocks.NewMockCoinsTransferer(ctrl)
				coinsTransferer.EXPECT().TransferCoins(gomock.Any(), "pooruser", "receiver", uint32(1000000)).Return(&domain.InsufficientBalanceError{Msg: "insufficient balance"})
				return coinsTransferer
			},
			expectedErr: &domain.InsufficientBalanceError{},
		},
		{
			name:         "user not found",
			fromUsername: "sender",
			toUsername:   "nonexistent",
			amount:       100,
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.CoinsTransferer {
				coinsTransferer := storemocks.NewMockCoinsTransferer(ctrl)
				coinsTransferer.EXPECT().TransferCoins(gomock.Any(), "sender", "nonexistent", uint32(100)).Return(&domain.UserNotFoundError{Msg: "user not found"})
				return coinsTransferer
			},
			expectedErr: &domain.UserNotFoundError{},
		},
		{
			name:         "internal error",
			fromUsername: "sender",
			toUsername:   "receiver",
			amount:       100,
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.CoinsTransferer {
				coinsTransferer := storemocks.NewMockCoinsTransferer(ctrl)
				coinsTransferer.EXPECT().TransferCoins(gomock.Any(), "sender", "receiver", uint32(100)).Return(assert.AnError)
				return coinsTransferer
			},
			expectedErr: assert.AnError,
		},
	}

	for _, tc := range tests {
		tt := tc
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			coinsTransferer := tt.prepareFn(t, ctrl)
			sendCoinsCase := NewSendCoinsCase(coinsTransferer)

			err := sendCoinsCase.SendCoins(t.Context(), tt.fromUsername, tt.toUsername, tt.amount)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
