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
		userFinder           *storemocks.MockUserFinder
		transactionProceeder *storemocks.MockTransactionProceeder
	}

	type testCase struct {
		name         string
		fromUsername string
		toUsername   string
		amount       uint32

		prepareFn func(t *testing.T, d *deps)

		expectedErr error
	}

	executeTxFn := func(ctx context.Context, txFn database.TxFunc) error {
		return txFn(ctx, nil)
	}

	tests := []testCase{
		{
			name:         "successful transfer",
			fromUsername: "sender",
			toUsername:   "receiver",
			amount:       100,
			prepareFn: func(t *testing.T, d *deps) {
				d.txManager.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(executeTxFn)
				d.userFinder.EXPECT().GetTargetUsers(gomock.Any(), nil, "sender", "receiver").
					Return([]domain.UserInfo{
						{Id: 1, Username: "sender", Balance: 500},
						{Id: 2, Username: "receiver", Balance: 200},
					}, nil)
				d.transactionProceeder.EXPECT().ProceedTransaction(gomock.Any(), nil, uint32(100),
					&domain.UserInfo{Id: 1, Username: "sender", Balance: 500},
					&domain.UserInfo{Id: 2, Username: "receiver", Balance: 200}).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:         "successful transfer with reversed user order",
			fromUsername: "sender",
			toUsername:   "receiver",
			amount:       100,
			prepareFn: func(t *testing.T, d *deps) {
				d.txManager.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(executeTxFn)
				d.userFinder.EXPECT().GetTargetUsers(gomock.Any(), nil, "sender", "receiver").
					Return([]domain.UserInfo{
						{Id: 2, Username: "receiver", Balance: 200},
						{Id: 1, Username: "sender", Balance: 500},
					}, nil)
				d.transactionProceeder.EXPECT().ProceedTransaction(gomock.Any(), nil, uint32(100),
					&domain.UserInfo{Id: 1, Username: "sender", Balance: 500},
					&domain.UserInfo{Id: 2, Username: "receiver", Balance: 200}).
					Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:         "same username error",
			fromUsername: "sameuser",
			toUsername:   "sameuser",
			amount:       100,
			prepareFn:    func(t *testing.T, d *deps) {},
			expectedErr:  &domain.InvalidArgumentsError{},
		},
		{
			name:         "insufficient balance",
			fromUsername: "pooruser",
			toUsername:   "receiver",
			amount:       1000,
			prepareFn: func(t *testing.T, d *deps) {
				d.txManager.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(executeTxFn)
				d.userFinder.EXPECT().GetTargetUsers(gomock.Any(), nil, "pooruser", "receiver").
					Return([]domain.UserInfo{
						{Id: 1, Username: "pooruser", Balance: 500},
						{Id: 2, Username: "receiver", Balance: 200},
					}, nil)
			},
			expectedErr: &domain.InsufficientBalanceError{},
		},
		{
			name:         "user not found",
			fromUsername: "sender",
			toUsername:   "nonexistent",
			amount:       100,
			prepareFn: func(t *testing.T, d *deps) {
				d.txManager.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(executeTxFn)
				d.userFinder.EXPECT().GetTargetUsers(gomock.Any(), nil, "sender", "nonexistent").
					Return(nil, &domain.UserNotFoundError{Msg: "user not found"})
			},
			expectedErr: &domain.UserNotFoundError{},
		},
		{
			name:         "proceed transaction error",
			fromUsername: "sender",
			toUsername:   "receiver",
			amount:       100,
			prepareFn: func(t *testing.T, d *deps) {
				d.txManager.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(executeTxFn)
				d.userFinder.EXPECT().GetTargetUsers(gomock.Any(), nil, "sender", "receiver").
					Return([]domain.UserInfo{
						{Id: 1, Username: "sender", Balance: 500},
						{Id: 2, Username: "receiver", Balance: 200},
					}, nil)
				d.transactionProceeder.EXPECT().ProceedTransaction(gomock.Any(), nil, uint32(100),
					&domain.UserInfo{Id: 1, Username: "sender", Balance: 500},
					&domain.UserInfo{Id: 2, Username: "receiver", Balance: 200}).
					Return(assert.AnError)
			},
			expectedErr: assert.AnError,
		},
		{
			name:         "internal error",
			fromUsername: "sender",
			toUsername:   "receiver",
			amount:       100,
			prepareFn: func(t *testing.T, d *deps) {
				d.txManager.EXPECT().WithinTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(executeTxFn)
				d.userFinder.EXPECT().GetTargetUsers(gomock.Any(), nil, "sender", "receiver").
					Return(nil, assert.AnError)
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
				userFinder:           storemocks.NewMockUserFinder(ctrl),
				transactionProceeder: storemocks.NewMockTransactionProceeder(ctrl),
			}

			tt.prepareFn(t, d)

			sendCoinsCase := NewSendCoinsCase(d.txManager, d.userFinder, d.transactionProceeder)
			err := sendCoinsCase.SendCoins(t.Context(), tt.fromUsername, tt.toUsername, tt.amount)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
