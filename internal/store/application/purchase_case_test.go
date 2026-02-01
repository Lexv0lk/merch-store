//go:generate mockgen
package application

import (
	"testing"

	storemocks "github.com/Lexv0lk/merch-store/gen/mocks/store"
	"github.com/Lexv0lk/merch-store/internal/store/domain"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestPurchaseCase_BuyItem(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name     string
		userId   int
		goodName string

		prepareFn func(t *testing.T, ctrl *gomock.Controller) domain.PurchaseHandler

		expectedErr error
	}

	tests := []testCase{
		{
			name:     "successful purchase",
			userId:   1,
			goodName: "t-shirt",
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.PurchaseHandler {
				purchaseHandler := storemocks.NewMockPurchaseHandler(ctrl)
				purchaseHandler.EXPECT().HandlePurchase(gomock.Any(), 1, "t-shirt").Return(nil)
				return purchaseHandler
			},
			expectedErr: nil,
		},
		{
			name:     "good not found",
			userId:   1,
			goodName: "nonexistent",
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.PurchaseHandler {
				purchaseHandler := storemocks.NewMockPurchaseHandler(ctrl)
				purchaseHandler.EXPECT().HandlePurchase(gomock.Any(), 1, "nonexistent").Return(&domain.GoodNotFoundError{Msg: "good nonexistent not found"})
				return purchaseHandler
			},
			expectedErr: &domain.GoodNotFoundError{},
		},
		{
			name:     "insufficient balance",
			userId:   1,
			goodName: "expensive-item",
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.PurchaseHandler {
				purchaseHandler := storemocks.NewMockPurchaseHandler(ctrl)
				purchaseHandler.EXPECT().HandlePurchase(gomock.Any(), 1, "expensive-item").Return(&domain.InsufficientBalanceError{Msg: "insufficient balance"})
				return purchaseHandler
			},
			expectedErr: &domain.InsufficientBalanceError{},
		},
		{
			name:     "user not found",
			userId:   999,
			goodName: "t-shirt",
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.PurchaseHandler {
				purchaseHandler := storemocks.NewMockPurchaseHandler(ctrl)
				purchaseHandler.EXPECT().HandlePurchase(gomock.Any(), 999, "t-shirt").Return(&domain.UserNotFoundError{Msg: "user not found"})
				return purchaseHandler
			},
			expectedErr: &domain.UserNotFoundError{},
		},
		{
			name:     "internal error",
			userId:   1,
			goodName: "t-shirt",
			prepareFn: func(t *testing.T, ctrl *gomock.Controller) domain.PurchaseHandler {
				purchaseHandler := storemocks.NewMockPurchaseHandler(ctrl)
				purchaseHandler.EXPECT().HandlePurchase(gomock.Any(), 1, "t-shirt").Return(assert.AnError)
				return purchaseHandler
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

			purchaseHandler := tt.prepareFn(t, ctrl)
			purchaseCase := NewPurchaseCase(purchaseHandler)

			err := purchaseCase.BuyItem(t.Context(), tt.userId, tt.goodName)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
