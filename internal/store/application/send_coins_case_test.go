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

func TestSendCoinsCase_SendCoins(t *testing.T) {
	t.Parallel()

	type deps struct {
		txManager            *dbmocks.MockTxManager
		userIDFetcher        *storemocks.MockUserIDFetcher
		balanceLocker        *storemocks.MockUserBalanceLocker
		balanceCreator       *storemocks.MockBalanceEnsurer
		transactionProceeder *storemocks.MockTransactionProceeder
	}

	type testCase struct {
		name       string
		fromUserID int
		toUsername string
		amount     uint32

		prepareFn func(t *testing.T, d *deps)

		expectedErr error
	}

	executeTxFn := func(ctx context.Context, txFn database.TxFunc) error {
		return txFn(ctx, nil)
	}

	tests := []testCase{
		{
			name:       "successful transfer",
			fromUserID: 1,
			toUsername: "receiver",
			amount:     100,
			prepareFn: func(t *testing.T, d *deps) {
				d.userIDFetcher.EXPECT().FetchUserID(gomock.Any(), "receiver").
					Return(2, nil)
				d.balanceCreator.EXPECT().EnsureBalanceCreated(gomock.Any(), 2, domain.StartBalance).
					Return(nil)
				d.txManager.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(executeTxFn)
				d.balanceLocker.EXPECT().LockAndGetUserBalance(gomock.Any(), nil, 1).
					Return(uint32(500), nil)
				d.transactionProceeder.EXPECT().ProceedTransaction(gomock.Any(), nil, uint32(100), 1, 2).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:       "same user error",
			fromUserID: 1,
			toUsername: "sender",
			amount:     100,
			prepareFn: func(t *testing.T, d *deps) {
				d.userIDFetcher.EXPECT().FetchUserID(gomock.Any(), "sender").
					Return(1, nil)
			},
			expectedErr: &domain.InvalidArgumentsError{},
		},
		{
			name:       "insufficient balance",
			fromUserID: 1,
			toUsername: "receiver",
			amount:     1000,
			prepareFn: func(t *testing.T, d *deps) {
				d.userIDFetcher.EXPECT().FetchUserID(gomock.Any(), "receiver").
					Return(2, nil)
				d.balanceCreator.EXPECT().EnsureBalanceCreated(gomock.Any(), 2, domain.StartBalance).
					Return(nil)
				d.txManager.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(executeTxFn)
				d.balanceLocker.EXPECT().LockAndGetUserBalance(gomock.Any(), nil, 1).
					Return(uint32(500), nil)
			},
			expectedErr: &domain.InsufficientBalanceError{},
		},
		{
			name:       "user not found",
			fromUserID: 1,
			toUsername: "nonexistent",
			amount:     100,
			prepareFn: func(t *testing.T, d *deps) {
				d.userIDFetcher.EXPECT().FetchUserID(gomock.Any(), "nonexistent").
					Return(0, &domain.UserNotFoundError{Msg: "user not found"})
			},
			expectedErr: &domain.UserNotFoundError{},
		},
		{
			name:       "proceed transaction error",
			fromUserID: 1,
			toUsername: "receiver",
			amount:     100,
			prepareFn: func(t *testing.T, d *deps) {
				d.userIDFetcher.EXPECT().FetchUserID(gomock.Any(), "receiver").
					Return(2, nil)
				d.balanceCreator.EXPECT().EnsureBalanceCreated(gomock.Any(), 2, domain.StartBalance).
					Return(nil)
				d.txManager.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(executeTxFn)
				d.balanceLocker.EXPECT().LockAndGetUserBalance(gomock.Any(), nil, 1).
					Return(uint32(500), nil)
				d.transactionProceeder.EXPECT().ProceedTransaction(gomock.Any(), nil, uint32(100), 1, 2).
					Return(assert.AnError)
			},
			expectedErr: assert.AnError,
		},
		{
			name:       "ensure balance error",
			fromUserID: 1,
			toUsername: "receiver",
			amount:     100,
			prepareFn: func(t *testing.T, d *deps) {
				d.userIDFetcher.EXPECT().FetchUserID(gomock.Any(), "receiver").
					Return(2, nil)
				d.balanceCreator.EXPECT().EnsureBalanceCreated(gomock.Any(), 2, domain.StartBalance).
					Return(assert.AnError)
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
				txManager:            dbmocks.NewMockTxManager(ctrl),
				userIDFetcher:        storemocks.NewMockUserIDFetcher(ctrl),
				balanceLocker:        storemocks.NewMockUserBalanceLocker(ctrl),
				balanceCreator:       storemocks.NewMockBalanceEnsurer(ctrl),
				transactionProceeder: storemocks.NewMockTransactionProceeder(ctrl),
			}

			tt.prepareFn(t, d)

			sendCoinsCase := NewSendCoinsCase(d.txManager, d.userIDFetcher, d.balanceLocker, d.balanceCreator, d.transactionProceeder)
			err := sendCoinsCase.SendCoins(t.Context(), tt.fromUserID, tt.toUsername, tt.amount)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
