package application

import (
	"context"
	"testing"

	dbmocks "github.com/Lexv0lk/merch-store/gen/mocks/database"
	storemocks "github.com/Lexv0lk/merch-store/gen/mocks/store"
	"github.com/Lexv0lk/merch-store/internal/pkg/database"
	"github.com/Lexv0lk/merch-store/internal/store/domain"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestPurchaseCase_BuyItem(t *testing.T) {
	t.Parallel()

	type deps struct {
		goodsRepository *storemocks.MockGoodsRepository
		balanceLocker   *storemocks.MockUserBalanceLocker
		purchaser       *storemocks.MockPurchaser
		txManager       *dbmocks.MockTxManager
	}

	type testCase struct {
		name     string
		userId   int
		goodName string

		prepareFn func(t *testing.T, d *deps)

		expectedErr error
	}

	// executeTxFn is a helper gomock.DoAndReturn that actually invokes the TxFunc callback
	executeTxFn := func(ctx context.Context, txFn database.TxFunc) error {
		return txFn(ctx, nil)
	}

	tests := []testCase{
		{
			name:     "successful purchase",
			userId:   1,
			goodName: "t-shirt",
			prepareFn: func(t *testing.T, d *deps) {
				d.goodsRepository.EXPECT().GetGoodInfo(gomock.Any(), "t-shirt").
					Return(domain.GoodInfo{Id: 10, Name: "t-shirt", Price: 80}, nil)
				d.txManager.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(executeTxFn)
				d.balanceLocker.EXPECT().LockAndGetUserBalance(gomock.Any(), nil, 1).
					Return(uint32(100), nil)
				d.purchaser.EXPECT().ProcessPurchase(gomock.Any(), nil, 1, domain.GoodInfo{Id: 10, Name: "t-shirt", Price: 80}).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:     "good not found",
			userId:   1,
			goodName: "nonexistent",
			prepareFn: func(t *testing.T, d *deps) {
				d.goodsRepository.EXPECT().GetGoodInfo(gomock.Any(), "nonexistent").
					Return(domain.GoodInfo{}, &domain.GoodNotFoundError{Msg: "good nonexistent not found"})
			},
			expectedErr: &domain.GoodNotFoundError{},
		},
		{
			name:     "user not found on balance lock",
			userId:   999,
			goodName: "t-shirt",
			prepareFn: func(t *testing.T, d *deps) {
				d.goodsRepository.EXPECT().GetGoodInfo(gomock.Any(), "t-shirt").
					Return(domain.GoodInfo{Id: 10, Name: "t-shirt", Price: 80}, nil)
				d.txManager.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(executeTxFn)
				d.balanceLocker.EXPECT().LockAndGetUserBalance(gomock.Any(), nil, 999).
					Return(uint32(0), &domain.UserNotFoundError{Msg: "user not found"})
			},
			expectedErr: &domain.UserNotFoundError{},
		},
		{
			name:     "insufficient balance",
			userId:   1,
			goodName: "expensive-item",
			prepareFn: func(t *testing.T, d *deps) {
				d.goodsRepository.EXPECT().GetGoodInfo(gomock.Any(), "expensive-item").
					Return(domain.GoodInfo{Id: 20, Name: "expensive-item", Price: 200}, nil)
				d.txManager.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(executeTxFn)
				d.balanceLocker.EXPECT().LockAndGetUserBalance(gomock.Any(), nil, 1).
					Return(uint32(50), nil)
			},
			expectedErr: &domain.InsufficientBalanceError{},
		},
		{
			name:     "process purchase error",
			userId:   1,
			goodName: "t-shirt",
			prepareFn: func(t *testing.T, d *deps) {
				d.goodsRepository.EXPECT().GetGoodInfo(gomock.Any(), "t-shirt").
					Return(domain.GoodInfo{Id: 10, Name: "t-shirt", Price: 80}, nil)
				d.txManager.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(executeTxFn)
				d.balanceLocker.EXPECT().LockAndGetUserBalance(gomock.Any(), nil, 1).
					Return(uint32(100), nil)
				d.purchaser.EXPECT().ProcessPurchase(gomock.Any(), nil, 1, domain.GoodInfo{Id: 10, Name: "t-shirt", Price: 80}).
					Return(assert.AnError)
			},
			expectedErr: assert.AnError,
		},
		{
			name:     "internal error",
			userId:   1,
			goodName: "t-shirt",
			prepareFn: func(t *testing.T, d *deps) {
				d.goodsRepository.EXPECT().GetGoodInfo(gomock.Any(), "t-shirt").
					Return(domain.GoodInfo{}, assert.AnError)
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

			d := &deps{
				goodsRepository: storemocks.NewMockGoodsRepository(ctrl),
				balanceLocker:   storemocks.NewMockUserBalanceLocker(ctrl),
				purchaser:       storemocks.NewMockPurchaser(ctrl),
				txManager:       dbmocks.NewMockTxManager(ctrl),
			}

			tt.prepareFn(t, d)

			purchaseCase := NewPurchaseCase(d.goodsRepository, d.balanceLocker, d.purchaser, d.txManager)
			err := purchaseCase.BuyItem(t.Context(), tt.userId, tt.goodName)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
